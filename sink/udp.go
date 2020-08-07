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

// Start initiates a generic UDP receiver
func (module *UDPForwardModule) Start(config *api.MinionConfig, broker api.Broker) error {
	listener := config.GetListener(module.Name)
	if listener == nil || !listener.Is(UDPForwardParser) {
		log.Warnf("UDP Module %s disabled", module.Name)
		return nil
	}

	var err error
	module.stopping = false
	module.broker = broker
	module.config = config

	module.conn, err = createUDPListener(listener.Port)
	if err != nil {
		return err
	}
	log.Infof("Starting %s receiver on port UDP %d", module.Name, config.TrapPort)
	go func() {
		payload := make([]byte, 1024)
		for {
			size, pktAddr, err := module.conn.ReadFromUDP(payload)
			if err != nil {
				if !module.stopping {
					log.Errorf("%s cannot read from UDP: %s", module.Name, err)
				}
				continue
			}
			payloadCut := make([]byte, size)
			copy(payloadCut, payload[0:size])
			log.Debugf("Received %d bytes from %s", size, pktAddr)
			if bytes := wrapMessageToTelemetry(config, pktAddr.IP.String(), uint32(pktAddr.Port), payloadCut); bytes != nil {
				sendBytes(module.GetID(), module.config, module.broker, bytes)
			}
		}
	}()
	return nil
}

// Stop shutdowns the sink module
func (module *UDPForwardModule) Stop() {
	log.Infof("Stopping %s receiver", module.Name)
	module.stopping = true
	if module.conn != nil {
		module.conn.Close()
	}
}

var graphiteModule = &UDPForwardModule{Name: "Graphite"}

func init() {
	api.RegisterSinkModule(graphiteModule)
}
