package sink

import (
	"github.com/agalue/gominion/api"
)

// CreateSinkRegistry creates a new Sink registry with all the available implementations
func CreateSinkRegistry() *api.SinkRegistry {
	registry := new(api.SinkRegistry)
	registry.Init()

	registry.RegisterModule(&NetflowModule{name: "Netflow-5", goflowID: "NetFlowV5"})
	registry.RegisterModule(&NetflowModule{name: "Netflow-9", goflowID: "NetFlow"})
	registry.RegisterModule(&NetflowModule{name: "IPFIX", goflowID: "NetFlow"})
	registry.RegisterModule(&NetflowModule{name: "SFlow", goflowID: "sFlow"})

	registry.RegisterModule(&HeartbeatModule{})
	registry.RegisterModule(&NxosGrpcModule{})
	registry.RegisterModule(&SyslogModule{})
	registry.RegisterModule(&SnmpTrapModule{})
	registry.RegisterModule(&UDPForwardModule{name: "Graphite"})

	return registry
}
