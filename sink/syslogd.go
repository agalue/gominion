package sink

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"gopkg.in/mcuadros/go-syslog.v2"
)

// SyslogModule represents the heartbeat module
type SyslogModule struct {
	broker  api.Broker
	config  *api.MinionConfig
	server  *syslog.Server
	channel syslog.LogPartsChannel
}

// GetID gets the ID of the sink module
func (module *SyslogModule) GetID() string {
	return "Syslog"
}

// Start initiates a Syslog UDP and TCP receiver
func (module *SyslogModule) Start(config *api.MinionConfig, broker api.Broker) error {
	if config.SyslogPort == 0 {
		log.Warnf("Syslog Module disabled")
		return nil
	}

	log.Infof("Starting Syslog receiver on port UDP/TCP %d", config.SyslogPort)

	module.config = config
	module.broker = broker

	listenAddr := fmt.Sprintf("0.0.0.0:%d", config.SyslogPort)
	module.channel = make(syslog.LogPartsChannel)
	module.server = syslog.NewServer()
	module.server.SetFormat(syslog.Automatic)
	module.server.SetHandler(syslog.NewChannelHandler(module.channel))
	if err := module.server.ListenUDP(listenAddr); err != nil {
		return fmt.Errorf("Cannot start Syslog UDP listener: %s", err)
	}
	if err := module.server.ListenTCP(listenAddr); err != nil {
		return fmt.Errorf("Cannot start Syslog TCP listener: %s", err)
	}
	if err := module.server.Boot(); err != nil {
		return fmt.Errorf("Cannot boot Syslog server: %s", err)
	}
	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			if messageLog := module.buildMessageLog(logParts); messageLog != nil {
				sendResponse(module.GetID(), module.config, module.broker, messageLog)
			}
		}
	}(module.channel)
	return nil
}

// Stop shutdowns the sink module
func (module *SyslogModule) Stop() {
	log.Warnf("Stopping Syslog receiver")
	if module.server != nil {
		close(module.channel)
		module.server.Kill()
	}
}

func (module *SyslogModule) buildMessageLog(logParts map[string]interface{}) *api.SyslogMessageLogDTO {
	if logParts["content"].(string) == "X" {
		return nil
	}
	clientParts := strings.Split(logParts["client"].(string), ":")
	clientPort, _ := strconv.Atoi(clientParts[1])
	messageLog := &api.SyslogMessageLogDTO{
		Location:      module.config.Location,
		SystemID:      module.config.ID,
		SourceAddress: clientParts[0],
		SourcePort:    clientPort,
	}
	timestamp := logParts["timestamp"].(time.Time)
	log.Debugf("Received Syslog message from %s\n", messageLog.SourceAddress)
	content := fmt.Sprintf("<%d>%s", logParts["priority"].(int), logParts["content"].(string))
	message := api.SyslogMessageDTO{
		Timestamp: timestamp.Format(api.TimeFormat),
		Content:   []byte(base64.StdEncoding.EncodeToString([]byte(content))),
	}
	messageLog.AddMessage(message)
	return messageLog
}

var syslogModule = &SyslogModule{}

func init() {
	api.RegisterSinkModule(syslogModule)
}
