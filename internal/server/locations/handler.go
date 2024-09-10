package locations

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexPavlikov/auth-service/internal/models"
	"github.com/alexPavlikov/auth-service/internal/service"
)

type Handler struct {
	Service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	var us models.UserPayLoad

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := dec.Decode(&us); err != nil {
		slog.Error("failed to decode body", "error", err)
		http.Error(w, "unauthorized", http.StatusForbidden)
	}

	ip := r.Header.Get("X-Forwarded-For")

	var user = models.User{
		UUID: us.UUID,
		IP:   ip,
	}

	token, refresh, err := h.Service.Auth(context.Background(), user)
	if err != nil {
		slog.Error("failed to get tokens", "error", err)
		http.Error(w, "unauthorized", http.StatusForbidden)
	}

	cookieAccess := http.Cookie{
		Name:     "access",
		Value:    token,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(8760 * time.Hour),
		MaxAge:   31556952,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookieAccess)

	cookieRefresh := http.Cookie{
		Name:     "refresh",
		Value:    refresh,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(8760 * time.Hour),
		MaxAge:   31556952,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookieRefresh)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Forwarded-For")

	accessCookie, err := r.Cookie("access")
	if err != nil {
		slog.Error("failed to get cookie access", "error", err)
		http.Error(w, "unauthorized", http.StatusForbidden)
	}

	refreshCookie, err := r.Cookie("refresh")
	if err != nil {
		slog.Error("failed to get cookie refresh", "error", err)
		http.Error(w, "unauthorized", http.StatusForbidden)
	}

	accessToken, err := h.Service.RefreshUserAuthToken(ip, accessCookie.Value, refreshCookie.Value)
	if err != nil {
		slog.Error("failed to confirm user auth", "error", err)
		http.Error(w, "unauthorized", http.StatusForbidden)
	}

	cookieAccess := http.Cookie{
		Name:     "access",
		Value:    accessToken,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(8760 * time.Hour),
		MaxAge:   31556952,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookieAccess)
}
