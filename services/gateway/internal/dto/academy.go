package dto

type AcademyHoldingAddressResponse struct {
	ID                     string                               `json:"id"`
	StreetAddress          string                               `json:"street_address"`
	Notes                  *string                              `json:"notes,omitempty"`
	Latitude               *float64                             `json:"latitude,omitempty"`
	Longitude              *float64                             `json:"longitude,omitempty"`
	AdministrativeDivision AdministrativeDivisionSimpleResponse `json:"administrative_division"`
}

type AdministrativeDivisionSimpleResponse struct {
	ID         string                                `json:"id"`
	CountryID  string                                `json:"country_id"`
	ParentID   *string                               `json:"parent_id,omitempty"`
	Name       string                                `json:"name"`
	Level      string                                `json:"level"`
	PostalCode string                                `json:"postal_code"`
	Country    *CountrySimpleResponse                `json:"country,omitempty"`
	Parent     *AdministrativeDivisionSimpleResponse `json:"parent,omitempty"`
}

type CountrySimpleResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ISOAlpha2    string `json:"iso_alpha_2"`
	ISOAlpha3    string `json:"iso_alpha_3"`
	PhoneCode    string `json:"phone_code"`
	CurrencyCode string `json:"currency_code"`
}

type AcademyHoldingResponse struct {
	ID                string                         `json:"id"`
	Name              string                         `json:"name"`
	Description       string                         `json:"description"`
	Email             string                         `json:"email"`
	PhoneNumber       string                         `json:"phone_number"`
	ImageAttachmentID *string                        `json:"image_attachment_id,omitempty"`
	Branches          []AcademyBranchSimpleResponse  `json:"branches"`
	Address           *AcademyHoldingAddressResponse `json:"address,omitempty"`
	Status            *StatusSimpleResponse          `json:"status,omitempty"`
	StatusID          string                         `json:"status_id"`
	CreatedAt         string                         `json:"created_at"`
	UpdatedAt         string                         `json:"updated_at"`
	CreatedBy         UserSimpleResponse             `json:"created_by"`
	UpdatedBy         UserSimpleResponse             `json:"updated_by"`
	DeletedBy         *UserSimpleResponse            `json:"deleted_by,omitempty"`
}

type AcademyHoldingSimpleResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Email             string  `json:"email"`
	PhoneNumber       string  `json:"phone_number"`
	ImageAttachmentID *string `json:"image_attachment_id,omitempty"`
	BranchesCount     int64   `json:"branches_count"`
	StatusID          string  `json:"status_id"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type AcademyBranchAddressResponse struct {
	ID                     string                               `json:"id"`
	BranchID               string                               `json:"branch_id"`
	IsPrimary              bool                                 `json:"is_primary"`
	StreetAddress          string                               `json:"street_address"`
	Notes                  *string                              `json:"notes,omitempty"`
	Latitude               *float64                             `json:"latitude,omitempty"`
	Longitude              *float64                             `json:"longitude,omitempty"`
	AdministrativeDivision AdministrativeDivisionSimpleResponse `json:"administrative_division"`
}

type SportSimpleResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Slug             string  `json:"slug"`
	IconAttachmentID *string `json:"icon_attachment_id,omitempty"`
	IsVerified       bool    `json:"is_verified"`
	RegulatorID      *string `json:"regulator_id,omitempty"`
	Tier             string  `json:"tier"`
}

type AcademyBranchResponse struct {
	ID          string                         `json:"id"`
	HoldingID   string                         `json:"holding_id"`
	SportID     string                         `json:"sport_id"`
	SportName   string                         `json:"sport_name"`
	Name        string                         `json:"name"`
	Email       string                         `json:"email"`
	PhoneNumber string                         `json:"phone_number"`
	StatusID    string                         `json:"status_id"`
	MemberCount int64                          `json:"member_count"`
	CreatedAt   string                         `json:"created_at"`
	UpdatedAt   string                         `json:"updated_at"`
	Holding     *AcademyHoldingSimpleResponse  `json:"holding,omitempty"`
	Sport       *SportSimpleResponse           `json:"sport,omitempty"`
	Status      *StatusSimpleResponse          `json:"status,omitempty"`
	Addresses   []AcademyBranchAddressResponse `json:"addresses"`
	CreatedBy   UserSimpleResponse             `json:"created_by"`
	UpdatedBy   UserSimpleResponse             `json:"updated_by"`
	DeletedBy   *UserSimpleResponse            `json:"deleted_by,omitempty"`
}

type AcademyBranchSimpleResponse struct {
	ID          string `json:"id"`
	HoldingID   string `json:"holding_id"`
	SportID     string `json:"sport_id"`
	SportName   string `json:"sport_name"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	StatusID    string `json:"status_id"`
	MemberCount int64  `json:"member_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type AcademyAdminResponse struct {
	ID           string                        `json:"id"`
	AcademyID    string                        `json:"academy_id"`
	BranchID     *string                       `json:"branch_id,omitempty"`
	UserID       string                        `json:"user_id"`
	RoleID       string                        `json:"role_id"`
	ApprovedAt   *string                       `json:"approved_at,omitempty"`
	ApprovedByID *string                       `json:"approved_by_id,omitempty"`
	CreatedAt    string                        `json:"created_at"`
	UpdatedAt    string                        `json:"updated_at"`
	DeletedAt    *string                       `json:"deleted_at,omitempty"`
	Academy      *AcademyHoldingSimpleResponse `json:"academy,omitempty"`
	Branch       *AcademyBranchSimpleResponse  `json:"branch,omitempty"`
	User         UserSimpleResponse            `json:"user"`
	Role         RoleSimpleResponse            `json:"role"`
	ApprovedBy   *UserSimpleResponse           `json:"approved_by,omitempty"`
	CreatedBy    UserSimpleResponse            `json:"created_by"`
	UpdatedBy    UserSimpleResponse            `json:"updated_by"`
	DeletedBy    *UserSimpleResponse           `json:"deleted_by,omitempty"`
}

type EnrollmentResponse struct {
	ID              string                       `json:"id"`
	AcademyBranchID string                       `json:"academy_branch_id"`
	AthleteID       string                       `json:"athlete_id"`
	JoinedAt        string                       `json:"joined_at"`
	LeftAt          *string                      `json:"left_at,omitempty"`
	ExpiresAt       *string                      `json:"expires_at,omitempty"`
	ApprovedAt      *string                      `json:"approved_at,omitempty"`
	ApprovedByID    *string                      `json:"approved_by_id,omitempty"`
	StatusID        string                       `json:"status_id"`
	CreatedAt       string                       `json:"created_at"`
	UpdatedAt       string                       `json:"updated_at"`
	DeletedAt       *string                      `json:"deleted_at,omitempty"`
	AcademyBranch   *AcademyBranchSimpleResponse `json:"academy_branch,omitempty"`
	Athlete         UserSimpleResponse           `json:"athlete"`
	ApprovedBy      *UserSimpleResponse          `json:"approved_by,omitempty"`
	Status          StatusSimpleResponse         `json:"status"`
	CreatedBy       UserSimpleResponse           `json:"created_by"`
	UpdatedBy       UserSimpleResponse           `json:"updated_by"`
	DeletedBy       *UserSimpleResponse          `json:"deleted_by,omitempty"`
}

type RosterResponse struct {
	ID              string                       `json:"id"`
	AcademyBranchID string                       `json:"academy_branch_id"`
	CompetitionID   *string                      `json:"competition_id,omitempty"`
	Name            string                       `json:"name"`
	TagID           string                       `json:"tag_id"`
	StatusID        string                       `json:"status_id"`
	MaxSize         int32                        `json:"max_size"`
	CreatedAt       string                       `json:"created_at"`
	UpdatedAt       string                       `json:"updated_at"`
	DeletedAt       *string                      `json:"deleted_at,omitempty"`
	MemberCount     int64                        `json:"member_count"`
	AcademyBranch   *AcademyBranchSimpleResponse `json:"academy_branch,omitempty"`
	Members         []RosterMemberResponse       `json:"members"`
	Tag             TagSimpleResponse            `json:"tag"`
	Status          StatusSimpleResponse         `json:"status"`
}

type RosterMemberResponse struct {
	ID            string               `json:"id"`
	RosterID      string               `json:"roster_id"`
	AthleteID     string               `json:"athlete_id"`
	JerseyNumber  *int32               `json:"jersey_number,omitempty"`
	PositionID    string               `json:"position_id"`
	StatusID      string               `json:"status_id"`
	AddedAt       string               `json:"added_at"`
	AddedByID     string               `json:"added_by_id"`
	RemovedAt     *string              `json:"removed_at,omitempty"`
	RemovedByID   *string              `json:"removed_by_id,omitempty"`
	RemovalReason *string              `json:"removal_reason,omitempty"`
	Athlete       UserSimpleResponse   `json:"athlete"`
	Position      TagSimpleResponse    `json:"position"`
	Status        StatusSimpleResponse `json:"status"`
	AddedBy       UserSimpleResponse   `json:"added_by"`
	RemovedBy     *UserSimpleResponse  `json:"removed_by,omitempty"`
}
