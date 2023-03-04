package commands

import (
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/rs/zerolog/log"

	"github.com/sumnerevans/tracktime/lib"
)

type List struct {
	Date     lib.Date     `arg:"-d,--date" help:"the date to list time entries for" default:"today"`
	Customer lib.Customer `arg:"-c,--customer" help:"list only time entries for the given customer"`
}

func (l *List) Run(config *lib.Config) error {
	entryList, err := lib.EntryListForDay(config, l.Date)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read entry list")
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("Entries for %s\n", l.Date.Format("2006-01-02"))
	green.Print("======================\n\n")

	tbl := table.New("#", "start", "stop", "project", "type", "task ID", "customer", "description")
	for _, entry := range entryList.EntriesForCustomer(l.Customer) {
		tbl.AddRow(entry.Index, entry.Start, entry.Stop, entry.Project, entry.Type, entry.TaskID, entry.Customer, entry.Description)
	}

	tbl.WithHeaderFormatter(green.SprintfFunc())
	tbl.WithFirstColumnFormatter(color.New(color.FgYellow).SprintfFunc())
	tbl.WithPadding(3)

	tbl.Print()

	duration := entryList.TotalTimeForCustomer(l.Customer)
	green.Printf("\nTotal: %d:%d\n", int(duration.Minutes())/60, int(duration.Minutes())%60)

	return err
}
