package router

import (
	"net/http"
	"time"

	repo "github.com/graciar/guestlist-api/internal/adapters/postgresql/sqlc"
	"github.com/graciar/guestlist-api/internal/auth"
	"github.com/graciar/guestlist-api/internal/event"
	"github.com/graciar/guestlist-api/internal/guest"
	"github.com/graciar/guestlist-api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

func AuthRoutes(db *pgxpool.Pool, oauthCfg *oauth2.Config) http.Handler {
	r := chi.NewRouter()
	userService := auth.NewService(repo.New(db), db)
	authHandler := auth.NewHandler(userService, oauthCfg)

	r.Post("/signup", authHandler.SignUp)
	r.Post("/signin", authHandler.SignIn)
	r.Post("/refresh", authHandler.RefreshToken)
	r.Post("/forgot-password", authHandler.RequestPasswordReset)
	r.Post("/reset-password", authHandler.ResetPassword)

	// oauth
	r.Get("/oauth", authHandler.GoogleLogin)
	r.Get("/callback", authHandler.GoogleCallback)
	return r
}

func PublicRoutes(db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	guestService := guest.NewService(repo.New(db), db)
	guestHandler := guest.NewHandler(guestService)

	r.With(httprate.Limit(
		10,
		60*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	)).Get("/guest/{eventSlug}/{id}", guestHandler.GetGuestTicket) // id: GuestID
	r.Post("/{id}/rsvp/attend", guestHandler.HandleRSVPAttend)   // id: GuestID
	r.Post("/{id}/rsvp/decline", guestHandler.HandleRSVPDecline) // id: GuestID
	return r
}

func EventRoutes(db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	eventService := event.NewService(repo.New(db), db)
	eventHandler := event.NewHandler(eventService)

	r.Post("/", eventHandler.CreateEvent)
	r.With(middleware.AdminOnly).Get("/", eventHandler.ListEvents)
	r.Put("/{id}", eventHandler.UpdateEvent)
	r.Delete("/{id}", eventHandler.DeleteEvent)
	r.Get("/{id}", eventHandler.GetEventByID)

	r.Get("/me", eventHandler.GetUserEvents)
	r.Get("/my-stats", eventHandler.GetUserEventStats)
	r.Get("/{id}/stats", eventHandler.GetEventStats)

	return r
}

func GuestRoutes(db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	guestService := guest.NewService(repo.New(db), db)
	guestHandler := guest.NewHandler(guestService)

	r.With(middleware.AdminOnly).Get("/", guestHandler.ListGuests)
	r.Post("/", guestHandler.CreateGuest)
	r.Put("/{id}", guestHandler.UpdateGuest)
	r.Delete("/{id}", guestHandler.DeleteGuest)
	r.Get("/{eventSlug}/{id}", guestHandler.GetGuestByID)

	r.Get("/event/{eventID}", guestHandler.GetGuestsForEvent)
	r.Post("/{id}/rsvp/attend", guestHandler.HandleRSVPAttend)
	r.Post("/{id}/rsvp/decline", guestHandler.HandleRSVPDecline)
	r.Post("/{id}/ticket/checkin", guestHandler.HandleCheckIn) // check in guest + 'sig' as query param

	return r
}
