package sink

import (
	"net"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
)

// UDPForwardParser represents org.opennms.netmgt.telemetry.protocols.common.parser.ForwardParser
const UDPForwardParser = "ForwardParser"

// UDPForwardModule represents a generic UDP forward module
// It starts a UDP Listener, and forwards the received data to OpenNMS without alteration
type UDPForwardModule struct {
	name     string
	sink     api.Sink
	config   *api.MinionConfig
	conn     *net.UDPConn
	stopping bool
}

// GetID gets the ID of the sink module
func (module *UDPForwardModule) GetID() string {
	return module.name
}

// Start initiates a generic UDP receiver
func (module *UDPForwardModule) Start(config *api.MinionConfig, sink api.Sink) error {
	listener := config.GetListener(module.name)
	if listener == nil || !listener.Is(UDPForwardParser) {
		log.Warnf("UDP Module %s disabled", module.name)
		return nil
	}

	var err error
	module.stopping = false
	module.sink = sink
	module.config = config

	module.conn, err = createUDPListener(listener.Port)
	if err != nil {
		return err
	}
	log.Infof("Starting %s receiver on port UDP %d", module.name, listener.Port)
	go func() {
		payload := make([]byte, 1024)
		for {
			size, pktAddr, err := module.conn.ReadFromUDP(payload)
			if err != nil {
				if !module.stopping {
					log.Errorf("%s cannot read from UDP: %s", module.name, err)
				}
				continue
			}
			payloadCut := make([]byte, size)
			copy(payloadCut, payload[0:size])
			log.Debugf("Received %d bytes from %s", size, pktAddr)
			messages := make([][]byte, 1)
			messages[0] = payloadCut
			if bytes := wrapMessageToTelemetry(module.config, pktAddr.IP.String(), uint32(pktAddr.Port), messages); bytes != nil {
				sendBytes(module.GetID(), module.config, module.sink, bytes)
			}
		}
	}()
	return nil
}

// Stop shutdowns the sink module
func (module *UDPForwardModule) Stop() {
	log.Warnf("Stopping %s receiver", module.name)
	module.stopping = true
	if module.conn != nil {
		module.conn.Close()
	}
}

func init() {
	api.RegisterSinkModule(&UDPForwardModule{name: "Graphite"})
}
