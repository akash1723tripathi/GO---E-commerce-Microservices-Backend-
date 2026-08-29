package graphql

import (
	"context"
	"errors"
	"github.com/akash1723tripathi/go-microservices/order"
	"time"
)

// import (
// 	"context"
// 	"log"
// 	"time"
// )

type mutationResolver struct {
	server *Server
}

func (r *mutationResolver) CreateAccount(ctx context.Context, in AccountInput) (*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	a, err := r.server.accountClient.PostAccount(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	return &Account{ID: a.ID, Name: a.Name}, nil
}
func (r *mutationResolver) CreateProduct(ctx context.Context, in ProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	p, err := r.server.catalogClient.PostProduct(ctx, in.Name, in.Description, in.Price)
	if err != nil {
		return nil, err
	}
	return &Product{ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price}, nil
}
func (r *mutationResolver) CreateOrder(ctx context.Context, in OrderInput) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if len(in.Products) == 0 {
		return nil, errors.New("order must contain products")
	}
	products := make([]order.OrderedProduct, 0, len(in.Products))
	for _, p := range in.Products {
		if p == nil || p.Quantity <= 0 {
			return nil, errors.New("product quantity must be positive")
		}
		products = append(products, order.OrderedProduct{ID: p.ID, Quantity: uint32(p.Quantity)})
	}
	o, err := r.server.orderClient.PostOrder(ctx, in.AccountID, products)
	if err != nil {
		return nil, err
	}
	return orderModel(*o), nil
}

func orderModel(o order.Order) *Order {
	products := make([]*OrderedProduct, 0, len(o.Products))
	for _, p := range o.Products {
		products = append(products, &OrderedProduct{ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price, Quantity: int(p.Quantity)})
	}
	return &Order{ID: o.ID, CreatedAt: o.CreatedAt, TotalPrice: o.TotalPrice, Products: products}
}

// func (r *mutationResolver) CreateAccount(ctx context.Context, in AccountInput) (*Account, error) {
// 	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
// 	defer cancel()

// 	a, err := r.server.accountClient.PostAccount(ctx, in.Name)
// 	if err != nil {
// 		log.Println(err)
// 		return nil, err
// 	}

// 	return &Account{
// 		ID:   a.ID,
// 		Name: a.Name,
// 	}, nil
// }

// func (r *mutationResolver) CreateProduct(ctx context.Context, in ProductInput) (*Product, error) {
// 	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
// 	defer cancel()

// 	p, err := r.server.catalogClient.PostProduct(ctx, in.Name, in.Description, in.Price)
// 	if err != nil {
// 		log.Println(err)
// 		return nil, err
// 	}

// 	return &Product{
// 		ID:          p.ID,
// 		Name:        p.Name,
// 		Description: p.Description,
// 		Price:       p.Price,
// 	}, nil
// }

// func (r *mutationResolver) CreateOrder(ctx context.Context, in OrderInput) (*Order, error) {
// 	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
// 	defer cancel()

// 	var products []order.OrderedProduct
// 	for _, p := range in.Products {
// 		if p.Quantity <= 0 {
// 			return nil, ErrInvalidParameter
// 		}
// 		products = append(products, order.OrderedProduct{
// 			ID:       p.ID,
// 			Quantity: uint32(p.Quantity),
// 		})
// 	}
// 	o, err := r.server.orderClient.PostOrder(ctx, in.AccountID, products)
// 	if err != nil {
// 		log.Println(err)
// 		return nil, err
// 	}

// 	return &Order{
// 		ID:         o.ID,
// 		CreatedAt:  o.CreatedAt,
// 		TotalPrice: o.TotalPrice,
// 	}, nil
// }
