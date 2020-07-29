package broker

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"google.golang.org/grpc"
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
		return fmt.Errorf("Cannot start gRPC client: %v", err)
	}

	cli.onms = ipc.NewOpenNMSIpcClient(cli.conn)

	cli.rpcStream, err = cli.onms.RpcStreaming(context.Background())
	if err != nil {
		return err
	}

	go func() {
		// Send Headers
		headers := &ipc.RpcResponseProto{
			ModuleId: "MINION_HEADERS",
			Location: cli.config.Location,
			SystemId: cli.config.ID,
			RpcId:    cli.config.ID,
		}
		log.Printf("Sending Minion Headers from SystemId %s to gRPC server", cli.config.ID)
		if err := cli.rpcStream.Send(headers); err != nil {
			log.Printf("Error while sending RPC headers: %v", err)
		}
		// Main Loop
		for {
			if request, err := cli.rpcStream.Recv(); err == nil {
				/*
					if request.ExpirationTime < uint64(time.Now().Unix()) {
						log.Printf("TTL already expired for the request id = %s, won't process", request.RpcId)
						continue
					}
				*/
				log.Printf("Received RPC request with ID %s for module %s at location %s", request.RpcId, request.ModuleId, request.Location)
				if module, ok := api.GetRPCModule(request.ModuleId); ok {
					go func() {
						response := module.Execute(request)
						if err := cli.rpcStream.Send(response); err != nil {
							log.Printf("Error while sending RPC response for module %s with ID %s: %v", request.ModuleId, request.RpcId, err)
						}
					}()
				} else {
					log.Printf("Error cannot find implementation for module %s, ignoring request with ID %s", request.ModuleId, request.RpcId)
				}
			} else {
				if err == io.EOF {
					return
				}
				log.Printf("Error while getting an RPC Request: %v", err)
			}
		}
	}()

	cli.sinkStream, err = cli.onms.SinkStreaming(context.Background())
	if err != nil {
		return err
	}

	for _, sinkModule := range api.GetAllSinkModules() {
		go func(module api.SinkModule) {
			module.Start(cli.config, cli.sinkStream)
		}(sinkModule)
	}

	return nil
}

// Stop finilizes the gRPC client
func (cli *GrpcClient) Stop() {
	log.Printf("Stopping gRPC client")
	if cli.rpcStream != nil {
		cli.rpcStream.CloseSend()
	}
	if cli.sinkStream != nil {
		cli.sinkStream.CloseSend()
	}
	if cli.conn != nil {
		cli.conn.Close()
	}
	log.Printf("Good bye")
}
