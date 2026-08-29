package main

import (
	"github.com/akash1723tripathi/go-microservices/account"
	"github.com/akash1723tripathi/go-microservices/catalog"
	"github.com/akash1723tripathi/go-microservices/order"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
	"log"
	"time"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}
	var r order.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = order.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer r.Close()
	accounts, err := account.NewClient(cfg.AccountURL)
	if err != nil {
		log.Fatal(err)
	}
	defer accounts.Close()
	products, err := catalog.NewClient(cfg.CatalogURL)
	if err != nil {
		log.Fatal(err)
	}
	defer products.Close()
	log.Println("Listening on port 8080...")
	log.Fatal(order.ListenGRPC(order.NewService(r, accounts, products), 8080))
}
