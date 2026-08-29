CREATE TABLE IF NOT EXISTS orders (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  total_price DOUBLE PRECISION NOT NULL CHECK (total_price >= 0)
);
CREATE TABLE IF NOT EXISTS order_products (
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  price DOUBLE PRECISION NOT NULL CHECK (price >= 0),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  PRIMARY KEY (order_id, product_id)
);
CREATE INDEX IF NOT EXISTS orders_account_created_idx ON orders(account_id, created_at DESC);
