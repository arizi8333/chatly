package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID         uuid.UUID `json:"id"`
	Room       string    `json:"room"`
	SenderID   uuid.UUID `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type MessageRepository interface {
	Save(ctx context.Context, msg *Message) error
	FindByRoom(ctx context.Context, room string, limit, offset int) ([]Message, error)
}
