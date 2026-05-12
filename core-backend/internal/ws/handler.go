package ws

import (
	"net/http"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub          *Hub
	tokenService token.TokenService
	authService  service.AuthService
}

func NewHandler(hub *Hub, tokenService token.TokenService, authService service.AuthService) *Handler {
	return &Handler{
		hub:          hub,
		tokenService: tokenService,
		authService:  authService,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenString := h.extractToken(r)
	if tokenString == "" {
		http.Error(w, "missing or invalid authorization", http.StatusUnauthorized)
		return
	}

	claims, err := h.tokenService.ValidateAccessToken(r.Context(), tokenString)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	output, err := h.authService.ValidateAccessSession(r.Context(), claims.SessionID, false)
	if err != nil {
		http.Error(w, "session is no longer active", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, output.AccountID)
	h.hub.Register(client, output.AccountID)

	go client.WritePump()
	go client.ReadPump()
}

func (h *Handler) extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		const bearerPrefix = "Bearer "
		if strings.HasPrefix(authHeader, bearerPrefix) {
			return strings.TrimPrefix(authHeader, bearerPrefix)
		}
	}

	// Fallback to query parameter (required for Flutter Web because browsers
	// cannot set custom headers on WebSocket connections).
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	return ""
}
