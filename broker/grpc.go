package broker

import (
	"context"
	"fmt"
	"io"

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

	cli.conn, err = grpc.Dial(config.BrokerURL, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("Cannot dial gRPC server: %v", err)
	}

	cli.onms = ipc.NewOpenNMSIpcClient(cli.conn)

	cli.sinkStream, err = cli.onms.SinkStreaming(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot start Sink Streaming: %v", err)
	}

	for _, module := range api.GetAllSinkModules() {
		if err := module.Start(cli.config, cli.sinkStream); err != nil {
			return fmt.Errorf("Cannot start Sink module %s: %v", module.GetID(), err)
		}
	}

	cli.rpcStream, err = cli.onms.RpcStreaming(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot start RPC Streaming: %v", err)
	}

	go func() {
		cli.sendHeaders()
		for {
			if request, err := cli.rpcStream.Recv(); err == nil {
				cli.processRequest(request)
			} else {
				if err == io.EOF {
					return
				}
				errStatus, _ := status.FromError(err)
				if errStatus.Code().String() != "Unavailable" {
					log.Errorf("Cannor receive RPC Request: code=%s, message=%s", errStatus.Code(), errStatus.Message())
				}
			}
		}
	}()

	return nil
}

// Stop finilizes the gRPC client
func (cli *GrpcClient) Stop() {
	log.Warnf("Stopping gRPC client")
	for _, module := range api.GetAllSinkModules() {
		module.Stop()
	}
	if cli.conn != nil {
		cli.conn.Close()
	}
	log.Infof("Good bye")
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
		log.Errorf("Cannor send RPC headers: %v", err)
	}
}

func (cli *GrpcClient) processRequest(request *ipc.RpcRequestProto) {
	log.Infof("Received RPC request with ID %s for module %s at location %s", request.RpcId, request.ModuleId, request.Location)
	if module, ok := api.GetRPCModule(request.ModuleId); ok {
		go func() {
			if response := module.Execute(request); response != nil {
				if err := cli.rpcStream.Send(response); err != nil {
					log.Errorf("Cannor send RPC response for module %s with ID %s: %v", request.ModuleId, request.RpcId, err)
				}
			} else {
				log.Errorf("Module %s returned an empty response for request %s, ignoring", request.ModuleId, request.RpcId)
			}
		}()
	} else {
		log.Errorf("Cannot find implementation for module %s, ignoring request with ID %s", request.ModuleId, request.RpcId)
	}
}
