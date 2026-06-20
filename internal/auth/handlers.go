package auth

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/graciar/guestlist-api/internal/env"
	"github.com/graciar/guestlist-api/internal/json"

	"golang.org/x/oauth2"
)

type handler struct {
	service Service
	config  *oauth2.Config
}

func NewHandler(service Service, config *oauth2.Config) *handler {
	return &handler{
		service: service,
		config:  config,
	}
}

// Helper utility to determine environment setup dynamically
func isProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}

func (h *handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var tempUser SignUpInput
	if err := json.Read(r, &tempUser); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdUser, err := h.service.SignUp(r.Context(), tempUser)
	if err != nil {
		log.Println(err)

		if err == ErrUserAlreadyExists {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusCreated, createdUser)
}

func (h *handler) SignIn(w http.ResponseWriter, r *http.Request) {
	clientPlatform := r.Header.Get("X-Client-Platform")
	log.Println("Client Platform : ", clientPlatform)

	var loginInput LoginInput
	if err := json.Read(r, &loginInput); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, refreshToken, err := h.service.SignIn(r.Context(), loginInput)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch clientPlatform {
	case "web":
		// 🌟 Dynamic configurations to prevent HTTP local development drops
		isProd := isProduction()

		sameSiteMode := http.SameSiteLaxMode
		if isProd {
			sameSiteMode = http.SameSiteStrictMode
		}

		cookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   isProd,       // 🌟 set to false locally over http://127.0.0.1
			SameSite: sameSiteMode, // 🌟 Lax locally so cross-origin ports don't discard it
		}
		http.SetCookie(w, cookie)

		json.Write(w, http.StatusOK, map[string]string{"token": token})
	case "ios", "android":
		mobileResponse := map[string]string{
			"token":         token,
			"refresh_token": refreshToken,
		}
		json.Write(w, http.StatusOK, mobileResponse)
	default:
		log.Printf("Rejected login: missing or invalid platform header '%s'", clientPlatform)
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := map[string]string{"error": "Missing or invalid X-Client-Platform header"}
		json.Write(w, http.StatusBadRequest, errorResponse)
		return
	}
}

func (h *handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req RequestPasswordResetInput
	if err := json.Read(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.RequestPasswordReset(r.Context(), req); err != nil {
		log.Println(err)
		http.Error(w, "Failed to process password reset request", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, map[string]string{"message": "A reset link has been sent."})
}

func (h *handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordInput
	if err := json.Read(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokenHash := r.URL.Query().Get("token")
	if tokenHash == "" {
		http.Error(w, "token is missing", http.StatusBadRequest)
		return
	}

	if err := h.service.ResetPassword(r.Context(), req, tokenHash); err != nil {
		http.Error(w, "Failed to process password reset request", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, map[string]string{"message": "Password successfully reset."})
}

func (h *handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	clientPlatform := r.Header.Get("X-Client-Platform")
	var refreshToken string

	switch clientPlatform {
	case "web":
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			log.Println("Refresh cookie missing:", err)
			http.Error(w, "Missing refresh token cookie", http.StatusUnauthorized)
			return
		}
		refreshToken = cookie.Value

	case "ios", "android":
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		refreshToken = strings.TrimPrefix(authHeader, "Bearer ")

	default:
		http.Error(w, "Missing or invalid X-Client-Platform header", http.StatusBadRequest)
		return
	}

	newToken, newRefreshToken, err := h.service.RotateTokens(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	if clientPlatform == "web" {
		isProd := isProduction()

		sameSiteMode := http.SameSiteLaxMode
		if isProd {
			sameSiteMode = http.SameSiteStrictMode
		}

		cookie := &http.Cookie{
			Name:  "refresh_token",
			Value: newRefreshToken,
			Path:  "/",
			// MaxAge:   7 * 24 * 60 * 60,
			HttpOnly: true,
			Secure:   isProd, // 🌟 Match local sandbox constraints
			SameSite: sameSiteMode,
		}
		http.SetCookie(w, cookie)

		json.Write(w, http.StatusOK, map[string]string{"token": newToken})
	} else {
		json.Write(w, http.StatusOK, map[string]string{
			"token":         newToken,
			"refresh_token": newRefreshToken,
		})
	}
}

func (h *handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Safety check to intercept nil pointers
	if h.config == nil {
		http.Error(w, "OAuth configuration was not properly initialized on the server", http.StatusInternalServerError)
		return
	}

	url := h.config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	t, err := h.config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := h.config.Client(r.Context(), t)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")

	var debugBuffer bytes.Buffer
	bodyBytes := io.TeeReader(resp.Body, &debugBuffer)

	// log.Println("Google payload response:", bodyBytes)

	var googleUser GoogleUserInfo
	if err := json.ReadFrom(bodyBytes, &googleUser); err != nil {
		log.Println("Failed to parse Google user payload:", err)
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	if !googleUser.VerifiedEmail {
		http.Error(w, "Email not verified", http.StatusUnauthorized)
		return
	}

	authResponse, err := h.service.HandleGoogleUser(r.Context(), googleUser)
	if err != nil {
		log.Println("OAuth processing error:", err)
		http.Error(w, "Internal authentication mapping failed", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    authResponse.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)

	client_url := env.GetString("CLIENT_DOMAIN", "")
	frontendRedirectURL := client_url + "/auth/callback?token=" + authResponse.AccessToken

	http.Redirect(w, r, frontendRedirectURL, http.StatusTemporaryRedirect)
}
