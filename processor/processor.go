package processor

import (
	"context"
	"log"
	"time"
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
	err := p.process(ctx)
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
			log.Println("triggered")
			err := p.process(ctx)
			if err != nil {
				return err
			}

		// Process activities on a regular interval.
		case <-ticker.C:
			log.Println("interval fired")
			err := p.process(ctx)
			if err != nil {
				return err
			}
		}
	}
}

// Check for unprocessed activities and process them.
// Single-thread operation only!
func (p Processor) process(ctx context.Context) error {
	log.Println("processing")
	defer func() { log.Println("done processing") }()
	return p.store.Process(ctx)
}
