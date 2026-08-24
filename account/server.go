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
	service Service
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp",fmt.Sprintf("%d", port))
	if err != nil{
		return nil, err
	}
	serv := grpc.NewServer()
	pb.RegisterAccountServiceServer(serv, &grpcServer{s})    // call register server function
	reflection.Register(serv)
	return serv.listen(lis)

}

func (s *grpcServer) PostAccount(ctx context.Context, r *pb.PostAccountRequest) (*pb.PostAccountResponse , error){   
	a, err := s.service.PostAccount(ctx, r.Name)
	if err != nil {
		return nil, err
	}

	return &pb.PostAccountResponse{Account: &pb.Account{
		Id: a.ID,
		Name : a.Name,
	}}, nil

}

func (s *grpcServer) GetAccount(ctx context.Context, r *pb.GetAccountRequest) (*pb.GetAccountResponse , error){ 
	a, err := s.service.GetAccount()(ctx, r.ID)
	if err != nil {
		return nil, err
	}

	return &pd.GetAccountResponse{
		Account: &pb.Account{
			Id: a.ID,
			Name : a.Name,
		},
	}, nil
}

func (s *grpcServer) GetAccounts(ctx context.Context, r *pb.GetAccountsequest) (*pb.GetAccountsResponse, error){ 
	res, err := s.service.GetAccounts()(ctx, r.ID)
	if err != nil {
		return nil, err
	}

	accounts := []*pb.Account{}
	for _,p := range res {
		accounts = append(accounts, 
			&pb.Account{
				Id: p.ID,
				Name : p.Name,
			},
		)
	}
	return &pd.GetAccountsResponse{Accounts : accounts, nil}

}