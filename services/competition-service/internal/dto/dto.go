package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Common Simple Responses ──────────────────────────────────────────────────

type UserSimpleResponse struct {
	ID       uuid.UUID
	Email    string
	Username string
	FullName string
}

type StatusSimpleResponse struct {
	ID   uuid.UUID
	Type string
	Name string
	Slug string
}

type TagSimpleResponse struct {
	ID   uuid.UUID
	Type string
	Name string
	Slug string
}

type SportSimpleResponse struct {
	ID               uuid.UUID
	Name             string
	Slug             string
	IconAttachmentID *uuid.UUID
	IsVerified       bool
	RegulatorID      *uuid.UUID
	Tier             string
}

type AcademyHoldingSimpleResponse struct {
	ID                string
	Name              string
	Description       string
	Email             string
	PhoneNumber       string
	ImageAttachmentID *string
	BranchesCount     int64
	StatusID          string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type AcademyBranchSimpleResponse struct {
	ID          string
	HoldingID   string
	SportID     string
	SportName   string
	Name        string
	Email       string
	PhoneNumber string
	StatusID    string
	MemberCount int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ─── Roster ───────────────────────────────────────────────────────────────────

type CreateRosterRequest struct {
	AcademyBranchID uuid.UUID
	CompetitionID   uuid.UUID
	Name            string
	TagID           uuid.UUID
	StatusID        uuid.UUID
	MaxSize         int
}

type UpdateRosterRequest struct {
	CompetitionID *uuid.UUID
	Name          *string
	TagID         *uuid.UUID
	StatusID      *uuid.UUID
	MaxSize       *int
}

type RosterResponse struct {
	ID              uuid.UUID
	AcademyBranchID uuid.UUID
	CompetitionID   uuid.UUID
	Name            string
	TagID           uuid.UUID
	StatusID        uuid.UUID
	MaxSize         int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	MemberCount     int64

	AcademyBranch *AcademyBranchSimpleResponse
	Members       []RosterMemberResponse
	Tag           TagSimpleResponse
	Status        StatusSimpleResponse
}

type ListRostersRequest struct {
	CompetitionID *uuid.UUID
	BranchID      *uuid.UUID
	StatusID      *uuid.UUID
	TagID         *uuid.UUID
	Search        *string
	Page          int
	PageSize      int
}

type ListRostersResponse struct {
	Rosters  []*RosterResponse
	Total    int64
	Page     int
	PageSize int
}

type DeleteRosterRequest struct {
	ID          uuid.UUID
	DeletedByID uuid.UUID
}

// ─── Roster Member ────────────────────────────────────────────────────────────

type AddRosterMemberRequest struct {
	RosterID     uuid.UUID
	AthleteID    uuid.UUID
	JerseyNumber *int
	PositionID   uuid.UUID
	StatusID     uuid.UUID
	AddedByID    uuid.UUID
}

type RemoveRosterMemberRequest struct {
	RemovedByID   uuid.UUID
	RemovalReason string
}

type RosterMemberResponse struct {
	ID            uuid.UUID
	RosterID      uuid.UUID
	AthleteID     uuid.UUID
	JerseyNumber  *int
	PositionID    uuid.UUID
	StatusID      uuid.UUID
	AddedAt       time.Time
	AddedByID     uuid.UUID
	RemovedAt     *time.Time
	RemovedByID   *uuid.UUID
	RemovalReason *string

	Athlete   UserSimpleResponse
	Position  TagSimpleResponse
	Status    StatusSimpleResponse
	AddedBy   UserSimpleResponse
	RemovedBy *UserSimpleResponse
}

// ─── Competition ─────────────────────────────────────────────────────────────

type CreateCompetitionRequest struct {
	SportID             uuid.UUID
	HostAcademyBranchID *uuid.UUID
	Name                string
	Description         string
	TierID              uuid.UUID
	StartDate           time.Time
	EndDate             *time.Time
	StatusID            uuid.UUID
}

type UpdateCompetitionRequest struct {
	SportID             *uuid.UUID
	HostAcademyBranchID *uuid.UUID
	Name                *string
	Description         *string
	TierID              *uuid.UUID
	StartDate           *time.Time
	EndDate             *time.Time
	StatusID            *uuid.UUID
}

type CompetitionResponse struct {
	ID                  uuid.UUID
	SportID             uuid.UUID
	HostAcademyBranchID *uuid.UUID
	Name                string
	Description         string
	TierID              uuid.UUID
	StartDate           time.Time
	EndDate             *time.Time
	StatusID            uuid.UUID
	CreatedByID         uuid.UUID
	UpdatedByID         uuid.UUID
	DeletedByID         *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time

	Sport       *SportSimpleResponse
	HostAcademy *AcademyBranchSimpleResponse
	TierTag     *TagSimpleResponse
	Status      *StatusSimpleResponse
	CreatedBy   *UserSimpleResponse
	UpdatedBy   *UserSimpleResponse
	Branches    []BranchResponse
}

type CompetitionFilters struct {
	BranchID *uuid.UUID
	SportID  *uuid.UUID
	TierID   *uuid.UUID
	StatusID *uuid.UUID
	Search   *string
}

// ─── Competition Branch ──────────────────────────────────────────────────────

type CreateBranchRequest struct {
	Name           string
	Description    string
	StatusID       uuid.UUID
	ParentBranchID *uuid.UUID
}

type UpdateBranchRequest struct {
	Name           *string
	Description    *string
	StatusID       *uuid.UUID
	ParentBranchID *uuid.UUID
}

type BranchResponse struct {
	ID             uuid.UUID
	CompetitionID  uuid.UUID
	ParentBranchID *uuid.UUID
	Name           string
	Description    string
	StatusID       uuid.UUID
	CreatedByID    uuid.UUID
	UpdatedByID    uuid.UUID
	DeletedByID    *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time

	ParentBranch  *BranchResponse
	Status        *StatusSimpleResponse
	ChildBranches []BranchResponse
	Matches       []MatchResponse
	CreatedBy     *UserSimpleResponse
	UpdatedBy     *UserSimpleResponse
}

// ─── Match ───────────────────────────────────────────────────────────────────

type CreateMatchRequest struct {
	ScheduledAt  time.Time
	Location     string
	Referee      string
	Notes        string
	Participants []CreateMatchParticipantRequest
}

type CreateMatchParticipantRequest struct {
	RosterID    uuid.UUID
	FormatTagID uuid.UUID
	ResultTagID uuid.UUID
	Score       int32
}

type UpdateScoreRequest struct {
	Participants []UpdateParticipantScoreRequest
}

type UpdateParticipantScoreRequest struct {
	ParticipantID uuid.UUID
	Score         int32
}

type MatchResponse struct {
	ID          uuid.UUID
	BranchID    uuid.UUID
	ScheduledAt time.Time
	StartedAt   *time.Time
	EndedAt     *time.Time
	Status      string
	Location    string
	Referee     string
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	Branch       *BranchResponse
	Participants []MatchParticipantResponse
}

type MatchParticipantResponse struct {
	ID          uuid.UUID
	MatchID     uuid.UUID
	RosterID    uuid.UUID
	FormatTagID uuid.UUID
	ResultTagID uuid.UUID
	Score       int32
	UpdatedAt   time.Time

	Roster    *RosterResponse
	FormatTag *TagSimpleResponse
	ResultTag *TagSimpleResponse
}

// ─── Stat ─────────────────────────────────────────────────────────────────────

type RecordStatRequest struct {
	AthleteID  uuid.UUID
	StatTypeID uuid.UUID
	Value      float64
}

type MatchStatResponse struct {
	ID           uuid.UUID
	MatchID      uuid.UUID
	AthleteID    uuid.UUID
	StatTypeID   uuid.UUID
	Value        float64
	RecordedByID uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Athlete    *UserSimpleResponse
	StatType   *TagSimpleResponse
	RecordedBy *UserSimpleResponse
}

type AggregateResponse struct {
	ID            uuid.UUID
	CompetitionID uuid.UUID
	BranchID      uuid.UUID
	AthleteID     uuid.UUID
	StatTypeID    uuid.UUID
	TotalMatches  int32
	AvgValue      float64
	SumValue      float64
	MaxValue      float64
	MinValue      float64
	LastUpdatedAt time.Time

	Athlete  *UserSimpleResponse
	StatType *TagSimpleResponse
}
