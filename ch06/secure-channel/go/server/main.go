// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc --go_out=plugins=grpc:golang/product_info product_info.proto
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/go/product_info
// Execute go run go/server/main.go

package main

import (
	"context"
	"crypto/tls"
	"errors"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	pb "github.com/grpc-up-and-running/samples/ch02/productinfo/go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"path/filepath"
)

const (
	port = ":50051"
	crtFile = filepath.Join("ch06", "secure-channel", "certs", "server.crt")
	keyFile = filepath.Join("ch06", "secure-channel", "certs", "server.key")
)

// server is used to implement ecommerce/product_info.
type server struct {
	productMap map[string]*pb.Product
}

// AddProduct implements ecommerce.AddProduct
func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*wrapper.StringValue, error) {
	out, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in
	return &wrapper.StringValue{Value: in.Id}, nil
}

// GetProduct implements ecommerce.GetProduct
func (s *server) GetProduct(ctx context.Context, in *wrapper.StringValue) (*pb.Product, error) {
	value, exists := s.productMap[in.Value]
	if exists {
		return value, nil
	}
	return nil, errors.New("Product does not exist for the ID" + in.Value)
}

func main() {
	//  读取和解析公钥–私钥对，并创建启用 TLS 的证书。
	cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}
	opts := []grpc.ServerOption{
		// Enable TLS for all incoming connections.
		// 添加证书作为 TLS 服务器凭证，从而为所有传入的连接启用TLS。
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}
	// 通过传入 TLS 服务器凭证来创建新的 gRPC 服务器实例。
	s := grpc.NewServer(opts...)
	// 通过调用生成的 API，将服务实现注册到新创建的 gRPC 服务器上。
	pb.RegisterProductInfoServer(s, &server{})
	// Register reflection service on gRPC server.
	//reflection.Register(s)
	// 在端口 50051 上创建 TCP 监听器。
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 绑定 gRPC 服务器到监听器，并开始监听端口 50051 上传入的消息。
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
