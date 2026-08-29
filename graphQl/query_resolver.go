package graphql

import (
	"context"
	accountpb "github.com/akash1723tripathi/go-microservices/account"
)

type queryResolver struct {
	server *Server
}

func (r *queryResolver) Accounts(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	skip, take := bounds(pagination)
	accounts, err := r.server.accountClient.GetAccounts(ctx, skip, take)
	if err != nil {
		return nil, err
	}
	if id != nil {
		account, err := r.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			return nil, err
		}
		accounts = []accountpb.Account{*account}
	}
	result := make([]*Account, 0, len(accounts))
	for _, a := range accounts {
		result = append(result, &Account{ID: a.ID, Name: a.Name})
	}
	return result, nil
}
func (r *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id *string) ([]*Product, error) {
	skip, take := bounds(pagination)
	q := ""
	if query != nil {
		q = *query
	}
	ids := []string{}
	if id != nil {
		ids = []string{*id}
	}
	products, err := r.server.catalogClient.GetProducts(ctx, skip, take, ids, q)
	if err != nil {
		return nil, err
	}
	result := make([]*Product, 0, len(products))
	for _, p := range products {
		result = append(result, &Product{ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price})
	}
	return result, nil
}

func bounds(p *PaginationInput) (uint64, uint64) {
	skipValue := uint64(0)
	takeValue := uint64(100)
	if p == nil {
		return skipValue, takeValue
	}
	if p.Skip != nil {
		skipValue = uint64(*p.Skip)
	}
	if p.Take != nil && *p.Take > 0 {
		takeValue = uint64(*p.Take)
	}
	return skipValue, takeValue
}
