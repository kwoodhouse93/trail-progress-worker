package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

var ErrFinished = errors.New("no more activity/route pairs to process")

type processable struct {
	ID         string
	ActivityID int
	RouteID    string
}

func (s Store) ProcessOne(ctx context.Context) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to acquire connection")
	}
	defer conn.Release()

	row := conn.QueryRow(ctx, nextUnprocessed)
	var unprocessed processable
	err = row.Scan(&unprocessed.ID, &unprocessed.ActivityID, &unprocessed.RouteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFinished
		}
		return errors.Wrap(err, "store: failed to query for next unprocessed activity")
	}
	// log.Printf("processing pair %s", unprocessed.ID)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return errors.Wrap(err, "store: failed to begin transaction")
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			log.Println("store: failed to rollback transaction", err)
		}
	}()

	// Finish early if activity has no map
	n, err := tx.Exec(ctx, processNullMapOne, unprocessed.ID)
	if err != nil {
		return errors.Wrapf(err, "store: failed to process null map for processing pair %s", unprocessed.ID)
	}
	if n.RowsAffected() > 1 {
		return errors.Errorf("store: processing null map marked more than one processing pair as processed (pair: %s)", unprocessed.ID)
	}
	if n.RowsAffected() == 1 {
		// log.Printf("store: processed pair %s", unprocessed.ID)
		err = tx.Commit(ctx)
		if err != nil {
			return errors.Wrap(err, "store: error committing transaction")
		}
		return nil
	}

	// Check if activity track is near route
	row = tx.QueryRow(ctx, populateRelevantActivitiesOne, unprocessed.ActivityID, unprocessed.RouteID)
	var relevant bool
	err = row.Scan(&relevant)
	if err != nil {
		return errors.Wrapf(err, "store: failed to populate relevant activity for processing pair %s", unprocessed.ID)
	}
	if !relevant {
		n, err := tx.Exec(ctx, markAsProcessedOne, unprocessed.ID)
		if err != nil {
			return errors.Wrap(err, "store: failed to mark as processed")
		}
		if n.RowsAffected() != 1 {
			return errors.New("store: failed to mark as processed")
		}
		// log.Printf("store: marked processing pair %s as processed", unprocessed.ID)
		err = tx.Commit(ctx)
		if err != nil {
			return errors.Wrap(err, "store: error committing transaction")
		}
		return nil
	}

	// Intersections
	_, err = tx.Exec(ctx, populateIntersectionsOne, unprocessed.ID)
	if err != nil {
		return errors.Wrap(err, "store: failed to populate intersections")
	}
	// log.Printf("store: populated %d intersections for processing pair %s", n.RowsAffected(), unprocessed.ID)

	// Route sections
	_, err = tx.Exec(ctx, populateRouteSectionsOne, unprocessed.ID)
	if err != nil {
		return errors.Wrap(err, "store: failed to populate route sections")
	}
	// log.Printf("store: populated %d route sections for processing pair %s", n.RowsAffected(), unprocessed.ID)

	// Update route stats
	n, err = tx.Exec(ctx, updateRouteStatsOne, unprocessed.RouteID)
	if err != nil {
		return errors.Wrap(err, "store: failed to populate route sections")
	}
	if n.RowsAffected() != 1 {
		return fmt.Errorf("store: failed to update route stats - affected %d rows", n.RowsAffected())
	}
	// log.Printf("store: populated route stats for route %s", unprocessed.RouteID)

	// Mark activity as processed
	n, err = tx.Exec(ctx, markAsProcessedOne, unprocessed.ID)
	if err != nil {
		return errors.Wrap(err, "store: failed to mark as processed")
	}
	if n.RowsAffected() != 1 {
		return errors.New("store: failed to mark as processed")
	}
	// log.Printf("store: marked processing pair %s as processed", unprocessed.ID)

	err = tx.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "store: error committing transaction")
	}
	return nil
}
