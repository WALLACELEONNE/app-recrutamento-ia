package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/username/app-recrutamento-ia/internal/usecase"
	"github.com/username/app-recrutamento-ia/templates/pages"
)

type AuthHandler struct {
	authUC *usecase.AuthUseCase
}

func NewAuthHandler(authUC *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// ServeLoginPage renders the login UI.
func (h *AuthHandler) ServeLoginPage(w http.ResponseWriter, r *http.Request) {
	// Se já estiver logado, redireciona para o dashboard
	cookie, err := r.Cookie("jwt_token")
	if err == nil && cookie.Value != "" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := pages.LoginPage()
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering login page", http.StatusInternalServerError)
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HandleLogin processes a login attempt and sets the JWT in an HttpOnly cookie.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Handles both JSON and Form submission
	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")
	}

	token, err := h.authUC.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			http.Error(w, "E-mail ou senha inválidos", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	// Set HttpOnly Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Responder JSON de sucesso para clientes Ajax (Alpine.js)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "redirect": "/dashboard"}`))
}

// HandleLogout clears the JWT cookie.
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	http.Redirect(w, r, "/login", http.StatusFound)
}
