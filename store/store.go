package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(connectionURL string) (*Store, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		return nil, err
	}

	return &Store{
		pool: pool,
	}, nil
}

type NotificationHandler func(payload string)

func (s Store) Listen(ctx context.Context, channel string, handler NotificationHandler) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channel))
	if err != nil {
		return err
	}
	log.Println("store: registered listener for channel", channel)

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}
		handler(notification.Payload)
	}
}

func (s Store) Cleanup() {
	s.pool.Close()
}
