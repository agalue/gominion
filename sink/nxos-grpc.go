package sink

import (
	"fmt"
	"io"
	"net"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/mdt_dialout"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// NxosGrpcModule represents the Cisco Nexus NX-OS Telemetry module via gRPC
type NxosGrpcModule struct {
	sink   api.Sink
	config *api.MinionConfig
	server *grpc.Server
	port   int
}

// GetID gets the ID of the sink module
func (module *NxosGrpcModule) GetID() string {
	return "NXOS"
}

// Start initiates a gRPC Server for NX-OS telemetry
func (module *NxosGrpcModule) Start(config *api.MinionConfig, sink api.Sink) error {
	listener := config.GetListenerByParser("NxosGrpcParser")
	if listener == nil || listener.Port == 0 {
		log.Warnf("NX-OS Telemetry Module disabled")
		return nil
	}

	module.config = config
	module.sink = sink
	module.port = listener.Port

	module.server = grpc.NewServer()
	mdt_dialout.RegisterGRPCMdtDialoutServer(module.server, module)

	log.Infof("Starting NX-OS telemetry gRPC server on port %d\n", listener.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", listener.Port))
	if err != nil {
		return fmt.Errorf("Error cannot start TCP listener: %s", err)
	}
	go func() {
		if err := module.server.Serve(lis); err != nil {
			log.Errorf("Cannot serve NX-OS gRPC: %v", err)
		}
	}()
	return nil
}

// Stop shutdowns the sink module
func (module *NxosGrpcModule) Stop() {
	log.Warnf("Stopping NX-OS telemetry gRPC server")
	if module.server != nil {
		module.server.Stop()
	}
}

// MdtDialout implements Cisco NX-OS streaming telemetry service
func (module *NxosGrpcModule) MdtDialout(stream mdt_dialout.GRPCMdtDialout_MdtDialoutServer) error {
	peer, peerOK := peer.FromContext(stream.Context())
	if peerOK {
		log.Debugf("Accepted Cisco MDT GRPC dialout connection from %s\n", peer.Addr)
	}
	for {
		dialoutArgs, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			if errStatus, ok := status.FromError(err); ok {
				return status.Errorf(errStatus.Code(), "error while receiving NX-OS data: %v ", errStatus.Message())
			}
			return fmt.Errorf("error while receiving NX-OS data: %v", err)
		}
		if len(dialoutArgs.Data) == 0 && len(dialoutArgs.Errors) != 0 {
			log.Errorf("Received zero data from client %s: %s\n", peer.Addr, dialoutArgs.Errors)
			continue
		}
		log.Debugf("Received request with ID %d of %d bytes from %s\n", dialoutArgs.ReqId, len(dialoutArgs.Data), peer.Addr)
		if bytes := wrapMessageToTelemetry(module.config, peer.Addr.String(), uint32(module.port), dialoutArgs.Data); bytes != nil {
			sendBytes(module.GetID(), module.config, module.sink, bytes)
		}
	}
	log.Warnf("Terminating NX-OS handler")
	return nil
}

var nxosGrpcModule = &NxosGrpcModule{}

func init() {
	api.RegisterSinkModule(nxosGrpcModule)
}
