package broker

import (
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/golang/protobuf/proto"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/agalue/gominion/protobuf/rpc"
	"github.com/agalue/gominion/protobuf/sink"
)

type KafkaClient struct {
	config        *api.MinionConfig
	registry      *api.SinkRegistry
	publisher     *kafka.Publisher
	subscriber    *kafka.Subscriber
	ctx           context.Context
	cancel        context.CancelFunc
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
		return fmt.Errorf("Minion configuration required")
	}
	if cli.registry == nil {
		return fmt.Errorf("Sink registry required")
	}
	if cli.metrics == nil {
		return fmt.Errorf("Prometheus Metrics required")
	}

	cli.msgBuffer = make(map[string][]byte)
	cli.chunkTracker = make(map[string]int32)

	// Maximum size of the buffer to split messages in chunks
	cli.maxBufferSize, err = strconv.Atoi(cli.config.BrokerProperties["max-buffer-size"])
	if err != nil {
		cli.maxBufferSize = 1024
	}

	// The OpenNMS Instance ID (org.opennms.instance.id), for Kafka topics
	cli.instanceID = cli.config.BrokerProperties["instance-id"]
	if cli.instanceID == "" {
		cli.instanceID = "OpenNMS"
	}

	if cli.traceCloser, err = initTracing(cli.config); err != nil {
		return err
	}

	subsConfig := kafka.DefaultSaramaSubscriberConfig()
	subsConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	subsConfig.Consumer.Offsets.AutoCommit.Enable = true
	subsConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	marshaler := kafka.NewWithPartitioningMarshaler(func(topic string, msg *message.Message) (string, error) {
		return msg.UUID, nil
	})

	cli.subscriber, err = kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               []string{cli.config.BrokerURL},
			Unmarshaler:           marshaler,
			OverwriteSaramaConfig: subsConfig,
			ConsumerGroup:         cli.config.Location,
			NackResendSleep:       kafka.NoSleep,
		},
		log.WatermillAdapter{},
	)
	if err != nil {
		return err
	}

	cli.publisher, err = kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   []string{cli.config.BrokerURL},
			Marshaler: marshaler,
		},
		log.WatermillAdapter{},
	)
	if err != nil {
		return err
	}

	if err := cli.registry.StartModules(cli.config, cli); err != nil {
		return err
	}

	cli.ctx, cli.cancel = context.WithCancel(context.Background())
	topic := fmt.Sprintf("%s.%s.rpc-request", cli.instanceID, cli.config.Location)
	rpcChannel, err := cli.subscriber.Subscribe(cli.ctx, topic)
	if err != nil {
		return err
	}
	go func(messages <-chan *message.Message) {
		for msg := range messages {
			msg.Ack()
			rpc := new(rpc.RpcMessageProto)
			if err := proto.Unmarshal(msg.Payload, rpc); err == nil {
				cli.metrics.RPCReqReceivedSucceeded.WithLabelValues(rpc.SystemId, rpc.ModuleId).Inc()
				cli.processRequest(rpc)
			} else {
				cli.metrics.RPCReqReceivedFailed.WithLabelValues(rpc.SystemId, rpc.ModuleId).Inc()
				log.Errorf("Cannot process RPC Request: %v", err)
			}
		}
	}(rpcChannel)

	return nil
}

// Stop finalizes the Kafka client and all its dependencies.
func (cli *KafkaClient) Stop() {
	cli.registry.StopModules()
	log.Warnf("Stopping Kafka client")
	cli.cancel()
	cli.subscriber.Close()
	cli.publisher.Close()
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
		data := message.NewMessage(msg.MessageId, bytes)
		err = cli.publisher.Publish(topic, data)
		if err != nil {
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
				err = fmt.Errorf("Module %s returned an empty response for request %s, ignoring", request.ModuleId, request.RpcId)
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
		data := message.NewMessage(response.RpcId, bytes)
		err := cli.publisher.Publish(topic, data)
		if err != nil {
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
