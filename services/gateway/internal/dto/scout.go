package dto

type CreateScoutProfileRequest struct {
	OrganizationName string  `json:"organization_name" validate:"required"`
	Bio              string  `json:"bio"`
	AvatarAttachment *string `json:"avatar_attachment_id"`
}

type UpdateScoutProfileRequest struct {
	OrganizationName *string `json:"organization_name"`
	Bio              *string `json:"bio"`
	AvatarAttachment *string `json:"avatar_attachment_id"`
}

type ScoutProfileResponse struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	OrganizationName string  `json:"organization_name"`
	Bio              string  `json:"bio"`
	AvatarAttachment *string `json:"avatar_attachment_id,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type AddToWatchlistRequest struct {
	AthleteID     string  `json:"athlete_id" validate:"required"`
	Notes         string  `json:"notes"`
	PriorityTagID *string `json:"priority_tag_id"`
}

type UpdateWatchlistEntryRequest struct {
	Notes         *string `json:"notes"`
	PriorityTagID *string `json:"priority_tag_id"`
}

type WatchlistEntryResponse struct {
	ID             string                  `json:"id"`
	ScoutID        string                  `json:"scout_id"`
	AthleteID      string                  `json:"athlete_id"`
	Notes          string                  `json:"notes"`
	PriorityTagID  *string                 `json:"priority_tag_id,omitempty"`
	AddedAt        string                  `json:"added_at"`
	UpdatedAt      string                  `json:"updated_at"`
	AthleteProfile *AthleteProfileResponse `json:"athlete_profile,omitempty"`
}

type AthleteProfileResponse struct {
	ID                      string  `json:"id"`
	AthleteID               string  `json:"athlete_id"`
	SportID                 string  `json:"sport_id"`
	CurrentAcademyID        *string `json:"current_academy_id,omitempty"`
	CurrentAcademyName      string  `json:"current_academy_name"`
	Age                     int32   `json:"age"`
	LeaderboardRank         int32   `json:"leaderboard_rank"`
	LeaderboardScore        float64 `json:"leaderboard_score"`
	HighestCompetitionLevel string  `json:"highest_competition_level"`
	TotalCompetitions       int32   `json:"total_competitions"`
	TotalAcademies          int32   `json:"total_academies"`
	YearsActive             int32   `json:"years_active"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
}

type LeaderboardEntryResponse struct {
	ID             string                  `json:"id"`
	SportID        string                  `json:"sport_id"`
	PeriodTagID    *string                 `json:"period_tag_id,omitempty"`
	Rank           int32                   `json:"rank"`
	AthleteID      string                  `json:"athlete_id"`
	Score          float64                 `json:"score"`
	PreviousScore  float64                 `json:"previous_score"`
	CalculatedAt   string                  `json:"calculated_at"`
	AthleteProfile *AthleteProfileResponse `json:"athlete_profile,omitempty"`
}

type LogScoutActivityRequest struct {
	Action       string  `json:"action" validate:"required"`
	ResourceID   *string `json:"resource_id"`
	ResourceType *string `json:"resource_type"`
}

type ScoutActivityLogResponse struct {
	ID           string  `json:"id"`
	ScoutID      string  `json:"scout_id"`
	Action       string  `json:"action"`
	ResourceID   *string `json:"resource_id,omitempty"`
	ResourceType *string `json:"resource_type,omitempty"`
	CreatedAt    string  `json:"created_at"`
}
