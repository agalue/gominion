package broker

import (
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/agalue/gominion/protobuf/rpc"
	"github.com/agalue/gominion/protobuf/sink"
	"github.com/google/uuid"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"google.golang.org/protobuf/proto"
)

// KafkaClient represents the Kafka client implementation for the OpenNMS IPC API.
type KafkaClient struct {
	config        *api.MinionConfig
	registry      *api.SinkRegistry
	producer      *kafka.Producer
	consumer      *kafka.Consumer
	traceCloser   io.Closer
	metrics       *api.Metrics
	maxBufferSize int
	instanceID    string
	msgBuffer     map[string][]byte
	chunkTracker  map[string]int32
}

// Start initializes the Kafka client.
// Returns an error when the configuration is incorrect or cannot connect to the server.
func (cli *KafkaClient) Start() error {
	var err error
	if cli.config == nil {
		return fmt.Errorf("minion configuration required")
	}
	if cli.registry == nil {
		return fmt.Errorf("sink registry required")
	}
	if cli.metrics == nil {
		return fmt.Errorf("prometheus Metrics required")
	}

	cli.msgBuffer = make(map[string][]byte)
	cli.chunkTracker = make(map[string]int32)

	// Maximum size of the buffer to split messages in chunks
	cli.maxBufferSize, err = strconv.Atoi(cli.config.GetBrokerProperty("max-buffer-size"))
	if err != nil {
		cli.maxBufferSize = 1024
	}

	// The OpenNMS Instance ID (org.opennms.instance.id), for Kafka topics
	cli.instanceID = cli.config.GetBrokerProperty("instance-id")
	if cli.instanceID == "" {
		cli.instanceID = "OpenNMS"
	}

	if cli.traceCloser, err = initTracing(cli.config); err != nil {
		return err
	}

	// Creating Kafka Producer
	// TODO Parse external settings
	producerCfg := &kafka.ConfigMap{
		"bootstrap.servers": cli.config.BrokerURL,
	}

	// enable SSL for producer
	if cli.config.GetBrokerProperty("tls-enabled") == "true" {
		log.Infof("Enabling TLS")
		producerCfg.SetKey("security.protocol", "ssl")
	}

	if cli.producer, err = kafka.NewProducer(producerCfg); err != nil {
		return fmt.Errorf("could not create producer: %v", err)
	}

	// Creating Kafka Consumer
	// TODO Parse external settings
	consumerCfg := &kafka.ConfigMap{
		"bootstrap.servers":       cli.config.BrokerURL,
		"group.id":                cli.config.Location,
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 1000,
	}

	// enable SSL for consumer
	if cli.config.GetBrokerProperty("tls-enabled") == "true" {
		log.Infof("Enabling TLS")
		consumerCfg.SetKey("security.protocol", "ssl")
	}

	if cli.consumer, err = kafka.NewConsumer(consumerCfg); err != nil {
		return fmt.Errorf("could not create consumer: %v", err)
	}

	// Starting Sink Modules
	if err := cli.registry.StartModules(cli.config, cli); err != nil {
		return err
	}

	// Subscribe to RPC Requests
	topic := fmt.Sprintf("%s.%s.rpc-request", cli.instanceID, cli.config.Location)
	if err := cli.consumer.Subscribe(topic, nil); err != nil {
		return fmt.Errorf("cannot subscribe to topic %s: %v", topic, err)
	}

	go func() {
		log.Infof("starting producer message logger")
		for e := range cli.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Errorf("kafka delivery failed: %v", ev.TopicPartition)
				}
			default:
				log.Debugf("kafka event: %s", ev)
			}
		}
	}()

	go func() {
		log.Infof("starting RPC consumer for location %s", cli.config.Location)
		for {
			event := cli.consumer.Poll(100)
			switch e := event.(type) {
			case *kafka.Message:
				rpc := new(rpc.RpcMessageProto)
				if err := proto.Unmarshal(e.Value, rpc); err == nil {
					cli.metrics.RPCReqReceivedSucceeded.WithLabelValues(rpc.SystemId, rpc.ModuleId).Inc()
					cli.processRequest(rpc)
				} else {
					cli.metrics.RPCReqReceivedFailed.WithLabelValues(rpc.SystemId, rpc.ModuleId).Inc()
					log.Errorf("Cannot process RPC Request: %v", err)
				}
			case kafka.Error:
				log.Errorf("kafka consumer error %v", e)
			}
		}
	}()

	return nil
}

// Stop finalizes the Kafka client and all its dependencies.
func (cli *KafkaClient) Stop() {
	cli.registry.StopModules()
	log.Warnf("Stopping Kafka client")
	cli.consumer.Unsubscribe()
	cli.consumer.Close()
	cli.producer.Close()
	if cli.traceCloser != nil {
		cli.traceCloser.Close()
	}
	log.Infof("Good bye")
}

// Send forwards a Sink API message to Kafka.
// Messages are discarded when the brokers are unavailable.
func (cli *KafkaClient) Send(msg *ipc.SinkMessage) error {
	trace := startSpanForSinkMessage(msg)
	defer trace.Finish()
	totalChunks := cli.getTotalChunks(msg.Content)
	var chunk int32
	var err error
	topic := fmt.Sprintf("%s.Sink.%s", cli.instanceID, msg.ModuleId)
	for chunk = 0; chunk < totalChunks; chunk++ {
		bytes := cli.wrapMessageToSink(msg, chunk, totalChunks)
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(uuid.New().String()),
			Value:          bytes,
		}
		if err = cli.producer.Produce(msg, nil); err != nil {
			break
		}
	}
	if err != nil {
		cli.metrics.SinkMsgDeliveryFailed.WithLabelValues(msg.SystemId, msg.ModuleId).Inc()
		trace.SetTag("failed", "true")
		trace.LogKV("event", err.Error())
		return fmt.Errorf("cannot send message to %s: %v", topic, err)
	}
	cli.metrics.SinkMsgDeliverySucceeded.WithLabelValues(msg.SystemId, msg.ModuleId).Inc()
	return nil
}

func (cli *KafkaClient) getTotalChunks(data []byte) int32 {
	if cli.maxBufferSize == 0 {
		return int32(1)
	}
	chunks := int32(math.Ceil(float64(len(data) / cli.maxBufferSize)))
	if len(data)%cli.maxBufferSize > 0 {
		chunks++
	}
	return chunks
}

func (cli *KafkaClient) wrapMessageToSink(request *ipc.SinkMessage, chunk, totalChunks int32) []byte {
	bufferSize := cli.getRemainingBufferSize(int32(len(request.Content)), chunk)
	offset := chunk * int32(cli.maxBufferSize)
	msg := request.Content[offset : offset+bufferSize]
	sinkMsg := &sink.SinkMessage{
		MessageId:          request.MessageId,
		CurrentChunkNumber: chunk,
		TotalChunks:        totalChunks,
		Content:            msg,
	}
	bytes, err := proto.Marshal(sinkMsg)
	if err != nil {
		return []byte{}
	}
	return bytes
}

func (cli *KafkaClient) getRemainingBufferSize(messageSize, chunk int32) int32 {
	if cli.maxBufferSize > 0 && messageSize > int32(cli.maxBufferSize) {
		remaining := messageSize - chunk*int32(cli.maxBufferSize)
		if remaining > int32(cli.maxBufferSize) {
			return int32(cli.maxBufferSize)
		}
		return remaining
	}
	return messageSize
}

// Processes an RPC API request sent by OpenNMS asynchronously within a goroutine and sends back the response from the module.
func (cli *KafkaClient) processRequest(request *rpc.RpcMessageProto) {
	// Process chunks
	chunk := request.CurrentChunkNumber + 1 // Chunks starts at 0
	log.Debugf("%s RPC chunk %d of %d for %s received", request.ModuleId, chunk, request.TotalChunks, request.RpcId)
	if chunk != request.TotalChunks {
		if cli.chunkTracker[request.RpcId] < chunk {
			// Adds partial message to the buffer
			cli.msgBuffer[request.RpcId] = append(cli.msgBuffer[request.RpcId], request.RpcContent...)
			cli.chunkTracker[request.RpcId] = chunk
		} else {
			log.Warnf("Chunk %d from %s was already processed, ignoring...", chunk, request.RpcId)
		}
		return
	}
	// Retrieve the complete message from the buffer
	var data []byte
	if request.TotalChunks == 1 { // Handle special case chunk == total == 1
		data = request.RpcContent
	} else {
		data = append(cli.msgBuffer[request.RpcId], request.RpcContent...)
	}
	delete(cli.msgBuffer, request.RpcId)
	delete(cli.chunkTracker, request.RpcId)
	// Process RPC request
	log.Debugf("Received RPC request with ID %s for module %s", request.RpcId, request.ModuleId)
	if module, ok := api.GetRPCModule(request.ModuleId); ok {
		go func() {
			var err error
			req := &ipc.RpcRequestProto{
				RpcId:          request.RpcId,
				SystemId:       request.SystemId,
				ModuleId:       request.ModuleId,
				ExpirationTime: request.ExpirationTime,
				RpcContent:     data,
				Location:       cli.config.Location,
				TracingInfo:    request.TracingInfo,
			}
			trace := startSpanFromRPCMessage(req)
			if response := module.Execute(req); response != nil {
				cli.metrics.RPCReqProcessedSucceeded.WithLabelValues(request.SystemId, request.ModuleId).Inc()
				err = cli.sendResponse(response)
			} else {
				cli.metrics.RPCReqProcessedFailed.WithLabelValues(request.SystemId, request.ModuleId).Inc()
				err = fmt.Errorf("module %s returned an empty response for request %s, ignoring", request.ModuleId, request.RpcId)
			}
			if err != nil {
				trace.SetTag("failed", "true")
				trace.LogKV("event", err.Error())
			}
			trace.Finish()
		}()
	} else {
		log.Errorf("Cannot find implementation for module %s, ignoring request with ID %s", request.ModuleId, request.RpcId)
	}
}

func (cli *KafkaClient) sendResponse(response *ipc.RpcResponseProto) error {
	totalChunks := cli.getTotalChunks(response.RpcContent)
	var chunk int32
	topic := fmt.Sprintf("%s.rpc-response", cli.instanceID)
	for chunk = 0; chunk < totalChunks; chunk++ {
		bytes := cli.wrapMessageToRPC(response, chunk, totalChunks)
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(response.RpcId),
			Value:          bytes,
		}
		if err := cli.producer.Produce(msg, nil); err != nil {
			cli.metrics.RPCResSentFailed.WithLabelValues(response.SystemId, response.ModuleId).Inc()
			return fmt.Errorf("cannot send message to %s: %v", topic, err)
		}
	}
	cli.metrics.RPCResSentSucceeded.WithLabelValues(response.SystemId, response.ModuleId).Inc()
	return nil
}

func (cli *KafkaClient) wrapMessageToRPC(response *ipc.RpcResponseProto, chunk, totalChunks int32) []byte {
	bufferSize := cli.getRemainingBufferSize(int32(len(response.RpcContent)), chunk)
	offset := chunk * int32(cli.maxBufferSize)
	msg := response.RpcContent[offset : offset+bufferSize]
	rpcMsg := &rpc.RpcMessageProto{
		RpcId:              response.RpcId,
		RpcContent:         msg,
		CurrentChunkNumber: chunk,
		TotalChunks:        totalChunks,
	}
	bytes, err := proto.Marshal(rpcMsg)
	if err != nil {
		return []byte{}
	}
	return bytes
}
