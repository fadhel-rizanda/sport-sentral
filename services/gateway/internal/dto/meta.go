package dto

type StatusSimpleResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type TagSimpleResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type StatusResponse struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Name      string              `json:"name"`
	Slug      string              `json:"slug"`
	CreatedBy UserSimpleResponse  `json:"created_by"`
	UpdatedBy *UserSimpleResponse `json:"updated_by,omitempty"`
	DeletedBy *UserSimpleResponse `json:"deleted_by,omitempty"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
	DeletedAt *string             `json:"deleted_at,omitempty"`
}

type TagResponse struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Name      string              `json:"name"`
	Slug      string              `json:"slug"`
	CreatedBy UserSimpleResponse  `json:"created_by"`
	UpdatedBy *UserSimpleResponse `json:"updated_by,omitempty"`
	DeletedBy *UserSimpleResponse `json:"deleted_by,omitempty"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
	DeletedAt *string             `json:"deleted_at,omitempty"`
}
