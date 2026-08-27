package handler

import (
	"github.com/gofiber/fiber/v2"
	authv1 "microservice-golang/gen/auth/v1"
	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
	"microservice-golang/services/gateway/internal/response"
)

type AuthHandler struct {
	authClient authv1.AuthServiceClient
	userClient userv1.UserServiceClient
}

func NewAuthHandler(
	authClient authv1.AuthServiceClient,
	userClient userv1.UserServiceClient,
) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
		userClient: userClient,
	}
}

func (h *AuthHandler) Routes(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/logout", h.Logout)
	auth.Post("/refresh", h.RefreshToken)
	auth.Post("/verify-email", h.SendVerifyEmail)
	auth.Get("/verify", h.VerifyAccount)
	auth.Post("/forgot-password", h.ForgotPassword)
	auth.Post("/reset-password", h.ResetPassword)
}

func (h *AuthHandler) MeRoutes(router fiber.Router, middlewares ...fiber.Handler) {
	me := router.Group("/me", middlewares...)
	me.Get("/", h.Me)
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new account with an initial role (athlete, scout, court_owner, academy_admin).
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "User Registration Data"
// @Success      201  {object}  response.Response{data=dto.UserResponse}
// @Failure      400  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
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

	return response.Created(c, fiber.Map{"user": mapper.ToUserResponse(resp.User)})
}

// Login godoc
// @Summary      User authentication & login
// @Description  Login using email & password, returning JWT access & refresh tokens and active role.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login Credentials"
// @Success      200  {object}  response.Response{data=dto.LoginResponseData}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body dto.LoginRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.authClient.Login(c.Context(), &authv1.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{
		"tokens": resp.Tokens,
		"user": dto.UserSimpleResponse{
			ID:       resp.User.Id,
			Email:    resp.User.Email,
			Username: resp.User.Username,
			FullName: resp.User.FullName,
		},
		"status": dto.StatusSimpleResponse{
			ID:   resp.Status.Id,
			Name: resp.Status.Name,
			Slug: resp.Status.Slug,
			Type: resp.Status.Type,
		},
		"active_role": dto.RoleSimpleResponse{
			ID:            resp.ActiveRole.Id,
			Name:          resp.ActiveRole.Name,
			Slug:          resp.ActiveRole.Slug,
			PermissionIDs: resp.ActiveRole.PermissionIds,
		},
	})
}

// Logout godoc
// @Summary      Logout from active session
// @Description  Revoke and blacklist the user's refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LogoutRequest true "Refresh token to revoke"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var body dto.LogoutRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.authClient.Logout(c.Context(), &authv1.LogoutRequest{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "logged out")
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Obtain a new access token using a valid refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshTokenRequest true "Refresh Token"
// @Success      200  {object}  response.Response{data=dto.TokenPairResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var body dto.RefreshTokenRequest
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.authClient.RefreshToken(c.Context(), &authv1.RefreshTokenRequest{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{
		"tokens": resp.Tokens,
	})
}

// SendVerifyEmail godoc
// @Summary      Resend verification email
// @Description  Send account verification link/token to user's email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.SendVerifyEmailRequest true "Target Email"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /auth/verify-email [post]
func (h *AuthHandler) SendVerifyEmail(c *fiber.Ctx) error {
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
// @Summary      Verify user account
// @Description  Verify email and activate account using verification token.
// @Tags         Auth
// @Produce      json
// @Param        token query string true "Account Verification Token"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /auth/verify [get]
func (h *AuthHandler) VerifyAccount(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return response.Error(c, fiber.StatusBadRequest, "verification token is required")
	}

	_, err := h.userClient.VerifyAccount(c.Context(), &userv1.VerifyAccountRequest{
		Token: token,
	})
	if err != nil {
		return err
	}

	return response.OKWithMessage(c, "account verified")
}

// ForgotPassword godoc
// @Summary      Forgot password request
// @Description  Send password reset instructions & token to registered email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Account Email"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
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

	return response.OKWithMessage(c, "reset password email sent")
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Reset user password using a valid password reset token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Password Reset Data"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
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

	return response.OKWithMessage(c, "password reset successful")
}

// ─── Me ───────────────────────────────────────────────────────────────────────

// Me godoc
// @Summary      Get current logged-in user profile
// @Description  Retrieve profile data, status, and active role of the currently authenticated user.
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  response.Response{data=dto.UserResponse}
// @Failure      401  {object}  response.Response
// @Router       /me [get]
// @Security     BearerAuth
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	resp, err := h.userClient.GetUser(c.Context(), &userv1.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{"user": mapper.ToUserResponse(resp.User)})
}
