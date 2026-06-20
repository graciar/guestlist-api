package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/graciar/guestlist-api/internal/adapters/email"
	repo "github.com/graciar/guestlist-api/internal/adapters/postgresql/sqlc"
	"github.com/graciar/guestlist-api/internal/env"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCreds      = errors.New("invalid credentials")
)

type svc struct {
	repo *repo.Queries
	db   *pgxpool.Pool
}

func NewService(repo *repo.Queries, db *pgxpool.Pool) Service {
	return &svc{repo: repo, db: db}
}

func (s *svc) SignUp(ctx context.Context, req SignUpInput) (repo.User, error) {
	_, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err == nil {
		return repo.User{}, ErrUserAlreadyExists
	}

	if req.Password == "" {
		return repo.User{}, errors.New("password is required")
	}

	// 1. Hash the password before saving it to the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return repo.User{}, err
	}

	// 2. Overwrite the plain-text password with the secure hash
	params := repo.CreateUserParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     repo.UserRole("user"),
	}

	// 3. Persist the user data via repository
	return s.repo.CreateUser(ctx, params)
}

func (s *svc) SignIn(ctx context.Context, req LoginInput) (string, string, error) {
	// 1. Retrieve the user records by email from the DB
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		// Return a generic error to prevent email enumeration attacks
		return "", "", errors.New("account does not exist")
	}

	if user.Password == "" {
		return "", "", errors.New("Login via Google")
	}

	// 2. Compare the stored database hash with the incoming login password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", "", errors.New("invalid password")
	}

	// 3. Call your utils package to generate Access and Refresh tokens
	accessToken, refreshToken, err := GenerateJWTToken(user.Email, user.Name, string(user.Role), user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateResetToken creates a raw token for the email and a hashed token for the DB
func GenerateResetToken() (rawToken string, hashedToken string, err error) {
	// 1. Generate 32 secure random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// 2. Convert raw bytes to a hex string (this goes in the email link)
	rawToken = hex.EncodeToString(bytes)

	// 3. Hash the raw token using SHA-256
	hash := sha256.Sum256([]byte(rawToken))

	// 4. Convert the hash to a hex string (this goes in the DB)
	hashedToken = hex.EncodeToString(hash[:])

	return rawToken, hashedToken, nil
}

func (s *svc) RequestPasswordReset(ctx context.Context, req RequestPasswordResetInput) error {
	// 1. Find the user
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	// 2. Generate 32 secure random bytes (increased from 16 for better security)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token bytes: %w", err)
	}
	// This is the raw token you will send in the email
	rawToken := hex.EncodeToString(tokenBytes)

	// 3. Hash the raw token using SHA-256 for the database
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	// 4. Save the HASHED token to the database
	_, err = s.repo.CreatePasswordReset(ctx, repo.CreatePasswordResetParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(1 * time.Hour), // 24 hours is too long for a password reset! 15m to 1h is standard.
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create password reset record: %w", err)
	}

	client_url := env.GetString("CLIENT_DOMAIN", "")
	if client_url == "" {
		return fmt.Errorf("client URL is not set in environment variables")
	}
	url := client_url + "/auth/reset-password?token=" + rawToken
	userData := struct {
		Name  string
		Token string
		url   string
	}{
		Name:  user.Name,
		Token: rawToken,
		url:   url,
	}

	htmlBody, err := email.LoadTemplates("internal/templates/email/password_reset.html", userData)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	err = email.SendEmail(user.Email, user.Name, "Password Reset", url, htmlBody)
	if err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}

func (s *svc) ResetPassword(ctx context.Context, req ResetPasswordInput, rawToken string) error {
	// 1. Hash the incoming raw token from the user's URL request
	hash := sha256.Sum256([]byte(rawToken))
	incomingHash := hex.EncodeToString(hash[:])

	// 2. Look up the token record by the HASH
	token, err := s.repo.GetPasswordResetByTokenHash(ctx, incomingHash)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	// 3. Hash the new password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	txRepo := s.repo.WithTx(tx)

	// 5. Invalidate the token so it can never be used again
	_, err = txRepo.InvalidatePasswordReset(ctx, token.ID)
	if err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	// 4. Update the user's password
	err = txRepo.UpdatePassword(ctx, repo.UpdatePasswordParams{
		ID:       token.UserID,
		Password: string(hashedPassword),
	})
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 6. Commit everything permanently to the database
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *svc) RotateTokens(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := ValidateToken(refreshToken, "refresh")
	if err != nil {
		return "", "", err
	}

	return GenerateJWTToken(claims.Email, claims.Name, claims.Role, claims.ID)
}

func (s *svc) HandleGoogleUser(ctx context.Context, googleUser GoogleUserInfo) (GoogleCallbackResponse, error) {
	// 1. Check if the user already exists in your database by email
	user, err := s.repo.FindUserByEmail(ctx, googleUser.Email)

	if err != nil {
		// If the error is genuinely "User Not Found" (or pgx.ErrNoRows), create a new user
		if errors.Is(err, ErrUserNotFound) || err.Error() == "no rows in result set" {

			createParams := repo.CreateUserParams{
				Name:     googleUser.Name,
				Email:    googleUser.Email,
				Role:     "user",
				Password: "",
			}

			user, err = s.repo.CreateUser(ctx, createParams)
			if err != nil {
				return GoogleCallbackResponse{}, fmt.Errorf("failed to create user from google profile: %w", err)
			}
		} else {
			// A real database connection error occurred
			return GoogleCallbackResponse{}, fmt.Errorf("database query failed: %w", err)
		}
	}

	// 2. The user now exists (either fetched or freshly created)!
	// Now, generate YOUR own application standard JWT tokens using their user ID.
	accessToken, refreshToken, err := GenerateJWTToken(user.Email, user.Name, string(user.Role), user.ID)
	if err != nil {
		return GoogleCallbackResponse{}, fmt.Errorf("failed to generate app tokens: %w", err)
	}

	// 3. Construct and return your GoogleCallbackResponse
	return GoogleCallbackResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		ExpiresIn:             3600,         // e.g., 1 hour in seconds
		RefreshTokenExpiresIn: 604800,       // e.g., 7 days in seconds
		Scope:                 "read write", // Your application scope constraints
	}, nil
}
