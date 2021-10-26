package api

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/agalue/gominion/protobuf/ipc"
)

// MinionListener represents a Minion Listener
type MinionListener struct {
	Name       string            `yaml:"name" json:"name"`
	Parser     string            `yaml:"parser" json:"parser"`
	Port       int               `yaml:"port" json:"port"`
	Properties map[string]string `yaml:"properties,omitempty" json:"properties,omitempty"`
}

// GetParser returns the simple class name for the parser implementation
func (listener *MinionListener) GetParser() string {
	if listener.Parser == "" {
		return ""
	}
	sections := strings.Split(listener.Parser, ".")
	return sections[len(sections)-1]
}

// Is returns true if the listener matches a given parser
func (listener *MinionListener) Is(parser string) bool {
	return strings.EqualFold(listener.GetParser(), parser)
}

// CircuitBreakerConfig Circuit Breaker Configuration
type CircuitBreakerConfig struct {
	MaxRequests uint32 `yaml:"maxRequests,omitempty" json:"maxRequests,omitempty"`
	Interval    int    `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     int    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// DNSConfig DNS Configuration
type DNSConfig struct {
	NameServer           string               `yaml:"nameServer,omitempty" json:"nameServer,omitempty"`
	Timeout              int                  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	CacheRefreshDuration int                  `yaml:"cacheRefreshDuration,omitempty" json:"cacheRefreshDuration,omitempty"`
	CircuitBreaker       CircuitBreakerConfig `yaml:"circuitBreaker" json:"circuitBreaker"`
}

// MinionConfig represents basic Minion Configuration
type MinionConfig struct {
	ID               string            `yaml:"id" json:"id"`
	Location         string            `yaml:"location" json:"location"`
	BrokerURL        string            `yaml:"brokerUrl" json:"brokerUrl"`
	BrokerType       string            `yaml:"brokerType" json:"brokerType"`
	BrokerProperties map[string]string `yaml:"brokerProperties,omitempty" json:"brokerProperties,omitempty"`
	TrapPort         int               `yaml:"trapPort" json:"traPort"`
	SyslogPort       int               `yaml:"syslogPort" json:"syslogPort"`
	StatsPort        int               `yaml:"statsPort" json:"statsPort"`
	LogLevel         string            `yaml:"logLevel" json:"logLevel"`
	DNS              *DNSConfig        `yaml:"dns,omitempty" json:"dns,omitempty"`
	Listeners        []MinionListener  `yaml:"listeners,omitempty" json:"listeners,omitempty"`
}

// ParseListeners parses an array of listeners in CSV format
func (cfg *MinionConfig) ParseListeners(csvs []string) error {
	for _, csv := range csvs {
		parts := strings.Split(csv, ",")
		if len(parts) == 3 {
			if port, err := strconv.Atoi(parts[1]); err == nil {
				listener := MinionListener{Name: parts[0], Parser: parts[2], Port: port}
				cfg.Listeners = append(cfg.Listeners, listener)
			} else {
				return fmt.Errorf("invalid port on listener CSV %s: %s", csv, err)
			}
		} else {
			return fmt.Errorf("invalid listener CSV %s", csv)
		}
	}
	return nil
}

func (cfg *MinionConfig) GetBrokerProperty(property string) string {
	if cfg.BrokerProperties == nil {
		return ""
	}
	if value, ok := cfg.BrokerProperties[property]; ok {
		return value
	}
	return ""
}

// GetListener gets a given listener by name
func (cfg *MinionConfig) GetListener(name string) *MinionListener {
	for _, listener := range cfg.Listeners {
		if strings.EqualFold(listener.Name, name) {
			return &listener
		}
	}
	return nil
}

// GetListenerByParser gets a given listener by parser name
func (cfg *MinionConfig) GetListenerByParser(parser string) *MinionListener {
	for _, listener := range cfg.Listeners {
		if strings.EqualFold(listener.GetParser(), parser) {
			return &listener
		}
	}
	return nil
}

func (cfg *MinionConfig) String() string {
	bytes, _ := json.MarshalIndent(cfg, "", "  ")
	return string(bytes)
}

// IsValid returns an error if the configuration is not valid
func (cfg *MinionConfig) IsValid() error {
	if cfg.ID == "" {
		return fmt.Errorf("minion ID required")
	}
	if cfg.Location == "" {
		return fmt.Errorf("location required")
	}
	if cfg.BrokerURL == "" {
		return fmt.Errorf("broker URL required")
	}
	if cfg.DNS != nil && cfg.DNS.NameServer != "" {
		ip := net.ParseIP(cfg.DNS.NameServer)
		if ip == nil {
			return fmt.Errorf("invalid DNS name server")
		}
	}
	return nil
}

// GetHeaderResponse builds an RPC response with the headers context
func (cfg *MinionConfig) GetHeaderResponse() *ipc.RpcResponseProto {
	return &ipc.RpcResponseProto{
		ModuleId: "MINION_HEADERS",
		Location: cfg.Location,
		SystemId: cfg.ID,
		RpcId:    cfg.ID,
	}
}
