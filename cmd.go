// Package cmd defines the cobra command structs and an execution method for adding an improved CLI to
// KrakenD based api gateways
package cmd

import (
	"fmt"
	"os"

	"github.com/devopsfaith/krakend/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Executor defines the function that requires a service description
type Executor func(config.ServiceConfig)

// Execute sets up the cmd package with the received configuration parser and executor and delegates
// the CLI execution to the cobra lib
func Execute(configParser config.Parser, f Executor) {
	DefaultRoot.Execute(configParser, f)
}

type Command struct {
	Cmd     *cobra.Command
	FlagSet *pflag.FlagSet
}

func NewCommand(command *cobra.Command) Command {
	return Command{Cmd: command, FlagSet: command.PersistentFlags()}
}

func NewRoot(root Command, subCommands ...Command) Root {
	r := Root{Command: root, SubCommands: subCommands}
	r.buildSubCommands()
	return r
}

type Root struct {
	Command
	SubCommands []Command
}

func (r *Root) buildSubCommands() {
	for i := range r.SubCommands {
		r.Cmd.AddCommand(r.SubCommands[i].Cmd)
	}
}

func (r Root) Execute(configParser config.Parser, f Executor) {
	parser = configParser
	run = f
	if err := r.Cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
