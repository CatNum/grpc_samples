// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc --go_out=plugins=grpc:golang/product_info product_info.proto
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/go/product_info
// Execute go run go/server/main.go

package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	pb "github.com/grpc-up-and-running/samples/ch02/productinfo/go/product_info"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"path/filepath"
	"strings"
)

// server is used to implement ecommerce/product_info.
type server struct {
	productMap map[string]*pb.Product
}

var (
	port = ":50051"
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid credentials")
)

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
	// 读取和解析公钥 - 私钥对，并创建启用 TLS 的证书
	cert, err := tls.LoadX509KeyPair(filepath.Join("ch06", "secure-channel", "certs", "server.crt"),
		filepath.Join("ch06", "secure-channel", "certs", "server.key"))
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	opts := []grpc.ServerOption{
		// Enable TLS for all incoming connections.
		// 添加证书作为 TLS 服务器凭证，从而为所有传入的连接启用 TLS
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
		// 通过 TLS 服务器证书添加新的服务器选项（grpc.ServerOption）。
		// grpc.UnaryInterceptor 是一个函数，
		// 我们在其中添加拦截器来拦截所有来自客户端的请求。
		// 我们向该函数传递个函数引用（ensureValidBasicCredentials），
		// 拦截器会将所有的客户端请求传递给该函数。
		grpc.UnaryInterceptor(ensureValidBasicCredentials),
	}
	// 通过传入 TLS 服务器凭证来创建新的 gRPC 服务器实例
	s := grpc.NewServer(opts...)
	// 通过调用生成的API，将服务实现注册到新创建的gRPC服务器上
	pb.RegisterProductInfoServer(s, &server{})
	// Register reflection service on gRPC server.
	//reflection.Register(s)
	// 在端口上创建 TCP 监听器
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 绑定 gRPC 服务器到监听器，并开始监听端口 50051 上传入的消息
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// valid validates the authorization.
// valid 验证授权
func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Basic ")
	return token == base64.StdEncoding.EncodeToString([]byte("admin:admin"))
}

// ensureValidToken ensures a valid token exists within a request's metadata. If
// the token is missing or invalid, the interceptor blocks execution of the
// handler and returns an error. Otherwise, the interceptor invokes the unary
// handler.
// ensureValidToken 确保请求的元数据中存在有效令牌。如果
// 令牌丢失或者无效，拦截器阻止执行处理程序并返回错误
// 否则，拦截器调用一元处理程序

// 定义名为 ensureValidBasicCredentials 的函数来校验调用者的身份。
// 在这里，context.Context 对象包含所需的元数据，
// 在请求的生命周期内，该元数据会一直存在。
func ensureValidBasicCredentials(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	// 从上下文中抽取元数据，获取 authentication 的值并校验凭证。
	// 由于 metadata.MD 中的键会被标准化为小写字母，因此需要检查键的值。
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}
	// The keys within metadata.MD are normalized to lowercase.
	// See: https://godoc.org/google.golang.org/grpc/metadata#New
	if !valid(md["authorization"]) {
		return nil, errInvalidToken
	}
	// Continue execution of handler after ensuring a valid token.
	return handler(ctx, req)
}
