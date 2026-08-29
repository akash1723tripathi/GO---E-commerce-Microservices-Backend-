package order

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("order not found")

type Repository interface {
	Close()
	PutOrder(context.Context, Order) error
	GetOrderByID(context.Context, string) (*Order, error)
	ListOrdersByAccountID(context.Context, string, uint64, uint64) ([]Order, error)
}

type postgresRepository struct{ db *sql.DB }

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return &postgresRepository{db: db}, nil
}

func (r *postgresRepository) Close() { _ = r.db.Close() }

func (r *postgresRepository) PutOrder(ctx context.Context, o Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO orders(id, account_id, created_at, total_price) VALUES ($1,$2,$3,$4)`, o.ID, o.AccountID, o.CreatedAt, o.TotalPrice)
	if err != nil {
		return err
	}
	for _, p := range o.Products {
		_, err = tx.ExecContext(ctx, `INSERT INTO order_products(order_id, product_id, name, description, price, quantity) VALUES ($1,$2,$3,$4,$5,$6)`, o.ID, p.ID, p.Name, p.Description, p.Price, p.Quantity)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

const orderSelect = `SELECT o.id, o.account_id, o.created_at, o.total_price,
 op.product_id, op.name, op.description, op.price, op.quantity
 FROM orders o JOIN order_products op ON op.order_id = o.id`

func scanOrders(rows *sql.Rows) ([]Order, error) {
	byID := make(map[string]int)
	orders := make([]Order, 0)
	for rows.Next() {
		var o Order
		var p OrderedProduct
		if err := rows.Scan(&o.ID, &o.AccountID, &o.CreatedAt, &o.TotalPrice, &p.ID, &p.Name, &p.Description, &p.Price, &p.Quantity); err != nil {
			return nil, err
		}
		idx, ok := byID[o.ID]
		if !ok {
			idx = len(orders)
			byID[o.ID] = idx
			orders = append(orders, o)
		}
		orders[idx].Products = append(orders[idx].Products, p)
	}
	return orders, rows.Err()
}

func (r *postgresRepository) GetOrderByID(ctx context.Context, id string) (*Order, error) {
	rows, err := r.db.QueryContext(ctx, orderSelect+` WHERE o.id=$1 ORDER BY op.product_id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders, err := scanOrders(rows)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrNotFound
	}
	return &orders[0], nil
}

func (r *postgresRepository) ListOrdersByAccountID(ctx context.Context, accountID string, skip, take uint64) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, orderSelect+` WHERE o.id IN (SELECT id FROM orders WHERE account_id=$1 ORDER BY created_at DESC OFFSET $2 LIMIT $3) ORDER BY o.created_at DESC, op.product_id`, accountID, skip, take)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}
