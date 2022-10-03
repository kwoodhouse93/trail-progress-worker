package store

import (
	"context"
	"strconv"
	"strings"

	"github.com/kwoodhouse93/trail-progress-worker/strava"
)

func (s Store) StoreActivity(ctx context.Context, athleteID int, activity strava.DetailedActivity) error {
	var summaryTrack *string = nil
	if activity.Map.SummaryPolyline != "" {
		summaryTrack = &activity.Map.SummaryPolyline
	}
	_, err := s.pool.Exec(
		ctx,
		insertActivityQuery,
		activity.ID,
		athleteID,
		activity.Name,
		activity.Distance,
		activity.MovingTime,
		activity.ElapsedTime,
		activity.TotalElevationGain,
		activity.SportType,
		activity.StartDate,
		activity.Timezone,
		summaryTrack,
		activity.StartLatLong[1],
		activity.StartLatLong[0],
		activity.EndLatLong[1],
		activity.EndLatLong[0],
		activity.ElevHigh,
		activity.ElevLow,
		activity.ExternalID,
	)
	if err != nil {
		return err
	}
	return nil
}

const insertActivityQuery = `
INSERT INTO activities (
	id,
	athlete_id,
	name,
	distance, 
	moving_time, 
	elapsed_time, 
	total_elevation_gain,
	activity_type,
	start_date,
	local_tz,
	summary_track,
	start_latlng,
	end_latlng,
	elev_high,
	elev_low,
	external_id
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	ST_LineFromEncodedPolyline($11),
	ST_Point($12, $13),
	ST_Point($14, $15),
	$16,
	$17,
	$18
) ON CONFLICT DO NOTHING`

func (s Store) UpdateActivity(ctx context.Context, athleteID, activityID int, title, activityType *string) error {
	paramCount := 3
	updates := []string{}
	values := []interface{}{athleteID, activityID}
	if title != nil {
		updates = append(updates, "name = $"+strconv.Itoa(paramCount))
		values = append(values, *title)
		paramCount++
	}
	if activityType != nil {
		updates = append(updates, "activity_type = $"+strconv.Itoa(paramCount))
		values = append(values, *activityType)
		paramCount++
	}
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE activities SET " + strings.Join(updates, ", ") + " WHERE athlete_id = $1 AND id = $2"
	_, err := s.pool.Exec(
		ctx,
		query,
		values...,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s Store) DeleteActivity(ctx context.Context, athleteID, activityID int) error {
	_, err := s.pool.Exec(ctx, deleteActivityQuery, athleteID, activityID)
	if err != nil {
		return err
	}
	return nil
}

const deleteActivityQuery = `
DELETE FROM activities
WHERE athlete_id = $1
AND id = $2
`
