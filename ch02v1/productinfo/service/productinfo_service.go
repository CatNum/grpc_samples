package main

import (
	"context"
	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "productinfo/service/ecommerce"  // 导入通过 protobuf 编译器所生成的代码所在的包
)

// 用来实现ecommerce/product_info的服务器
// server 结构体是对服务器的抽象，可以通过它将服务方法附加到服务器上
type server struct{
	productMap map[string]*pb.Product
}

// 实现 ecommerce.AddProduct的AddProduct方法
func  (s *server)AddProduct(ctx context.Context,
	in *pb.Product)(*pb.ProductID,error){
	out,err := uuid.NewV4()
	if err != nil {
		return nil,status.Errorf(codes.Internal,
			"Error while generating Product ID",err)
	}
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in
	return &pb.ProductID{Value:in.Id},status.New(codes.OK,"").Err()

}

// 实现ecommerce.GetProduct的GetProduct方法
func (s *server) GetProduct(ctx context.Context,
	in *pb.ProductID)(*pb.Product,error){
	value,exists := s.productMap[in.Value]
	if exists {
		return value,status.New(codes.OK,"").Err()
	}
	return nil,status.Errorf(codes.NotFound,"Product does not exists.",in.Value)
}

