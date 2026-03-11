package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppSettingRepo struct {
	pool *pgxpool.Pool
}

func NewAppSettingRepo(pool *pgxpool.Pool) *AppSettingRepo {
	return &AppSettingRepo{pool: pool}
}

func (r *AppSettingRepo) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx, `SELECT value FROM app_setting WHERE key = $1`, key).Scan(&value)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get app setting %q: %w", key, err)
	}
	return value, nil
}

func (r *AppSettingRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO app_setting (key, value, updated_at) VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`,
		key, value)
	if err != nil {
		return fmt.Errorf("set app setting %q: %w", key, err)
	}
	return nil
}

func (r *AppSettingRepo) GetInt(ctx context.Context, key string, fallback int) (int, error) {
	v, err := r.Get(ctx, key)
	if err != nil {
		return fallback, err
	}
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback, nil
	}
	return n, nil
}
