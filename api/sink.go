package api

// SinkModule represents the Sink Module interface
type SinkModule interface {
	GetID() string
	Start(config *MinionConfig, broker Broker)
	Stop()
}
