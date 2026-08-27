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

type CreateStatusRequest struct {
	Type string `json:"type" validate:"required" example:"ACADEMY"`
	Name string `json:"name" validate:"required" example:"Active"`
	Slug string `json:"slug" validate:"required" example:"active"`
}

type UpdateStatusRequest struct {
	Type *string `json:"type,omitempty" example:"ACADEMY"`
	Name *string `json:"name,omitempty" example:"Suspended"`
	Slug *string `json:"slug,omitempty" example:"suspended"`
}

type CreateTagRequest struct {
	Type string `json:"type" validate:"required" example:"SPORTS"`
	Name string `json:"name" validate:"required" example:"3x3"`
	Slug string `json:"slug" validate:"required" example:"3x3"`
}

type UpdateTagRequest struct {
	Type *string `json:"type,omitempty" example:"SPORTS"`
	Name *string `json:"name,omitempty" example:"5x5 Full Court"`
	Slug *string `json:"slug,omitempty" example:"5x5"`
}
