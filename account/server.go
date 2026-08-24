package account

import (
	"context"
	"fmt"
	"net"

	"github.com/akash1723tripathi/go-microservices/account/pb/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service Service
}

func ListenGRPC(service Service, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	pb.RegisterAccountServiceServer(server, &grpcServer{service: service})
	reflection.Register(server)
	return server.Serve(listener)

}

func (s *grpcServer) PostAccount(ctx context.Context, request *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	account, err := s.service.PostAccount(ctx, request.GetName())
	if err != nil {
		return nil, err
	}

	return &pb.PostAccountResponse{Account: &pb.Account{
		Id:   account.ID,
		Name: account.Name,
	}}, nil

}

func (s *grpcServer) GetAccount(ctx context.Context, request *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	account, err := s.service.GetAccount(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.GetAccountResponse{
		Account: &pb.Account{
			Id:   account.ID,
			Name: account.Name,
		},
	}, nil
}

func (s *grpcServer) GetAccounts(ctx context.Context, request *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	result, err := s.service.GetAccounts(ctx, request.GetSkip(), request.GetTake())
	if err != nil {
		return nil, err
	}

	accounts := make([]*pb.Account, 0, len(result))
	for _, account := range result {
		accounts = append(accounts, &pb.Account{
			Id:   account.ID,
			Name: account.Name,
		})
	}
	return &pb.GetAccountsResponse{Accounts: accounts}, nil

}
