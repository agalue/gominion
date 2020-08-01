package sink

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
	"github.com/soniah/gosnmp"
)

// SnmpTrapModule represents the SNMP trap receiver module
type SnmpTrapModule struct {
	broker   api.Broker
	config   *api.MinionConfig
	listener *gosnmp.TrapListener
}

// GetID gets the ID of the sink module
func (module *SnmpTrapModule) GetID() string {
	return "Trap"
}

// Start initiates a blocking loop with the SNMP trap listener
func (module *SnmpTrapModule) Start(config *api.MinionConfig, broker api.Broker) {
	if config.TrapPort == 0 {
		log.Printf("Trap Module disabled")
		return
	}

	log.Printf("Starting SNMP Trap receiver on port UDP %d", config.TrapPort)

	module.config = config
	module.broker = broker
	module.listener = gosnmp.NewTrapListener()
	module.listener.OnNewTrap = module.trapHandler
	module.listener.Params = gosnmp.Default
	err := module.listener.Listen(fmt.Sprintf("0.0.0.0:%d", config.TrapPort))
	if err != nil {
		log.Fatalf("Cannot start SNMP trap listener: %s", err)
	}
}

// Stop shutdowns the sink module
func (module *SnmpTrapModule) Stop() {
	if module.listener != nil {
		module.listener.Close()
	}
}

func (module *SnmpTrapModule) trapHandler(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
	version := fmt.Sprintf("v%s", packet.Version)
	log.Printf("Received SNMP%s trap (type: 0x%X) from %s\n", version, packet.PDUType, addr.IP)

	trap := api.TrapDTO{
		AgentAddress: addr.IP.String(),
		PDULength:    len(packet.Variables),
		CreationTime: time.Now().Unix() * 1000,
		Timestamp:    int64(packet.Timestamp),
		Community:    packet.Community,
		Version:      version,
	}

	trapLog := api.TrapLogDTO{
		Location: module.config.Location,
		SystemID: module.config.ID,
	}

	if packet.PDUType == gosnmp.Trap {
		trap.TrapIdentity = &api.TrapIdentityDTO{
			EnterpriseID: packet.Enterprise,
			Generic:      packet.GenericTrap,
			Specific:     packet.SpecificTrap,
		}
		trapLog.TrapAddress = packet.AgentAddress
	} else {
		trapLog.TrapAddress = addr.IP.String()
	}

	for _, pdu := range packet.Variables {
		switch pdu.Name {
		case ".1.3.6.1.2.1.1.3.0":
			trap.Timestamp = gosnmp.ToBigInt(pdu.Value).Int64()
		case ".1.3.6.1.6.3.1.1.4.1.0":
			trap.TrapIdentity = module.extractTrapIdentity(pdu)
		default:
			result := tools.GetResultForPDU(pdu, pdu.Name)
			trap.AddResult(result)
		}
	}

	trapLog.AddTrap(trap)
	sendResponse(module.GetID(), module.config, module.broker, trapLog)
}

func (module *SnmpTrapModule) extractTrapIdentity(pdu gosnmp.SnmpPDU) *api.TrapIdentityDTO {
	if pdu.Name != ".1.3.6.1.6.3.1.1.4.1.0" {
		return nil
	}
	value := pdu.Value.(string)
	parts := strings.Split(value, ".")
	specific, _ := strconv.Atoi(parts[len(parts)-1])
	return &api.TrapIdentityDTO{
		EnterpriseID: strings.Join(parts[0:len(parts)-2], "."),
		Generic:      6,
		Specific:     specific,
	}
}

var snmpTrapModule = &SnmpTrapModule{}

func init() {
	api.RegisterSinkModule(snmpTrapModule)
}
