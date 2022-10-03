package store

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func (s Store) Process(ctx context.Context) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "store: error acquiring connection")
	}
	defer conn.Release()

	// Setting IsoLevel to RepeatableRead ensures we only process activities
	// that were available when we started running the process. It effectively
	// uses a snapshot of the database for the duration of the transaction.
	//
	// This is important as the later queries - intersections and route_sections -
	// do not produce 1 row per activity, but 0 to many rows per activity.
	//
	// The upshot of this is we cannot easily tell whether an activity has been
	// processed. Hence, our last query sets all remaining activities to
	// `processed = true`. If we don't use RepeatableRead, and more activities
	// are committed to the DB before we finish processing, we will mark those
	// activities as processed without processing them. This means they will never
	// be processed.
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return errors.Wrap(err, "store: error beginning transaction")
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			log.Println("store: failed to rollback transaction", err)
		}
	}()

	row := tx.QueryRow(ctx, "SELECT COUNT(id) FROM processing WHERE processed = false")
	var count int
	err = row.Scan(&count)
	if err != nil {
		return errors.Wrap(err, "store: error getting unprocessed activity/route pairs")
	}
	log.Printf("store: %d activity/route pairs to process\n", count)

	totalProcessed := 0

	n, err := tx.Exec(ctx, processNullMaps)
	if err != nil {
		return errors.Wrap(err, "store: error executing process null maps")
	}
	totalProcessed += int(n.RowsAffected())
	log.Printf("store: marked as processed %d activity/route pairs with null maps\n", n.RowsAffected())

	n, err = tx.Exec(ctx, populateRelevantActivities)
	if err != nil {
		return errors.Wrap(err, "store: error executing populate relevant activities")
	}
	log.Printf("store: populated %d relevant_activities\n", n.RowsAffected())

	n, err = tx.Exec(ctx, processIrrelevantActivities)
	if err != nil {
		return errors.Wrap(err, "store: error executing process irrelevant activities")
	}
	totalProcessed += int(n.RowsAffected())
	log.Printf("store: marked as processed %d irrelevant activity/route pairs\n", n.RowsAffected())

	n, err = tx.Exec(ctx, populateIntersections)
	if err != nil {
		return errors.Wrap(err, "store: error executing populate intersections")
	}
	log.Printf("store: populated %d intersections\n", n.RowsAffected())

	n, err = tx.Exec(ctx, populateRouteSections)
	if err != nil {
		return errors.Wrap(err, "store: error executing populate route sections")
	}
	log.Printf("store: populated %d route sections\n", n.RowsAffected())

	n, err = tx.Exec(ctx, processRemaining)
	if err != nil {
		return errors.Wrap(err, "store: error executing process remaining")
	}
	totalProcessed += int(n.RowsAffected())
	log.Printf("store: marked as processed %d activity/route pairs\n", n.RowsAffected())

	n, err = tx.Exec(ctx, populateRouteStats)
	if err != nil {
		return errors.Wrap(err, "store: error executing populate route stats")
	}
	log.Printf("store: populated %d route stats\n", n.RowsAffected())

	log.Printf("store: total activity/route pairs marked as processed: %d\n", totalProcessed)

	err = tx.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "store: error committing transaction")
	}
	return nil
}
