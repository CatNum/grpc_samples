package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "productinfo/service/ecommerce"   // 导入protobuf生成的代码所在的包
)

const (
	port = ":50051"
)

func main(){
	// 希望由gRPC服务器所绑定的TCP监听器在给定的端口上创建
	lis,err := net.Listen("tcp",port)
	if err != nil{
		log.Fatalf("failed to listen: %v",err)
	}
	// 通过调用 gRPC Go API 创建新的gRPC服务器实例
	s := grpc.NewServer()
	// 通过调用生成的API，将之前生成的服务注册到新创建的gRPC服务器上
	pb.RegisterProductInfoServer(s,&server{})

	log.Printf("Starting gRPC listener on port" + port)
	// 在指定的端口上开始监听传入的消息
	if err := s.Serve(lis);err != nil{
		log.Fatalf("failed to serve:%v",err)
	}
}
