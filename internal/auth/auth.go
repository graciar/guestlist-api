package auth

import (
	"context"

	repo "github.com/graciar/guestlist-api/internal/adapters/postgresql/sqlc"
)

type SignUpInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RequestPasswordResetInput struct {
	Email string `json:"email"`
}

type ResetPasswordInput struct {
	NewPassword string `json:"newPassword"`
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	// Picture       string `json:"picture"`
}

type GoogleCallbackResponse struct {
	AccessToken           string `json:"accessToken"`
	RefreshToken          string `json:"refreshToken"`
	Token                 string `json:"token"`
	ExpiresIn             int    `json:"expiresIn"`
	RefreshTokenExpiresIn int    `json:"refreshTokenExpiresIn"`
	Scope                 string `json:"scope"`
	TokenType             string `json:"tokenType"`
}

type Service interface {
	SignUp(ctx context.Context, req SignUpInput) (repo.User, error)
	SignIn(ctx context.Context, req LoginInput) (string, string, error)
	HandleGoogleUser(ctx context.Context, googleUser GoogleUserInfo) (GoogleCallbackResponse, error)

	RequestPasswordReset(ctx context.Context, req RequestPasswordResetInput) error
	ResetPassword(ctx context.Context, req ResetPasswordInput, tokenHash string) error
	RotateTokens(ctx context.Context, refreshToken string) (string, string, error)
}
