package broker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GrpcClient represents the gRPC client implementation
type GrpcClient struct {
	config     *api.MinionConfig
	conn       *grpc.ClientConn
	onms       ipc.OpenNMSIpcClient
	rpcStream  ipc.OpenNMSIpc_RpcStreamingClient
	sinkStream ipc.OpenNMSIpc_SinkStreamingClient
}

// Start initializes the gRPC client
func (cli *GrpcClient) Start(config *api.MinionConfig) error {
	var err error

	cli.config = config

	cli.conn, err = grpc.Dial(config.BrokerURL,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("Cannot dial gRPC server: %v", err)
	}
	cli.onms = ipc.NewOpenNMSIpcClient(cli.conn)

	if err := cli.startSinkStream(); err != nil {
		return err
	}

	for _, module := range api.GetAllSinkModules() {
		if err := module.Start(cli.config, cli); err != nil {
			return fmt.Errorf("Cannot start Sink API module %s: %v", module.GetID(), err)
		}
	}

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
		if err := cli.startSinkStream(); err != nil {
			return err
		}
	}
	err := cli.sinkStream.Send(msg)
	if err == io.EOF {
		cli.startSinkStream()
		return fmt.Errorf("Sink API Stream has restarted")
	}
	return nil
}

func (cli *GrpcClient) startSinkStream() error {
	var err error
	if cli.sinkStream != nil {
		cli.sinkStream.CloseSend()
	}
	log.Infof("Starting Sink API Stream")
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

	log.Infof("Starting RPC API Stream")
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
				cli.processRequest(request)
			} else {
				if err == io.EOF {
					break
				}
				errStatus, _ := status.FromError(err)
				if errStatus.Code().String() != "Unavailable" {
					log.Errorf("Cannot receive RPC Request: code=%s, message=%s", errStatus.Code(), errStatus.Message())
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
				if err := cli.rpcStream.Send(response); err != nil {
					log.Errorf("Cannot send RPC response for module %s with ID %s: %v", request.ModuleId, request.RpcId, err)
				}
			} else {
				log.Errorf("Module %s returned an empty response for request %s, ignoring", request.ModuleId, request.RpcId)
			}
		}()
	} else {
		log.Errorf("Cannot find implementation for module %s, ignoring request with ID %s", request.ModuleId, request.RpcId)
	}
}
