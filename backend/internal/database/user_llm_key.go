package database

import (
	"context"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserLLMKeyRepo struct {
	pool *pgxpool.Pool
}

func NewUserLLMKeyRepo(pool *pgxpool.Pool) *UserLLMKeyRepo {
	return &UserLLMKeyRepo{pool: pool}
}

func (r *UserLLMKeyRepo) Upsert(ctx context.Context, key *domain.UserLLMKey) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_llm_key (id, user_id, provider, encrypted_key, key_nonce, key_hint, model, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, provider) DO UPDATE SET
			id = EXCLUDED.id,
			encrypted_key = EXCLUDED.encrypted_key,
			key_nonce = EXCLUDED.key_nonce,
			key_hint = EXCLUDED.key_hint,
			model = EXCLUDED.model,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`, key.ID, key.UserID, key.Provider, key.EncryptedKey, key.KeyNonce,
		key.KeyHint, key.Model, key.IsActive, key.CreatedAt, key.UpdatedAt)
	return err
}

func (r *UserLLMKeyRepo) GetByUserAndProvider(ctx context.Context, userID, provider string) (*domain.UserLLMKey, error) {
	var k domain.UserLLMKey
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, encrypted_key, key_nonce, key_hint, model, is_active, created_at, updated_at
		FROM user_llm_key
		WHERE user_id = $1 AND provider = $2
	`, userID, provider).Scan(
		&k.ID, &k.UserID, &k.Provider, &k.EncryptedKey, &k.KeyNonce,
		&k.KeyHint, &k.Model, &k.IsActive, &k.CreatedAt, &k.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

func (r *UserLLMKeyRepo) ListByUser(ctx context.Context, userID string) ([]domain.UserLLMKey, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, provider, encrypted_key, key_nonce, key_hint, model, is_active, created_at, updated_at
		FROM user_llm_key
		WHERE user_id = $1
		ORDER BY provider
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []domain.UserLLMKey
	for rows.Next() {
		var k domain.UserLLMKey
		if err := rows.Scan(
			&k.ID, &k.UserID, &k.Provider, &k.EncryptedKey, &k.KeyNonce,
			&k.KeyHint, &k.Model, &k.IsActive, &k.CreatedAt, &k.UpdatedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (r *UserLLMKeyRepo) Update(ctx context.Context, userID, provider string, isActive *bool, model *string) error {
	if isActive == nil && model == nil {
		return nil
	}
	now := time.Now()
	if isActive != nil && model != nil {
		_, err := r.pool.Exec(ctx, `
			UPDATE user_llm_key SET is_active = $1, model = $2, updated_at = $3
			WHERE user_id = $4 AND provider = $5
		`, *isActive, *model, now, userID, provider)
		return err
	}
	if isActive != nil {
		_, err := r.pool.Exec(ctx, `
			UPDATE user_llm_key SET is_active = $1, updated_at = $2
			WHERE user_id = $3 AND provider = $4
		`, *isActive, now, userID, provider)
		return err
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE user_llm_key SET model = $1, updated_at = $2
		WHERE user_id = $3 AND provider = $4
	`, *model, now, userID, provider)
	return err
}

func (r *UserLLMKeyRepo) Delete(ctx context.Context, userID, provider string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_llm_key WHERE user_id = $1 AND provider = $2
	`, userID, provider)
	return err
}
