package store

const (
	nextUnprocessed = `
UPDATE processing SET processing_started_at = NOW()
WHERE processing.id = (
	SELECT p.id FROM processing as p
	WHERE p.processing_started_at IS NULL
	ORDER BY p.created_at ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
RETURNING processing.id, processing.activity_id, processing.route_id
`

	markAsProcessedOne = `
UPDATE processing SET processed = true
WHERE id = $1
`

	processNullMapOne = `
UPDATE processing
SET processed = true
FROM activities
WHERE
	processing.activity_id = activities.id AND
	processing.id = $1 AND
	activities.summary_track IS NULL
`

	populateRelevantActivitiesOne = `
INSERT INTO relevant_activities (
	activity_id,
	route_id,
	relevant
)
SELECT
	activities.id AS activity_id,
	routes.id AS route_id,
	ST_DWithin(
		ST_Simplify(
			activities.summary_track::geometry,
			0.001
		)::geography,
		ST_Simplify(
			routes.track::geometry,
			0.001
		)::geography,
		200, -- buffer distance
		false -- Use sphere for speed
	) AS relevant
FROM
	activities
CROSS JOIN routes
JOIN processing ON processing.activity_id = activities.id AND processing.route_id = routes.id
WHERE processing.activity_id = $1 AND processing.route_id = $2
RETURNING relevant
`

	populateIntersectionsOne = `
WITH
	activity AS (
		SELECT
			activities.id AS activity_id,
			activities.summary_track AS activity_track,
			routes.id AS route_id,
			routes.track AS route_track
		FROM activities
		JOIN processing ON processing.activity_id = activities.id
		JOIN routes ON processing.route_id = routes.id
		WHERE
			processing.id = $1
	)
INSERT INTO intersections (
	activity_id,
	route_id,
	intersection_track
)
SELECT
	activity_id,
	route_id,
	(ST_Dump(
		ST_Intersection(
			activity_track,
			ST_Buffer(
				route_track,
				200 -- buffer distance
			)
		)::geometry
	)).geom AS intersection_track
FROM activity
`

	populateRouteSectionsOne = `
WITH
	pick AS (
		SELECT
			intersections.id AS intersection_id,
			intersections.activity_id,
			intersections.route_id,
			intersections.intersection_track,
			routes.track AS route_track
		FROM activities
		JOIN intersections ON intersections.activity_id = activities.id
		JOIN routes ON intersections.route_id = routes.id
		JOIN processing ON processing.activity_id = activities.id AND processing.route_id = routes.id
		WHERE processing.id = $1
	),
	start_ends AS (
		SELECT
			intersection_id,
			activity_id,
			route_id,
			route_track,
			ST_LineLocatePoint(
				route_track::geometry,
				(ST_Dump(
					ST_Boundary(
						intersection_track::geometry
					)
				)).geom
			) AS start_end_points
		FROM pick
	),
	sections AS (
		SELECT
			activity_id,
			route_id,
			ST_LineSubstring(
				route_track::geometry,
				MIN(start_end_points),
				MAX(start_end_points)
			) AS section_track
		FROM start_ends
		GROUP BY activity_id, route_id, intersection_id, route_track
	)
INSERT INTO route_sections (
	activity_id,
	route_id,
	section_track
)
SELECT
	activity_id,
	route_id,
	section_track
FROM sections
WHERE GeometryType(section_track) = 'LINESTRING'
`

	updateRouteStatsOne = `
WITH rs as (
		SELECT
			route_sections.section_track,
			route_sections.route_id,
			activities.athlete_id
		FROM route_sections
		JOIN activities ON activities.id = route_sections.activity_id
	)
	INSERT INTO route_stats (
		route_id,
		athlete_id,
		covered_length
	)
	SELECT
		routes.id AS route_id,
		athletes.id AS athlete_id,
		COALESCE(
			ST_Length(
				ST_Union(
					rs.section_track::geometry
				)::geography
			),
			0
		) AS covered_length
	FROM routes
	CROSS JOIN athletes
	LEFT OUTER JOIN rs
		ON rs.route_id = routes.id
		AND rs.athlete_id = athletes.id
	WHERE routes.id = $1
	GROUP BY routes.id, athletes.id
	ON CONFLICT (athlete_id, route_id) DO UPDATE
	SET covered_length = EXCLUDED.covered_length
`
)
