package server

import (
	"KnowLedger/internal/server/handler"
	"KnowLedger/internal/server/middleware"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(
	app *fiber.App,
	adminHandler *handler.AdminHandler,
	adminApiHandler *handler.AdminApiHandler,
	publicHandler *handler.PublicHandler,
	authHandler *handler.AuthHandler,
) {
	admin := app.Group("/admin")
	admin.Use(middleware.RequireAuth)

	adminApi := admin.Group("/api")
	adminApi.Use(middleware.RequireAuth)

	auth := app.Group("/auth")
	public := app.Group("/")

	admin.Get("/facts", adminHandler.ShowFunFacts).Name("Show Fun Facts")
	admin.Get("/facts/create", adminHandler.ShowCreateFunFact).Name("Show Create Fun Fact")
	admin.Post("/facts/create", adminHandler.CreateFunFact).Name("Create Fun Fact")
	admin.Get("/facts/:id/edit", adminHandler.ShowEditFunFact).Name("Show Edit Fun Fact")
	admin.Post("/facts/:id/edit", adminHandler.EditFunFact).Name("Edit Fun Fact")
	admin.Get("/tags", adminHandler.ShowTags).Name("Show Tags")

	adminApi.Delete("/facts/:id", adminApiHandler.DeleteFunFact).Name("API - Delete One Fun Fact")
	adminApi.Post("/media", adminApiHandler.UploadMedia).Name("API - Upload Media")
	adminApi.Delete("/tags/:id", adminApiHandler.DeleteTag).Name("API - Delete One Tag")

	public.Get("/", publicHandler.PublicShowIndex).Name("Public Index")

	auth.Get("/admin", authHandler.ShowLogin).Name("Show Admin Login")
	auth.Post("/admin", authHandler.Login).Name("Admin Login")
}
