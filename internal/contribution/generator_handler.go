package contribution

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nati-darse/edirlink-backend/internal/auth"
)

func (h *Handler) GenerateByEdir(w http.ResponseWriter, r *http.Request) {
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

	created, err := h.repository.GenerateForEdir(r.Context(), edirID)
	if err != nil {
		log.Printf("generate contributions: %v", err)
		http.Error(w, "could not generate contributions", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]int{"created": created})
}
