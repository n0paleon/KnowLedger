/*
Copyright © 2026 n0paleon <nopaleon@proton.me>
*/
package cmd

import (
	"KnowLedger/internal/database"
	"KnowLedger/internal/repository"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var envFile string // flag --env

var rootCmd = &cobra.Command{
	Use:   "admincli",
	Short: "KnowLedger Admin CLI",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := godotenv.Load(envFile); err != nil {
			return fmt.Errorf("failed to load env file '%s': %w", envFile, err)
		}

		db, err := database.Connect(os.Getenv("DATABASE_DSN"))
		if err != nil {
			return fmt.Errorf("database connection error: %w", err)
		}

		app = &App{
			AdminRepo: repository.NewAdminRepository(db),
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&envFile, "env", ".env", "Path ke env config file")
}
