package store

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

type batch []processable

func (b batch) IDs() []string {
	ids := []string{}
	for _, p := range b {
		ids = append(ids, p.ID)
	}
	return ids
}

func (s Store) ProcessBatch(ctx context.Context, batchSize int) (int, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to acquire connection")
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, nextUnprocessedBatch, batchSize)
	if err != nil {
		return 0, errors.Wrap(err, "store: failed to get unprocessed batch")
	}
	unprocessed := batch{}
	for rows.Next() {
		var p processable
		err = rows.Scan(&p.ID, &p.ActivityID, &p.RouteID)
		if err != nil {
			return 0, errors.Wrap(err, "store: failed to scan processable")
		}
		unprocessed = append(unprocessed, p)
	}
	if len(unprocessed) == 0 {
		return 0, ErrFinished
	}
	log.Printf("processing %d pairs", len(unprocessed))

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return 0, errors.Wrap(err, "store: failed to begin transaction")
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			log.Println("store: failed to rollback transaction", err)
		}
	}()

	// Finish early if activities have no map
	n, err := tx.Exec(ctx, processNullMapBatch, unprocessed.IDs())
	if err != nil {
		return 0, errors.Wrapf(err, "store: failed to process null maps")
	}
	// If all pairs were processed, we're done
	if n.RowsAffected() == int64(batchSize) {
		err = tx.Commit(ctx)
		if err != nil {
			return 0, errors.Wrap(err, "store: error committing transaction")
		}
		return len(unprocessed), nil
	}

	// Check if activity tracks are near route
	_, err = tx.Exec(ctx, populateRelevantActivitiesBatch, unprocessed.IDs())
	if err != nil {
		return 0, errors.Wrapf(err, "store: failed to populate relevant activities")
	}

	// Intersections
	_, err = tx.Exec(ctx, populateIntersectionsBatch, unprocessed.IDs())
	if err != nil {
		return 0, errors.Wrap(err, "store: failed to populate intersections")
	}
	// log.Printf("store: populated %d intersections for processing pair %s", n.RowsAffected(), unprocessed.ID)

	// Route sections
	_, err = tx.Exec(ctx, populateRouteSectionsBatch, unprocessed.IDs())
	if err != nil {
		return 0, errors.Wrap(err, "store: failed to populate route sections")
	}
	// log.Printf("store: populated %d route sections for processing pair %s", n.RowsAffected(), unprocessed.ID)

	// Update route stats
	_, err = tx.Exec(ctx, updateRouteStatsBatch)
	if err != nil {
		return 0, errors.Wrap(err, "store: failed to populate route sections")
	}
	// log.Printf("store: populated route stats for route %s", unprocessed.RouteID)

	// Mark activity as processed
	_, err = tx.Exec(ctx, markAsProcessedBatch, unprocessed.IDs())
	if err != nil {
		return 0, errors.Wrap(err, "store: failed to mark as processed")
	}
	// log.Printf("store: marked processing pair %s as processed", unprocessed.ID)

	err = tx.Commit(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "store: error committing transaction")
	}
	return len(unprocessed), nil
}
