package guest

import (
	"log"
	"net/http"

	"github.com/graciar/guestlist-api/internal/json"

	"github.com/go-chi/chi/v5"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateGuest(w http.ResponseWriter, r *http.Request) {
	var tempGuest CreateGuestInput
	if err := json.Read(r, &tempGuest); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdGuest, err := h.service.CreateGuest(r.Context(), tempGuest)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusCreated, createdGuest)
}

func (h *handler) UpdateGuest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}

	var tempGuest UpdateGuestInput
	if err := json.Read(r, &tempGuest); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedGuest, err := h.service.UpdateGuest(r.Context(), id, tempGuest)
	if err != nil {
		log.Println(err)
		log.Println("abctest")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusCreated, updatedGuest)
}

func (h *handler) ListGuests(w http.ResponseWriter, r *http.Request) {
	guests, err := h.service.ListGuests(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, guests)
}

func (h *handler) GetGuestByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing guest ID", http.StatusBadRequest)
		return
	}

	eventSlug := chi.URLParam(r, "eventSlug")
	if eventSlug == "" {
		http.Error(w, "Missing event slug", http.StatusBadRequest)
		return
	}

	guest, err := h.service.GetGuestByID(r.Context(), id, eventSlug)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, guest)
}

func (h *handler) DeleteGuest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}
	err := h.service.DeleteGuest(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, "guest deleted successfully")
}

func (h *handler) GetGuestsForEvent(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "eventID")

	guests, err := h.service.GetGuestsForEvent(r.Context(), eventIDStr)
	if err != nil {
		log.Println(err) // Good practice to log internal errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, guests)
}

func (h *handler) GetGuestTicket(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing guest ID", http.StatusBadRequest)
		return
	}

	eventSlug := chi.URLParam(r, "eventSlug")
	if eventSlug == "" {
		http.Error(w, "Missing event slug", http.StatusBadRequest)
		return
	}

	guest, err := h.service.GetGuestTicket(r.Context(), id, eventSlug)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, guest)
}

func (h *handler) HandleRSVPAttend(w http.ResponseWriter, r *http.Request) {
	h.executeRSVPChange(w, r, "attending")
}

func (h *handler) HandleRSVPDecline(w http.ResponseWriter, r *http.Request) {
	h.executeRSVPChange(w, r, "declined")
}

// Private helper to keep code DRY (Don't Repeat Yourself)
func (h *handler) executeRSVPChange(w http.ResponseWriter, r *http.Request, status string) {
	token := r.URL.Query().Get("rsvp_token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing guest ID", http.StatusBadRequest)
		return
	}

	err := h.service.HandleRSVPResponse(r.Context(), id, token, status)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, map[string]string{"message": "guest RSVP updated successfully"})
}

func (h *handler) HandleCheckIn(w http.ResponseWriter, r *http.Request) {
	sig := r.URL.Query().Get("sig")
	if sig == "" {
		http.Error(w, "Missing signature", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing guest ID", http.StatusBadRequest)
		return
	}
	err := h.service.HandleCheckIn(r.Context(), id, sig)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, "guest checked in successfully")
}
