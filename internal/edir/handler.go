package edir

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nati-darse/edirlink-backend/internal/auth"
)

type Handler struct {
	edirs *Repository
}

func NewHandler(edirs *Repository) *Handler {
	return &Handler{
		edirs: edirs,
	}
}

type CreateRequest struct {
	Name                      string  `json:"name"`
	MonthlyContributionAmount float64 `json:"monthly_contribution_amount"`
	DueEthiopianDay           int     `json:"due_ethiopian_day"`
	SupportAmount             float64 `json:"support_amount"`
	PenaltyPercentage         float64 `json:"penalty_percentage"`
}

func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if request.Name == "" {
		http.Error(
			w,
			"name is required",
			http.StatusBadRequest,
		)
		return
	}

	if request.MonthlyContributionAmount < 0 {
		http.Error(
			w,
			"monthly contribution amount cannot be negative",
			http.StatusBadRequest,
		)
		return
	}

	if request.DueEthiopianDay < 1 || request.DueEthiopianDay > 30 {
		http.Error(
			w,
			"due Ethiopian day must be between 1 and 30",
			http.StatusBadRequest,
		)
		return
	}

	if request.SupportAmount < 0 {
		http.Error(
			w,
			"support amount cannot be negative",
			http.StatusBadRequest,
		)
		return
	}

	if request.PenaltyPercentage < 0 {
		http.Error(
			w,
			"penalty percentage cannot be negative",
			http.StatusBadRequest,
		)
		return
	}

	claims, ok := auth.GetClaims(r.Context())

	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	if claims.Role != "ORGANIZER" {
		http.Error(
			w,
			"only organizers can create an edir",
			http.StatusForbidden,
		)
		return
	}

	edir, err := h.edirs.Create(
		r.Context(),
		request.Name,
		claims.UserID,
		request.MonthlyContributionAmount,
		request.DueEthiopianDay,
		request.SupportAmount,
		request.PenaltyPercentage,
	)

	if err != nil {
		http.Error(
			w,
			"failed to create edir",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                          edir.ID,
		"name":                        edir.Name,
		"organizer_id":                edir.OrganizerID,
		"monthly_contribution_amount": edir.MonthlyContributionAmount,
		"due_ethiopian_day":           edir.DueEthiopianDay,
		"support_amount":              edir.SupportAmount,
		"penalty_percentage":          edir.PenaltyPercentage,
		"status":                      edir.Status,
	})
}

func (h *Handler) requireSuperAdmin(
	w http.ResponseWriter,
	r *http.Request,
) bool {

	claims, ok := auth.GetClaims(r.Context())

	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return false
	}

	if claims.Role != "SUPER_ADMIN" {
		http.Error(
			w,
			"super admin access required",
			http.StatusForbidden,
		)
		return false
	}

	return true
}

func (h *Handler) ListPending(
	w http.ResponseWriter,
	r *http.Request,
) {

	if !h.requireSuperAdmin(w, r) {
		return
	}

	edirs, err := h.edirs.FindPending(
		r.Context(),
	)

	if err != nil {
		http.Error(
			w,
			"failed to fetch pending edirs",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(edirs)
}

func (h *Handler) Approve(
	w http.ResponseWriter,
	r *http.Request,
) {

	if !h.requireSuperAdmin(w, r) {
		return
	}

	edirID := chi.URLParam(r, "id")

	if edirID == "" {
		http.Error(
			w,
			"edir id is required",
			http.StatusBadRequest,
		)
		return
	}

	err := h.edirs.Approve(
		r.Context(),
		edirID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to approve edir",
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "edir approved successfully",
	})
}

func (h *Handler) Reject(
	w http.ResponseWriter,
	r *http.Request,
) {

	if !h.requireSuperAdmin(w, r) {
		return
	}

	edirID := chi.URLParam(r, "id")

	if edirID == "" {
		http.Error(
			w,
			"edir id is required",
			http.StatusBadRequest,
		)
		return
	}

	var request struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if request.Reason == "" {
		http.Error(
			w,
			"rejection reason is required",
			http.StatusBadRequest,
		)
		return
	}

	err := h.edirs.Reject(
		r.Context(),
		edirID,
		request.Reason,
	)

	if err != nil {
		http.Error(
			w,
			"failed to reject edir",
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "edir rejected successfully",
	})
}
