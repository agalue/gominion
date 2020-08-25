package broker

import (
	"io"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
)

// Initializes the OpenTracing integration using Jaeger.
// Overrides the global tracer when Jaeger is available.
func initTracing(cfg *api.MinionConfig) (io.Closer, error) {
	jcfg := jaegercfg.Configuration{
		ServiceName: cfg.Location + "@" + cfg.ID,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1, // JAEGER_SAMPLER_PARAM
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := jcfg.NewTracer(
		jaegercfg.Logger(jaegerlog.NullLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}

// Starts a tracing span for an RPC API request
func startSpanFromRPCMessage(request *ipc.RpcRequestProto) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	tags := getTagsForRPC(request)
	ctx, err := tracer.Extract(opentracing.TextMap, request.TracingInfo)
	if err == nil {
		return tracer.StartSpan(request.ModuleId, opentracing.FollowsFrom(ctx), tags)
	}
	return tracer.StartSpan(request.ModuleId, tags)
}

// Gets the tracing span tags for an RPC API request
func getTagsForRPC(request *ipc.RpcRequestProto) opentracing.Tags {
	var tags = opentracing.Tags{"location": request.Location}
	if request.SystemId != "" {
		tags["systemId"] = request.SystemId
	}
	for key, value := range request.TracingInfo {
		tags[key] = value
	}
	return tags
}

// Starts a tracing span for a Sink API message
func startSpanForSinkMessage(msg *ipc.SinkMessage) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	tags := getTagsForSink(msg)
	ctx, err := tracer.Extract(opentracing.TextMap, msg.TracingInfo)
	if err == nil {
		return tracer.StartSpan(msg.ModuleId, opentracing.FollowsFrom(ctx), tags)
	}
	return tracer.StartSpan(msg.ModuleId, tags)
}

// Gets the tracing span tags for a Sink API message
func getTagsForSink(msg *ipc.SinkMessage) opentracing.Tags {
	var tags = opentracing.Tags{"location": msg.Location}
	if msg.SystemId != "" {
		tags["systemId"] = msg.SystemId
	}
	for key, value := range msg.TracingInfo {
		tags[key] = value
	}
	return tags
}
