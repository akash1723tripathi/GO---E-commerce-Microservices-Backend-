package order

import (
	"context"
	"fmt"
	"net"

	"github.com/akash1723tripathi/go-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service Service
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	pb.RegisterOrderServiceServer(server, &grpcServer{service: s})
	reflection.Register(server)
	return server.Serve(lis)
}
func toProto(o Order) *pb.Order {
	ps := make([]*pb.OrderedProduct, 0, len(o.Products))
	for _, p := range o.Products {
		ps = append(ps, &pb.OrderedProduct{Id: p.ID, Name: p.Name, Description: p.Description, Price: p.Price, Quantity: p.Quantity})
	}
	return &pb.Order{Id: o.ID, AccountId: o.AccountID, CreatedAt: timestamppb.New(o.CreatedAt), TotalPrice: o.TotalPrice, Products: ps}
}
func fromProto(ps []*pb.OrderedProduct) []OrderedProduct {
	out := make([]OrderedProduct, 0, len(ps))
	for _, p := range ps {
		out = append(out, OrderedProduct{ID: p.GetId(), Quantity: p.GetQuantity()})
	}
	return out
}
func (s *grpcServer) PostOrder(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	o, err := s.service.PostOrder(ctx, r.GetAccountId(), fromProto(r.GetProducts()))
	if err != nil {
		return nil, err
	}
	return &pb.PostOrderResponse{Order: toProto(*o)}, nil
}
func (s *grpcServer) GetOrder(ctx context.Context, r *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	o, err := s.service.GetOrder(ctx, r.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetOrderResponse{Order: toProto(*o)}, nil
}
func (s *grpcServer) GetOrders(ctx context.Context, r *pb.GetOrdersRequest) (*pb.GetOrdersResponse, error) {
	os, err := s.service.GetOrders(ctx, r.GetAccountId(), r.GetSkip(), r.GetTake())
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Order, 0, len(os))
	for _, o := range os {
		out = append(out, toProto(o))
	}
	return &pb.GetOrdersResponse{Orders: out}, nil
}
