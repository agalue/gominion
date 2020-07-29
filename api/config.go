package api

import (
	"encoding/json"
	"fmt"
)

// MinionConfig represents basic Minion Configuration
type MinionConfig struct {
	ID               string            `yaml:"id" json:"id"`
	Location         string            `yaml:"location" json:"location"`
	OnmsURL          string            `yaml:"onms_url" json:"onms_url"`
	BrokerURL        string            `yaml:"broker_url" json:"broker_url"`
	BrokerProperties map[string]string `yaml:"broker_properties,omitempty" json:"broker_properties,omitempty"`
	TrapPort         int               `yaml:"trap_port" json:"trap_port"`
	SyslogPort       int               `yaml:"syslog_port" json:"syslog_port"`
}

func (cfg *MinionConfig) String() string {
	bytes, _ := json.MarshalIndent(cfg, "", "  ")
	return string(bytes)
}

// IsValid returns an error if the configuration is not valid
func (cfg *MinionConfig) IsValid() error {
	if cfg.Location == "" {
		return fmt.Errorf("Location required")
	}
	if cfg.OnmsURL == "" {
		return fmt.Errorf("OpenNMS URL required")
	}
	if cfg.BrokerURL == "" {
		return fmt.Errorf("Broker URL required")
	}
	if cfg.TrapPort == 0 {
		return fmt.Errorf("SNMP Trap port required")
	}
	if cfg.SyslogPort == 0 {
		return fmt.Errorf("Syslog port required")
	}
	return nil
}
