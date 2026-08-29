package graphql

import (
	gqlgraphql "github.com/99designs/gqlgen/graphql"
	"github.com/akash1723tripathi/go-microservices/account"
	"github.com/akash1723tripathi/go-microservices/catalog"
	"github.com/akash1723tripathi/go-microservices/order"
)

type Server struct {
	accountClient *account.Client
	catalogClient *catalog.Client
	orderClient   *order.Client
}

func NewGraphQLServer(accountUrl, catalogURL, orderURL string) (*Server, error) {
	// Connect to account service
	accountClient, err := account.NewClient(accountUrl)
	if err != nil {
		return nil, err
	}

	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		accountClient.Close()
		return nil, err
	}

	orderClient, err := order.NewClient(orderURL)
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return nil, err
	}

	return &Server{
		accountClient: accountClient,
		catalogClient: catalogClient,
		orderClient:   orderClient,
	}, nil
}

func (s *Server) Account() AccountResolver {
	return &accountResolver{server: s}
}

func (s *Server) Mutation() MutationResolver {
	return &mutationResolver{server: s}
}

func (s *Server) Query() QueryResolver {
	return &queryResolver{server: s}
}

func (s *Server) ToExecutableSchema() gqlgraphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: s,
	})
}
