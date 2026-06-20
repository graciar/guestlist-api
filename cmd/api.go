package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/graciar/guestlist-api/internal/env"
	authencation "github.com/graciar/guestlist-api/internal/middleware"
	"github.com/graciar/guestlist-api/internal/router"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

type dbConfig struct {
	dsn string
}

type application struct {
	config      config
	db          *pgxpool.Pool
	googleOAuth *oauth2.Config
}

type oauthConfig struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

type config struct {
	addr  string
	db    dbConfig
	oauth oauthConfig
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	allowedOriginsStr := env.GetString("ALLOWED_ORIGINS", "")
	allowedOrigins := strings.Split(allowedOriginsStr, ",")

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Client-Platform"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// set a timeout value on the request context(ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I'm alive"))
	})

	// r.Get("/docs", docs.Handler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/auth", router.AuthRoutes(app.db, app.googleOAuth))

		r.Mount("/public", router.PublicRoutes(app.db))

		r.Group(func(r chi.Router) {
			r.Use(authencation.Authenticate)
			r.Mount("/event", router.EventRoutes(app.db))
			r.Mount("/guest", router.GuestRoutes(app.db))
		})
	})

	return r
}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:        app.config.addr,
		Handler:     h,
		ReadTimeout: 10 * time.Second,
		// ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Printf("server has started at addr %s", app.config.addr)
	return srv.ListenAndServe()
}
