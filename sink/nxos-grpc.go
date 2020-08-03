package sink

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/mdt_dialout"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// NxosGrpcModule represents the Cisco Nexus NX-OS Telemetry module via gRPC
type NxosGrpcModule struct {
	broker api.Broker
	config *api.MinionConfig
	server *grpc.Server
	port   int
}

// GetID gets the ID of the sink module
func (module *NxosGrpcModule) GetID() string {
	return "NXOS"
}

// Start initiates a blocking loop that starts the gRPC Server for NX-OS telemetry
func (module *NxosGrpcModule) Start(config *api.MinionConfig, broker api.Broker) {
	listener := config.GetListenerByParser("NxosGrpcParser")
	if listener == nil || listener.Port == 0 {
		log.Printf("NX-OS Telemetry Module disabled")
		return
	}

	module.config = config
	module.broker = broker
	module.port = listener.Port

	log.Printf("Starting NX-OS Telemetry Module")

	module.server = grpc.NewServer()
	mdt_dialout.RegisterGRPCMdtDialoutServer(module.server, module)

	log.Printf("Starting NX-OS gRPC server on port %d\n", listener.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", listener.Port))
	if err != nil {
		log.Fatalf("Error cannot start TCP listener: %s", err)
	}
	if err := module.server.Serve(lis); err != nil {
		log.Fatalf("Error could not serve: %v", err)
	}
}

// Stop shutdowns the sink module
func (module *NxosGrpcModule) Stop() {
	if module.server != nil {
		module.server.Stop()
	}
}

// MdtDialout implements Cisco NX-OS streaming telemetry service
func (module *NxosGrpcModule) MdtDialout(stream mdt_dialout.GRPCMdtDialout_MdtDialoutServer) error {
	peer, peerOK := peer.FromContext(stream.Context())
	if peerOK {
		log.Printf("Accepted Cisco MDT GRPC dialout connection from %s\n", peer.Addr)
	}
	for {
		dialoutArgs, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Error NX-OS session closed")
			} else {
				log.Println("Error NX-OS session error")
			}
			continue
		}
		if len(dialoutArgs.Data) == 0 && len(dialoutArgs.Errors) != 0 {
			log.Printf("Error from client %s, %s\n", peer.Addr, dialoutArgs.Errors)
			continue
		}
		log.Printf("Received request with ID %d of %d bytes from %s\n", dialoutArgs.ReqId, len(dialoutArgs.Data), peer.Addr)
		if bytes := wrapMessageToTelemetry(module.config, peer.Addr.String(), uint32(module.port), dialoutArgs.Data); bytes != nil {
			sendBytes(module.GetID(), module.config, module.broker, bytes)
		}
	}
}

var nxosGrpcModule = &NxosGrpcModule{}

func init() {
	api.RegisterSinkModule(nxosGrpcModule)
}
