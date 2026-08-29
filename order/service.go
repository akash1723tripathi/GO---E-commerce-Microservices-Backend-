package order

import (
	"context"
	"errors"
	"time"

	"github.com/akash1723tripathi/go-microservices/account"
	"github.com/akash1723tripathi/go-microservices/catalog"
	"github.com/segmentio/ksuid"
)

var (
	ErrInvalidParameter = errors.New("invalid order parameter")
	ErrAccountNotFound  = errors.New("account not found")
	ErrProductNotFound  = errors.New("product not found")
)

type AccountVerifier interface {
	GetAccount(context.Context, string) (*account.Account, error)
}
type ProductVerifier interface {
	GetProducts(context.Context, uint64, uint64, []string, string) ([]catalog.Product, error)
}

type OrderedProduct struct {
	ID, Name, Description string
	Price                 float64
	Quantity              uint32
}
type Order struct {
	ID, AccountID string
	CreatedAt     time.Time
	TotalPrice    float64
	Products      []OrderedProduct
}

type Service interface {
	PostOrder(context.Context, string, []OrderedProduct) (*Order, error)
	GetOrder(context.Context, string) (*Order, error)
	GetOrders(context.Context, string, uint64, uint64) ([]Order, error)
}
type orderService struct {
	repository Repository
	accounts   AccountVerifier
	products   ProductVerifier
}

func NewService(r Repository, accounts AccountVerifier, products ProductVerifier) Service {
	return &orderService{r, accounts, products}
}

func (s *orderService) PostOrder(ctx context.Context, accountID string, requested []OrderedProduct) (*Order, error) {
	if accountID == "" || len(requested) == 0 {
		return nil, ErrInvalidParameter
	}
	if _, err := s.accounts.GetAccount(ctx, accountID); err != nil {
		return nil, err
	}
	ids := make([]string, len(requested))
	quantities := make(map[string]uint32, len(requested))
	for i, p := range requested {
		if p.ID == "" || p.Quantity == 0 || quantities[p.ID] != 0 {
			return nil, ErrInvalidParameter
		}
		ids[i] = p.ID
		quantities[p.ID] = p.Quantity
	}
	products, err := s.products.GetProducts(ctx, 0, uint64(len(ids)), ids, "")
	if err != nil {
		return nil, err
	}
	if len(products) != len(ids) {
		return nil, ErrProductNotFound
	}
	o := &Order{ID: ksuid.New().String(), AccountID: accountID, CreatedAt: time.Now().UTC(), Products: make([]OrderedProduct, 0, len(products))}
	for _, p := range products {
		q := quantities[p.ID]
		if q == 0 {
			return nil, ErrProductNotFound
		}
		o.Products = append(o.Products, OrderedProduct{ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price, Quantity: q})
		o.TotalPrice += p.Price * float64(q)
	}
	if err := s.repository.PutOrder(ctx, *o); err != nil {
		return nil, err
	}
	return o, nil
}
func (s *orderService) GetOrder(ctx context.Context, id string) (*Order, error) {
	if id == "" {
		return nil, ErrInvalidParameter
	}
	return s.repository.GetOrderByID(ctx, id)
}
func (s *orderService) GetOrders(ctx context.Context, accountID string, skip, take uint64) ([]Order, error) {
	if accountID == "" {
		return nil, ErrInvalidParameter
	}
	if take == 0 || take > 100 {
		take = 100
	}
	return s.repository.ListOrdersByAccountID(ctx, accountID, skip, take)
}
