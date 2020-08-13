package broker

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics represents the broker metric set per module
type Metrics struct {
	SinkMsgDeliverySucceeded *prometheus.CounterVec // Sink messages successfully delivered
	SinkMsgDeliveryFailed    *prometheus.CounterVec // Failed attempts to send Sink messages
	RPCReqReceivedSucceeded  *prometheus.CounterVec // RPC requests successfully received
	RPCReqReceivedFailed     *prometheus.CounterVec // Failed attempts to receive RPC requests
	RPCReqProcessedSucceeded *prometheus.CounterVec // RPC requests successfully processed
	RPCReqProcessedFailed    *prometheus.CounterVec // Failed attempts to process RPC requests
	RPCResSentSucceeded      *prometheus.CounterVec // RPC responses successfully sent
	RPCResSentFailed         *prometheus.CounterVec // Failed attempts to send RPC responses
}

// Register register all prometheus metrics
func (m *Metrics) Register() {
	prometheus.MustRegister(
		m.SinkMsgDeliverySucceeded,
		m.SinkMsgDeliveryFailed,
		m.RPCReqReceivedSucceeded,
		m.RPCReqReceivedFailed,
		m.RPCReqProcessedSucceeded,
		m.RPCReqProcessedFailed,
		m.RPCResSentSucceeded,
		m.RPCResSentFailed,
	)
}

// NewMetrics returns a new Metrics object
func NewMetrics() Metrics {
	return Metrics{
		SinkMsgDeliverySucceeded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "onms_sink_messages_delivery_succeeded",
			Help: "The total number of Sink messages successfully delivered per module",
		}, []string{"module"}),
		SinkMsgDeliveryFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "onms_sink_messages_delivery_failed",
			Help: "The total number of failed attempts to send Sink messages per module",
		}, []string{"module"}),
		RPCReqReceivedSucceeded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "onms_rpc_requests_received_succeeded",
			Help: "The total number of RPC requests successfully received per module",
		}, []string{"module"}),
		RPCReqReceivedFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "onms_rpc_requests_received_failed",
			Help: "The total number of failed attempts to receive RPC messages per module",
		}, []string{"module"}),
		RPCReqProcessedSucceeded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "onms_rpc_requests_processed_succeeded",
			Help: "The total number of RPC requests successfully processed per module",
		}, []string{"module"}),
		RPCReqProcessedFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "onms_rpc_requests_processed_failed",
			Help: "The total number of failed attempts to process RPC messages per module",
		}, []string{"module"}),
		RPCResSentSucceeded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "onms_rpc_responses_sent_succeeded",
			Help: "The total number of RPC responses successfully sent per module",
		}, []string{"module"}),
		RPCResSentFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "onms_rpc_responses_sent_failed",
			Help: "The total number of failed attempts to send RPC responses per module",
		}, []string{"module"}),
	}
}
