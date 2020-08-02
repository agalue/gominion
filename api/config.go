package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MinionListener represents a Minion Listener
type MinionListener struct {
	Name   string `yaml:"name" json:"name"`
	Parser string `yaml:"parser" json:"parser"`
	Port   int    `yaml:"port" json:"port"`
}

// GetParser returns the simple class name for the parser implementation
func (listener *MinionListener) GetParser() string {
	if listener.Parser == "" {
		return ""
	}
	sections := strings.Split(listener.Parser, ".")
	return sections[len(sections)-1]
}

// MinionConfig represents basic Minion Configuration
type MinionConfig struct {
	ID               string            `yaml:"id" json:"id"`
	Location         string            `yaml:"location" json:"location"`
	BrokerURL        string            `yaml:"broker_url" json:"broker_url"`
	BrokerType       string            `yaml:"broker_type" json:"broker_type"`
	BrokerProperties map[string]string `yaml:"broker_properties,omitempty" json:"broker_properties,omitempty"`
	TrapPort         int               `yaml:"trap_port" json:"trap_port"`
	SyslogPort       int               `yaml:"syslog_port" json:"syslog_port"`
	NxosGrpcPort     int               `yaml:"nxos_grpc_port" json:"nxos_grpc_port"`
	Listeners        []MinionListener  `yaml:"listeners" json:"listeners"`
}

// GetListener gets a given listener by name
func (cfg *MinionConfig) GetListener(name string) *MinionListener {
	for _, listener := range cfg.Listeners {
		if listener.Name == name {
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
		return fmt.Errorf("Minion ID required")
	}
	if cfg.Location == "" {
		return fmt.Errorf("Location required")
	}
	if cfg.BrokerURL == "" {
		return fmt.Errorf("Broker URL required")
	}
	return nil
}
