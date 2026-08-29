package order

import (
	"context"
	"testing"

	"github.com/akash1723tripathi/go-microservices/account"
	"github.com/akash1723tripathi/go-microservices/catalog"
)

type accountVerifier struct{ err error }

func (a accountVerifier) GetAccount(context.Context, string) (*account.Account, error) {
	return &account.Account{ID: "a"}, a.err
}

type productVerifier struct {
	products []catalog.Product
	err      error
}

func (p productVerifier) GetProducts(context.Context, uint64, uint64, []string, string) ([]catalog.Product, error) {
	return p.products, p.err
}

type orderRepository struct{ saved *Order }

func (r *orderRepository) Close()                                               {}
func (r *orderRepository) PutOrder(_ context.Context, o Order) error            { r.saved = &o; return nil }
func (r *orderRepository) GetOrderByID(context.Context, string) (*Order, error) { return r.saved, nil }
func (r *orderRepository) ListOrdersByAccountID(context.Context, string, uint64, uint64) ([]Order, error) {
	if r.saved == nil {
		return nil, nil
	}
	return []Order{*r.saved}, nil
}

func TestPostOrderCalculatesTotalAndSnapshotsProducts(t *testing.T) {
	repo := &orderRepository{}
	s := NewService(repo, accountVerifier{}, productVerifier{products: []catalog.Product{{ID: "p1", Name: "Book", Price: 12.5}, {ID: "p2", Name: "Pen", Price: 2}}})
	o, err := s.PostOrder(context.Background(), "a", []OrderedProduct{{ID: "p1", Quantity: 2}, {ID: "p2", Quantity: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if o.TotalPrice != 31 {
		t.Fatalf("total = %v, want 31", o.TotalPrice)
	}
	if len(o.Products) != 2 || o.Products[0].Name != "Book" || o.Products[0].Quantity != 2 {
		t.Fatalf("unexpected products: %+v", o.Products)
	}
	if repo.saved == nil || repo.saved.ID != o.ID {
		t.Fatal("order was not persisted")
	}
}

func TestPostOrderRejectsInvalidQuantity(t *testing.T) {
	s := NewService(&orderRepository{}, accountVerifier{}, productVerifier{})
	_, err := s.PostOrder(context.Background(), "a", []OrderedProduct{{ID: "p1", Quantity: 0}})
	if err != ErrInvalidParameter {
		t.Fatalf("err = %v, want ErrInvalidParameter", err)
	}
}

func TestPostOrderRejectsMissingProduct(t *testing.T) {
	s := NewService(&orderRepository{}, accountVerifier{}, productVerifier{products: []catalog.Product{}})
	_, err := s.PostOrder(context.Background(), "a", []OrderedProduct{{ID: "missing", Quantity: 1}})
	if err != ErrProductNotFound {
		t.Fatalf("err = %v, want ErrProductNotFound", err)
	}
}
