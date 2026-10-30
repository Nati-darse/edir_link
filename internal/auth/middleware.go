package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/nati-darse/edirlink-backend/internal/jwt"
)

type contextKey string

const claimsKey contextKey = "claims"

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(
					w,
					"authorization required",
					http.StatusUnauthorized,
				)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)

			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(
					w,
					"invalid authorization header",
					http.StatusUnauthorized,
				)
				return
			}

			claims, err := jwt.Validate(
				parts[1],
				secret,
			)

			if err != nil {
				http.Error(
					w,
					"invalid or expired token",
					http.StatusUnauthorized,
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				claimsKey,
				claims,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

func GetClaims(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*jwt.Claims)
	return claims, ok
}
