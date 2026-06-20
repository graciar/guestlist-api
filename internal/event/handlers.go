package event

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
	return &handler{service: service}
}

func (h *handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var tempEvent CreateEventInput
	if err := json.Read(r, &tempEvent); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdEvent, err := h.service.CreateEvent(r.Context(), tempEvent)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusCreated, createdEvent)
}

func (h *handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}

	var tempEvent UpdateEventInput
	if err := json.Read(r, &tempEvent); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedEvent, err := h.service.UpdateEvent(r.Context(), idStr, tempEvent)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusCreated, updatedEvent)
}

func (h *handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}
	err := h.service.DeleteEvent(r.Context(), idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusAccepted, nil)
}

func (h *handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	listedEvents, err := h.service.ListEvents(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, listedEvents)
}

func (h *handler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}
	event, err := h.service.GetEventByID(r.Context(), idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, event)
}

func (h *handler) GetUserEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.GetUserEvents(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, events)
}

func (h *handler) GetUserEventStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetUserEventStats(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, stats)
}

func (h *handler) GetEventStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}
	stats, err := h.service.GetEventStats(r.Context(), idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusOK, stats)

}
