package repository

import (
	"context"
	"fmt"
	"time"

	"switch-server/internal/db"
	"switch-server/models"
)

type RelayTokenRepository struct {
	db *db.DB
}

func NewRelayTokenRepository(database *db.DB) *RelayTokenRepository {
	return &RelayTokenRepository{db: database}
}

func (r *RelayTokenRepository) CreateRelayToken(ctx context.Context, tokenHash, tokenPrefix string, userID, tenantID int64) (*models.RelayToken, error) {
	row, err := r.db.CreateRelayToken(ctx, db.CreateRelayTokenParams{
		TokenHash:   tokenHash,
		UserID:      userID,
		TenantID:    tenantID,
		TokenPrefix: tokenPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create relay token: %w", err)
	}

	return sqlcRelayTokenToModel(&row), nil
}

func (r *RelayTokenRepository) GetRelayTokenByHash(ctx context.Context, tokenHash string) (*models.RelayToken, error) {
	row, err := r.db.GetRelayTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get relay token by hash: %w", err)
	}

	return sqlcRelayTokenToModel(&row), nil
}

func (r *RelayTokenRepository) GetRelayTokenByUserID(ctx context.Context, userID int64) (*models.RelayToken, error) {
	row, err := r.db.GetRelayTokenByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relay token by user ID: %w", err)
	}

	return sqlcRelayTokenToModel(&row), nil
}

func (r *RelayTokenRepository) RevokeRelayTokenByUserID(ctx context.Context, userID int64) error {
	err := r.db.RevokeRelayTokenByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke relay token: %w", err)
	}
	return nil
}

func (r *RelayTokenRepository) UpdateRelayTokenLastUsed(ctx context.Context, id int64) error {
	err := r.db.UpdateRelayTokenLastUsed(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to update relay token last used: %w", err)
	}
	return nil
}

func sqlcRelayTokenToModel(row *db.RelayToken) *models.RelayToken {
	m := &models.RelayToken{
		ID:          row.ID,
		TokenHash:   row.TokenHash,
		UserID:      row.UserID,
		TenantID:    row.TenantID,
		TokenPrefix: row.TokenPrefix,
	}
	if row.CreatedAt.Valid {
		m.CreatedAt = row.CreatedAt.Time
	}
	if row.LastUsedAt.Valid {
		m.LastUsedAt = &row.LastUsedAt.Time
	}
	if row.RevokedAt.Valid {
		m.RevokedAt = &row.RevokedAt.Time
	}
	return m
}

// RelayTokenRepositoryInterface defines the contract for relay token repository operations
type RelayTokenRepositoryInterface interface {
	CreateRelayToken(ctx context.Context, tokenHash, tokenPrefix string, userID, tenantID int64) (*models.RelayToken, error)
	GetRelayTokenByHash(ctx context.Context, tokenHash string) (*models.RelayToken, error)
	GetRelayTokenByUserID(ctx context.Context, userID int64) (*models.RelayToken, error)
	RevokeRelayTokenByUserID(ctx context.Context, userID int64) error
	UpdateRelayTokenLastUsed(ctx context.Context, id int64) error
}

// Compile-time check
var _ RelayTokenRepositoryInterface = (*RelayTokenRepository)(nil)

// Ensure time is available for model conversion
var _ time.Time
