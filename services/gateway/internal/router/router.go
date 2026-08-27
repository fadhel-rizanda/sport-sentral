package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	authv1 "microservice-golang/gen/auth/v1"
	_ "microservice-golang/services/gateway/docs"
	"microservice-golang/services/gateway/internal/handler"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/shared/pkg/jwt"
)

func Setup(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	profileHandler *handler.ProfileHandler,
	jwtManager *jwt.Manager,
	authClient authv1.AuthServiceClient,
	statusHandler *handler.StatusHandler,
	tagHandler *handler.TagHandler,
	roleHandler *handler.RoleHandler,
	permissionHandler *handler.PermissionHandler,
	academyHandler *handler.AcademyHandler,
	sportHandler *handler.SportHandler,
	competitionHandler *handler.CompetitionHandler,
	venueHandler *handler.VenueHandler,
	scoutHandler *handler.ScoutHandler,
	logHandler *handler.LogHandler,
	attachmentHandler *handler.AttachmentHandler,
) {
	// Swagger UI Route
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api/v1")

	api.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Welcome to Sport Sentral V1 API Gateway")
	})

	// define middleware sekali di sini
	auth := middleware.Auth(jwtManager, authClient)
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
	venueHandler.Routes(api, auth, adminOnly)
	scoutHandler.Routes(api, auth, adminOnly)
	logHandler.Routes(api, auth, adminOnly)
	attachmentHandler.Routes(api, auth, adminOnly)
}
