package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/agalue/gominion/api"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"gotest.tools/v3/assert"
)

func TestMinionConfig(t *testing.T) {
	var err error
	configYAML := []byte(`---
id: go-minion1
location: Test
brokerUrl: 10.0.0.100:8990
brokerProperties:
  tls-enabled: "true"
trapPort: 11162
syslogPort: 11514
dns:
  nameServer: 8.8.8.8
  timeout: 1000
  cacheRefreshDuration: 300000
  circuitBreaker:
    timeout: 1000
    interval: 60000
    maxRequests: 5
listeners:
- name: Netflow-5
  port: 18877
  parser: Netflow5UdpParser
  properties:
    workers: "4"
- name: Netflow-9
  port: 14729
  parser: Netflow9UdpParser
`)

	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(configYAML))
	assert.NilError(t, err)

	config := &api.MinionConfig{
		DNS: &api.DNSConfig{
			CircuitBreaker: api.CircuitBreakerConfig{
				MaxRequests: 1,
			},
		},
	}
	err = viper.Unmarshal(config)
	assert.NilError(t, err)

	bytes, err := yaml.Marshal(config)
	assert.NilError(t, err)
	fmt.Println(string(bytes))

	assert.NilError(t, config.IsValid())
	assert.Equal(t, "go-minion1", config.ID)
	assert.Equal(t, 2, len(config.Listeners))
	assert.Assert(t, config.BrokerProperties != nil)
	assert.Equal(t, "true", config.BrokerProperties["tls-enabled"])

	assert.Equal(t, "8.8.8.8", config.DNS.NameServer)
	assert.Equal(t, uint32(5), config.DNS.CircuitBreaker.MaxRequests)

	netflow := config.GetListener("Netflow-5")
	if netflow == nil {
		t.FailNow()
	} else {
		assert.Equal(t, 18877, netflow.Port)
		assert.Assert(t, netflow.Is("Netflow5UdpParser"))
		assert.Assert(t, netflow.Properties != nil)
		assert.Equal(t, "4", netflow.Properties["workers"])
	}
}
