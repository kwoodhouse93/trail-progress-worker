package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type Store struct {
	conn *pgx.Conn
}

func New(connectionURL string) (*Store, error) {
	ctx := context.Background()

	config, err := pgx.ParseConfig(connectionURL)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &Store{
		conn: conn,
	}, nil
}

type NotificationHandler func(payload string)

func (s Store) Listen(ctx context.Context, channel string, handler NotificationHandler) error {
	_, err := s.conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channel))
	if err != nil {
		return err
	}
	log.Println("registered listener for channel", channel)

	for {
		notification, err := s.conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		handler(notification.Payload)
	}
}

func (s Store) Process(ctx context.Context) error {
	rows, err := s.conn.Query(ctx, "SELECT id FROM activities")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return err
		}
		log.Println("processing activity", id)
	}

	return nil
}

func (s Store) Cleanup() error {
	return s.conn.Close(context.Background())
}
