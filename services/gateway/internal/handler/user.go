package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
	"time"
)

type UserHandler struct {
	userClient userv1.UserServiceClient
}

func NewUserHandler(userClient userv1.UserServiceClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
	}
}

func (h *UserHandler) Routes(router fiber.Router, auth fiber.Handler) {
	me := router.Group("/me", auth)
	me.Put("/", h.UpdateUser)
	me.Delete("/", h.DeleteUser)

	users := router.Group("/users", auth)
	users.Get("/", h.ListUsers)
	users.Get("/:id", h.GetUser)

	//adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	//adminGroup := router.Group("/admin/users", adminMiddlewares...)
	//adminGroup.Get("/", h.ListUsers)
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.userClient.GetUser(context.Background(), &userv1.GetUserRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"user": toUserResponse(resp.User)})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		FullName *string `json:"full_name"`
		Username *string `json:"username"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.userClient.UpdateUser(context.Background(), &userv1.UpdateUserRequest{
		Id:       userID,
		FullName: body.FullName,
		Username: body.Username,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"user": toUserResponse(resp.User)})
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)

	resp, err := h.userClient.ListUsers(context.Background(), &userv1.ListUsersRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return err
	}

	users := make([]UserResponse, len(resp.Users))
	for i, u := range resp.Users {
		users[i] = toUserResponse(u)
	}

	return response.OKWithMeta(c, users, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		Password string `json:"password"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.userClient.DeleteUser(context.Background(), &userv1.DeleteUserRequest{
		Id:       userID,
		Password: body.Password,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "user deleted")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

type UserResponse struct {
	ID             string         `json:"id"`
	Email          string         `json:"email"`
	Username       string         `json:"username"`
	FullName       string         `json:"full_name"`
	StatusName     string         `json:"status_name"`
	StatusID       string         `json:"status_id"`
	ActiveRoleName string         `json:"active_role_name"`
	ActiveRoleID   string         `json:"active_role_id"`
	Roles          []RoleResponse `json:"roles"`
	VerifiedAt     *string        `json:"verified_at"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
	DeletedAt      *string        `json:"deletedAt"`
}

type RoleResponse struct {
	RoleID     string `json:"role_id"`
	RoleName   string `json:"role_name"`
	IsActive   bool   `json:"is_active"`
	StatusName string `json:"status_name"`
	StatusID   string `json:"status_id"`
}

func toUserResponse(u *userv1.User) UserResponse {
	roles := make([]RoleResponse, len(u.Roles))
	for i, p := range u.Roles {
		roles[i] = RoleResponse{
			RoleID:     p.RoleId,
			RoleName:   p.RoleName,
			IsActive:   p.IsActive,
			StatusID:   p.StatusId,
			StatusName: p.StatusName,
		}
	}

	var verifiedAt *string
	if u.VerifiedAt != nil {
		t := u.VerifiedAt.AsTime().UTC().Format(time.RFC3339)
		verifiedAt = &t
	}

	res := UserResponse{
		ID:             u.Id,
		Email:          u.Email,
		Username:       u.Username,
		FullName:       u.FullName,
		StatusName:     u.StatusName,
		StatusID:       u.StatusId,
		ActiveRoleName: u.ActiveRoleName,
		ActiveRoleID:   u.ActiveRoleId,
		Roles:          roles,
		VerifiedAt:     verifiedAt,
		CreatedAt:      u.CreatedAt.AsTime().UTC().Format(time.RFC3339),
		UpdatedAt:      u.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
	}
	if u.DeletedAt != nil {
		formattedDate := u.DeletedAt.AsTime().Format(time.RFC3339)
		res.DeletedAt = &formattedDate
	}
	return res
}
