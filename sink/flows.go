package sink

import (
	"net"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/netflow"
	"github.com/golang/protobuf/proto"

	decoder "github.com/cloudflare/goflow/v3/decoders"
	goflowMsg "github.com/cloudflare/goflow/v3/pb"
	goflow "github.com/cloudflare/goflow/v3/utils"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
)

// UDPNetflow5Parser represents the UDP Netflow5 parser name
const UDPNetflow5Parser = "Netflow5UdpParser"

// UDPNetflow9Parser represents the UDP Netflow9 parser name
const UDPNetflow9Parser = "Netflow9UdpParser"

// UDPIpfixParser represents the UDP IPFIX parser name
const UDPIpfixParser = "IpfixUdpParser"

// UDPSFlowParser represents the UDP SFlow parser name
const UDPSFlowParser = "SFlowUdpParser"

// TCPIpfixParser represents the TCP IPFIX parser name
const TCPIpfixParser = "IpfixUdpParser"

// Custom Logger implementation for goflow
type flowLogger struct{}

func (logger flowLogger) Printf(format string, params ...interface{}) {
	log.Infof(format, params...)
}

func (logger flowLogger) Errorf(format string, params ...interface{}) {
	log.Errorf(format, params...)
}

func (logger flowLogger) Warnf(format string, params ...interface{}) {
	log.Warnf(format, params...)
}

func (logger flowLogger) Infof(format string, params ...interface{}) {
	log.Infof(format, params...)
}

func (logger flowLogger) Debugf(format string, params ...interface{}) {
	log.Debugf(format, params...)
}

func (logger flowLogger) Fatalf(format string, params ...interface{}) {
	log.Fatalf(format, params...)
}

func (logger flowLogger) Warn(params ...interface{})  {}
func (logger flowLogger) Error(params ...interface{}) {}
func (logger flowLogger) Debug(params ...interface{}) {}

// NetflowModule represents a generic UDP forward module
// It starts a UDP Listener, and forwards the received data to OpenNMS without alteration
type NetflowModule struct {
	Name      string
	broker    api.Broker
	config    *api.MinionConfig
	listener  *api.MinionListener
	conn      *net.UDPConn
	processor *decoder.Processor
	stopping  bool
}

// GetID gets the ID of the sink module
func (module *NetflowModule) GetID() string {
	return module.Name
}

// Start initiates a Netflow UDP receiver
func (module *NetflowModule) Start(config *api.MinionConfig, broker api.Broker) error {
	module.stopping = false
	module.broker = broker
	module.config = config
	module.listener = config.GetListener(module.Name)
	if module.listener == nil {
		log.Warnf("Flow Module %s disabled", module.Name)
		return nil
	}
	var err error
	if module.conn, err = createUDPListener(module.listener.Port); err != nil {
		return err
	}
	var handler = module.getDecoderHandler()
	if handler == nil {
		log.Warnf("Flow Module %s disabled", module.Name)
		return nil
	}

	go func() {
		module.startProcessor(handler)
		payload := make([]byte, 9000)
		for {
			size, pktAddr, err := module.conn.ReadFromUDP(payload)
			if err != nil {
				if !module.stopping {
					log.Errorf("%s Cannot read from UDP: %s", module.Name, err)
				}
				continue
			}
			payloadCut := make([]byte, size)
			copy(payloadCut, payload[0:size])
			baseMessage := goflow.BaseMessage{
				Src:     pktAddr.IP,
				Port:    pktAddr.Port,
				Payload: payloadCut,
			}
			module.processor.ProcessMessage(baseMessage)
		}
	}()
	return nil
}

// Stop shutdowns the sink module
func (module *NetflowModule) Stop() {
	module.stopping = true
	if module.processor != nil {
		module.processor.Stop()
	}
	if module.conn != nil {
		module.conn.Close()
	}
}

// Publish represents the Transport interface implementation used by goflow
func (module *NetflowModule) Publish(msgs []*goflowMsg.FlowMessage) {
	log.Debugf("Received %d %s messages", len(msgs), module.Name)
	for _, flowmsg := range msgs {
		msg := module.convertToNetflow(flowmsg)
		buffer, _ := proto.Marshal(msg)
		if bytes := wrapMessageToTelemetry(module.config, net.IP(flowmsg.SamplerAddress).String(), uint32(module.listener.Port), buffer); bytes != nil {
			sendBytes("Telemetry-"+module.listener.Name, module.config, module.broker, bytes)
		}
	}
}

func (module *NetflowModule) getDecoderHandler() decoder.DecoderFunc {
	if module.listener == nil {
		return nil
	}
	if module.listener.Is(UDPNetflow5Parser) {
		netflow := goflow.StateNFLegacy{
			Transport: module,
			Logger:    flowLogger{},
		}
		return netflow.DecodeFlow
	} else if module.listener.Is(UDPNetflow9Parser) || module.listener.Is(UDPIpfixParser) {
		netflow := goflow.StateNetFlow{
			Transport: module,
			Logger:    flowLogger{},
		}
		netflow.InitTemplates()
		return netflow.DecodeFlow
	} else if module.listener.Is(UDPSFlowParser) {
		sflow := goflow.StateSFlow{
			Transport: module,
			Logger:    flowLogger{},
		}
		return sflow.DecodeFlow
	}
	return nil
}

func (module *NetflowModule) startProcessor(handler decoder.DecoderFunc) {
	if module.processor != nil {
		return
	}
	ecb := goflow.DefaultErrorCallback{
		Logger: flowLogger{},
	}
	decoderParams := decoder.DecoderParams{
		DecoderFunc:   handler,
		DoneCallback:  goflow.DefaultAccountCallback,
		ErrorCallback: ecb.Callback,
	}
	processor := decoder.CreateProcessor(1, decoderParams, module.Name)
	module.processor = &processor
	module.processor.Start()
}

func (module *NetflowModule) lookup(addr string) string {
	hostnames, err := net.LookupAddr(addr)
	if err != nil && len(hostnames) == 0 {
		return addr
	}
	return hostnames[0]
}

func (module *NetflowModule) convertToNetflow(flowmsg *goflowMsg.FlowMessage) *netflow.FlowMessage {
	srcAddress := net.IP(flowmsg.SrcAddr).String()
	dstAddress := net.IP(flowmsg.DstAddr).String()
	nextHopeAddress := net.IP(flowmsg.NextHop).String()
	var version netflow.NetflowVersion
	switch flowmsg.Type {
	case goflowMsg.FlowMessage_NETFLOW_V5:
		version = netflow.NetflowVersion_V5
	case goflowMsg.FlowMessage_NETFLOW_V9:
		version = netflow.NetflowVersion_V9
	case goflowMsg.FlowMessage_IPFIX:
		version = netflow.NetflowVersion_IPFIX
	}
	msg := &netflow.FlowMessage{
		NetflowVersion:    version,
		Direction:         netflow.Direction(flowmsg.FlowDirection),
		Timestamp:         flowmsg.TimeReceived * 1000,
		SrcAddress:        srcAddress,
		SrcHostname:       module.lookup(srcAddress),
		SrcPort:           &wrappers.UInt32Value{Value: flowmsg.SrcPort},
		SrcAs:             &wrappers.UInt64Value{Value: uint64(flowmsg.SrcAS)},
		SrcMaskLen:        &wrappers.UInt32Value{Value: flowmsg.SrcNet},
		DstAddress:        dstAddress,
		DstHostname:       module.lookup(dstAddress),
		DstPort:           &wrappers.UInt32Value{Value: flowmsg.DstPort},
		DstAs:             &wrappers.UInt64Value{Value: uint64(flowmsg.DstAS)},
		DstMaskLen:        &wrappers.UInt32Value{Value: flowmsg.DstNet},
		NextHopAddress:    nextHopeAddress,
		NextHopHostname:   module.lookup(nextHopeAddress),
		InputSnmpIfindex:  &wrappers.UInt32Value{Value: flowmsg.InIf},
		OutputSnmpIfindex: &wrappers.UInt32Value{Value: flowmsg.OutIf},
		FirstSwitched:     &wrappers.UInt64Value{Value: flowmsg.TimeFlowStart * 1000},
		LastSwitched:      &wrappers.UInt64Value{Value: flowmsg.TimeFlowEnd * 1000},
		TcpFlags:          &wrappers.UInt32Value{Value: flowmsg.TCPFlags},
		Protocol:          &wrappers.UInt32Value{Value: flowmsg.Proto},
		Tos:               &wrappers.UInt32Value{Value: flowmsg.IPTos},
		FlowSeqNum:        &wrappers.UInt64Value{Value: uint64(flowmsg.SequenceNum)},
		SamplingInterval:  &wrappers.DoubleValue{Value: float64(flowmsg.SamplingRate)},
		NumBytes:          &wrappers.UInt64Value{Value: flowmsg.Bytes},
		NumPackets:        &wrappers.UInt64Value{Value: flowmsg.Packets},
		EngineType:        &wrappers.UInt32Value{Value: flowmsg.Etype},
		Vlan:              &wrappers.UInt32Value{Value: flowmsg.VlanId},
	}
	return msg
}

var netflow5Module = &NetflowModule{Name: "Netflow-5"}
var netflow9Module = &NetflowModule{Name: "Netflow-9"}
var ipfixModule = &NetflowModule{Name: "IPFIX"}
var sflowModule = &NetflowModule{Name: "SFlow"}

func init() {
	api.RegisterSinkModule(netflow5Module)
	api.RegisterSinkModule(netflow9Module)
	api.RegisterSinkModule(ipfixModule)
	api.RegisterSinkModule(sflowModule)
}
