package broker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/prometheus/client_golang/prometheus"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// GrpcClient represents the gRPC client implementation
type GrpcClient struct {
	config     *api.MinionConfig
	conn       *grpc.ClientConn
	onms       ipc.OpenNMSIpcClient
	rpcStream  ipc.OpenNMSIpc_RpcStreamingClient
	sinkStream ipc.OpenNMSIpc_SinkStreamingClient

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
	var err error

	cli.config = config
	options := []grpc.DialOption{grpc.WithBlock()}
	if config.BrokerProperties == nil {
		options = append(options, grpc.WithInsecure())
	} else {
		// TODO add client certificate for authentication
		tlsEnabled, ok := config.BrokerProperties["tls-enabled"]
		if ok && tlsEnabled == "true" {
			systemRoots, err := cli.getCertPool()
			if err != nil {
				return err
			}
			cred := credentials.NewTLS(&tls.Config{
				RootCAs: systemRoots,
			})
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
	if err := cli.startSinkStream(); err != nil {
		return err
	}

	for _, module := range api.GetAllSinkModules() {
		if err := module.Start(cli.config, cli); err != nil {
			return fmt.Errorf("Cannot start Sink API module %s: %v", module.GetID(), err)
		}
	}

	log.Infof("Starting RPC API Stream")
	if err := cli.startRPCStream(); err != nil {
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
	log.Infof("Good bye")
}

// Send sends a Sink API message
func (cli *GrpcClient) Send(msg *ipc.SinkMessage) error {
	if cli.sinkStream == nil {
		// Try to restart the Sink stream
		if err := cli.startSinkStream(); err != nil {
			return err
		}
	}
	err := cli.sinkStream.Send(msg)
	if err == nil {
		cli.metricSinkMsgDeliverySucceeded.WithLabelValues(msg.ModuleId).Inc()
	} else if err == io.EOF {
		// Try to restart the Sink stream on server error
		cli.startSinkStream()
		return fmt.Errorf("Server unreachable; restarting Sink API Stream")
	} else {
		cli.metricSinkMsgDeliveryFailed.WithLabelValues(msg.ModuleId).Inc()
	}
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

func (cli *GrpcClient) getCertPool() (*x509.CertPool, error) {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	srvCert := make([]byte, 0)
	if srvCertPath, ok := cli.config.BrokerProperties["server-certificate-path"]; ok {
		data, err := ioutil.ReadFile(srvCertPath)
		if err != nil {
			return nil, fmt.Errorf("Cannot read server certificate from file %s: %v", srvCertPath, err)
		}
		srvCert = data
	}
	if len(srvCert) == 0 {
		if data, ok := cli.config.BrokerProperties["server-certificate"]; ok {
			srvCert = []byte(data)
		}
	}
	if len(srvCert) > 0 {
		cert, err := x509.ParseCertificate(srvCert)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse server certificate: %v", err)
		}
		systemRoots.AddCert(cert)
	}
	return systemRoots, nil
}

func (cli *GrpcClient) startSinkStream() error {
	var err error
	if cli.sinkStream != nil {
		cli.sinkStream.CloseSend()
	}

	cli.sinkStream, err = cli.onms.SinkStreaming(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot start Sink API Stream: %v", err)
	}

	return nil
}

func (cli *GrpcClient) startRPCStream() error {
	var err error
	if cli.rpcStream != nil {
		cli.rpcStream.CloseSend()
	}

	cli.rpcStream, err = cli.onms.RpcStreaming(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot start RPC API Stream: %v", err)
	}

	go func() {
		cli.sendHeaders()
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
				if errStatus.Code().String() != "Unavailable" {
					log.Errorf("Cannot receive RPC Request: code=%s, message=%s", errStatus.Code(), errStatus.Message())
					cli.metricRPCReqReceivedFailed.WithLabelValues(request.ModuleId).Inc()
				}
			}
		}
		log.Warnf("Terminating RPC API handler")
	}()

	go func() {
		<-cli.rpcStream.Context().Done()
		for {
			err := cli.startRPCStream()
			if err == nil {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func (cli *GrpcClient) sendHeaders() {
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
			if response := module.Execute(request); response != nil {
				cli.metricRPCReqProcessedSucceeded.WithLabelValues(request.ModuleId).Inc()
				if err := cli.rpcStream.Send(response); err == nil {
					cli.metricRPCResSentSucceeded.WithLabelValues(request.ModuleId).Inc()
				} else {
					cli.metricRPCResSentFailed.WithLabelValues(request.ModuleId).Inc()
					log.Errorf("Cannot send RPC response for module %s with ID %s: %v", request.ModuleId, request.RpcId, err)
				}
			} else {
				log.Errorf("Module %s returned an empty response for request %s, ignoring", request.ModuleId, request.RpcId)
				cli.metricRPCReqProcessedFailed.WithLabelValues(request.ModuleId).Inc()
			}
		}()
	} else {
		log.Errorf("Cannot find implementation for module %s, ignoring request with ID %s", request.ModuleId, request.RpcId)
	}
}
