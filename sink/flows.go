package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/netflow"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/dnscache"
	"github.com/sony/gobreaker"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	decoder "github.com/cloudflare/goflow/v3/decoders"
	goflowMsg "github.com/cloudflare/goflow/v3/pb"
	goflow "github.com/cloudflare/goflow/v3/utils"
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
	name      string
	goflowID  string
	sink      api.Sink
	config    *api.MinionConfig
	listener  *api.MinionListener
	conn      *net.UDPConn
	processor *decoder.Processor
	stopping  bool
	resolver  *dnscache.Resolver
	breaker   *gobreaker.CircuitBreaker
}

// GetID gets the ID of the sink module
func (module *NetflowModule) GetID() string {
	return module.name
}

// Start initiates a Netflow UDP receiver
func (module *NetflowModule) Start(config *api.MinionConfig, sink api.Sink) error {
	module.stopping = false
	module.sink = sink
	module.config = config
	module.listener = config.GetListener(module.name)
	if module.listener == nil {
		log.Warnf("Flow Module %s disabled", module.name)
		return nil
	}
	var err error
	if module.conn, err = createUDPListener(module.listener.Port); err != nil {
		return err
	}
	var handler = module.getDecoderHandler()
	if handler == nil {
		log.Warnf("Flow Module %s disabled", module.name)
		return nil
	}
	log.Infof("Starting %s flow receiver on port UDP %d", module.name, module.listener.Port)
	module.initDNSResolver()
	module.initCircuitBreaker()
	module.startProcessor(handler)

	localIP := module.conn.LocalAddr().String()

	go func() {
		payload := make([]byte, 9000)
		for {
			size, pktAddr, err := module.conn.ReadFromUDP(payload)
			if err != nil {
				if !module.stopping {
					log.Errorf("%s Cannot read from UDP: %s", module.name, err)
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
			if module.config.StatsPort > 0 {
				goflow.MetricTrafficBytes.With(
					prometheus.Labels{
						"remote_ip":   pktAddr.IP.String(),
						"remote_port": strconv.Itoa(pktAddr.Port),
						"local_ip":    localIP,
						"local_port":  strconv.Itoa(module.listener.Port),
						"type":        module.goflowID,
					}).
					Add(float64(size))
				goflow.MetricTrafficPackets.With(
					prometheus.Labels{
						"remote_ip":   pktAddr.IP.String(),
						"remote_port": strconv.Itoa(pktAddr.Port),
						"local_ip":    localIP,
						"local_port":  strconv.Itoa(module.listener.Port),
						"type":        module.goflowID,
					}).
					Inc()
				goflow.MetricPacketSizeSum.With(
					prometheus.Labels{
						"remote_ip":   pktAddr.IP.String(),
						"remote_port": strconv.Itoa(pktAddr.Port),
						"local_ip":    localIP,
						"local_port":  strconv.Itoa(module.listener.Port),
						"type":        module.goflowID,
					}).
					Observe(float64(size))
			}
		}
	}()
	return nil
}

// Stop shutdowns the sink module
func (module *NetflowModule) Stop() {
	log.Warnf("Stopping %s flow receiver", module.name)
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
	messages := make([][]byte, len(msgs))
	sourceAddress := ""
	for idx, flowmsg := range msgs {
		if sourceAddress == "" {
			sourceAddress = net.IP(flowmsg.SamplerAddress).String()
		}
		if flowmsg.Type == goflowMsg.FlowMessage_SFLOW_5 {
			// DEBUG: start
			if bytes, err := json.MarshalIndent(flowmsg, "", "  "); err != nil {
				log.Debugf("SFlow message received %s", string(bytes))
			}
			// DEBUG: end
		} else {
			msg := module.convertToNetflow(flowmsg)
			buffer, _ := proto.Marshal(msg)
			messages[idx] = buffer
		}
	}
	if bytes := wrapMessageToTelemetry(module.config, sourceAddress, uint32(module.listener.Port), messages); bytes != nil {
		sendBytes("Telemetry-"+module.listener.Name, module.config, module.sink, bytes)
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
	processor := decoder.CreateProcessor(module.getWorkers(), decoderParams, module.goflowID)
	module.processor = &processor
	module.processor.Start()
}

// DNS processing can slow down flow processing, which is why reverse DNS is disabled by default
func (module *NetflowModule) lookup(addr string) ([]string, error) {
	body, err := module.breaker.Execute(func() (interface{}, error) {
		return module.resolver.LookupAddr(context.Background(), addr)
	})
	if err != nil {
		return nil, err
	}
	return body.([]string), err
}

func (module *NetflowModule) initDNSResolver() {
	if module.resolver != nil {
		return
	}
	dns := module.config.DNS
	module.resolver = &dnscache.Resolver{}
	if dns != nil {
		if dns.Timeout > 0 {
			module.resolver.Timeout = time.Duration(dns.Timeout) * time.Microsecond
		}
		if dns.NameServer != "" {
			module.resolver.Resolver = &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{}
					return d.DialContext(ctx, "udp", fmt.Sprintf("%s:%d", dns.NameServer, 53))
				},
			}
		}
	}
	go func() {
		duration := 30 * time.Minute
		if dns != nil && dns.CacheRefreshDuration > 0 {
			duration = time.Duration(dns.CacheRefreshDuration) * time.Microsecond
		}
		t := time.NewTicker(duration)
		defer t.Stop()
		for range t.C {
			module.resolver.Refresh(true)
		}
	}()
}

func (module *NetflowModule) initCircuitBreaker() {
	if module.breaker != nil {
		return
	}
	config := gobreaker.Settings{Name: module.GetID()}
	dns := module.config.DNS
	if dns != nil {
		cb := dns.CircuitBreaker
		config.MaxRequests = cb.MaxRequests
		if cb.Interval > 0 {
			config.Interval = time.Duration(cb.Interval) * time.Millisecond
		}
		if cb.Timeout > 0 {
			config.Timeout = time.Duration(cb.Timeout) * time.Millisecond
		}
	}
	module.breaker = gobreaker.NewCircuitBreaker(config)
}

func (module *NetflowModule) isReverseDNSEnabled() bool {
	value, ok := module.listener.Properties["dnsLookupsEnabled"]
	return ok && value == "true"
}

func (module *NetflowModule) getWorkers() int {
	value, ok := module.listener.Properties["workers"]
	if ok {
		w, err := strconv.Atoi(value)
		if err != nil && w > 0 {
			return w
		}
	}
	return runtime.NumCPU()
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
		SrcPort:           &wrapperspb.UInt32Value{Value: flowmsg.SrcPort},
		SrcAs:             &wrapperspb.UInt64Value{Value: uint64(flowmsg.SrcAS)},
		SrcMaskLen:        &wrapperspb.UInt32Value{Value: flowmsg.SrcNet},
		DstAddress:        dstAddress,
		DstPort:           &wrapperspb.UInt32Value{Value: flowmsg.DstPort},
		DstAs:             &wrapperspb.UInt64Value{Value: uint64(flowmsg.DstAS)},
		DstMaskLen:        &wrapperspb.UInt32Value{Value: flowmsg.DstNet},
		NextHopAddress:    nextHopeAddress,
		InputSnmpIfindex:  &wrapperspb.UInt32Value{Value: flowmsg.InIf},
		OutputSnmpIfindex: &wrapperspb.UInt32Value{Value: flowmsg.OutIf},
		FirstSwitched:     &wrapperspb.UInt64Value{Value: flowmsg.TimeFlowStart * 1000},
		LastSwitched:      &wrapperspb.UInt64Value{Value: flowmsg.TimeFlowEnd * 1000},
		TcpFlags:          &wrapperspb.UInt32Value{Value: flowmsg.TCPFlags},
		Protocol:          &wrapperspb.UInt32Value{Value: flowmsg.Proto},
		IpProtocolVersion: &wrapperspb.UInt32Value{Value: flowmsg.Etype},
		Tos:               &wrapperspb.UInt32Value{Value: flowmsg.IPTos},
		FlowSeqNum:        &wrapperspb.UInt64Value{Value: uint64(flowmsg.SequenceNum)},
		SamplingInterval:  &wrapperspb.DoubleValue{Value: float64(flowmsg.SamplingRate)},
		NumBytes:          &wrapperspb.UInt64Value{Value: flowmsg.Bytes},
		NumPackets:        &wrapperspb.UInt64Value{Value: flowmsg.Packets},
		Vlan:              &wrapperspb.UInt32Value{Value: flowmsg.VlanId},
	}
	if module.isReverseDNSEnabled() {
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			if array, err := module.lookup(srcAddress); err != nil && len(array) > 0 {
				msg.SrcHostname = array[0]
			}
		}()
		go func() {
			defer wg.Done()
			if array, err := module.lookup(dstAddress); err != nil && len(array) > 0 {
				msg.DstHostname = array[0]
			}
		}()
		wg.Wait()
	}
	return msg
}
