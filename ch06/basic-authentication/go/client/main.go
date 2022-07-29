// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc --go_out=plugins=grpc:golang/product_info product_info.proto
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/golang/product_info
// Execute go run go/client/main.go

package main

import (
	"context"
	"encoding/base64"
	"google.golang.org/grpc/credentials"
	"log"
	"path/filepath"
	"time"

	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/grpc-up-and-running/samples/ch02/productinfo/go/product_info"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	// 读取并解析公开证书，创建启用 TLS 的证书
	creds, err := credentials.NewClientTLSFromFile(filepath.Join("ch06", "secure-channel", "certs", "server.crt"),
		"localhost")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}
	// 使用有效的用户凭证（用户名和密码）初始化 auth 变量。
	// auth 变量存放了我们要使用的值。
	auth := basicAuth{
		username: "admin",
		password: "admin",
	}
	// 以 DialOption 的形式添加传输凭证
	opts := []grpc.DialOption{
		// 传递 auth 变量给 grpc.WithPerRPCCredentials 函数。
		// 该函数接受一个接口作为参数。
		// 因为我们定义的认证结构符合该接口，所以可以传递变量。
		grpc.WithPerRPCCredentials(auth),
		// transport credentials.
		grpc.WithTransportCredentials(creds),
	}

	// Set up a connection to the server.
	// 通过传入 dial 选项，建立到服务器的安全连接
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// 延迟关闭
	defer conn.Close()
	// 传入连接并创建存根。该存根实例包含了调用服务器的所有远程方法
	c := pb.NewProductInfoClient(conn)

	// Contact the server and print out its response.
	name := "Sumsung S10"
	description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2019"
	price := float32(700.0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.AddProduct(ctx, &pb.Product{Name: name, Description: description, Price: price})
	if err != nil {
		log.Fatalf("Could not add product: %v", err)
	}
	log.Printf("Product ID: %s added successfully", r.Value)

	product, err := c.GetProduct(ctx, &wrapper.StringValue{Value: r.Value})
	if err != nil {
		log.Fatalf("Could not get product: %v", err)
	}
	log.Printf("Product: ", product.String())
}

//  定义结构体来存放要注入 RPC 的字段集合
// （在我们的场景中，也就是用户的凭证，如用户名和密码）。
type basicAuth struct {
	username string
	password string
}

// 实现 GetRequestMetadata 方法，并将用户凭证转换成请求元数据。
// 在我们的场景中，键是 Authorization，
// 值则由 Basic 和加上 <用户名 >:< 密码 > 的 base64 算法计算结果所组成。
func (b basicAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	auth := b.username + ":" + b.password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{
		"authorization": "Basic " + enc,
	}, nil
}

//声明在传递凭证时是否需要启用通道安全性。如前所述，建议启用。
func (b basicAuth) RequireTransportSecurity() bool {
	return true
}
