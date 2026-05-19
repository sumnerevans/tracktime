package commands

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
	"github.com/sumnerevans/tracktime/internal/types"
)

type List struct {
	Date     types.Date         `arg:"-d,--date" help:"the date to list time entries for" default:"today"`
	Customer timeentry.Customer `arg:"-c,--customer" help:"list only time entries for the given customer"`
}

func (l *List) Run(_ context.Context, config *config.Config) error {
	entryList, err := timeentry.EntryListForDay(config, l.Date)
	if err != nil {
		return fmt.Errorf("failed to read entry list: %w", err)
	}

	greenUnderline := color.New(color.FgGreen, color.Bold, color.Underline)
	greenUnderline.Printf("Entries for %s\n\n", l.Date.Format("2006-01-02"))

	tbl := table.
		New("#", "start", "stop", "project", "type", "task ID", "customer", "description").
		WithHeaderFormatter(greenUnderline.SprintfFunc()).
		WithFirstColumnFormatter(color.New(color.FgYellow).SprintfFunc()).
		WithPadding(3)

	for _, entry := range entryList.EntriesForCustomer(l.Customer) {
		tbl.AddRow(entry.Index, entry.Start, entry.Stop, entry.Project, entry.Type, entry.TaskID, entry.Customer, entry.Description)
	}

	tbl.Print()

	duration := entryList.TotalTimeForCustomer(l.Customer)

	color.New(color.FgGreen, color.Bold).
		Printf("\nTotal: %d:%02d\n", int(duration.Minutes())/60, int(duration.Minutes())%60)

	return nil
}
