package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/graciar/guestlist-api/internal/env"

	"github.com/golang-jwt/jwt/v5"
)

type SignedDetails struct {
	Email string
	Name  string
	Role  string
	ID    string
	jwt.RegisteredClaims
}

var ACCESS_SECRET_KEY = env.GetString("JWT_ACCESS_SECRET", "")
var REFRESH_SECRET_KEY = env.GetString("JWT_REFRESH_SECRET", "")

// FIXED: Updated return signature to (string, string, error) to return both tokens
func GenerateJWTToken(email string, name string, role string, id string) (string, string, error) {
	claims := &SignedDetails{
		Email: email,
		Name:  name,
		Role:  role,
		ID:    id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "auth-service",
		},
	}

	refreshClaims := &SignedDetails{
		Email: email,
		Name:  name,
		Role:  role,
		ID:    id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)), // 7 days
			Issuer:    "auth-service",
		},
	}

	// Generate access token
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(ACCESS_SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(REFRESH_SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

// FIXED: Cleaned up error string handling
func ValidateToken(signedToken string, tokenType string) (claims *SignedDetails, err error) {
	token, parseErr := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			// Ensure token signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			if tokenType == "access" {
				return []byte(ACCESS_SECRET_KEY), nil
			} else {
				return []byte(REFRESH_SECRET_KEY), nil
			}
		},
	)

	if parseErr != nil {
		// Detect specifically if the error is due to expiration
		if errors.Is(parseErr, jwt.ErrTokenExpired) {
			return nil, errors.New("token is expired")
		}
		return nil, parseErr
	}

	// Safely assert the claims type
	claims, ok := token.Claims.(*SignedDetails)
	if !ok || !token.Valid {
		return nil, errors.New("the token is invalid")
	}

	return claims, nil
}
