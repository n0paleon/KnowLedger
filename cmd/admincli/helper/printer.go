package helper

import (
	"KnowLedger/internal/model"

	"github.com/pterm/pterm"
)

func PrintAdmins(admins []model.Admin) {
	if len(admins) == 0 {
		pterm.Warning.Println("No admins found")
		return
	}

	tableData := pterm.TableData{
		{"ID", "Username", "Password (Hash)", "Created At", "Updated At"},
	}

	for _, a := range admins {
		tableData = append(tableData, []string{
			a.ID,
			a.Username,
			a.Password,
			formatTime(a.CreatedAt),
			formatTime(a.UpdatedAt),
		})
	}

	_ = pterm.DefaultTable.
		WithHasHeader().
		WithBoxed(true).
		WithData(tableData).
		Render()

	pterm.Info.Printf("Total: %d\n", len(admins))
}

func PrintAdminsWithRawPassword(admins []model.Admin) {
	if len(admins) == 0 {
		pterm.Warning.Println("No admins found")
		return
	}

	tableData := pterm.TableData{
		{"ID", "Username", "Password (Raw)", "Created At", "Updated At"},
	}

	for _, a := range admins {
		tableData = append(tableData, []string{
			a.ID,
			a.Username,
			a.Password,
			formatTime(a.CreatedAt),
			formatTime(a.UpdatedAt),
		})
	}

	_ = pterm.DefaultTable.
		WithHasHeader().
		WithBoxed(true).
		WithData(tableData).
		Render()

	pterm.Info.Printf("Total: %d\n", len(admins))
}
