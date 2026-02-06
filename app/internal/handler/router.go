package handler

import (
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/RoGogDBD/GQLGo/internal/qraphql/graph"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
)

func NewRouter(resolver *graph.Resolver) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](1000)})

	r.GET("/", gin.WrapH(playground.Handler("GraphQL playground", "/query")))
	r.POST("/query", gin.WrapH(srv))
	r.GET("/query", gin.WrapH(srv))
	return r
}
