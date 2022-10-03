package strava

import "time"

type MetaAthlete struct {
	ID int `json:"id"`
}

type LatLong [2]float64

type ActivityType string

const (
	ActivityTypeAlpineSki       ActivityType = "AlpineSki"
	ActivityTypeBackcountrySki  ActivityType = "BackcountrySki"
	ActivityTypeCanoeing        ActivityType = "Canoeing"
	ActivityTypeCrossfit        ActivityType = "Crossfit"
	ActivityTypeEBikeRide       ActivityType = "EBikeRide"
	ActivityTypeElliptical      ActivityType = "Elliptical"
	ActivityTypeGolf            ActivityType = "Golf"
	ActivityTypeHandcycle       ActivityType = "Handcycle"
	ActivityTypeHike            ActivityType = "Hike"
	ActivityTypeIceSkate        ActivityType = "IceSkate"
	ActivityTypeInlineSkate     ActivityType = "InlineSkate"
	ActivityTypeKayaking        ActivityType = "Kayaking"
	ActivityTypeKitesurf        ActivityType = "Kitesurf"
	ActivityTypeNordicSki       ActivityType = "NordicSki"
	ActivityTypeRide            ActivityType = "Ride"
	ActivityTypeRockClimbing    ActivityType = "RockClimbing"
	ActivityTypeRollerSki       ActivityType = "RollerSki"
	ActivityTypeRowing          ActivityType = "Rowing"
	ActivityTypeRun             ActivityType = "Run"
	ActivityTypeSail            ActivityType = "Sail"
	ActivityTypeSkateboard      ActivityType = "Skateboard"
	ActivityTypeSnowboard       ActivityType = "Snowboard"
	ActivityTypeSnowshoe        ActivityType = "Snowshoe"
	ActivityTypeSoccer          ActivityType = "Soccer"
	ActivityTypeStairStepper    ActivityType = "StairStepper"
	ActivityTypeStandUpPaddling ActivityType = "StandUpPaddling"
	ActivityTypeSurfing         ActivityType = "Surfing"
	ActivityTypeSwim            ActivityType = "Swim"
	ActivityTypeVelomobile      ActivityType = "Velomobile"
	ActivityTypeVirtualRide     ActivityType = "VirtualRide"
	ActivityTypeVirtualRun      ActivityType = "VirtualRun"
	ActivityTypeWalk            ActivityType = "Walk"
	ActivityTypeWeightTraining  ActivityType = "WeightTraining"
	ActivityTypeWheelchair      ActivityType = "Wheelchair"
	ActivityTypeWindsurf        ActivityType = "Windsurf"
	ActivityTypeWorkout         ActivityType = "Workout"
	ActivityTypeYoga            ActivityType = "Yoga"
)

type SportType string

const (
	SportTypeAlpineSki         = "AlpineSki"
	SportTypeBackcountrySki    = "BackcountrySki"
	SportTypeCanoeing          = "Canoeing"
	SportTypeCrossfit          = "Crossfit"
	SportTypeEBikeRide         = "EBikeRide"
	SportTypeElliptical        = "Elliptical"
	SportTypeEMountainBikeRide = "EMountainBikeRide"
	SportTypeGolf              = "Golf"
	SportTypeGravelRide        = "GravelRide"
	SportTypeHandcycle         = "Handcycle"
	SportTypeHike              = "Hike"
	SportTypeIceSkate          = "IceSkate"
	SportTypeInlineSkate       = "InlineSkate"
	SportTypeKayaking          = "Kayaking"
	SportTypeKitesurf          = "Kitesurf"
	SportTypeMountainBikeRide  = "MountainBikeRide"
	SportTypeNordicSki         = "NordicSki"
	SportTypeRide              = "Ride"
	SportTypeRockClimbing      = "RockClimbing"
	SportTypeRollerSki         = "RollerSki"
	SportTypeRowing            = "Rowing"
	SportTypeRun               = "Run"
	SportTypeSail              = "Sail"
	SportTypeSkateboard        = "Skateboard"
	SportTypeSnowboard         = "Snowboard"
	SportTypeSnowshoe          = "Snowshoe"
	SportTypeSoccer            = "Soccer"
	SportTypeStairStepper      = "StairStepper"
	SportTypeStandUpPaddling   = "StandUpPaddling"
	SportTypeSurfing           = "Surfing"
	SportTypeSwim              = "Swim"
	SportTypeTrailRun          = "TrailRun"
	SportTypeVelomobile        = "Velomobile"
	SportTypeVirtualRide       = "VirtualRide"
	SportTypeVirtualRun        = "VirtualRun"
	SportTypeWalk              = "Walk"
	SportTypeWeightTraining    = "WeightTraining"
	SportTypeWheelchair        = "Wheelchair"
	SportTypeWindsurf          = "Windsurf"
	SportTypeWorkout           = "Workout"
	SportTypeYoga              = "Yoga"
)

type PolylineMap struct {
	ID              string `json:"id"`
	Polyline        string `json:"polyline"`
	SummaryPolyline string `json:"summary_polyline"`
}

// Ignoring some fields because we don't care about them.
type DetailedActivity struct {
	ID                   int          `json:"id"`
	ExternalID           string       `json:"external_id"`
	UploadID             int          `json:"upload_id"`
	Athlete              MetaAthlete  `json:"athlete"`
	Name                 string       `json:"name"`
	Distance             float64      `json:"distance"`
	MovingTime           int          `json:"moving_time"`
	ElapsedTime          int          `json:"elapsed_time"`
	TotalElevationGain   float64      `json:"total_elevation_gain"`
	ElevHigh             float64      `json:"elev_high"`
	ElevLow              float64      `json:"elev_low"`
	Type                 ActivityType `json:"type"` // Deprecated
	SportType            SportType    `json:"sport_type"`
	StartDate            time.Time    `json:"start_date"`
	StartDateLocal       time.Time    `json:"start_date_local"`
	Timezone             string       `json:"timezone"`
	StartLatLong         LatLong      `json:"start_latlng"`
	EndLatLong           LatLong      `json:"end_latlng"`
	AchievementCount     int          `json:"achievement_count"`
	KudosCount           int          `json:"kudos_count"`
	CommentCount         int          `json:"comment_count"`
	AthleteCount         int          `json:"athlete_count"`
	PhotoCount           int          `json:"photo_count"`
	TotalPhotoCount      int          `json:"total_photo_count"`
	Map                  PolylineMap  `json:"map"`
	Trainer              bool         `json:"trainer"`
	Commute              bool         `json:"commute"`
	Manual               bool         `json:"manual"`
	Private              bool         `json:"private"`
	Flagged              bool         `json:"flagged"`
	WorkoutType          int          `json:"workout_type"`
	UploadIDStr          string       `json:"upload_id_str"`
	AverageSpeed         float64      `json:"average_speed"`
	MaxSpeed             float64      `json:"max_speed"`
	HasKudoed            bool         `json:"has_kudoed"`
	HideFromHome         bool         `json:"hide_from_home"`
	GearID               string       `json:"gear_id"`
	Kilojoules           float64      `json:"kilojoules"`
	AverageWatts         float64      `json:"average_watts"`
	DeviceWatts          bool         `json:"device_watts"`
	MaxWatts             int          `json:"max_watts"`
	WeightedAverageWatts int          `json:"weighted_average_watts"`
	Description          string       `json:"description"`
	// Photos               PhotosSummary           `json:"photos"`
	// Gear                 SummaryGear             `json:"gear"`
	Calories float64 `json:"calories"`
	// SegmentEfforts       []DetailedSegmentEffort `json:"segment_efforts"`
	DeviceName string `json:"device_name"`
	EmbedToken string `json:"embed_token"`
	// SplitsMetric         []SplitMetric           `json:"splits_metric"`
	// SplitsStandard       []SplitStandard         `json:"splits_standard"`
	// Laps                 []Lap                   `json:"laps"`
	// BestEfforts          []DetailedSegmentEffort `json:"best_efforts"`
}
