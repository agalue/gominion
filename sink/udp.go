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
	Name     string
	broker   api.Broker
	config   *api.MinionConfig
	conn     *net.UDPConn
	stopping bool
}

// GetID gets the ID of the sink module
func (module *UDPForwardModule) GetID() string {
	return module.Name
}

// Start initiates a blocking loop that forwards data received via UDP to OpenNMS
func (module *UDPForwardModule) Start(config *api.MinionConfig, broker api.Broker) {
	module.stopping = false
	module.broker = broker
	module.config = config
	listener := config.GetListener(module.Name)
	if listener == nil || !listener.Is(UDPForwardParser) {
		log.Printf("UDP Module %s disabled", module.Name)
		return
	}
	module.conn = startUDPServer(module.Name, listener.Port)
	payload := make([]byte, 1024)
	for {
		size, pktAddr, err := module.conn.ReadFromUDP(payload)
		if err != nil {
			if !module.stopping {
				log.Printf("Error while reading from %s: %s", module.Name, err)
			}
			continue
		}
		payloadCut := make([]byte, size)
		copy(payloadCut, payload[0:size])
		log.Printf("Received %d bytes from %s", size, pktAddr)
		if bytes := wrapMessageToTelemetry(config, pktAddr.IP.String(), uint32(pktAddr.Port), payloadCut); bytes != nil {
			sendBytes(module.GetID(), module.config, module.broker, bytes)
		}
	}
}

// Stop shutdowns the sink module
func (module *UDPForwardModule) Stop() {
	module.stopping = true
	if module.conn != nil {
		module.conn.Close()
	}
}

var graphiteModule = &UDPForwardModule{Name: "Graphite"}

func init() {
	api.RegisterSinkModule(graphiteModule)
}
