package member

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nati-darse/edirlink-backend/internal/auth"
	"github.com/nati-darse/edirlink-backend/internal/edir"
	"github.com/nati-darse/edirlink-backend/internal/password"
)

type Handler struct {
	repository     *Repository
	edirRepository *edir.Repository
}

func NewHandler(
	repository *Repository,
	edirRepository *edir.Repository,
) *Handler {
	return &Handler{
		repository:     repository,
		edirRepository: edirRepository,
	}
}

type CreateRequest struct {
	FullName string  `json:"full_name"`
	Phone    string  `json:"phone"`
	Email    *string `json:"email"`
	Gender   *string `json:"gender"`
	Address  *string `json:"address"`
}

type AddExistingMemberRequest struct {
	MemberID string  `json:"member_id"`
	Gender   *string `json:"gender"`
	Address  *string `json:"address"`
}

func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {

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
			"only organizers can create members",
			http.StatusForbidden,
		)
		return
	}

	edirID := chi.URLParam(r, "edirID")

	if edirID == "" {
		http.Error(
			w,
			"edir id is required",
			http.StatusBadRequest,
		)
		return
	}

	edirRecord, err := h.edirRepository.FindByID(
		r.Context(),
		edirID,
	)

	if err != nil {
		http.Error(
			w,
			"edir not found",
			http.StatusNotFound,
		)
		return
	}

	if edirRecord.OrganizerID != claims.UserID {
		http.Error(
			w,
			"you do not have access to this edir",
			http.StatusForbidden,
		)
		return
	}

	if edirRecord.Status != "ACTIVE" {
		http.Error(
			w,
			"edir is not active",
			http.StatusForbidden,
		)
		return
	}

	var req CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if req.FullName == "" {
		http.Error(
			w,
			"full_name is required",
			http.StatusBadRequest,
		)
		return
	}

	if req.Phone == "" {
		http.Error(
			w,
			"phone is required",
			http.StatusBadRequest,
		)
		return
	}

	exists, err := h.repository.CheckPhoneInEdir(
		r.Context(),
		edirID,
		req.Phone,
	)

	if err != nil {
		http.Error(
			w,
			"failed to check phone",
			http.StatusInternalServerError,
		)
		return
	}

	if exists {
		http.Error(
			w,
			"member with this phone already exists in this edir",
			http.StatusConflict,
		)
		return
	}

	temporaryPassword := generateTemporaryPassword()

	// Hash password

	passwordHash, err := password.Hash(
		temporaryPassword,
	)

	if err != nil {
		http.Error(
			w,
			"failed to secure password",
			http.StatusInternalServerError,
		)
		return
	}

	// Create user

	member, err := h.repository.CreateUser(
		r.Context(),
		req.FullName,
		req.Phone,
		req.Email,
		passwordHash,
	)

	if err != nil {
		http.Error(
			w,
			"failed to create member",
			http.StatusInternalServerError,
		)
		return
	}

	// Create membership

	err = h.repository.CreateMembership(
		r.Context(),
		edirID,
		member.ID,
		req.Gender,
		req.Address,
	)

	if err != nil {
		http.Error(
			w,
			"failed to create membership",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"message": "member created successfully",
			"member": map[string]interface{}{
				"id":        member.ID,
				"full_name": member.FullName,
				"phone":     member.Phone,
				"email":     member.Email,
				"gender":    req.Gender,
				"address":   req.Address,
				"is_active": member.IsActive,
			},
			"temporary_password":   temporaryPassword,
			"must_change_password": true,
		},
	)
}

func generateTemporaryPassword() string {
	return fmt.Sprintf(
		"Edir@%d",
		100000+randomNumber(),
	)
}

func randomNumber() int {
	return 123456
}

func (h *Handler) List(
	w http.ResponseWriter,
	r *http.Request,
) {

	// Get authenticated user
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
			"only organizers can list members",
			http.StatusForbidden,
		)
		return
	}

	// Get Edir ID
	edirID := chi.URLParam(r, "edirID")

	if edirID == "" {
		http.Error(
			w,
			"edir id is required",
			http.StatusBadRequest,
		)
		return
	}
	edirRecord, err := h.edirRepository.FindByID(
		r.Context(),
		edirID,
	)

	if err != nil {
		http.Error(
			w,
			"edir not found",
			http.StatusNotFound,
		)
		return
	}

	if edirRecord.OrganizerID != claims.UserID {
		http.Error(
			w,
			"you do not have access to this edir",
			http.StatusForbidden,
		)
		return
	}
	if edirRecord.Status != "ACTIVE" {
		http.Error(
			w,
			"edir is not active",
			http.StatusForbidden,
		)
		return
	}

	members, err := h.repository.ListMembers(
		r.Context(),
		edirID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to list members",
			http.StatusInternalServerError,
		)
		return
	}
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"members": members,
		},
	)
}

func (h *Handler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
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
			"organizer access required",
			http.StatusForbidden,
		)
		return
	}

	edirID := chi.URLParam(r, "edirID")
	memberID := chi.URLParam(r, "memberID")

	edirRecord, err := h.edirRepository.FindByID(
		r.Context(),
		edirID,
	)

	if err != nil {
		http.Error(
			w,
			"edir not found",
			http.StatusNotFound,
		)
		return
	}

	if edirRecord.OrganizerID != claims.UserID {
		http.Error(
			w,
			"you do not have access to this edir",
			http.StatusForbidden,
		)
		return
	}

	if edirRecord.Status != "ACTIVE" {
		http.Error(
			w,
			"edir is not active",
			http.StatusForbidden,
		)
		return
	}

	member, err := h.repository.FindByIDInEdir(
		r.Context(),
		edirID,
		memberID,
	)

	if err != nil {
		http.Error(
			w,
			"member not found in this edir",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(member)
}
func (h *Handler) AddExisting(
	w http.ResponseWriter,
	r *http.Request,
) {
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
			"organizer access required",
			http.StatusForbidden,
		)
		return
	}

	edirID := chi.URLParam(r, "edirID")

	edirRecord, err := h.edirRepository.FindByID(
		r.Context(),
		edirID,
	)

	if err != nil {
		http.Error(w, "edir not found", http.StatusNotFound)
		return
	}

	if edirRecord.OrganizerID != claims.UserID {
		http.Error(
			w,
			"you do not have access to this edir",
			http.StatusForbidden,
		)
		return
	}

	if edirRecord.Status != "ACTIVE" {
		http.Error(
			w,
			"edir is not active",
			http.StatusForbidden,
		)
		return
	}

	var request AddExistingMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if request.MemberID == "" {
		http.Error(
			w,
			"member_id is required",
			http.StatusBadRequest,
		)
		return
	}

	err = h.repository.AddExistingMember(
		r.Context(),
		edirID,
		request.MemberID,
		request.Gender,
		request.Address,
	)

	if err != nil {
		http.Error(
			w,
			"failed to add member",
			http.StatusConflict,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "member added to edir successfully",
	})
}

func (h *Handler) Remove(
	w http.ResponseWriter,
	r *http.Request,
) {
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
			"organizer access required",
			http.StatusForbidden,
		)
		return
	}

	edirID := chi.URLParam(r, "edirID")
	memberID := chi.URLParam(r, "memberID")

	edirRecord, err := h.edirRepository.FindByID(
		r.Context(),
		edirID,
	)

	if err != nil {
		http.Error(
			w,
			"edir not found",
			http.StatusNotFound,
		)
		return
	}

	if edirRecord.OrganizerID != claims.UserID {
		http.Error(
			w,
			"you do not have access to this edir",
			http.StatusForbidden,
		)
		return
	}

	if edirRecord.Status != "ACTIVE" {
		http.Error(
			w,
			"edir is not active",
			http.StatusForbidden,
		)
		return
	}

	err = h.repository.RemoveFromEdir(
		r.Context(),
		edirID,
		memberID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to remove member",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]string{
			"message": "member removed from edir successfully",
		},
	)
}

func (h *Handler) ListMyEdirs(
	w http.ResponseWriter,
	r *http.Request,
) {

	claims, ok := auth.GetClaims(r.Context())

	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	if claims.Role != "MEMBER" {
		http.Error(
			w,
			"member access required",
			http.StatusForbidden,
		)
		return
	}

	edirs, err := h.repository.ListMyEdirs(
		r.Context(),
		claims.UserID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to load edirs",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"edirs": edirs,
		},
	)
}
