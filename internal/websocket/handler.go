package websocket

import (
	"log"
	"net/http"

	"chatly/internal/config"
	"chatly/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, check against allowed origins
	},
}

type Handler struct {
	hub       *Hub
	userRepo  domain.UserRepository
	onMessage func(domain.Message)
}

func NewHandler(hub *Hub, userRepo domain.UserRepository, onMessage func(domain.Message)) *Handler {
	return &Handler{
		hub:       hub,
		userRepo:  userRepo,
		onMessage: onMessage,
	}
}

func (h *Handler) HandleWS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		room := c.Query("room")
		if room == "" {
			room = "general"
		}

		// Auth via query param 'token' for WS
		userIDString, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, _ := uuid.Parse(userIDString.(string))
		user, err := h.userRepo.FindByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("upgrade error: %v", err)
			return
		}

		client := &Client{
			Hub:  h.hub,
			Conn: conn,
			Send: make(chan domain.Message, 256),
			Room: room,
			User: user,
		}

		client.Hub.Register <- client

		go client.WritePump()
		go client.ReadPump(h.onMessage)
	}
}
