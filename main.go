package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"io"
	"net"
	"strconv"
	"sxp-server/helper"
	"sxp-server/model"
	"sxp-server/pb"
	"sxp-server/service"
	"sxp-server/tracer"
	"time"
)

type modelSever struct {
	pb.UnimplementedModelServer
}

func (m *modelSever) GetModel(ctx context.Context, request *pb.ModelRequest) (res *pb.ModelResponse, err error) {
	// 设置trailer
	defer func() {
		if err = helper.TrailerResponse(ctx); err != nil {
			return
		}
	}()
	time.Sleep(100 * time.Millisecond)
	err = helper.HeadResponse(ctx, "1")
	if err != nil {
		return
	}
	pm := model.ProductMap
	name := pm[request.GetProductId()].Name
	res = &pb.ModelResponse{}
	res.Product = name
	fmt.Println(res)
	return
}

// UpdateModel
//
//	@Description: 更新product model
//	@receiver m
//	@param ctx
//	@param request
//	@return res
//	@return err
func (m *modelSever) UpdateModel(ctx context.Context, request *pb.UpdateRequest) (res *pb.UpdateResponse, err error) {
	// 设置trailer
	defer func() {
		if err = helper.TrailerResponse(ctx); err != nil {
			return
		}
	}()
	err = helper.HeadResponse(ctx, "1")
	if err != nil {
		return
	}

	pro := model.ProductMap[request.GetProductId()]
	pro.Name = request.GetProduct()
	model.ProductMap[request.GetProductId()] = pro
	res = &pb.UpdateResponse{
		Message: "更新数据成功",
	}
	fmt.Println(model.ProductMap)
	return
}

func (m *modelSever) GetByStatus(stream pb.Model_GetByStatusServer) (err error) {
	// 在defer中创建trailer记录函数的返回时间.
	defer func() {
		if err = helper.TrailerResponse(stream.Context()); err != nil {
			return
		}
	}()
	err = helper.HeadResponse(stream.Context(), "1")
	if err != nil {
		return
	}
	data := make([]model.Product, 0)
	for {
		request, er := stream.Recv()
		if er != nil && er != io.EOF {
			fmt.Println("Recv error:", er.Error())
			continue
		} else if er == io.EOF {
			fmt.Println("Recv EOF")
			break
		}
		pm := model.ProductMap
		for _, v := range pm {
			if v.Status == request.GetStatus() {
				data = append(data, v)
				stream.Send(&pb.StatusResponse{
					ProductId: strconv.Itoa(v.Id),
					Product:   v.Name,
					Status:    v.Status,
				})
			}
		}
		break
	}
	return
}

func main() {
	model.Init()
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	// 初始化tracer
	trace, _, err := tracer.NewJaegerTracer("sxp-server", "192.168.111.143:6831")
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(service.UnaryInterceptor,
			tracer.UnaryTraceInterceptor(trace))),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_middleware.ChainStreamServer(
			service.StreamInterceptor,
			tracer.StreamTraceInterceptor(trace))))) // 创建gRPC服务器
	pb.RegisterModelServer(s, &modelSever{}) // 在gRPC服务端注册服务
	// 启动服务
	err = s.Serve(lis)

}
