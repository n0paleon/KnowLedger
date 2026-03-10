package server

import (
	"KnowLedger/internal/server/handler"
	"KnowLedger/internal/server/middleware"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/memory/v2"
)

func SetupRoutes(
	app *fiber.App,
	adminHandler *handler.AdminHandler,
	adminApiHandler *handler.AdminApiHandler,
	publicHandler *handler.PublicHandler,
	authHandler *handler.AuthHandler,
) {
	setupSessionStorage(app) // setup storage for sessions

	admin := app.Group("/admin")
	admin.Use(middleware.RequireAuth)

	adminApi := admin.Group("/api")
	adminApi.Use(middleware.RequireAuth)

	auth := app.Group("/auth")
	public := app.Group("/")

	admin.Get("/", adminHandler.ShowDashboardIndex).Name("Show Admin Dashboard")
	admin.Get("/create", adminHandler.ShowCreateFunFact).Name("Show Create Fun Fact")
	admin.Post("/create", adminHandler.CreateFunFact).Name("Create Fun Fact")
	admin.Get("/facts/:id/edit", adminHandler.ShowEditFunFact).Name("Show Edit Fun Fact")
	admin.Post("/facts/:id/edit", adminHandler.EditFunFact).Name("Edit Fun Fact")

	adminApi.Delete("/facts/:id", adminApiHandler.DeleteFunFact).Name("API - Delete One Fun Fact")
	adminApi.Post("/media", adminApiHandler.UploadMedia).Name("API - Upload Media")

	public.Get("/", publicHandler.PublicShowIndex).Name("Public Index")

	auth.Get("/admin", authHandler.ShowLogin).Name("Show Admin Login")
	auth.Post("/admin", authHandler.Login).Name("Admin Login")
}

func setupSessionStorage(app *fiber.App) {
	app.Use(session.New(session.Config{
		Storage: memory.New(memory.Config{
			GCInterval: 1 * time.Minute,
		}),
		CookieSecure:    true,           // HTTPS only
		CookieHTTPOnly:  true,           // Prevent XSS
		CookieSameSite:  "Lax",          // CSRF protection
		IdleTimeout:     8 * time.Hour,  // Session timeout, after N-minute of inactivity, session will be auto expire
		AbsoluteTimeout: 48 * time.Hour, // Maximum session life, force expire after N-hours regardless of activity
		Extractor:       extractors.FromCookie("__Host-session_id"),
	}))
}
