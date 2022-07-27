package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"io"
	"log"
	"net"
	/*"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"*/
	pb "ordermgt/service/ecommerce"
	"strings"
)

const (
	port           = ":50051"
	orderBatchSize = 3
)

var orderMap = make(map[string]pb.Order)

type server struct {
	orderMap map[string]*pb.Order
}

// Simple RPC
func (s *server) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrapper.StringValue, error) {
	log.Printf("Order Added. ID : %v", orderReq.Id)
	orderMap[orderReq.Id] = *orderReq
	return &wrapper.StringValue{Value: "Order Added: " + orderReq.Id}, nil
}

// Simple RPC
func (s *server) GetOrder(ctx context.Context, orderId *wrapper.StringValue) (*pb.Order, error) {
	ord, exists := orderMap[orderId.Value]
	if exists {
		return &ord, status.New(codes.OK, "").Err()
	}

	return nil, status.Errorf(codes.NotFound, "Order does not exist. : ", orderId)


}

// Server-side Streaming RPC
func (s *server) SearchOrders(searchQuery *wrappers.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {

	for key, order := range orderMap {
		log.Print(key, order)
		for _, itemStr := range order.Items {
			log.Print(itemStr)
			if strings.Contains(itemStr, searchQuery.Value) {
				// Send the matching orders in a stream
				// 在流中发送匹配的订单
				err := stream.Send(&order)
				if err != nil {
					return fmt.Errorf("error sending message to stream : %v", err)
				}
				log.Print("Matching Order Found : " + key)
				break
			}
		}
	}
	return nil
}

// Client-side Streaming RPC
func (s *server) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {

	ordersStr := "Updated Order IDs : "
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			// Finished reading the order stream.
			// 完成读取订单流
			return stream.SendAndClose(&wrapper.StringValue{Value: "Orders processed " + ordersStr})
		}

		if err != nil {
			return err
		}
		// Update order
		orderMap[order.Id] = *order

		log.Printf("Order ID : %s - %s", order.Id, "Updated")
		ordersStr += order.Id + ", "
	}
}

// Bi-directional Streaming RPC
// 订单处理功能，
// 用户可以发送连续的订单集合（订单流），
// 并根据投递地址将它们进行组合发货
// 每次服务器端以三个为一组发送消息
func (s *server) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {
	// 表示三个订单一组
	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)
	for {
		// 从流中读取订单ID
		orderId, err := stream.Recv()
		log.Printf("Reading Proc order : %s", orderId)
		// 持续读取，直到流结束
		if err == io.EOF {
			// Client has sent all the messages
			// Send remaining shipments
			// 客户端已经发送了所有消息，发送剩余货物
			log.Printf("EOF : %s", orderId)
			for _, shipment := range combinedShipmentMap {
				// 当流结束，将所有剩余的发货组合发送给客户端
				if err := stream.Send(&shipment); err != nil {
					return err
				}
			}
			return nil
		}
		if err != nil {
			log.Println(err)
			return err
		}
		// 根据目的地
		// 将订单放到一组
		destination := orderMap[orderId.GetValue()].Destination
		shipment, found := combinedShipmentMap[destination]
		// 如果存在该目的地组
		if found {
			ord := orderMap[orderId.GetValue()]
			// 将同一目的地的订单放到一起
			shipment.OrdersList = append(shipment.OrdersList, &ord)
			// 将所有目的地的订单组放到一个 map 中
			combinedShipmentMap[destination] = shipment
		} else {
			comShip := pb.CombinedShipment{Id: "cmb - " + (orderMap[orderId.GetValue()].Destination), Status: "Processed!", }
			ord := orderMap[orderId.GetValue()]
			comShip.OrdersList = append(shipment.OrdersList, &ord)
			combinedShipmentMap[destination] = comShip
			log.Print(len(comShip.OrdersList), comShip.GetId())
		}

		if batchMarker == orderBatchSize {
			// 将组合后的订单以流的形式分批发送给客户端
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping : %v -> %v" , comb.Id, len(comb.OrdersList))
				// 将发货组合发送到客户端
				if err := stream.Send(&comb); err != nil {
					return err
				}
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}


func main() {
	initSampleData()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrderManagementServer(s, &server{})
	// Register reflection service on gRPC server.
	// reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initSampleData() {
	orderMap["102"] = pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	orderMap["103"] = pb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	orderMap["104"] = pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	orderMap["105"] = pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	orderMap["106"] = pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}
