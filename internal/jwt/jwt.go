package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID             string `json:"user_id"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"must_change_password"`
	jwt.RegisteredClaims
}

func Generate(
	userID string,
	role string,
	mustChangePassword bool,
	secret string,
	duration time.Duration,
) (string, error) {

	now := time.Now()

	claims := Claims{
		UserID:             userID,
		Role:               role,
		MustChangePassword: mustChangePassword,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(secret))
}

func Validate(
	tokenString string,
	secret string,
) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
