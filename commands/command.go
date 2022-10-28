package commands

import "github.com/sumnerevans/tracktime/config"

type Command interface {
	Run(*config.Config) error
}
