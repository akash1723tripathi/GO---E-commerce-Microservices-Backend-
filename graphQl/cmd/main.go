package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	graphqlserver "github.com/akash1723tripathi/go-microservices/graphql"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL string `envconfig:"CATALOG_SERVICE_URL"`
	OrderURL   string `envconfig:"ORDER_SERVICE_URL"`
}

func main() {
	var cfg AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	s, err := graphqlserver.NewGraphQLServer(cfg.AccountURL, cfg.CatalogURL, cfg.OrderURL)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/graphql", handler.New(s.ToExecutableSchema()))
	http.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	log.Fatal(http.ListenAndServe(":8080", nil))

	fmt.Println("Working")
}
