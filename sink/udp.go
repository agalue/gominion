package sink

import (
	"log"
	"net"

	"github.com/agalue/gominion/api"
)

// UDPForwardParser represents org.opennms.netmgt.telemetry.protocols.common.parser.ForwardParser
const UDPForwardParser = "ForwardParser"

// UDPForwardModule represents a generic UDP forward module
// It starts a UDP Listener, and forwards the received data to OpenNMS without alteration
type UDPForwardModule struct {
	broker api.Broker
	config *api.MinionConfig
	Name   string
	conn   *net.UDPConn
}

// GetID gets the ID of the sink module
func (module *UDPForwardModule) GetID() string {
	return module.Name
}

// Start initiates a blocking loop that forwards data received via UDP to OpenNMS
func (module *UDPForwardModule) Start(config *api.MinionConfig, broker api.Broker) {
	module.broker = broker
	module.config = config
	listener := config.GetListener(module.Name)
	if listener == nil || listener.GetParser() != UDPForwardParser {
		log.Printf("UDP Module %s disabled", module.Name)
		return
	}
	module.conn = startUDPServer(module.Name, listener.Port)
	for {
		buffer := make([]byte, 1024)
		n, addr, err := module.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error while reading from %s: %s", module.Name, err)
			continue
		}
		log.Printf("Received %d bytes from %s", n, addr)
		if bytes := wrapMessageToTelemetry(config, addr.IP.String(), uint32(listener.Port), buffer); bytes != nil {
			sendBytes(module.GetID(), module.config, module.broker, bytes)
		}
	}
}

// Stop shutdowns the sink module
func (module *UDPForwardModule) Stop() {
	if module.conn != nil {
		module.conn.Close()
	}
}

var graphiteModule = &UDPForwardModule{Name: "Graphite"}

func init() {
	api.RegisterSinkModule(graphiteModule)
}
