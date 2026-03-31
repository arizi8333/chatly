package repository

import (
	"context"
	"database/sql"

	"chatly/internal/domain"
)

type pgMessageRepository struct {
	db *sql.DB
}

func NewPostgresMessageRepository(db *sql.DB) domain.MessageRepository {
	return &pgMessageRepository{db: db}
}

func (r *pgMessageRepository) Save(ctx context.Context, msg *domain.Message) error {
	query := `INSERT INTO messages (id, room, sender_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, msg.ID, msg.Room, msg.SenderID, msg.Content, msg.CreatedAt)
	return err
}

func (r *pgMessageRepository) FindByRoom(ctx context.Context, room string, limit, offset int) ([]domain.Message, error) {
	query := `SELECT m.id, m.room, m.sender_id, u.username, m.content, m.created_at 
	          FROM messages m 
	          JOIN users u ON m.sender_id = u.id 
	          WHERE m.room = $1 
	          ORDER BY m.created_at DESC 
	          LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, room, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.Room, &msg.SenderID, &msg.SenderName, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
