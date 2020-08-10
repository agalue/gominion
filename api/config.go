package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	return strings.ToLower(listener.GetParser()) == strings.ToLower(parser)
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
				return fmt.Errorf("Invalid port on listener CSV %s: %s", csv, err)
			}
		} else {
			return fmt.Errorf("Invalid listener CSV %s", csv)
		}
	}
	return nil
}

// GetListener gets a given listener by name
func (cfg *MinionConfig) GetListener(name string) *MinionListener {
	for _, listener := range cfg.Listeners {
		if strings.ToLower(listener.Name) == strings.ToLower(name) {
			return &listener
		}
	}
	return nil
}

// GetListenerByParser gets a given listener by parser name
func (cfg *MinionConfig) GetListenerByParser(parser string) *MinionListener {
	for _, listener := range cfg.Listeners {
		if strings.ToLower(listener.GetParser()) == strings.ToLower(parser) {
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
