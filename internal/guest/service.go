package guest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	repo "github.com/graciar/guestlist-api/internal/adapters/postgresql/sqlc"
	"github.com/graciar/guestlist-api/internal/auth"
	"github.com/graciar/guestlist-api/internal/event"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type svc struct {
	repo *repo.Queries
	db   *pgxpool.Pool
}

func NewService(repo *repo.Queries, db *pgxpool.Pool) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) CreateGuest(ctx context.Context, req CreateGuestInput) (Guest, error) {
	// 1. Clean the input name
	cleanedName := strings.TrimSpace(req.GuestName)
	if cleanedName == "" {
		return Guest{}, fmt.Errorf("guest name cannot be empty")
	}

	// 2. Find the Event
	Event, err := s.repo.FindEventByID(ctx, req.EventId)
	if err != nil {
		return Guest{}, err
	}

	// 3. NEW: Check if name already exists for this event
	exists, _ := s.repo.CheckGuestExists(ctx, repo.CheckGuestExistsParams{
		EventID: req.EventId,
		Btrim:   cleanedName,
	})
	if exists {
		return Guest{}, fmt.Errorf("a guest named '%s' is already registered for this event", cleanedName)
	}

	var rvsp_status repo.RsvpStatus
	var rsvpTokenStr string
	var signatureStr string

	if Event.Type == "public" || Event.Type == "private" {
		rvsp_status = repo.RsvpStatusAttending
		tokenBytes := make([]byte, 16)
		if _, err := rand.Read(tokenBytes); err != nil {
			return Guest{}, fmt.Errorf("failed to generate ticket signature: %w", err)
		}
		signatureStr = hex.EncodeToString(tokenBytes)
	} else {
		rvsp_status = repo.RsvpStatusPending

		tokenBytes := make([]byte, 16)
		if _, err := rand.Read(tokenBytes); err != nil {
			return Guest{}, err
		}
		rsvpTokenStr = hex.EncodeToString(tokenBytes)
	}

	dbGuest, err := s.repo.CreateGuest(ctx, repo.CreateGuestParams{
		EventID: req.EventId,
		Name:    cleanedName,
		Email: pgtype.Text{
			String: req.GuestEmail,
			Valid:  req.GuestEmail != "",
		},
		Phone: pgtype.Text{
			String: req.GuestPhone,
			Valid:  req.GuestPhone != "",
		},
		RsvpStatus:  rvsp_status,
		RsvpToken:   pgtype.Text{String: rsvpTokenStr, Valid: rsvpTokenStr != ""},
		TicketCode:  pgtype.Text{String: signatureStr, Valid: signatureStr != ""},
		IsCheckedIn: false,
	})

	if err != nil {
		return Guest{}, err
	}
	return Guest{
		GuestId:     dbGuest.ID,
		EventId:     dbGuest.EventID,
		GuestName:   dbGuest.Name,
		GuestEmail:  dbGuest.Email.String,
		GuestPhone:  dbGuest.Phone.String,
		RsvpStatus:  string(dbGuest.RsvpStatus),
		RsvpToken:   dbGuest.RsvpToken.String,
		TicketCode:  dbGuest.TicketCode.String,
		IsCheckedIn: dbGuest.IsCheckedIn,
	}, nil
}

func (s *svc) UpdateGuest(ctx context.Context, ID string, req UpdateGuestInput) (Guest, error) {
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return Guest{}, fmt.Errorf("unauthenticated request")
	}

	guest, err := s.repo.FindGuestByID(ctx, ID)
	if err != nil {
		return Guest{}, err
	}
	if guest.HostID != requestingUserID {
		return Guest{}, fmt.Errorf("forbidden: you do not own this event")
	}

	dbGuest, err := s.repo.UpdateGuest(ctx, repo.UpdateGuestParams{
		ID:      ID,
		EventID: guest.EventID,
		Name:    req.GuestName,
		Email: pgtype.Text{
			String: req.GuestEmail,
			Valid:  true,
		},
		Phone: pgtype.Text{
			String: req.GuestPhone,
			Valid:  true,
		},
		RsvpStatus: guest.RsvpStatus,
	})
	if err != nil {
		return Guest{}, err
	}
	return Guest{
		GuestId:     dbGuest.ID,
		EventId:     dbGuest.EventID,
		GuestName:   dbGuest.Name,
		GuestEmail:  dbGuest.Email.String,
		GuestPhone:  dbGuest.Phone.String,
		RsvpStatus:  string(dbGuest.RsvpStatus),
		TicketCode:  dbGuest.TicketCode.String,
		IsCheckedIn: dbGuest.IsCheckedIn,
	}, nil
}

func (s *svc) ListGuests(ctx context.Context) ([]Guest, error) {
	guests, err := s.repo.ListGuests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list guests from database: %w", err)
	}

	GuestLists := make([]Guest, len(guests))
	for i, guest := range guests {
		GuestLists[i] = Guest{
			GuestId:     guest.ID,
			EventId:     guest.EventID,
			GuestName:   guest.Name,
			GuestEmail:  guest.Email.String,
			GuestPhone:  guest.Phone.String,
			IsCheckedIn: guest.IsCheckedIn,
		}
	}

	return GuestLists, nil
}

func (s *svc) GetGuestByID(ctx context.Context, ID string, eventSlug string) (Guest, error) {
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return Guest{}, fmt.Errorf("unauthorized request")
	}

	guest, err := s.repo.FindGuestByID(ctx, ID)
	if err != nil {
		return Guest{}, err
	}

	if guest.EventSlug != eventSlug {
		return Guest{}, fmt.Errorf("forbidden: you do not own this event")
	}

	userRole, ok := auth.UserRoleFromContext(ctx)
	if !ok {
		return Guest{}, fmt.Errorf("unauthorized request")
	}

	if guest.HostID != requestingUserID && userRole != "admin" {
		return Guest{}, fmt.Errorf("forbidden: you do not own this event")
	}

	return Guest{
		GuestId:     guest.GuestID,
		EventId:     guest.EventID,
		GuestName:   guest.GuestName,
		GuestEmail:  guest.GuestEmail.String,
		GuestPhone:  guest.GuestPhone.String,
		RsvpStatus:  string(guest.RsvpStatus),
		TicketCode:  guest.TicketCode.String,
		IsCheckedIn: guest.IsCheckedIn,
	}, nil
}

func (s *svc) DeleteGuest(ctx context.Context, ID string) error {
	// 1. Get the authenticated user ID from the context
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("unauthorized request")
	}

	// 2. Find the existing event to check who owns it
	event, err := s.repo.FindEventByID(ctx, ID)
	if err != nil {
		return fmt.Errorf("failed to find event in database: %w", err)
	}

	userRole, ok := auth.UserRoleFromContext(ctx)
	if !ok {
		return fmt.Errorf("unauthorized request")
	}

	// 3. Authorization Check: Does the requester own this event?
	if event.UserID != requestingUserID && userRole != "admin" {
		return fmt.Errorf("forbidden: you do not own this event")
	}

	// 4. Delete the event
	_, err = s.repo.DeleteEvent(ctx, ID)
	if err != nil {
		return fmt.Errorf("failed to delete event from database: %w", err)
	}

	return nil
}

func (s *svc) GetGuestsForEvent(ctx context.Context, eventID string) (EventGuests, error) {
	// 1. Get the current logged-in user from the context backpack
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return EventGuests{}, fmt.Errorf("unauthorized")
	}

	userRole, ok := auth.UserRoleFromContext(ctx)
	if !ok {
		return EventGuests{}, fmt.Errorf("unauthorized")
	}

	// 2. Cross-domain validation: Verify this event belongs to the requesting user
	// Your SQLC repo can execute a query like: FindEventByID
	dbEvent, err := s.repo.FindEventByID(ctx, eventID)
	if err != nil {
		return EventGuests{}, fmt.Errorf("event not found or query failed: %w", err)
	}
	if dbEvent.UserID != userID && userRole != "admin" {
		return EventGuests{}, fmt.Errorf("forbidden: you do not own this event")
	}

	// 3. Fetch the guests since validation passed
	dbGuests, err := s.repo.GetGuestsByEventID(ctx, eventID)
	if err != nil {
		return EventGuests{}, fmt.Errorf("failed to fetch guests: %w", err)
	}

	Guests := make([]Guest, len(dbGuests))

	for i, guest := range dbGuests {
		Guests[i] = Guest{
			GuestId:     guest.ID,
			EventId:     guest.EventID,
			GuestName:   guest.Name,
			GuestEmail:  guest.Email.String,
			GuestPhone:  guest.Phone.String,
			RsvpStatus:  string(guest.RsvpStatus),
			TicketCode:  guest.TicketCode.String,
			IsCheckedIn: guest.IsCheckedIn,
		}
	}

	GuestsList := EventGuests{
		Event: event.Event{
			ID:          dbEvent.ID,
			HostID:      dbEvent.UserID,
			Title:       dbEvent.Title,
			Slug:        dbEvent.Slug,
			Description: dbEvent.Description,
			Location:    dbEvent.Location,
			Time:        dbEvent.StartTime.Time,
			Type:        string(dbEvent.Type),
			Status:      string(dbEvent.Status),
		},
		Guests: Guests,
	}

	return GuestsList, nil
}

func (s *svc) GetGuestTicket(ctx context.Context, ID string, eventSlug string) (GuestTicket, error) {
	guest, err := s.repo.FindGuestByID(ctx, ID)
	if err != nil {
		return GuestTicket{}, err
	}

	if guest.EventSlug != eventSlug {
		return GuestTicket{}, fmt.Errorf("forbidden: you do not own this event")
	}

	dbEvent, err := s.repo.FindEventByID(ctx, guest.EventID)
	if err != nil {
		return GuestTicket{}, fmt.Errorf("event not found or query failed: %w", err)
	}

	return GuestTicket{
		GuestId:   guest.GuestID,
		EventId:   dbEvent.ID,
		GuestName: guest.GuestName,
		Ticket:    guest.TicketCode.String,
		EventName: dbEvent.Title,
	}, nil
}

func (s *svc) HandleRSVPResponse(ctx context.Context, ID string, token string, status string) error {
	if token == "" || ID == "" {
		return fmt.Errorf("id and token are required")
	}

	// Start the Database Transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Defer a rollback. If the function returns early due to an error,
	// tx.Rollback() ensures the database connection cleans up safely.
	// If tx.Commit() is called first, Rollback() does nothing.
	defer tx.Rollback(ctx)

	// Create a repository instance that runs inside this specific transaction
	txRepo := s.repo.WithTx(tx) // SQLC automatically generates .WithTx() for you!

	// Find the guest (Row is now LOCKED by FOR UPDATE)
	guest, err := txRepo.GetRsvpGuestByToken(ctx, repo.GetRsvpGuestByTokenParams{
		ID:        ID,
		RsvpToken: pgtype.Text{String: token, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to find guest: %w", err)
	}

	// Add a business rule check while the row is locked
	if guest.RsvpStatus == repo.RsvpStatusAttending || guest.RsvpStatus == repo.RsvpStatusDeclined {
		return fmt.Errorf("guest has already rsvp'd")
	}

	// Update the guest status safely
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate ticket signature: %w", err)
	}
	signatureStr := hex.EncodeToString(tokenBytes)

	_, err = txRepo.UpdateGuest(ctx, repo.UpdateGuestParams{
		ID:         ID,
		EventID:    guest.EventID,
		RsvpStatus: repo.RsvpStatus(status),
		TicketCode: pgtype.Text{String: signatureStr, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update guest: %w", err)
	}

	// Commit everything permanently to the database
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the updated data mapped to your application model
	return nil
}

func (s *svc) HandleCheckIn(ctx context.Context, ID string, sig string) error {
	guest, err := s.repo.FindGuestByID(ctx, ID)
	if err != nil {
		return fmt.Errorf("failed to find guest: %w", err)
	}

	if guest.RsvpStatus != repo.RsvpStatusAttending {
		return fmt.Errorf("guest has not rsvp'd")
	}

	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("unauthorized request")
	}

	userRole, ok := auth.UserRoleFromContext(ctx)
	if !ok {
		return fmt.Errorf("unauthorized request")
	}

	if guest.HostID != requestingUserID && userRole != "admin" {
		return fmt.Errorf("forbidden: you do not own this event")
	}

	if guest.TicketCode.String != sig {
		return fmt.Errorf("invalid signature")
	}

	_, err = s.repo.UpdateGuest(ctx, repo.UpdateGuestParams{
		ID:          ID,
		EventID:     guest.EventID,
		Name:        guest.GuestName,
		RsvpStatus:  repo.RsvpStatusAttending,
		IsCheckedIn: true,
		CheckedInAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to check in guest: %w", err)
	}
	return nil
}
