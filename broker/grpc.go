package broker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// GrpcClient represents the gRPC client implementation for the OpenNMS IPC API.
// This should be equivalent to MinionGrpcClient.java
type GrpcClient struct {
	config      *api.MinionConfig
	conn        *grpc.ClientConn
	onms        ipc.OpenNMSIpcClient
	rpcStream   ipc.OpenNMSIpc_RpcStreamingClient
	sinkStream  ipc.OpenNMSIpc_SinkStreamingClient
	traceCloser io.Closer
	metrics     Metrics
	sinkMutex   *sync.Mutex
	rpcMutex    *sync.Mutex
}

// Start initializes the gRPC client.
// Returns an error when the configuration is incorrect or cannot connect to the server.
func (cli *GrpcClient) Start(config *api.MinionConfig) error {
	cli.config = config
	var err error

	cli.metrics = NewMetrics()
	cli.sinkMutex = new(sync.Mutex)
	cli.rpcMutex = new(sync.Mutex)

	if cli.traceCloser, err = initTracing(cli.config); err != nil {
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
			cred, err := cli.getTransportCredentials()
			if err != nil {
				return err
			}
			options = append(options, grpc.WithTransportCredentials(cred))
		}
	}

	if config.StatsPort > 0 {
		cli.metrics.Register()
		options = append(options, grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	}

	cli.conn, err = grpc.Dial(config.BrokerURL, options...)
	if err != nil {
		return fmt.Errorf("Cannot dial gRPC server: %v", err)
	}
	cli.onms = ipc.NewOpenNMSIpcClient(cli.conn)

	log.Infof("Starting Sink API Stream")
	if err = cli.initSinkStream(); err != nil {
		return err
	}

	for _, module := range api.GetAllSinkModules() {
		if err = module.Start(cli.config, cli); err != nil {
			return fmt.Errorf("Cannot start Sink API module %s: %v", module.GetID(), err)
		}
	}

	log.Infof("Starting RPC API Stream")
	if err = cli.initRPCStream(); err != nil {
		return err
	}

	return nil
}

// Stop finalizes the gRPC client and all its dependencies.
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

// Send forwards a Sink API message to the OpenNMS gRPC server.
// Attempts to restart the client when the stream is unavailable or the connection is not ready.
// Messages are discarded when the server is unavailable.
func (cli *GrpcClient) Send(msg *ipc.SinkMessage) error {
	if cli.sinkStream == nil || cli.conn.GetState() != connectivity.Ready {
		// Try to restart the Sink stream
		if err := cli.initSinkStream(); err != nil {
			return err
		}
		log.Warnf("Sink API stream restarted")
	}
	trace := startSpanForSinkMessage(msg)
	defer trace.Finish()
	cli.sinkMutex.Lock()
	err := cli.sinkStream.Send(msg)
	cli.sinkMutex.Unlock()
	if err == nil {
		cli.metrics.SinkMsgDeliverySucceeded.WithLabelValues(msg.ModuleId).Inc()
		return nil
	}
	if err == io.EOF {
		err = fmt.Errorf("Server unreachable")
	}
	cli.metrics.SinkMsgDeliveryFailed.WithLabelValues(msg.ModuleId).Inc()
	trace.SetTag("failed", "true")
	trace.LogKV("event", err.Error())
	return err
}

// Initializes the Sink API stream
func (cli *GrpcClient) initSinkStream() error {
	var err error

	cli.sinkMutex.Lock()
	defer cli.sinkMutex.Unlock()

	if cli.sinkStream != nil {
		cli.sinkStream.CloseSend()
	}

	cli.sinkStream, err = cli.onms.SinkStreaming(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot initialize Sink API Stream: %v", err)
	}

	return nil
}

// Initializes the RPC API stream
func (cli *GrpcClient) initRPCStream() error {
	var err error

	cli.rpcMutex.Lock()
	defer cli.rpcMutex.Unlock()

	if cli.rpcStream != nil {
		cli.rpcStream.CloseSend()
	}

	cli.rpcStream, err = cli.onms.RpcStreaming(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot initialize RPC API Stream: %v", err)
	}

	// Goroutine to handle RPC API requests from the gRPC server.
	go func() {
		cli.sendMinionHeaders()
		for {
			if cli.rpcStream == nil || cli.conn.GetState() != connectivity.Ready {
				break
			}
			if request, err := cli.rpcStream.Recv(); err == nil {
				cli.processRequest(request)
				cli.metrics.RPCReqReceivedSucceeded.WithLabelValues(request.ModuleId).Inc()
			} else {
				if err == io.EOF {
					break
				}
				if errStatus, _ := status.FromError(err); errStatus.Code() != codes.Unavailable {
					log.Errorf("Cannot receive RPC Request: %v", err)
				}
				cli.metrics.RPCReqReceivedFailed.WithLabelValues(request.ModuleId).Inc()
			}
		}
		log.Warnf("Terminating RPC API handler")
	}()

	// Detects the termination of the stream and try to restart it until success
	go func() {
		<-cli.rpcStream.Context().Done()
		for {
			if err := cli.initRPCStream(); err == nil {
				log.Warnf("RPC API stream restarted")
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return nil
}

// Gets the TLS transport credentials from a file or a string.
func (cli *GrpcClient) getTransportCredentials() (credentials.TransportCredentials, error) {
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

// Sends the Minion headers as an RPC API response, to register the Minion as a client.
// Executes this every time the RPC API Stream is created.
func (cli *GrpcClient) sendMinionHeaders() {
	headers := cli.config.GetHeaderResponse()
	log.Infof("Sending Minion Headers from SystemId %s to gRPC server", cli.config.ID)
	cli.rpcMutex.Lock()
	if err := cli.rpcStream.Send(headers); err != nil {
		log.Errorf("Cannot send RPC headers: %v", err)
	}
	cli.rpcMutex.Unlock()
}

// Processes an RPC API request sent by OpenNMS asynchronously within a goroutine and sends back the response from the module.
func (cli *GrpcClient) processRequest(request *ipc.RpcRequestProto) {
	log.Debugf("Received RPC request with ID %s for module %s at location %s", request.RpcId, request.ModuleId, request.Location)
	if module, ok := api.GetRPCModule(request.ModuleId); ok {
		go func() {
			trace := startSpanFromRPCMessage(request)
			var err error
			if response := module.Execute(request); response != nil {
				cli.metrics.RPCReqProcessedSucceeded.WithLabelValues(request.ModuleId).Inc()
				err = cli.sendResponse(response)
			} else {
				cli.metrics.RPCReqProcessedFailed.WithLabelValues(request.ModuleId).Inc()
				err = fmt.Errorf("Module %s returned an empty response for request %s, ignoring", request.ModuleId, request.RpcId)
			}
			if err != nil {
				trace.SetTag("failed", "true")
				trace.LogKV("event", err.Error())
			}
			trace.Finish()
		}()
	} else {
		log.Errorf("Cannot find implementation for module %s, ignoring request with ID %s", request.ModuleId, request.RpcId)
	}
}

// Sends an RPC API response to OpenNMS
func (cli *GrpcClient) sendResponse(response *ipc.RpcResponseProto) error {
	if cli.rpcStream != nil && cli.conn.GetState() == connectivity.Ready {
		cli.rpcMutex.Lock()
		err := cli.rpcStream.Send(response)
		cli.rpcMutex.Unlock()
		if err == nil {
			cli.metrics.RPCResSentSucceeded.WithLabelValues(response.ModuleId).Inc()
			return nil
		}
		cli.metrics.RPCResSentFailed.WithLabelValues(response.ModuleId).Inc()
		return fmt.Errorf("Cannot send RPC response for module %s with ID %s: %v", response.ModuleId, response.RpcId, err)
	}
	cli.metrics.RPCResSentFailed.WithLabelValues(response.ModuleId).Inc()
	return fmt.Errorf("Cannot connect to the server, ignoring RPC request for module %s with ID %s", response.ModuleId, response.RpcId)
}
