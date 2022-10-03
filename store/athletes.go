package store

import "context"

func (s Store) DeleteAthlete(ctx context.Context, athleteID int) error {
	_, err := s.pool.Exec(ctx, deleteAthleteQuery, athleteID)
	if err != nil {
		return err
	}
	return nil
}

const deleteAthleteQuery = `
DELETE FROM athletes
WHERE athlete_id = $1
`
