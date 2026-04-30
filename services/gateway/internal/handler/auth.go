package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	authv1 "microservice-golang/gen/auth/v1"
	userv1 "microservice-golang/gen/user/v1"
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

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
		RoleName string `json:"role_name"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.userClient.CreateUser(context.Background(), &userv1.CreateUserRequest{
		Email:    body.Email,
		Username: body.Username,
		FullName: body.FullName,
		Password: body.Password,
		RoleName: body.RoleName,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.Created(c, fiber.Map{"user": toUserResponse(resp.User)})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.authClient.Login(context.Background(), &authv1.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OK(c, fiber.Map{
		"tokens":         resp.Tokens,
		"user_id":        resp.UserId,
		"email":          resp.Email,
		"username":       resp.Username,
		"active_profile": resp.ActiveProfile,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.authClient.Logout(context.Background(), &authv1.LogoutRequest{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "logged out")
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	resp, err := h.authClient.RefreshToken(context.Background(), &authv1.RefreshTokenRequest{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OK(c, fiber.Map{
		"tokens": resp.Tokens,
	})
}

func (h *AuthHandler) SendVerifyEmail(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.userClient.SendVerifyEmail(context.Background(), &userv1.SendVerifyEmailRequest{
		Email: body.Email,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "verification email sent")
}

func (h *AuthHandler) VerifyAccount(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return response.Error(c, fiber.StatusBadRequest, "verification token is required")
	}

	_, err := h.userClient.VerifyAccount(context.Background(), &userv1.VerifyAccountRequest{
		Token: token,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "account verified")
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.userClient.ForgotPassword(context.Background(), &userv1.ForgotPasswordRequest{
		Email: body.Email,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "reset password email sent")
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := request.Parse(c, &body); err != nil {
		return err
	}

	_, err := h.userClient.ResetPassword(context.Background(), &userv1.ResetPasswordRequest{
		Token:    body.Token,
		Password: body.Password,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OKWithMessage(c, "password reset successful")
}

// ─── Me ───────────────────────────────────────────────────────────────────────

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	resp, err := h.userClient.GetUser(context.Background(), &userv1.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return grpcError(c, err)
	}

	return response.OK(c, fiber.Map{"user": toUserResponse(resp.User)})
}
