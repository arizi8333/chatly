package delivery

import (
	"net/http"
	"strconv"

	"chatly/internal/domain"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageRepo domain.MessageRepository
}

func NewMessageHandler(messageRepo domain.MessageRepository) *MessageHandler {
	return &MessageHandler{messageRepo: messageRepo}
}

// GetHistory godoc
// @Summary Get chat history
// @Description Retrieve paginated chat history for a specific room
// @Tags messages
// @Produce json
// @Param room query string false "Room name (default: general)"
// @Param limit query int false "Max messages to return (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {array} domain.Message
// @Failure 500 {object} map[string]string
// @Router /api/messages [get]
func (h *MessageHandler) GetHistory(c *gin.Context) {
	room := c.Query("room")
	if room == "" {
		room = "general"
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.messageRepo.FindByRoom(c.Request.Context(), room, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
