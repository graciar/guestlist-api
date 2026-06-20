package event

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	repo "github.com/graciar/guestlist-api/internal/adapters/postgresql/sqlc"
	"github.com/graciar/guestlist-api/internal/auth"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEventNotFound   = errors.New("event not found")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrNotOwner        = errors.New("forbidden: you do not own this event")
)

type svc struct {
	repo *repo.Queries
	db   *pgxpool.Pool
}

func NewService(repo *repo.Queries, db *pgxpool.Pool) Service {
	return &svc{repo: repo, db: db}
}

func randomSuffix() string {
	b := make([]byte, 2) // 2 bytes = 4 hex characters
	if _, err := rand.Read(b); err != nil {
		return "x" // fallback
	}
	return hex.EncodeToString(b)
}

func generateSlug(title string) string {
	// 1. Lowercase the string
	slug := strings.ToLower(title)

	// 2. Replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")

	// 3. Trim leading/trailing hyphens
	return strings.Trim(slug, "-")
}

func (s *svc) CreateEvent(ctx context.Context, tempEvent CreateEventInput) (Event, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return Event{}, fmt.Errorf("unauthorized")
	}

	userEvent, _ := s.repo.GetEventByHostIDAndTitle(ctx, repo.GetEventByHostIDAndTitleParams{
		UserID: userID,
		Title:  tempEvent.Title,
	})

	if userEvent.ID != "" {
		return Event{}, fmt.Errorf("you already have an event with this title")
	}

	baseSlug := generateSlug(tempEvent.Title)

	// e.g., "my-awesome-event-a3f2"
	computedSlug := fmt.Sprintf("%s-%s", baseSlug, randomSuffix())

	// 1. Execute the database query and catch the SQLC row result
	dbEvent, err := s.repo.CreateEvent(ctx, repo.CreateEventParams{
		UserID:      userID,
		Title:       tempEvent.Title,
		Slug:        computedSlug,
		Description: tempEvent.Description,
		Location:    tempEvent.Location,
		StartTime: pgtype.Timestamp{
			Time:  tempEvent.Time,
			Valid: true,
		},
		Type:   repo.EventType(tempEvent.Type),
		Status: repo.EventStatus(tempEvent.Status),
	})
	if err != nil {
		return Event{}, fmt.Errorf("failed to create event in database: %w", err)
	}

	// 2. Explicitly map the repo.Event to your domain Event
	return Event{
		ID:          dbEvent.ID,
		HostID:      dbEvent.UserID,
		Title:       dbEvent.Title,
		Slug:        dbEvent.Slug,
		Description: dbEvent.Description,
		Location:    dbEvent.Location,
		Time:        dbEvent.StartTime.Time,
		Type:        string(dbEvent.Type),
		Status:      string(dbEvent.Status),
	}, nil
}

func (s *svc) UpdateEvent(ctx context.Context, ID string, tempEvent UpdateEventInput) (Event, error) {
	// 1. Get the authenticated user ID from the context
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return Event{}, ErrUnauthenticated
	}

	// 2. Find the existing event to check who owns it
	event, err := s.repo.FindEventByID(ctx, ID)
	if err != nil {
		return Event{}, fmt.Errorf("failed to find event in database: %w", err)
	}

	userRole, ok := auth.UserRoleFromContext(ctx)
	if !ok {
		return Event{}, ErrUnauthenticated
	}

	// 3. Authorization Check: Does the requester own this event?
	if event.UserID != requestingUserID && userRole != "admin" {
		return Event{}, ErrNotOwner
	}

	// 1. Execute the database query and catch the SQLC row result
	dbEvent, err := s.repo.UpdateEvent(ctx, repo.UpdateEventParams{
		ID:          ID,
		Title:       tempEvent.Title,
		Description: tempEvent.Description,
		Location:    tempEvent.Location,
		StartTime: pgtype.Timestamp{
			Time:  tempEvent.Time,
			Valid: true,
		},
		Type:   repo.EventType(tempEvent.Type),
		Status: repo.EventStatus(tempEvent.Status),
	})
	if err != nil {
		return Event{}, fmt.Errorf("failed to create event in database: %w", err)
	}

	return Event{
		ID:          dbEvent.ID,
		HostID:      dbEvent.UserID,
		Title:       dbEvent.Title,
		Slug:        dbEvent.Slug,
		Description: dbEvent.Description,
		Location:    dbEvent.Location,
		Time:        dbEvent.StartTime.Time,
		Type:        string(dbEvent.Type),
		Status:      string(dbEvent.Status),
	}, nil
}

func (s *svc) DeleteEvent(ctx context.Context, ID string) error {
	// 1. Get the authenticated user ID from the context
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return ErrUnauthenticated
	}

	// 2. Find the existing event to check who owns it
	event, err := s.repo.FindEventByID(ctx, ID)
	if err != nil {
		return fmt.Errorf("failed to find event in database: %w", err)
	}

	userRole, ok := auth.UserRoleFromContext(ctx)
	if !ok {
		return ErrUnauthenticated
	}

	// 3. Authorization Check: Does the requester own this event?
	if event.UserID != requestingUserID && userRole != "admin" {
		return ErrNotOwner
	}

	// 4. Delete the event
	_, err = s.repo.DeleteEvent(ctx, ID)
	if err != nil {
		return fmt.Errorf("failed to delete event from database: %w", err)
	}

	return nil
}

func (s *svc) ListEvents(ctx context.Context) ([]Event, error) {
	events, err := s.repo.ListEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list events from database: %w", err)
	}

	EventLists := make([]Event, len(events))
	for i, event := range events {
		EventLists[i] = Event{
			ID:          event.ID,
			HostID:      event.UserID,
			Title:       event.Title,
			Slug:        event.Slug,
			Description: event.Description,
			Location:    event.Location,
			Time:        event.StartTime.Time,
			Type:        string(event.Type),
			Status:      string(event.Status),
		}
	}

	return EventLists, nil
}

func (s *svc) GetEventByID(ctx context.Context, ID string) (Event, error) {
	// 1. Get the authenticated user ID from the context
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return Event{}, ErrUnauthenticated
	}

	userRole, ok := auth.UserRoleFromContext(ctx)
	if !ok {
		return Event{}, ErrUnauthenticated
	}

	// 2. Find the existing event to check who owns it
	event, err := s.repo.FindEventByID(ctx, ID)
	if err != nil {
		return Event{}, fmt.Errorf("failed to find event in database: %w", err)
	}

	if event.UserID != requestingUserID && userRole != "admin" {
		return Event{}, ErrNotOwner
	}

	return Event{
		ID:          event.ID,
		HostID:      event.UserID,
		Title:       event.Title,
		Description: event.Description,
		Location:    event.Location,
		Time:        event.StartTime.Time,
		Type:        string(event.Type),
		Status:      string(event.Status),
	}, nil
}

func (s *svc) GetUserEvents(ctx context.Context) ([]Event, error) {
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthenticated request")
	}

	events, err := s.repo.GetUserEvents(ctx, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user events from database: %w", err)
	}

	EventLists := make([]Event, len(events))
	for i, event := range events {
		EventLists[i] = Event{
			ID:          event.ID,
			HostID:      event.UserID,
			Title:       event.Title,
			Slug:        event.Slug,
			Description: event.Description,
			Location:    event.Location,
			Time:        event.StartTime.Time,
			Type:        string(event.Type),
			Status:      string(event.Status),
		}
	}

	return EventLists, nil
}

func (s *svc) GetUserEventStats(ctx context.Context) (UserEventStats, error) {
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return UserEventStats{}, fmt.Errorf("unauthenticated request")
	}

	stat, err := s.repo.GetUserEventStats(ctx, requestingUserID)
	if err != nil {
		// Check if pgx/database/sql returned ErrNoRows
		if errors.Is(err, sql.ErrNoRows) {
			// Return a blank UserEventStats struct and a nil error
			// because 0 stats is perfectly valid for a brand-new user!
			return UserEventStats{}, nil
		}
		return UserEventStats{}, fmt.Errorf("failed to get user event stats from database: %w", err)
	}

	return UserEventStats{
		TotalEvents:              stat.TotalEvents,
		OpenEvents:               stat.OpenEvents,
		ClosedEvents:             stat.ClosedEvents,
		CancelledEvents:          stat.CancelledEvents,
		TotalGuestsInvited:       stat.TotalGuestsInvited,
		TotalGuestsConfirmed:     stat.TotalGuestsConfirmed,
		TotalGuestsCheckedIn:     stat.TotalGuestsCheckedIn,
		AttendanceRatePercentage: float64(stat.AttendanceRatePercentage),
	}, nil
}

func (s *svc) GetEventStats(ctx context.Context, eventID string) (EventStatsResponse, error) {
	requestingUserID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return EventStatsResponse{}, fmt.Errorf("unauthenticated request")
	}

	event, err := s.repo.FindEventByID(ctx, eventID)
	if err != nil {
		return EventStatsResponse{}, fmt.Errorf("failed to find event in database: %w", err)
	}

	if event.UserID != requestingUserID {
		return EventStatsResponse{}, fmt.Errorf("unauthenticated request")
	}

	stat, err := s.repo.GetEventStats(ctx, eventID)
	if err != nil {
		return EventStatsResponse{}, fmt.Errorf("failed to get user event stats from database: %w", err)
	}

	return EventStatsResponse{
		EventID:                  stat.EventID,
		Title:                    stat.Title,
		Slug:                     stat.Slug,
		EventStatus:              string(stat.EventStatus),
		StartTime:                stat.StartTime.Time,
		TotalGuestsInvited:       stat.TotalGuestsInvited,
		TotalGuestsConfirmed:     stat.TotalGuestsConfirmed,
		TotalGuestsCheckedIn:     stat.TotalGuestsCheckedIn,
		AttendanceRatePercentage: float64(stat.AttendanceRatePercentage),
	}, nil
}
