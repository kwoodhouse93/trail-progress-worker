package store

const (
	processNullMaps = `
UPDATE processing
SET processed = true
FROM activities
WHERE
	processing.activity_id = activities.id AND
	processing.processed = false AND
	activities.summary_track IS NULL
`

	populateRelevantActivities = `
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
WHERE processing.processed = false
ON CONFLICT DO NOTHING
`

	processIrrelevantActivities = `
UPDATE processing
SET processed = true
FROM relevant_activities
WHERE
	relevant_activities.activity_id = processing.activity_id
	AND relevant_activities.route_id = processing.route_id
	AND processing.processed = false
	AND relevant_activities.relevant = false
`

	populateIntersections = `
WITH
	relevants AS (
		SELECT
			activities.id AS activity_id,
			activities.summary_track AS activity_track,
			routes.id AS route_id,
			routes.track AS route_track
		FROM activities
		JOIN relevant_activities ON relevant_activities.activity_id = activities.id
		JOIN routes ON relevant_activities.route_id = routes.id
		JOIN processing ON processing.activity_id = activities.id AND processing.route_id = routes.id
		WHERE
			processing.processed = false AND
			relevant_activities.relevant = true
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
FROM relevants
`

	populateRouteSections = `
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
		WHERE processing.processed = false
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

	populateRouteStats = `
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
GROUP BY routes.id, athletes.id
ON CONFLICT (athlete_id, route_id) DO UPDATE
SET covered_length = EXCLUDED.covered_length
`

	processRemaining = `
UPDATE processing
SET processed = true
WHERE processed = false
`
)
