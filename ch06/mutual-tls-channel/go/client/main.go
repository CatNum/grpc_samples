// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc --go_out=plugins=grpc:golang/product_info product_info.proto
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/golang/product_info
// Execute go run go/client/main.go

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/grpc-up-and-running/samples/ch02/productinfo/go/proto"
	"google.golang.org/grpc"
)

var (
	address = "localhost:50051"
	hostname = "localhost"
	crtFile = filepath.Join("ch06", "mutual-tls-channel", "certs", "client.crt")
	keyFile = filepath.Join("ch06", "mutual-tls-channel", "certs", "client.key")
	caFile = filepath.Join("ch06", "mutual-tls-channel", "certs", "ca.crt")
)

func main() {
	// Load the client certificates from disk
	// 通过服务器端的证书和密钥直接创建 X.509 密钥对。
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("could not load client key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	// 通过 CA 创建证书池。
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	// 将来自 CA 的客户端证书附加到证书池中。
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certs")
	}

	opts := []grpc.DialOption{
		// transport credentials.
		// 添加传输凭证作为连接选项。这里，ServerName 必须与证书中的 Common Name 一致。
		grpc.WithTransportCredentials( credentials.NewTLS(&tls.Config{
			ServerName:   hostname, // NOTE: this is required!
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})),
	}

	// Set up a connection to the server.
	// 传入连接选项，搭建到服务器的安全连接。
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	// 传入连接并创建存根。该存根实例包含调用服务器的所有远程方法。
	c := pb.NewProductInfoClient(conn)

	// Contact the server and print out its response.
	// 连接服务器并打印响应
	name := "Samsung S10"
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
	log.Printf("Product: %s", product.String())
}
