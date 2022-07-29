// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc --go_out=plugins=grpc:golang/product_info product_info.proto
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/go/product_info
// Execute go run go/server/main.go

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	pb "github.com/grpc-up-and-running/samples/ch02/productinfo/go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
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

var (
	port = ":50051"
    crtFile = filepath.Join("ch06", "mutual-tls-channel", "certs", "server.crt")
    keyFile = filepath.Join("ch06", "mutual-tls-channel", "certs", "server.key")
    caFile = filepath.Join("ch06", "mutual-tls-channel", "certs", "ca.crt")
)

func main() {
	// 通过服务器端的证书和密钥直接创建 X.509 密钥对。
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	// 通过 CA 创建证书池。
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	// 将来自 CA 的客户端证书附加到证书池中。
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append client certs")
	}

	opts := []grpc.ServerOption{
		// Enable TLS for all incoming connections.
		// 通过创建 TLS 凭证为所有传入的连接启用 TLS。
		grpc.Creds(    // Create the TLS credentials
			credentials.NewTLS(&tls.Config {
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{certificate},
				ClientCAs:    certPool,
				},
			)),
	}

	// 通过传入的 TLS 服务器凭证创建新的 gRPC 服务器实例。
	s := grpc.NewServer(opts...)
	// 通过调用生成的 API 将 gRPC 服务注册到新创建的 gRPC 服务器上。
	pb.RegisterProductInfoServer(s, &server{})
	// Register reflection service on gRPC server.
	//reflection.Register(s)

	// 在端口 50051 上创建 TCP 监听器。
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//绑定 gRPC 服务器到监听器，并开始在端口 50051 上监听传入的消息。
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
