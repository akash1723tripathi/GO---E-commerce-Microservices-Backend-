package order

import (
	"context"

	"github.com/akash1723tripathi/go-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, service: pb.NewOrderServiceClient(conn)}, nil
}
func (c *Client) Close() { _ = c.conn.Close() }
func fromOrder(p *pb.Order) *Order {
	o := &Order{ID: p.GetId(), AccountID: p.GetAccountId(), TotalPrice: p.GetTotalPrice()}
	if t := p.GetCreatedAt(); t != nil {
		o.CreatedAt = timestamppb.New(t.AsTime()).AsTime()
	}
	for _, x := range p.GetProducts() {
		o.Products = append(o.Products, OrderedProduct{ID: x.GetId(), Name: x.GetName(), Description: x.GetDescription(), Price: x.GetPrice(), Quantity: x.GetQuantity()})
	}
	return o
}
func (c *Client) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	ps := make([]*pb.OrderedProduct, 0, len(products))
	for _, p := range products {
		ps = append(ps, &pb.OrderedProduct{Id: p.ID, Quantity: p.Quantity})
	}
	r, err := c.service.PostOrder(ctx, &pb.PostOrderRequest{AccountId: accountID, Products: ps})
	if err != nil {
		return nil, err
	}
	return fromOrder(r.GetOrder()), nil
}
func (c *Client) GetOrder(ctx context.Context, id string) (*Order, error) {
	r, err := c.service.GetOrder(ctx, &pb.GetOrderRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return fromOrder(r.GetOrder()), nil
}
func (c *Client) GetOrders(ctx context.Context, accountID string, skip, take uint64) ([]Order, error) {
	r, err := c.service.GetOrders(ctx, &pb.GetOrdersRequest{AccountId: accountID, Skip: skip, Take: take})
	if err != nil {
		return nil, err
	}
	out := make([]Order, 0, len(r.GetOrders()))
	for _, o := range r.GetOrders() {
		out = append(out, *fromOrder(o))
	}
	return out, nil
}
