package repository

import (
	"context"
	"database/sql"
	"errors"

	"chatly/internal/domain"

	"github.com/google/uuid"
)

type pgTokenRepository struct {
	db *sql.DB
}

func NewPostgresTokenRepository(db *sql.DB) domain.TokenRepository {
	return &pgTokenRepository{db: db}
}

func (r *pgTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.Revoked, token.CreatedAt)
	return err
}

func (r *pgTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, expires_at, revoked, created_at FROM refresh_tokens WHERE token_hash = $1`
	row := r.db.QueryRowContext(ctx, query, hash)

	var token domain.RefreshToken
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.Revoked, &token.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("refresh token not found")
	}
	return &token, err
}

func (r *pgTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *pgTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
