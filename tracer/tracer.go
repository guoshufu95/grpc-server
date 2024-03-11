package tracer

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"time"
)

// NewJaegerTracer
//
//	@Description: 构建tracer
//	@param serviceName
//	@param jaegerHostPort
//	@return opentracing.Tracer
//	@return io.Closer
//	@return error
func NewJaegerTracer(serviceName string, jaegerHostPort string) (opentracing.Tracer, io.Closer, error) {
	cfg := &jaegerConfig.Configuration{
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const", //固定采样
			Param: 1,       //1=全采样、0=不采样
		},

		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jaegerHostPort,
		},

		ServiceName: serviceName,
	}
	tracer, closer, err := cfg.NewTracer(jaegerConfig.Logger(jaeger.StdLogger))
	if err != nil {
		return nil, nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer, err
}

type MDReaderWriter struct {
	metadata.MD
}

func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// UnaryTraceInterceptor
//
//	@Description: 服务端链路追踪middleware
//	@param tracer
//	@return grpc.UnaryServerInterceptor
func UnaryTraceInterceptor(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
		span := opentracing.GlobalTracer().StartSpan(
			info.FullMethod,
			ext.RPCServerOption(spanContext),
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCServer,
		)
		defer span.Finish()
		ctx1 := opentracing.ContextWithSpan(ctx, span)
		time.Sleep(150 * time.Millisecond)
		return handler(ctx1, req)
	}
}

// wrappedStream
// @Description: 重写stream的方法
type wrappedStream struct {
	grpc.ServerStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	return w.ServerStream.SendMsg(m)
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}

// StreamTraceInterceptor
//
//	@Description: 流式拦截器
//	@param tracer
//	@return grpc.StreamServerInterceptor
func StreamTraceInterceptor(tracer opentracing.Tracer) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
		if err != nil {
			fmt.Println(err.Error())
		}
		span := tracer.StartSpan(
			info.FullMethod,
			ext.RPCServerOption(spanContext),
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCServer,
		)
		defer span.Finish()
		ss = &openTracingServerStream{
			ServerStream: ss,
			ctx:          opentracing.ContextWithSpan(ss.Context(), span),
		}
		return handler(srv, ss)
	}
}

type openTracingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *openTracingServerStream) Context() context.Context {
	return ss.ctx
}
