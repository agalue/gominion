package sink

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/google/uuid"
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
	return "Syslogd"
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
	module.server.ListenUDP(listenAddr)
	module.server.ListenTCP(listenAddr)
	module.server.Boot()

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			module.handleLogParts(logParts)
		}
	}(channel)
}

// Stop shutdowns the sink module
func (module *SyslogModule) Stop() {
	module.server.Kill()
}

/*
{
  "client": "127.0.0.1:64557",
  "content": ": 2020 Jul 29 17:07:29 EDT: %ETHPORT-5-IF_DOWN_LINK_FAILURE: Interface eth1 is down (Link failure)",
  "facility": 23,
  "hostname": "127.0.0.1",
  "priority": 189,
  "severity": 5,
  "tag": "",
  "timestamp": "2020-07-29T17:54:29-04:00",
  "tls_peer": ""
}
*/
func (module *SyslogModule) handleLogParts(logParts map[string]interface{}) {
	clientParts := strings.Split(logParts["client"].(string), ":")
	clientPort, _ := strconv.Atoi(clientParts[1])
	messageLog := api.SyslogMessageLogDTO{
		Location:      module.config.Location,
		SystemID:      module.config.ID,
		SourceAddress: clientParts[0],
		SourcePort:    clientPort,
	}
	timestamp := logParts["timestamp"].(time.Time)
	if logParts["content"].(string) == "X" {
		return
	}
	content := fmt.Sprintf("<%d>%s", logParts["priority"].(int), logParts["content"].(string))
	message := api.SyslogMessageDTO{
		Timestamp: timestamp.Format(api.TimeFormat),
		Content:   []byte(base64.StdEncoding.EncodeToString([]byte(content))),
	}
	messageLog.AddMessage(message)
	module.sendSinkMessage(messageLog)
}

func (module *SyslogModule) sendSinkMessage(messageLog api.SyslogMessageLogDTO) {
	bytes, _ := xml.MarshalIndent(messageLog, "", "   ")
	msg := &ipc.SinkMessage{
		MessageId: uuid.New().String(),
		ModuleId:  "Syslog",
		SystemId:  module.config.ID,
		Location:  module.config.Location,
		Content:   bytes,
	}
	if err := module.broker.Send(msg); err != nil {
		log.Printf("Error while sending Syslog message: %v", err)
	}
}

var syslogModule = &SyslogModule{}

func init() {
	api.RegisterSinkModule(syslogModule)
}
