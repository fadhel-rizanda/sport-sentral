package handler

import (
	"github.com/gofiber/fiber/v2"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type UserHandler struct {
	userClient userv1.UserServiceClient
}

func NewUserHandler(userClient userv1.UserServiceClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
	}
}

func (h *UserHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	router.Post("/register", h.CreateUser)
	router.Post("/verify-email", h.SendVerifyEmail)
	router.Post("/verify-account", h.VerifyAccount)
	router.Post("/forgot-password", h.ForgotPassword)
	router.Post("/reset-password", h.ResetPassword)

	me := router.Group("/me", auth)
	me.Put("/", h.UpdateUser)
	me.Delete("/", h.DeleteUser)

	// TODO RAPIHIN MIDDLEWARE
	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminGroup := router.Group("/users", adminMiddlewares...)

	adminGroup.Get("/", h.ListUsers)
	adminGroup.Get("/:id", h.GetUser)

	adminGroup.Post("/assign-roles", h.AssignRolesToUser)
	adminGroup.Post("/remove-roles", h.RemoveRolesFromUser)
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.userClient.GetUser(c.Context(), &userv1.GetUserRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"user": mapper.ToUserResponse(resp.User)})
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

	resp, err := h.userClient.UpdateUser(c.Context(), &userv1.UpdateUserRequest{
		Id:       userID,
		FullName: body.FullName,
		Username: body.Username,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"user": mapper.ToUserResponse(resp.User)})
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)

	resp, err := h.userClient.ListUsers(c.Context(), &userv1.ListUsersRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return err
	}

	users := make([]dto.UserResponse, len(resp.Users))
	for i, u := range resp.Users {
		users[i] = mapper.ToUserResponse(u)
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

	_, err := h.userClient.DeleteUser(c.Context(), &userv1.DeleteUserRequest{
		Id:       userID,
		Password: body.Password,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "user deleted")
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
		RoleID   string `json:"role_id"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}
	resp, err := h.userClient.CreateUser(c.Context(), &userv1.CreateUserRequest{
		Email:    body.Email,
		Username: body.Username,
		FullName: body.FullName,
		Password: body.Password,
		RoleId:   body.RoleID,
	})
	if err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"user": mapper.ToUserResponse(resp.User)})
}

func (h *UserHandler) SendVerifyEmail(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}
	_, err := h.userClient.SendVerifyEmail(c.Context(), &userv1.SendVerifyEmailRequest{
		Email: body.Email,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "verification email sent")
}

func (h *UserHandler) VerifyAccount(c *fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}
	_, err := h.userClient.VerifyAccount(c.Context(), &userv1.VerifyAccountRequest{
		Token: body.Token,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "account verified")
}

func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}
	_, err := h.userClient.ForgotPassword(c.Context(), &userv1.ForgotPasswordRequest{
		Email: body.Email,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "Forgot password email sent")
}

func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}
	_, err := h.userClient.ResetPassword(c.Context(), &userv1.ResetPasswordRequest{
		Token:    body.Token,
		Password: body.Password,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "password changed")
}

func (h *UserHandler) AssignRolesToUser(c *fiber.Ctx) error {
	var body struct {
		UserID string   `json:"user_id"`
		Roles  []string `json:"roles"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}
	_, err := h.userClient.AssignRolesToUser(c.Context(), &userv1.AssignRolesToUserRequest{
		UserId:  body.UserID,
		RoleIds: body.Roles,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "assigned roles to user")
}

func (h *UserHandler) RemoveRolesFromUser(c *fiber.Ctx) error {
	var body struct {
		UserID string   `json:"user_id"`
		Roles  []string `json:"roles"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}
	_, err := h.userClient.RemoveRolesFromUser(c.Context(), &userv1.RemoveRolesFromUserRequest{
		UserId:  body.UserID,
		RoleIds: body.Roles,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "removed roles from user")
}
