package router

import (
	"github.com/gofiber/fiber/v2"
	authv1 "microservice-golang/gen/auth/v1"
	"microservice-golang/services/gateway/internal/handler"
	"microservice-golang/services/gateway/internal/middleware"
)

func Setup(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	profileHandler *handler.ProfileHandler,
	authClient authv1.AuthServiceClient,
	statusHandler *handler.StatusHandler,
	tagHandler *handler.TagHandler,
	roleHandler *handler.RoleHandler,
	permissionHandler *handler.PermissionHandler,
	academyHandler *handler.AcademyHandler,
	sportHandler *handler.SportHandler,
	competitionHandler *handler.CompetitionHandler,
) {
	api := app.Group("/api/v1")

	// define middleware sekali di sini
	auth := middleware.Auth(authClient)
	adminOnly := middleware.RequireRole("platform_admin")

	authHandler.Routes(api)
	userHandler.Routes(api, auth)
	authHandler.MeRoutes(api, auth)
	profileHandler.Routes(api, auth, adminOnly)
	statusHandler.Routes(api, auth, adminOnly)
	tagHandler.Routes(api, auth, adminOnly)
	roleHandler.Routes(api, auth, adminOnly)
	permissionHandler.Routes(api, auth, adminOnly)
	academyHandler.Routes(api, auth, adminOnly)
	sportHandler.Routes(api, auth, adminOnly)
	competitionHandler.Routes(api, auth, adminOnly)
}
