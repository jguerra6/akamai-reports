package command

import (
	"flag"
)

type Command struct {
	flagSet *flag.FlagSet
	flags   map[string]*string
}

type Flag struct {
	Name     string
	Value    string
	Usage    string
	Required bool
}

func NewCommand(name string, flags map[string]Flag) *Command {
	command := &Command{
		flagSet: flag.NewFlagSet(name, flag.ContinueOnError),
		flags:   map[string]*string{},
	}

	for _, f := range flags {
		flagN := f.Name
		command.flagSet.StringVar(&flagN, f.Name, f.Value, f.Usage)
		command.flags[f.Name] = &flagN
	}

	return command
}

func (c *Command) Name() string {
	return c.flagSet.Name()
}

func (c *Command) Flags() map[string]*string {
	return c.flags
}

func (c *Command) Init(args []string) error {
	return c.flagSet.Parse(args)
}
