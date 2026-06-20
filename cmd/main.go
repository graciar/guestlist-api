package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/graciar/guestlist-api/internal/env"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	godotenv.Load()
	ctx := context.Background()

	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", ""),
		},
		oauth: oauthConfig{
			clientID:     env.GetString("GOOGLE_CLIENT_ID", ""),
			clientSecret: env.GetString("GOOGLE_CLIENT_SECRET", ""),
			redirectURL:  env.GetString("GOOGLE_REDIRECT_URL", ""),
		},
	}

	googleOAuth := &oauth2.Config{
		ClientID:     env.GetString("GOOGLE_CLIENT_ID", ""),
		ClientSecret: env.GetString("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  env.GetString("GOOGLE_REDIRECT_URL", ""),
		Scopes:       []string{"email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// database
	conn, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	logger.Info("connected to database", "dsn", cfg.db.dsn)

	api := application{
		config:      cfg,
		db:          conn,
		googleOAuth: googleOAuth,
	}

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := api.run(api.mount()); err != nil {
		// log.Println("Error running server:", err)
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
