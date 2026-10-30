package contribution

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nati-darse/edirlink-backend/internal/auth"
)

type Handler struct {
	repository *Repository
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if claims.Role != "MEMBER" {
		http.Error(w, "members only", http.StatusForbidden)
		return
	}

	items, err := h.repository.ListByMember(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("list member contributions: %v", err)
		http.Error(w, "could not retrieve contributions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, items)
}

func (h *Handler) ListByEdir(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if claims.Role != "ORGANIZER" {
		http.Error(w, "organizers only", http.StatusForbidden)
		return
	}

	edirID := chi.URLParam(r, "edirID")
	owns, err := h.repository.OrganizerOwnsEdir(r.Context(), edirID, claims.UserID)
	if err != nil {
		log.Printf("check Edir ownership: %v", err)
		http.Error(w, "could not verify Edir ownership", http.StatusInternalServerError)
		return
	}
	if !owns {
		http.Error(w, "Edir not found or access denied", http.StatusForbidden)
		return
	}

	items, err := h.repository.ListByEdir(r.Context(), edirID)
	if err != nil {
		log.Printf("list Edir contributions: %v", err)
		http.Error(w, "could not retrieve contributions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, items)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}
