package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nati-darse/edirlink-backend/internal/jwt"
	"github.com/nati-darse/edirlink-backend/internal/password"
	"github.com/nati-darse/edirlink-backend/internal/user"
)

type Handler struct {
	users     *user.Repository
	jwtSecret string
}

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type MemberLoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (h *Handler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if request.FullName == "" ||
		request.Email == "" ||
		request.Password == "" {

		http.Error(
			w,
			"full_name, email and password are required",
			http.StatusBadRequest,
		)
		return
	}

	ctx := r.Context()

	_, err := h.users.FindByEmail(
		ctx,
		request.Email,
	)

	if err == nil {
		http.Error(
			w,
			"email already registered",
			http.StatusConflict,
		)
		return
	}

	passwordHash, err := password.Hash(
		request.Password,
	)

	if err != nil {
		http.Error(
			w,
			"failed to hash password",
			http.StatusInternalServerError,
		)
		return
	}

	newUser, err := h.users.Create(
		ctx,
		request.FullName,
		request.Email,
		passwordHash,
		"ORGANIZER",
		false,
	)

	if err != nil {
		http.Error(
			w,
			"failed to create organizer",
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

		"id":        newUser.ID,
		"full_name": newUser.FullName,
		"email":     newUser.Email,
		"role":      newUser.Role,
		"is_active": newUser.IsActive,
	})
}
func NewHandler(
	users *user.Repository,
	jwtSecret string,
) *Handler {
	return &Handler{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *Handler) ChangePassword(
	w http.ResponseWriter,
	r *http.Request,
) {

	// 1. Get authenticated user
	claims, ok := GetClaims(r.Context())
	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	// 2. Parse request
	var req ChangePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	// 3. Validate passwords
	if req.CurrentPassword == "" {
		http.Error(
			w,
			"current_password is required",
			http.StatusBadRequest,
		)
		return
	}

	if req.NewPassword == "" {
		http.Error(
			w,
			"new_password is required",
			http.StatusBadRequest,
		)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(
			w,
			"new password must be at least 8 characters",
			http.StatusBadRequest,
		)
		return
	}

	// 4. Find current user
	user, err := h.users.FindByID(
		r.Context(),
		claims.UserID,
	)

	if err != nil {
		http.Error(
			w,
			"user not found",
			http.StatusNotFound,
		)
		return
	}

	// 5. Verify current password
	if err := password.Compare(
		user.PasswordHash,
		req.CurrentPassword,
	); err != nil {
		http.Error(
			w,
			"current password is incorrect",
			http.StatusUnauthorized,
		)
		return
	}

	// 6. Hash new password
	newPasswordHash, err := password.Hash(
		req.NewPassword,
	)

	if err != nil {
		http.Error(
			w,
			"failed to secure new password",
			http.StatusInternalServerError,
		)
		return
	}

	// 7. Update password
	if err := h.users.ChangePassword(
		r.Context(),
		claims.UserID,
		newPasswordHash,
	); err != nil {
		http.Error(
			w,
			"failed to change password",
			http.StatusInternalServerError,
		)
		return
	}

	// 8. Response
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"message":              "password changed successfully",
			"must_change_password": false,
		},
	)
}

func (h *Handler) MemberLogin(
	w http.ResponseWriter,
	r *http.Request,
) {

	var req MemberLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
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

	if req.Password == "" {
		http.Error(
			w,
			"password is required",
			http.StatusBadRequest,
		)
		return
	}

	user, err := h.users.FindMemberByPhone(
		r.Context(),
		req.Phone,
	)

	if err != nil {
		http.Error(
			w,
			"invalid phone or password",
			http.StatusUnauthorized,
		)
		return
	}

	if !user.IsActive {
		http.Error(
			w,
			"account is inactive",
			http.StatusForbidden,
		)
		return
	}

	if err := password.Compare(
		user.PasswordHash,
		req.Password,
	); err != nil {
		http.Error(
			w,
			"invalid phone or password",
			http.StatusUnauthorized,
		)
		return
	}

	token, err := jwt.Generate(
		user.ID,
		user.Role,
		user.MustChangePassword,
		h.jwtSecret,
		15*time.Minute,
	)

	if err != nil {
		http.Error(
			w,
			"failed to generate token",
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
			"access_token":         token,
			"must_change_password": user.MustChangePassword,
			"user": map[string]interface{}{
				"id":        user.ID,
				"full_name": user.FullName,
				"phone":     user.Phone,
				"email":     user.Email,
				"role":      user.Role,
			},
		},
	)
}

func (h *Handler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.Email == "" || request.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	foundUser, err := h.users.FindByEmail(ctx, request.Email)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	if !foundUser.IsActive {
		http.Error(w, "user account is inactive", http.StatusForbidden)
		return
	}

	if err := password.Compare(
		foundUser.PasswordHash,
		request.Password,
	); err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	accessToken, err := jwt.Generate(
		foundUser.ID,
		foundUser.Role,
		foundUser.MustChangePassword,
		h.jwtSecret,
		15*time.Minute,
	)

	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":         accessToken,
		"must_change_password": foundUser.MustChangePassword,
		"user": map[string]interface{}{
			"id":        foundUser.ID,
			"full_name": foundUser.FullName,
			"email":     foundUser.Email,
			"role":      foundUser.Role,
		},
	})
}
