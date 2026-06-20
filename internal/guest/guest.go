package guest

import (
	"context"

	"github.com/graciar/guestlist-api/internal/event"
)

type Guest struct {
	EventId     string `json:"eventId"`
	GuestId     string `json:"guestId"`
	GuestName   string `json:"guestName"`
	GuestEmail  string `json:"guestEmail,omitempty"`
	GuestPhone  string `json:"guestPhone,omitempty"`
	RsvpStatus  string `json:"rsvpStatus,omitempty"`
	RsvpToken   string `json:"rsvpToken,omitempty"`
	TicketCode  string `json:"ticketCode,omitempty"`
	IsCheckedIn bool   `json:"isCheckedIn"`
}

type CreateGuestInput struct {
	EventId    string `json:"eventId"`
	GuestName  string `json:"guestName"`
	GuestEmail string `json:"guestEmail,omitempty"`
	GuestPhone string `json:"guestPhone,omitempty"`
}

type UpdateGuestInput struct {
	GuestName   string `json:"guestName"`
	GuestEmail  string `json:"guestEmail,omitempty"`
	GuestPhone  string `json:"guestPhone,omitempty"`
	RsvpStatus  string `json:"rsvpStatus,omitempty"`
	IsCheckedIn bool   `json:"isCheckedIn,omitempty"`
}

type GuestTicket struct {
	EventId   string `json:"eventId"`
	GuestId   string `json:"guestId"`
	Ticket    string `json:"ticket"`
	EventName string `json:"eventName"`
	GuestName string `json:"guestName"`
}

type EventGuests struct {
	Event  event.Event `json:"event"`
	Guests []Guest     `json:"guests"`
}

type Service interface {
	CreateGuest(ctx context.Context, req CreateGuestInput) (Guest, error)
	ListGuests(ctx context.Context) ([]Guest, error)
	UpdateGuest(ctx context.Context, ID string, req UpdateGuestInput) (Guest, error)
	DeleteGuest(ctx context.Context, ID string) error
	GetGuestByID(ctx context.Context, ID string, eventSlug string) (Guest, error)

	GetGuestsForEvent(ctx context.Context, eventID string) (EventGuests, error)
	GetGuestTicket(ctx context.Context, ID string, eventSlug string) (GuestTicket, error)
	HandleRSVPResponse(ctx context.Context, ID string, token string, status string) error
	HandleCheckIn(ctx context.Context, ID string, sig string) error
}
