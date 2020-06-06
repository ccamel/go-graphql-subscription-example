package server

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"time"

	"github.com/ccamel/go-graphql-subscription-example/server/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"github.com/ccamel/go-graphql-subscription-example/static"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-transport-ws/graphqlws"
	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"

	// Register available consumers
	_ "github.com/ccamel/go-graphql-subscription-example/server/consumer"
)

type Server struct {
	cfg *Configuration
	log zerolog.Logger
}

func NewServer(cfg *Configuration) *Server {
	return &Server{
		cfg,
		log.NewLogger(),
	}
}

func (s *Server) Start() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := http.NewServeMux()
	router.Handle("/graphql", withMiddleware(s.log, s.graphqlApp()))
	router.Handle("/graphiql", withMiddleware(s.log, s.graphiqlApp()))
	router.Handle("/metrics", withMiddleware(s.log, promhttp.Handler()))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	s.log.
		Info().
		Uint16("port", s.cfg.Port).
		Msg("Ready to handle requests")

	if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		s.log.
			Error().
			Err(err).
			Uint16("port", s.cfg.Port).
			Msg("Could not start server")
	}
}

func (s *Server) graphiqlApp() http.Handler {
	t := template.Must(template.New("graphiql").Parse(static.FSMustString(false, "/static/graphiql/graphiql.html")))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, s.cfg.Port); err != nil {
			s.log.
				Error().
				Err(err).
				Str("template", t.Name()).
				Msg("Failed to serve template")
		}
	})
}

func (s *Server) graphqlApp() http.Handler {
	resolver, err := NewResolver(s.cfg, s.log)
	if err != nil {
		s.log.
			Fatal().
			Err(err).
			Msg("Failed to create resolver")
	}

	schema := graphql.MustParseSchema(static.FSMustString(false, "/static/graphql/schema/subscription-api.graphql"), resolver)

	graphQLHandler := graphqlws.NewHandlerFunc(schema, &relay.Handler{Schema: schema})

	return graphQLHandler
}

func withMiddleware(log zerolog.Logger, handler http.Handler) http.Handler {
	return alice.
		New().
		Append(hlog.NewHandler(log)).
		Append(hlog.URLHandler("url")).
		Append(hlog.MethodHandler("method")).
		Append(hlog.RemoteAddrHandler("ip")).
		Append(hlog.UserAgentHandler("user_agent")).
		Append(hlog.RefererHandler("referer")).
		Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hlog.
				FromRequest(r).
				Info().
				Int64("size", r.ContentLength).
				Msg("⚡ incoming request")

			handler.ServeHTTP(w, r)
		}))
}
