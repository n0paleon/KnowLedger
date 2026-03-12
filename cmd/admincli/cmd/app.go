package cmd

import "KnowLedger/internal/repository"

type App struct {
	AdminRepo *repository.AdminRepository
}

var app *App
