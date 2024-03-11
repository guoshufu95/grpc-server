package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sxp-server/helper"
)

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

// StreamInterceptor
//
//	@Description: 校验token
//	@param srv
//	@param ss
//	@param info
//	@param handler
//	@return error
func StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "获取不到metadata")
	}
	err, flag := helper.CheckToken(md) //token校验
	if err != nil || !flag {
		helper.HeadResponse(ctx, "0")
		return err
	}
	s := newWrappedStream(ss)

	err = handler(srv, s)
	if err != nil {
		fmt.Printf("RPC failed with error %v\n", err)
	}
	return err
}

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "获取不到metadata")
	}
	err, flag := helper.CheckToken(md) //token校验
	if err != nil || !flag {
		helper.HeadResponse(ctx, "0")
		return nil, err
	}
	m, err := handler(ctx, req)
	if err != nil {
		fmt.Printf("RPC failed with error %v\n", err)
	}
	return m, err
}
