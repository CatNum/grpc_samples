package main

import (
	"context"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"log"
	pb "order-service/client/ecommerce"
	"time"
)

const (
	address = "localhost:50051"
)

func main() {
	// Setting up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewOrderManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	//// Add Order
	//order1 := pb.Order{Id: "101", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}
	//res, _ := client.AddOrder(ctx, &order1)
	//if res != nil {
	//	log.Print("AddOrder Response -> ", res.Value)
	//}

	// Get Order
	retrievedOrder , err := client.GetOrder(ctx, &wrapper.StringValue{Value: "106"})
	log.Print("GetOrder Response -> : ", retrievedOrder)
}
