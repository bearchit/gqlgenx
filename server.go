package gqlgenx

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"couse/pkg/gqlgenx/graphiql"

	"github.com/99designs/gqlgen/graphql/handler/lru"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/bearchit/goboost/graphql/voyager"
	"github.com/creasty/defaults"
	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/rs/cors"
)

type ServerOption struct {
	Addr          string `validate:"required" default:":8080"`
	Endpoint      string `validate:"required" default:"/graphql"`
	GraphiQL      *bool  `validate:"required" default:"true"`
	Playground    *bool  `validate:"required" default:"false"`
	Voyager       *bool  `validate:"required" default:"true"`
	Introspection *bool  `validate:"required" default:"true"`
	CORS          *struct {
		AllowedOrigins   []string
		AllowCredentials bool `default:"true"`
	}
	HealthCheck struct {
		Path string `validate:"required" default:"/health_check"`
	}
}

func (opt ServerOption) FullEndpoint() string {
	u := url.URL{Scheme: "http", Host: opt.Addr, Path: opt.Endpoint}
	return u.String()
}

func (opt ServerOption) BasePath() string {
	u := url.URL{Scheme: "http", Host: opt.Addr}
	return u.String()
}

type Server struct {
	option  ServerOption
	handler http.Handler
}

func (s Server) Option() ServerOption {
	return s.option
}

func (s Server) Serve() error {
	return http.ListenAndServe(s.option.Addr, s.handler)
}

func NewServer(es graphql.ExecutableSchema, opt ServerOption) (*Server, error) {
	if err := defaults.Set(&opt); err != nil {
		return nil, fmt.Errorf("server option error: %w", err)
	}

	if err := validator.New().Struct(&opt); err != nil {
		return nil, fmt.Errorf("server option error: %w", err)
	}

	srv := handler.New(es)
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.SetQueryCache(lru.New(1000))
	if *opt.Introspection {
		srv.Use(extension.Introspection{})
	}
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	r := chi.NewRouter()
	if opt.CORS != nil {
		r.Use(cors.New(cors.Options{
			AllowedOrigins:   opt.CORS.AllowedOrigins,
			AllowCredentials: opt.CORS.AllowCredentials,
		}).Handler)
	}

	r.Get(opt.HealthCheck.Path, func(w http.ResponseWriter, req *http.Request) {
		if _, err := w.Write([]byte("I'm healthy")); err != nil {
			panic(err)
		}
	})

	if *opt.GraphiQL {
		r.Handle("/", graphiql.Handler("GraphiQL", opt.Endpoint))
	}

	if *opt.Playground {
		r.Handle("/playground", playground.Handler("GraphQL Playground", opt.Endpoint))
	}

	if *opt.Voyager {
		r.Handle("/voyager", voyager.Handler("GraphQL Voyager", opt.Endpoint))
	}

	r.Handle(opt.Endpoint, srv)

	return &Server{
		option:  opt,
		handler: r,
	}, nil
}
