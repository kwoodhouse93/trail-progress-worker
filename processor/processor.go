package processor

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kwoodhouse93/trail-progress-worker/store"
)

type Store interface {
	Process(ctx context.Context) error
	ProcessOne(ctx context.Context) error
	ProcessBatch(ctx context.Context, batchSize int) (int, error)
}

type processorState int

const (
	idle processorState = iota
	processing
)

type Processor struct {
	store     Store
	trigger   chan struct{}
	interval  time.Duration
	state     processorState
	batchSize int
}

func New(store Store, trigger chan struct{}, interval time.Duration, batchSize int) *Processor {
	return &Processor{
		store:     store,
		trigger:   trigger,
		interval:  interval,
		state:     idle,
		batchSize: batchSize,
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
func (p *Processor) process(ctx context.Context, recur int) error {
	if p.state == processing {
		log.Println("processor: already processing")
		return nil
	}
	p.state = processing

	processed := 0
	log.Println("processor: processing")
	defer func() { log.Printf("processor: done processing - processed %d", processed) }()

	retries := 0
	for p.state == processing {
		n, err := p.store.ProcessBatch(ctx, p.batchSize)
		if err != nil {
			if errors.Is(err, store.ErrFinished) {
				log.Println("processor: finished processing")
				p.state = idle
				return nil
			}
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01") {
				if retries > 3 {
					return err
				}
				// Serialization error or deadlock - retry.
				log.Printf("processor: serialization error (%s), retrying", pgErr.Code)
				retries++
				continue
			}
			return err
		}
		processed += n
	}
	return nil
}
