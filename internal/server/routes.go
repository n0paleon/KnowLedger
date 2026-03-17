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
	internalApiHandler *handler.InternalAPIHandler,
) {
	admin := app.Group("/admin")
	admin.Use(middleware.RequireAuth)

	adminApi := admin.Group("/api")
	adminApi.Use(middleware.RequireAuth)

	auth := app.Group("/auth")
	public := app.Group("/")
	internalApi := app.Group("/internal/api")
	internalApi.Use(internalApiHandler.AuthMiddleware)

	admin.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().Route("Show Fun Facts")
	})
	admin.Get("/facts", adminHandler.ShowFunFacts).Name("Show Fun Facts")
	admin.Get("/facts/create", adminHandler.ShowCreateFunFact).Name("Show Create Fun Fact")
	admin.Post("/facts/create", adminHandler.CreateFunFact).Name("Create Fun Fact")
	admin.Get("/facts/:id/edit", adminHandler.ShowEditFunFact).Name("Show Edit Fun Fact")
	admin.Post("/facts/:id/edit", adminHandler.EditFunFact).Name("Edit Fun Fact")
	admin.Get("/tags", adminHandler.ShowTags).Name("Show Tags")
	admin.Get("/profile", adminHandler.ShowProfile).Name("Show Profile")
	admin.Get("/gc/jobs", adminHandler.ShowGCJobs).Name("Show GC Jobs")
	admin.Get("/gc/jobs/:job_id", adminHandler.ShowGCJobDetails).Name("Show GC Job Details")

	adminApi.Delete("/facts/:id", adminApiHandler.DeleteFunFact).Name("API - Delete One Fun Fact")
	adminApi.Post("/media", adminApiHandler.UploadMedia).Name("API - Upload Media")
	adminApi.Delete("/tags/:id", adminApiHandler.DeleteTag).Name("API - Delete One Tag")
	adminApi.Get("/tags/suggestions", adminApiHandler.GetTagSuggestions).Name("API - Get Tag Suggestions")
	adminApi.Post("/profile/reset-apikey", adminApiHandler.ResetApiKey).Name("API - Reset API key")
	adminApi.Post("/profile/change-password", adminApiHandler.ChangePassword).Name("API - Change Password")
	adminApi.Post("/gc/execute", adminApiHandler.TriggerManualGC).Name("API - Trigger Manual GC")
	adminApi.Get("/gc/jobs/:job_id/logs", adminApiHandler.GetLogs).Name("API - Get Job Logs")

	public.Get("/", publicHandler.PublicShowIndex).Name("Public Index")

	auth.Get("/admin", authHandler.ShowLogin).Name("Show Admin Login")
	auth.Post("/admin", authHandler.Login).Name("Admin Login")
	auth.Get("/logout", authHandler.Logout).Name("Admin Logout")

	internalApi.Post("/upload-media", internalApiHandler.UploadMedia).Name("Internal API - Upload Media")
	internalApi.Post("/facts/create", internalApiHandler.CreateFunFact).Name("Internal API - Create Fun Facts")
}
