package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
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
	log.Println("registered listener for channel", channel)

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}
		handler(notification.Payload)
	}
}

func (s Store) Process(ctx context.Context) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			log.Println("failed to rollback transaction", err)
		}
	}()

	row := tx.QueryRow(ctx, "SELECT COUNT(id) FROM activities WHERE processed = false")
	var count int
	err = row.Scan(&count)
	if err != nil {
		return err
	}
	log.Printf("%d activities to process\n", count)

	totalProcessed := 0

	n, err := tx.Exec(ctx, processNullMaps)
	if err != nil {
		return err
	}
	totalProcessed += int(n.RowsAffected())
	log.Printf("marked as processed %d activities with null maps\n", n.RowsAffected())

	n, err = tx.Exec(ctx, populateRelevantActivities)
	if err != nil {
		return err
	}
	log.Printf("populated %d relevant_activities\n", n.RowsAffected())

	n, err = tx.Exec(ctx, processIrrelevantActivities)
	if err != nil {
		return err
	}
	totalProcessed += int(n.RowsAffected())
	log.Printf("marked as processed %d irrelevant activities\n", n.RowsAffected())

	n, err = tx.Exec(ctx, populateIntersections)
	if err != nil {
		return err
	}
	log.Printf("populated %d intersections\n", n.RowsAffected())

	n, err = tx.Exec(ctx, populateRouteSections)
	if err != nil {
		return err
	}
	log.Printf("populated %d route sections\n", n.RowsAffected())

	n, err = tx.Exec(ctx, processRemaining)
	if err != nil {
		return err
	}
	totalProcessed += int(n.RowsAffected())
	log.Printf("marked as processed %d activities\n", n.RowsAffected())

	log.Printf("total activities marked as processed: %d\n", totalProcessed)
	return tx.Commit(ctx)
}

func (s Store) Cleanup() {
	s.pool.Close()
}
