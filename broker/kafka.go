package broker

import (
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
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
	publisher     *kafka.Publisher
	subscriber    *kafka.Subscriber
	ctx           context.Context
	cancel        context.CancelFunc
	traceCloser   io.Closer
	metrics       Metrics
	maxBufferSize int
	instanceID    string
}

// Start initializes the Kafka client.
// Returns an error when the configuration is incorrect or cannot connect to the server.
func (cli *KafkaClient) Start(config *api.MinionConfig) error {
	cli.config = config
	var err error

	cli.metrics = NewMetrics()

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

	cli.subscriber, err = kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               []string{config.BrokerURL},
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: subsConfig,
			ConsumerGroup:         "gominion",
		},
		log.WatermillAdapter{},
	)
	if err != nil {
		return err
	}

	cli.publisher, err = kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:               []string{config.BrokerURL},
			Marshaler:             kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: kafka.DefaultSaramaSubscriberConfig(),
		},
		log.WatermillAdapter{},
	)
	if err != nil {
		return err
	}

	if config.StatsPort > 0 {
		cli.metrics.Register()
	}

	for _, module := range api.GetAllSinkModules() {
		if err = module.Start(cli.config, cli); err != nil {
			return fmt.Errorf("Cannot start Sink API module %s: %v", module.GetID(), err)
		}
	}

	cli.ctx, cli.cancel = context.WithCancel(context.Background())
	topic := fmt.Sprintf("%s.%s.rpc-request", cli.instanceID, cli.config.Location)
	rpcChannel, err := cli.subscriber.Subscribe(cli.ctx, topic)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-cli.ctx.Done():
				return
			case msg := <-rpcChannel:
				rpc := new(rpc.RpcMessageProto)
				if err := proto.Unmarshal(msg.Payload, rpc); err != nil {
					cli.processRequest(rpc)
				}
				msg.Ack()
			}
		}
	}()

	return nil
}

// Stop finalizes the Kafka client and all its dependencies.
func (cli *KafkaClient) Stop() {
	for _, module := range api.GetAllSinkModules() {
		module.Stop()
	}
	log.Warnf("Stopping Kafka client")
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
		bytes := cli.wrapMessageToSink(chunk, totalChunks, msg.Content)
		data := message.NewMessage(watermill.NewUUID(), bytes)
		err = cli.publisher.Publish(topic, data)
		if err != nil {
			break
		}
	}
	if err != nil {
		cli.metrics.SinkMsgDeliveryFailed.WithLabelValues(msg.ModuleId).Inc()
		trace.SetTag("failed", "true")
		trace.LogKV("event", err.Error())
		return fmt.Errorf("cannot send message to %s: %v", topic, err)
	}
	cli.metrics.SinkMsgDeliverySucceeded.WithLabelValues(msg.ModuleId).Inc()
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

func (cli *KafkaClient) wrapMessageToSink(chunk, totalChunks int32, data []byte) []byte {
	bufferSize := cli.getRemainingBufferSize(int32(len(data)), chunk)
	offset := chunk * int32(cli.maxBufferSize)
	msg := data[offset : offset+bufferSize]
	sinkMsg := &sink.SinkMessage{
		MessageId:          watermill.NewUUID(),
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
	log.Debugf("Received RPC request with ID %s for module %s", request.RpcId, request.ModuleId)
	if module, ok := api.GetRPCModule(request.ModuleId); ok {
		go func() {
			var err error
			req := &ipc.RpcRequestProto{
				RpcId:          request.RpcId,
				SystemId:       request.SystemId,
				ModuleId:       request.ModuleId,
				ExpirationTime: request.ExpirationTime,
				RpcContent:     request.RpcContent,
				Location:       cli.config.Location,
				TracingInfo:    request.TracingInfo,
			}
			trace := startSpanFromRPCMessage(req)
			if response := module.Execute(req); response != nil {
				cli.metrics.RPCReqProcessedSucceeded.WithLabelValues(request.ModuleId).Inc()
				err = cli.sendResponse(response)
			} else {
				cli.metrics.RPCReqProcessedFailed.WithLabelValues(request.ModuleId).Inc()
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
		bytes := cli.wrapMessageToSink(chunk, totalChunks, response.RpcContent)
		data := message.NewMessage(watermill.NewUUID(), bytes)
		err := cli.publisher.Publish(topic, data)
		if err != nil {
			return fmt.Errorf("cannot send message to %s: %v", topic, err)
		}
	}
	return nil
}

func (cli *KafkaClient) wrapMessageToRPC(response *ipc.RpcResponseProto, chunk, totalChunks int32, data []byte) []byte {
	bufferSize := cli.getRemainingBufferSize(int32(len(data)), chunk)
	offset := chunk * int32(cli.maxBufferSize)
	msg := data[offset : offset+bufferSize]
	rpcMsg := &rpc.RpcMessageProto{
		RpcId:              response.RpcId,
		RpcContent:         msg,
		SystemId:           response.SystemId,
		ExpirationTime:     uint64(time.Now().Unix()),
		CurrentChunkNumber: chunk,
		TotalChunks:        totalChunks,
	}
	bytes, err := proto.Marshal(rpcMsg)
	if err != nil {
		return []byte{}
	}
	return bytes
}
