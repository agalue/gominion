package broker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"

	"github.com/prometheus/client_golang/prometheus"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// GrpcClient represents the gRPC client implementation
type GrpcClient struct {
	config      *api.MinionConfig
	conn        *grpc.ClientConn
	onms        ipc.OpenNMSIpcClient
	rpcStream   ipc.OpenNMSIpc_RpcStreamingClient
	sinkStream  ipc.OpenNMSIpc_SinkStreamingClient
	traceCloser io.Closer

	// Prometheus metrics per module
	metricSinkMsgDeliverySucceeded *prometheus.CounterVec // Sink messages successfully delivered
	metricSinkMsgDeliveryFailed    *prometheus.CounterVec // Failed attempts to send Sink messages
	metricRPCReqReceivedSucceeded  *prometheus.CounterVec // RPC requests successfully received
	metricRPCReqReceivedFailed     *prometheus.CounterVec // Failed attempts to receive RPC requests
	metricRPCReqProcessedSucceeded *prometheus.CounterVec // RPC requests successfully processed
	metricRPCReqProcessedFailed    *prometheus.CounterVec // Failed attempts to process RPC requests
	metricRPCResSentSucceeded      *prometheus.CounterVec // RPC responses successfully sent
	metricRPCResSentFailed         *prometheus.CounterVec // Failed attempts to send RPC responses
}

// Start initializes the gRPC client
func (cli *GrpcClient) Start(config *api.MinionConfig) error {
	cli.config = config
	var err error

	if err = cli.initTracing(); err != nil {
		return err
	}

	options := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithStreamInterceptor(grpc_zap.StreamClientInterceptor(log.GetLogger())),
	}

	if config.BrokerProperties == nil {
		options = append(options, grpc.WithInsecure())
	} else {
		// TODO add client certificate for authentication
		tlsEnabled, ok := config.BrokerProperties["tls-enabled"]
		if ok && tlsEnabled == "true" {
			cred, err := cli.getCredentials()
			if err != nil {
				return err
			}
			options = append(options, grpc.WithTransportCredentials(cred))
		}
	}

	if config.StatsPort > 0 {
		cli.initPrometheusMetrics()
		options = append(options, grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	}

	cli.conn, err = grpc.Dial(config.BrokerURL, options...)
	if err != nil {
		return fmt.Errorf("Cannot dial gRPC server: %v", err)
	}
	cli.onms = ipc.NewOpenNMSIpcClient(cli.conn)

	log.Infof("Starting Sink API Stream")
	if err := cli.initSinkStream(); err != nil {
		return err
	}

	for _, module := range api.GetAllSinkModules() {
		if err := module.Start(cli.config, cli); err != nil {
			return fmt.Errorf("Cannot start Sink API module %s: %v", module.GetID(), err)
		}
	}

	log.Infof("Starting RPC API Stream")
	if err := cli.initRPCStream(); err != nil {
		return err
	}

	return nil
}

// Stop finilizes the gRPC client
func (cli *GrpcClient) Stop() {
	for _, module := range api.GetAllSinkModules() {
		module.Stop()
	}
	log.Warnf("Stopping gRPC client")
	if cli.rpcStream != nil {
		cli.rpcStream.CloseSend()
	}
	if cli.rpcStream != nil {
		cli.rpcStream.CloseSend()
	}
	if cli.conn != nil {
		cli.conn.Close()
	}
	if cli.traceCloser != nil {
		cli.traceCloser.Close()
	}
	log.Infof("Good bye")
}

// Send sends a Sink API message
func (cli *GrpcClient) Send(msg *ipc.SinkMessage) error {
	if cli.sinkStream == nil || cli.conn.GetState() != connectivity.Ready {
		// Try to restart the Sink stream
		if err := cli.initSinkStream(); err != nil {
			return err
		}
	}
	trace := cli.buildSpanForSinkMessage(msg)
	err := cli.sinkStream.Send(msg)
	if err == nil {
		cli.metricSinkMsgDeliverySucceeded.WithLabelValues(msg.ModuleId).Inc()
	} else if err == io.EOF {
		// Try to restart the Sink stream on server error
		cli.initSinkStream()
		return fmt.Errorf("Server unreachable; restarting Sink API Stream")
	} else {
		cli.metricSinkMsgDeliveryFailed.WithLabelValues(msg.ModuleId).Inc()
		trace.SetTag("failed", "true")
		trace.LogKV("event", err.Error())
	}
	trace.Finish()
	return err
}

// Initialize prometheus counters. Should be called once.
func (cli *GrpcClient) initPrometheusMetrics() {
	cli.metricSinkMsgDeliverySucceeded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "onms_sink_messages_delivery_succeeded",
		Help: "The total number of Sink messages successfully delivered per module",
	}, []string{"module"})
	cli.metricSinkMsgDeliveryFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "onms_sink_messages_delivery_failed",
		Help: "The total number of failed attempts to send Sink messages per module",
	}, []string{"module"})
	cli.metricRPCReqReceivedSucceeded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "onms_rpc_requests_received_succeeded",
		Help: "The total number of RPC requests successfully received per module",
	}, []string{"module"})
	cli.metricRPCReqReceivedFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "onms_rpc_requests_received_failed",
		Help: "The total number of failed attempts to receive RPC messages per module",
	}, []string{"module"})
	cli.metricRPCReqProcessedSucceeded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "onms_rpc_requests_processed_succeeded",
		Help: "The total number of RPC requests successfully processed per module",
	}, []string{"module"})
	cli.metricRPCReqProcessedFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "onms_rpc_requests_processed_failed",
		Help: "The total number of failed attempts to process RPC messages per module",
	}, []string{"module"})
	cli.metricRPCResSentSucceeded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "onms_rpc_responses_sent_succeeded",
		Help: "The total number of RPC responses successfully sent per module",
	}, []string{"module"})
	cli.metricRPCResSentFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "onms_rpc_responses_sent_failed",
		Help: "The total number of failed attempts to send RPC responses per module",
	}, []string{"module"})

	prometheus.MustRegister(
		cli.metricSinkMsgDeliverySucceeded,
		cli.metricSinkMsgDeliveryFailed,
		cli.metricRPCReqReceivedSucceeded,
		cli.metricRPCReqReceivedFailed,
		cli.metricRPCReqProcessedSucceeded,
		cli.metricRPCReqProcessedFailed,
		cli.metricRPCResSentSucceeded,
		cli.metricRPCResSentFailed,
	)
}

func (cli *GrpcClient) initTracing() error {
	cfg := jaegercfg.Configuration{
		ServiceName: cli.config.Location + "@" + cli.config.ID,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1, // JAEGER_SAMPLER_PARAM
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaegerlog.NullLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	cli.traceCloser = closer
	return nil
}

func (cli *GrpcClient) getCredentials() (credentials.TransportCredentials, error) {
	if srvCertPath, ok := cli.config.BrokerProperties["server-certificate-path"]; ok {
		return credentials.NewClientTLSFromFile(srvCertPath, "")
	}
	if data, ok := cli.config.BrokerProperties["server-certificate"]; ok {
		cert, err := x509.ParseCertificate([]byte(data))
		if err != nil {
			return nil, fmt.Errorf("Cannot parse server certificate: %v", err)
		}
		tlsCert := &tls.Certificate{
			Certificate: [][]byte{cert.Raw},
		}
		return credentials.NewServerTLSFromCert(tlsCert), nil
	}
	return nil, fmt.Errorf("Cannot find server certificate")
}

func (cli *GrpcClient) initSinkStream() error {
	var err error
	if cli.sinkStream != nil {
		cli.sinkStream.CloseSend()
	}

	cli.sinkStream, err = cli.onms.SinkStreaming(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot initialize Sink API Stream: %v", err)
	}

	return nil
}

func (cli *GrpcClient) initRPCStream() error {
	var err error
	if cli.rpcStream != nil {
		cli.rpcStream.CloseSend()
	}

	cli.rpcStream, err = cli.onms.RpcStreaming(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot initialize RPC API Stream: %v", err)
	}

	go func() {
		cli.sendMinionHeaders()
		for {
			if cli.rpcStream == nil {
				break
			}
			if request, err := cli.rpcStream.Recv(); err == nil {
				cli.metricRPCReqReceivedSucceeded.WithLabelValues(request.ModuleId).Inc()
				cli.processRequest(request)
			} else {
				if err == io.EOF {
					break
				}
				errStatus, _ := status.FromError(err)
				if errStatus.Code() != codes.Unavailable {
					log.Errorf("Cannot receive RPC Request: code=%s, message=%s", errStatus.Code(), errStatus.Message())
					cli.metricRPCReqReceivedFailed.WithLabelValues(request.ModuleId).Inc()
				}
			}
		}
		log.Warnf("Terminating RPC API handler")
	}()

	// Detect termination of the stream and try to restart it until success
	go func() {
		<-cli.rpcStream.Context().Done()
		for {
			if err := cli.initRPCStream(); err == nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return nil
}

func (cli *GrpcClient) sendMinionHeaders() {
	headers := &ipc.RpcResponseProto{
		ModuleId: "MINION_HEADERS",
		Location: cli.config.Location,
		SystemId: cli.config.ID,
		RpcId:    cli.config.ID,
	}
	log.Infof("Sending Minion Headers from SystemId %s to gRPC server", cli.config.ID)
	if err := cli.rpcStream.Send(headers); err != nil {
		log.Errorf("Cannot send RPC headers: %v", err)
	}
}

func (cli *GrpcClient) processRequest(request *ipc.RpcRequestProto) {
	log.Debugf("Received RPC request with ID %s for module %s at location %s", request.RpcId, request.ModuleId, request.Location)
	if module, ok := api.GetRPCModule(request.ModuleId); ok {
		go func() {
			trace := cli.buildSpanFromRPCMessage(request)
			if response := module.Execute(request); response != nil {
				cli.metricRPCReqProcessedSucceeded.WithLabelValues(request.ModuleId).Inc()
				if err := cli.rpcStream.Send(response); err == nil {
					cli.metricRPCResSentSucceeded.WithLabelValues(request.ModuleId).Inc()
				} else {
					trace.SetTag("failed", "true")
					trace.LogKV("event", err.Error())
					cli.metricRPCResSentFailed.WithLabelValues(request.ModuleId).Inc()
					log.Errorf("Cannot send RPC response for module %s with ID %s: %v", request.ModuleId, request.RpcId, err)
				}
			} else {
				log.Errorf("Module %s returned an empty response for request %s, ignoring", request.ModuleId, request.RpcId)
				cli.metricRPCReqProcessedFailed.WithLabelValues(request.ModuleId).Inc()
			}
			trace.Finish()
		}()
	} else {
		log.Errorf("Cannot find implementation for module %s, ignoring request with ID %s", request.ModuleId, request.RpcId)
	}
}

func (cli *GrpcClient) buildSpanFromRPCMessage(request *ipc.RpcRequestProto) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	tags := cli.getTagsForRPC(request)
	ctx, err := tracer.Extract(opentracing.TextMap, request.TracingInfo)
	if err == nil {
		return tracer.StartSpan(request.ModuleId, opentracing.FollowsFrom(ctx), tags)
	}
	return tracer.StartSpan(request.ModuleId, tags)
}

func (cli *GrpcClient) getTagsForRPC(request *ipc.RpcRequestProto) opentracing.Tags {
	var tags = opentracing.Tags{"location": request.Location}
	if request.SystemId != "" {
		tags["systemId"] = request.SystemId
	}
	for key, value := range request.TracingInfo {
		tags[key] = value
	}
	return tags
}

func (cli *GrpcClient) buildSpanForSinkMessage(msg *ipc.SinkMessage) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	tags := cli.getTagsForSink(msg)
	ctx, err := tracer.Extract(opentracing.TextMap, msg.TracingInfo)
	if err == nil {
		return tracer.StartSpan(msg.ModuleId, opentracing.FollowsFrom(ctx), tags)
	}
	return tracer.StartSpan(msg.ModuleId, tags)
}

func (cli *GrpcClient) getTagsForSink(msg *ipc.SinkMessage) opentracing.Tags {
	var tags = opentracing.Tags{"location": msg.Location}
	if msg.SystemId != "" {
		tags["systemId"] = msg.SystemId
	}
	for key, value := range msg.TracingInfo {
		tags[key] = value
	}
	return tags
}
