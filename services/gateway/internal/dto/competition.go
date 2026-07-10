package dto

// ─── COMPETITION DTOS ────────────────────────────────────────────────────────

type CompetitionResponse struct {
	ID                  string                       `json:"id"`
	SportID             string                       `json:"sport_id"`
	HostAcademyBranchID string                       `json:"host_academy_branch_id"`
	Name                string                       `json:"name"`
	Description         string                       `json:"description"`
	TierID              string                       `json:"tier_id"`
	StartDate           string                       `json:"start_date"`
	EndDate             *string                      `json:"end_date,omitempty"`
	StatusID            string                       `json:"status_id"`
	CreatedByID         string                       `json:"created_by_id"`
	UpdatedByID         string                       `json:"updated_by_id"`
	DeletedByID         *string                      `json:"deleted_by_id,omitempty"`
	CreatedAt           string                       `json:"created_at"`
	UpdatedAt           string                       `json:"updated_at"`
	DeletedAt           *string                      `json:"deleted_at,omitempty"`
	Sport               *SportSimpleResponse         `json:"sport,omitempty"`
	HostAcademy         *AcademyBranchSimpleResponse `json:"host_academy,omitempty"`
	TierTag             *TagSimpleResponse           `json:"tier_tag,omitempty"`
	Status              *StatusSimpleResponse        `json:"status,omitempty"`
	CreatedBy           *UserSimpleResponse          `json:"created_by,omitempty"`
	UpdatedBy           *UserSimpleResponse          `json:"updated_by,omitempty"`
	Branches            []CompetitionBranchResponse  `json:"branches,omitempty"`
}

type CreateCompetitionRequest struct {
	SportID             string  `json:"sport_id" validate:"required,uuid"`
	HostAcademyBranchID *string `json:"host_academy_branch_id" validate:"omitempty,uuid"`
	Name                string  `json:"name" validate:"required,min=1"`
	Description         string  `json:"description"`
	TierID              string  `json:"tier_id" validate:"required,uuid"`
	StartDate           string  `json:"start_date" validate:"required"` // RFC3339
	EndDate             *string `json:"end_date"`                       // RFC3339
	StatusID            string  `json:"status_id" validate:"required,uuid"`
}

type UpdateCompetitionRequest struct {
	SportID             *string `json:"sport_id" validate:"omitempty,uuid"`
	HostAcademyBranchID *string `json:"host_academy_branch_id" validate:"omitempty,uuid"`
	Name                *string `json:"name" validate:"omitempty,min=1"`
	Description         *string `json:"description"`
	TierID              *string `json:"tier_id" validate:"omitempty,uuid"`
	StartDate           *string `json:"start_date"` // RFC3339
	EndDate             *string `json:"end_date"`   // RFC3339
	StatusID            *string `json:"status_id" validate:"omitempty,uuid"`
}

// ─── COMPETITION BRANCH DTOS ──────────────────────────────────────────────────

type CompetitionBranchResponse struct {
	ID             string                      `json:"id"`
	CompetitionID  string                      `json:"competition_id"`
	ParentBranchID *string                     `json:"parent_branch_id,omitempty"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	StatusID       string                      `json:"status_id"`
	CreatedByID    string                      `json:"created_by_id"`
	UpdatedByID    string                      `json:"updated_by_id"`
	DeletedByID    *string                     `json:"deleted_by_id,omitempty"`
	CreatedAt      string                      `json:"created_at"`
	UpdatedAt      string                      `json:"updated_at"`
	DeletedAt      *string                     `json:"deleted_at,omitempty"`
	ParentBranch   *CompetitionBranchResponse  `json:"parent_branch,omitempty"`
	Status         *StatusSimpleResponse       `json:"status,omitempty"`
	ChildBranches  []CompetitionBranchResponse `json:"child_branches,omitempty"`
	Matches        []MatchResponse             `json:"matches,omitempty"`
	CreatedBy      *UserSimpleResponse         `json:"created_by,omitempty"`
	UpdatedBy      *UserSimpleResponse         `json:"updated_by,omitempty"`
}

type CreateCompetitionBranchRequest struct {
	CompetitionID  string  `json:"competition_id" validate:"required,uuid"`
	Name           string  `json:"name" validate:"required,min=1"`
	Description    string  `json:"description"`
	StatusID       string  `json:"status_id" validate:"required,uuid"`
	ParentBranchID *string `json:"parent_branch_id" validate:"omitempty,uuid"`
}

type UpdateCompetitionBranchRequest struct {
	Name           *string `json:"name" validate:"omitempty,min=1"`
	Description    *string `json:"description"`
	StatusID       *string `json:"status_id" validate:"omitempty,uuid"`
	ParentBranchID *string `json:"parent_branch_id" validate:"omitempty,uuid"`
}

// ─── MATCH DTOS ──────────────────────────────────────────────────────────────

type MatchParticipantResponse struct {
	ID          string             `json:"id"`
	MatchID     string             `json:"match_id"`
	RosterID    string             `json:"roster_id"`
	FormatTagID string             `json:"format_tag_id"`
	ResultTagID string             `json:"result_tag_id"`
	Score       int32              `json:"score"`
	UpdatedAt   string             `json:"updated_at"`
	Roster      *RosterResponse    `json:"roster,omitempty"`
	FormatTag   *TagSimpleResponse `json:"format_tag,omitempty"`
	ResultTag   *TagSimpleResponse `json:"result_tag,omitempty"`
}

type MatchResponse struct {
	ID           string                     `json:"id"`
	BranchID     string                     `json:"branch_id"`
	ScheduledAt  string                     `json:"scheduled_at"`
	StartedAt    *string                    `json:"started_at,omitempty"`
	EndedAt      *string                    `json:"ended_at,omitempty"`
	Status       string                     `json:"status"`
	Location     string                     `json:"location"`
	Referee      string                     `json:"referee"`
	Notes        string                     `json:"notes"`
	CreatedAt    string                     `json:"created_at"`
	UpdatedAt    string                     `json:"updated_at"`
	DeletedAt    *string                    `json:"deleted_at,omitempty"`
	Participants []MatchParticipantResponse `json:"participants,omitempty"`
}

type CreateMatchParticipantRequest struct {
	RosterID    string `json:"roster_id" validate:"required,uuid"`
	FormatTagID string `json:"format_tag_id" validate:"required,uuid"`
	ResultTagID string `json:"result_tag_id" validate:"required,uuid"`
	Score       int32  `json:"score"`
}

type CreateMatchRequest struct {
	BranchID     string                          `json:"branch_id" validate:"required,uuid"`
	ScheduledAt  string                          `json:"scheduled_at" validate:"required"` // RFC3339
	Location     string                          `json:"location"`
	Referee      string                          `json:"referee"`
	Notes        string                          `json:"notes"`
	Participants []CreateMatchParticipantRequest `json:"participants" validate:"required,min=1,dive"`
}

type UpdateParticipantScoreRequest struct {
	ParticipantID string `json:"participant_id" validate:"required,uuid"`
	Score         int32  `json:"score"`
}

type UpdateMatchScoreRequest struct {
	Participants []UpdateParticipantScoreRequest `json:"participants" validate:"required,min=1,dive"`
}

// ─── STATS DTOS ──────────────────────────────────────────────────────────────

type MatchStatResponse struct {
	ID           string              `json:"id"`
	MatchID      string              `json:"match_id"`
	AthleteID    string              `json:"athlete_id"`
	StatTypeID   string              `json:"stat_type_id"`
	Value        float64             `json:"value"`
	RecordedByID string              `json:"recorded_by_id"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	Athlete      *UserSimpleResponse `json:"athlete,omitempty"`
	StatType     *TagSimpleResponse  `json:"stat_type,omitempty"`
	RecordedBy   *UserSimpleResponse `json:"recorded_by,omitempty"`
}

type AthleteStatsAggregateResponse struct {
	ID            string              `json:"id"`
	CompetitionID string              `json:"competition_id"`
	BranchID      string              `json:"branch_id"`
	AthleteID     string              `json:"athlete_id"`
	StatTypeID    string              `json:"stat_type_id"`
	TotalMatches  int32               `json:"total_matches"`
	SumValue      float64             `json:"sum_value"`
	AvgValue      float64             `json:"avg_value"`
	MaxValue      float64             `json:"max_value"`
	MinValue      float64             `json:"min_value"`
	LastUpdatedAt string              `json:"last_updated_at"`
	Athlete       *UserSimpleResponse `json:"athlete,omitempty"`
	StatType      *TagSimpleResponse  `json:"stat_type,omitempty"`
}

type RecordMatchStatRequest struct {
	MatchID    string  `json:"match_id" validate:"required,uuid"`
	AthleteID  string  `json:"athlete_id" validate:"required,uuid"`
	StatTypeID string  `json:"stat_type_id" validate:"required,uuid"`
	Value      float64 `json:"value"`
}

type UpdateMatchStatRequest struct {
	Value float64 `json:"value"`
}
