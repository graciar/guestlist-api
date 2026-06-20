package event

import (
	"context"
	"time"
)

type Event struct {
	ID          string    `json:"id"`
	HostID      string    `json:"hostId"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug,omitempty"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	Time        time.Time `json:"time"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateEventInput struct {
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	Time        time.Time `json:"time"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
}

type UpdateEventInput struct {
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	Time        time.Time `json:"time"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserEventStats struct {
	TotalEvents              int64   `json:"total_events"`
	OpenEvents               int64   `json:"open_events"`
	ClosedEvents             int64   `json:"closed_events"`
	CancelledEvents          int64   `json:"cancelled_events"`
	TotalGuestsInvited       int64   `json:"total_guests_invited"`
	TotalGuestsConfirmed     int64   `json:"total_guests_confirmed"`
	TotalGuestsCheckedIn     int64   `json:"total_guests_checked_in"`
	AttendanceRatePercentage float64 `json:"attendance_rate_percentage"`
}

type EventStatsResponse struct {
	EventID                  string    `json:"event_id"`
	Title                    string    `json:"title"`
	Slug                     string    `json:"slug"`
	EventStatus              string    `json:"event_status"`
	StartTime                time.Time `json:"start_time"`
	TotalGuestsInvited       int64     `json:"total_guests_invited"`
	TotalGuestsConfirmed     int64     `json:"total_guests_confirmed"`
	TotalGuestsCheckedIn     int64     `json:"total_guests_checked_in"`
	AttendanceRatePercentage float64   `json:"attendance_rate_percentage"`
}

type Service interface {
	CreateEvent(ctx context.Context, req CreateEventInput) (Event, error)
	UpdateEvent(ctx context.Context, ID string, req UpdateEventInput) (Event, error)
	DeleteEvent(ctx context.Context, ID string) error
	ListEvents(ctx context.Context) ([]Event, error)
	GetEventByID(ctx context.Context, ID string) (Event, error)

	GetUserEvents(ctx context.Context) ([]Event, error)
	GetUserEventStats(ctx context.Context) (UserEventStats, error)
	GetEventStats(ctx context.Context, eventID string) (EventStatsResponse, error)
}
