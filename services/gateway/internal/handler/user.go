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

// GetUser godoc
// @Summary      Get user by ID
// @Description  Retrieve detailed user profile information by user ID (UUID).
// @Tags         Users
// @Produce      json
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.UserResponse}
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /users/{id} [get]
// @Security     BearerAuth
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

// UpdateUser godoc
// @Summary      Update current user profile
// @Description  Update username or full name of the currently logged-in account.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdateUserRequest true "Profile Update Data"
// @Success      200  {object}  response.Response{data=dto.UserResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /me [put]
// @Security     BearerAuth
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body dto.UpdateUserRequest
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

// ListUsers godoc
// @Summary      Get list of users (Admin)
// @Description  Retrieve paginated list of all users (platform admin only).
// @Tags         Users
// @Produce      json
// @Param        page      query int false "Page number (default 1)"
// @Param        page_size query int false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.UserResponse}
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /users [get]
// @Security     BearerAuth
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

// DeleteUser godoc
// @Summary      Delete current user account
// @Description  Soft-delete current user account with password confirmation.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.DeleteUserRequest true "Password Confirmation"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /me [delete]
// @Security     BearerAuth
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body dto.DeleteUserRequest
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

// CreateUser godoc
// @Summary      Create new user (Direct register)
// @Description  Create a new user through the root registration endpoint.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "New User Data"
// @Success      200  {object}  response.Response{data=dto.UserResponse}
// @Failure      400  {object}  response.Response
// @Router       /register [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var body dto.RegisterRequest
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

// SendVerifyEmail godoc
// @Summary      Send verification email (Root alias)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.SendVerifyEmailRequest true "Target Email"
// @Success      200  {object}  response.Response
// @Router       /verify-email [post]
func (h *UserHandler) SendVerifyEmail(c *fiber.Ctx) error {
	var body dto.SendVerifyEmailRequest
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

// VerifyAccount godoc
// @Summary      Verify account (POST Body)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyAccountRequest true "Verification Token"
// @Success      200  {object}  response.Response
// @Router       /verify-account [post]
func (h *UserHandler) VerifyAccount(c *fiber.Ctx) error {
	var body dto.VerifyAccountRequest
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

// ForgotPassword godoc
// @Summary      Forgot password (Root alias)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Account Email"
// @Success      200  {object}  response.Response
// @Router       /forgot-password [post]
func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var body dto.ForgotPasswordRequest
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

// ResetPassword godoc
// @Summary      Reset password (Root alias)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Password Reset Data"
// @Success      200  {object}  response.Response
// @Router       /reset-password [post]
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	var body dto.ResetPasswordRequest
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

// AssignRolesToUser godoc
// @Summary      Assign roles to user (Admin)
// @Description  Assign a list of roles to a specific user account.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.AssignRolesRequest true "Assign Roles Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /users/assign-roles [post]
// @Security     BearerAuth
func (h *UserHandler) AssignRolesToUser(c *fiber.Ctx) error {
	var body dto.AssignRolesRequest
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

// RemoveRolesFromUser godoc
// @Summary      Remove roles from user (Admin)
// @Description  Remove a list of roles from a specific user account.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request body dto.RemoveRolesRequest true "Remove Roles Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /users/remove-roles [post]
// @Security     BearerAuth
func (h *UserHandler) RemoveRolesFromUser(c *fiber.Ctx) error {
	var body dto.RemoveRolesRequest
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
