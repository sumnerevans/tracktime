package commands

import (
	"fmt"

	"github.com/sumnerevans/tracktime/lib"
)

type Start struct {
	Description string `arg:"positional" placeholder:"DESC" help:"the descripiton of the time entry"`
	Start       string `arg:"-s,--start" help:"the start time of the time entry" default:"now"`
	Type        string `arg:"-t,--type" help:"the type of the time entry"`
	Project     string `arg:"-p,--project" help:"the project of the time entry"`
	Customer    string `arg:"-c,--customer" help:"the customer of the time entry"`
	TaskID      string `arg:"-i,--taskid" help:"the task ID of the time entry"`
}

func (s *Start) Run(config *lib.Config) error {
	fmt.Println(config)
	fmt.Println(s)
	return nil
}
