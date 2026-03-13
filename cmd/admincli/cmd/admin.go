/*
Copyright © 2026 n0paleon <nopaleon@proton.me>
*/
package cmd

import (
	"KnowLedger/cmd/admincli/helper"
	"KnowLedger/internal/model"
	"KnowLedger/pkg/utils"
	"context"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// adminCmd represents the admin command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin Management commands",
}

// admincli admin list
var adminListCmd = &cobra.Command{
	Use:   "list",
	Short: "Get all admins",
	RunE: func(cmd *cobra.Command, args []string) error {
		spinner, _ := pterm.DefaultSpinner.Start("Getting all admins")

		admins, err := app.AdminRepo.FindAll(context.Background())
		if err != nil {
			spinner.Fail("Failed to get all admins")
			return err
		}

		spinner.Success("Found all admins")
		helper.PrintAdmins(admins)
		pterm.Println()

		return nil
	},
}

var adminCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new admin account",
	RunE: func(cmd *cobra.Command, args []string) error {
		usernameInput, _ := pterm.DefaultInteractiveTextInput.Show("Username")

		pterm.Println()

		passwordInput, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Password")

		spinner, _ := pterm.DefaultSpinner.Start("Creating new admin account")

		hashedPassword, _ := utils.GeneratePasswordHash(passwordInput)
		admin, err := app.AdminRepo.Create(context.Background(), usernameInput, hashedPassword)
		if err != nil {
			spinner.Fail("Failed to create admin account")
			return err
		}

		spinner.Success("Created new admin account")
		helper.PrintAdmins([]model.Admin{*admin})
		pterm.Println()

		return nil
	},
}

var adminDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete admin account",
	RunE: func(cmd *cobra.Command, args []string) error {
		usernameInput, _ := pterm.DefaultInteractiveTextInput.Show("Admin Username")
		pterm.Println()

		spinner, _ := pterm.DefaultSpinner.Start("Searching admin account")
		admin, err := app.AdminRepo.FindByUsername(context.Background(), usernameInput)
		if err != nil {
			spinner.Fail(fmt.Sprintf("Username %s not found", usernameInput))
			return err
		}

		spinner.Info("Username found")
		helper.PrintAdmins([]model.Admin{*admin})
		pterm.Println()

		confirm, _ := pterm.DefaultInteractiveContinue.Show("Are you sure you want to delete this admin account? [Y/n]")
		pterm.Println()

		confirm = strings.ToLower(confirm)
		if confirm != "y" && confirm != "yes" {
			spinner.Success("Action cancelled")
			return nil
		}

		if err := app.AdminRepo.Delete(context.Background(), model.Admin{Username: usernameInput}); err != nil {
			spinner.Fail("Failed to delete admin account")
			return err
		}

		spinner.Success("Admin account deleted successfully")
		pterm.Println()

		return nil
	},
}

var adminUpdatePasswordCmd = &cobra.Command{
	Use:   "update-password",
	Short: "Update admin account password",
	RunE: func(cmd *cobra.Command, args []string) error {
		usernameInput, _ := pterm.DefaultInteractiveTextInput.Show("Account username")
		pterm.Println()

		spinner, _ := pterm.DefaultSpinner.Start("Searching admin account")

		admin, err := app.AdminRepo.FindByUsername(context.Background(), usernameInput)
		if err != nil || admin == nil || admin.Username != usernameInput {
			spinner.Fail(fmt.Sprintf("Username %s not found", usernameInput))
			return err
		}

		spinner.Info("Account found")
		helper.PrintAdmins([]model.Admin{*admin})
		pterm.Println()

		newPassword, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("New password")
		newPasswordConfirm, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Confirm new password")
		pterm.Println()

		spinner.Info("Updating admin password")
		if newPassword != newPasswordConfirm {
			spinner.Fail("New passwords do not match")
			return nil
		}

		hashedPassword, _ := utils.GeneratePasswordHash(newPassword)
		err = app.AdminRepo.UpdatePasswordByUsername(context.Background(), usernameInput, hashedPassword)
		if err != nil {
			spinner.Fail("Failed to update admin password")
			return err
		}

		spinner.Success("Admin account updated successfully")
		pterm.Println()

		return nil
	},
}

var adminResetPasswordCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Reset admin account password",
	RunE: func(cmd *cobra.Command, args []string) error {
		usernameInput, _ := pterm.DefaultInteractiveTextInput.Show("Account username")
		pterm.Println()

		spinner, _ := pterm.DefaultSpinner.Start("Searching admin account")

		admin, err := app.AdminRepo.FindByUsername(context.Background(), usernameInput)
		if err != nil || admin == nil || admin.Username != usernameInput {
			spinner.Fail(fmt.Sprintf("Username %s not found", usernameInput))
			return err
		}

		spinner.Info("Account found")
		pterm.Println()

		newPassword, err := helper.GenerateRandomPassword(16)
		if err != nil {
			spinner.Fail("Failed to generate new password")
			return err
		}

		spinner.Info("Updating admin password")
		pterm.Println()

		hashedPassword, _ := utils.GeneratePasswordHash(newPassword)
		err = app.AdminRepo.UpdatePasswordByUsername(context.Background(), usernameInput, hashedPassword)
		if err != nil {
			spinner.Fail("Failed to update admin password")
			return err
		}

		admin.Password = newPassword
		helper.PrintAdminsWithRawPassword([]model.Admin{*admin})
		pterm.Println()

		spinner.Success("Admin account updated successfully")
		pterm.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(adminCmd)

	adminCmd.AddCommand(adminListCmd)
	adminCmd.AddCommand(adminCreateCmd)
	adminCmd.AddCommand(adminDeleteCmd)
	adminCmd.AddCommand(adminUpdatePasswordCmd)
	adminCmd.AddCommand(adminResetPasswordCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// adminCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// adminCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
