package store

import (
	"context"
	"time"
)

type AccessToken struct {
	Token     string
	ExpiresAt time.Time
}

func (a AccessToken) IsExpired() bool {
	return a.ExpiresAt.Before(time.Now())
}

func (s Store) GetAccessToken(ctx context.Context, athleteID int) (*AccessToken, error) {
	row := s.pool.QueryRow(ctx, "SELECT access_token, expires_at FROM access_tokens WHERE athlete_id = $1", athleteID)

	var accessToken string
	var expiresAt time.Time
	err := row.Scan(&accessToken, &expiresAt)
	if err != nil {
		return nil, err
	}
	return &AccessToken{
		Token:     accessToken,
		ExpiresAt: expiresAt,
	}, nil
}

type RefreshToken struct {
	Token string
}

func (s Store) GetRefreshToken(ctx context.Context, athleteID int) (*RefreshToken, error) {
	row := s.pool.QueryRow(ctx, "SELECT refresh_token FROM refresh_tokens WHERE athlete_id = $1", athleteID)

	var accessToken string
	err := row.Scan(&accessToken)
	if err != nil {
		return nil, err
	}
	return &RefreshToken{
		Token: accessToken,
	}, nil
}

func (s Store) StoreTokens(ctx context.Context, athleteID int, accessToken, refreshToken string, expiresAt time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "UPDATE access_tokens SET access_token = $1, expires_at = $2 WHERE athlete_id = $3", accessToken, expiresAt, athleteID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "UPDATE refresh_tokens SET refresh_token = $1 WHERE athlete_id = $2", refreshToken, athleteID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
