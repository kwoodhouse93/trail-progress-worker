package processor

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type Store interface {
	Process(ctx context.Context) error
}

type Processor struct {
	store    Store
	trigger  chan struct{}
	interval time.Duration
}

func New(store Store, trigger chan struct{}, interval time.Duration) *Processor {
	return &Processor{
		store:    store,
		trigger:  trigger,
		interval: interval,
	}
}

func (p *Processor) Serve(ctx context.Context) error {
	err := p.process(ctx, 0)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()

		// Process activities when triggered.
		case <-p.trigger:
			log.Println("processor: triggered")
			err := p.process(ctx, 0)
			if err != nil {
				return err
			}

		// Process activities on a regular interval.
		case <-ticker.C:
			log.Println("processor: interval fired")
			err := p.process(ctx, 0)
			if err != nil {
				return err
			}
		}
	}
}

// Check for unprocessed activities and process them.
// Single-thread operation only!
func (p Processor) process(ctx context.Context, recur int) error {
	if recur > 5 {
		return errors.New("processor: too many retries")
	}

	log.Println("processor: processing")
	defer func() { log.Println("processor: done processing") }()
	err := p.store.Process(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Serialization error - retry.
			if pgErr.Code == "40001" {
				log.Println("processor: serialization error, retrying")
				return p.process(ctx, recur+1)
			}
		}
		return err
	}
	return nil
}
