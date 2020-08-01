package sink

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/agalue/gominion/api"
	"gopkg.in/mcuadros/go-syslog.v2"
)

// SyslogModule represents the heartbeat module
type SyslogModule struct {
	broker api.Broker
	config *api.MinionConfig
	server *syslog.Server
}

// GetID gets the ID of the sink module
func (module *SyslogModule) GetID() string {
	return "Syslog"
}

// Start initiates a blocking loop with the Syslog Listener
func (module *SyslogModule) Start(config *api.MinionConfig, broker api.Broker) {
	log.Printf("Starting Syslog receiver on port UDP/TCP %d", config.SyslogPort)
	module.config = config
	module.broker = broker

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	listenAddr := fmt.Sprintf("0.0.0.0:%d", config.SyslogPort)
	module.server = syslog.NewServer()
	module.server.SetFormat(syslog.Automatic)
	module.server.SetHandler(handler)
	if err := module.server.ListenUDP(listenAddr); err != nil {
		log.Fatalf("Cannot start Syslog UDP listener: %s", err)
	}
	if err := module.server.ListenTCP(listenAddr); err != nil {
		log.Fatalf("Cannot start Syslog TCP listener: %s", err)
	}
	module.server.Boot()

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			if messageLog := module.buildMessageLog(logParts); messageLog != nil {
				sendResponse(module.GetID(), module.config, module.broker, messageLog)
			}
		}
	}(channel)
}

// Stop shutdowns the sink module
func (module *SyslogModule) Stop() {
	module.server.Kill()
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
	log.Printf("Received Syslog message from %s\n", messageLog.SourceAddress)
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
