package server

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"time"

	"github.com/ccamel/go-graphql-subscription-example/static"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-transport-ws/graphqlws"
	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"
)

func StartServer(cfg *Configuration) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := http.NewServeMux()
	router.Handle("/graphql", withMiddleware(graphqlApp(cfg)))
	router.Handle("/graphiql", withMiddleware(graphiqlApp(cfg)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.
		Info().
		Uint16("port", cfg.Port).
		Msg("Ready to handle requests")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.
			Error().
			Err(err).
			Uint16("port", cfg.Port).
			Msg("Could not start server")
	}
}

func graphiqlApp(cfg *Configuration) http.Handler {
	t := template.Must(template.New("graphiql").Parse(static.FSMustString(false, "/static/graphiql/graphiql.html")))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, cfg.Port); err != nil {
			log.
				Error().
				Err(err).
				Str("template", t.Name()).
				Msg("Failed to serve template")
		}
	})
}

func graphqlApp(cfg *Configuration) http.Handler {
	s := graphql.MustParseSchema(static.FSMustString(false, "/static/graphql/schema/subscription-api.graphql"), NewResolver(cfg))

	graphQLHandler := graphqlws.NewHandlerFunc(s, &relay.Handler{Schema: s})

	return graphQLHandler
}

func withMiddleware(handler http.Handler) http.Handler {
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
