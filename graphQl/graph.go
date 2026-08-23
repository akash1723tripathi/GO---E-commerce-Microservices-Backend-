package graphql

import gqlgraphql "github.com/99designs/gqlgen/graphql"

type Server struct{
	// accountClient *account.client
	// catalogClient *catalog.client
	// orderClient  *order.client
}

func NewGraphQLServer(accountUrl, catalogURL, orderURL string) (*Server, error) {
	// // Connect to account service
	// accountClient, err := account.NewClient(accountUrl)
	// if err != nil {
	// 	return nil, err
	// }

	// // Connect to product service
	// catalogClient, err := catalog.NewClient(catalogURL)
	// if err != nil {
	// 	accountClient.Close()
	// 	return nil, err
	// }

	// // Connect to order service
	// orderClient, err := order.NewClient(orderURL)
	// if err != nil {
	// 	accountClient.Close()
	// 	catalogClient.Close()
	// 	return nil, err
	// }

	return &Server{
		// accountClient,
		// catalogClient,
		// orderClient,
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
