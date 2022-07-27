package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	pb "productinfo/client/ecommerce"  // 导入protobuf生成的代码所在的包
	"time"
)

const (
	address = "localhost:50051"
)

func main() {
	// 根据提供的地址（localhost：50051）创建到服务器端的连接
	// 这里创建了一个客户端和服务器端之间的连接，但它目前不安全
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect:%v", err)
	}
	// 所有事情都完成后，关闭连接
	defer conn.Close()
	// 传递连接并创建存根文件，这个实例包含可调用服务器的所有远程方法
	c := pb.NewProductInfoClient(conn)

	name := "Apple iPhone 11"
	description := `Meeeetdsakhdosahosgih sdhiusfhifid sdfoihio`
	price := float32(1000.0)
	// 创建 Context 以传递给远程调用。这里的Context对象包含一些元数据，
	// 如终端用户的标识，授权令牌以及请求的截止时间，该对象后在请求的生命周期内一直存在
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 使用商品的详情信息调用AddProduct方法，如果操作成功完成，就会返回一个商品ID，否则返回一个错误
	r, err := c.AddProduct(ctx,
		&pb.Product{Name: name, Description: description, Price: price})
	if err != nil {
		log.Fatalf("Could not add product:%v", err)
	}
	log.Printf("Product ID:%s added successfully", r.Value)
	// 使用商品ID来调用GetProduct方法。如果操作成功完成，将返回商品详情，否则返回一个错误
	product, err := c.GetProduct(ctx, &pb.ProductID{Value: r.Value})
	if err != nil {
		log.Fatalf("Could not get product:%v", err)
	}
	log.Printf("Product:", product.String())
}
