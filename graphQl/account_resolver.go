package graphql

import "context"

type accountResolver struct {
	server *Server
}

func (r *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	if obj.Orders != nil {
		return obj.Orders, nil
	}
	orders, err := r.server.orderClient.GetOrders(ctx, obj.ID, 0, 100)
	if err != nil {
		return nil, err
	}
	result := make([]*Order, 0, len(orders))
	for _, o := range orders {
		result = append(result, orderModel(o))
	}
	return result, nil
}
