// Package cmd defines the cobra command structs and an execution method for adding an improved CLI to
// KrakenD based api gateways
package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/spf13/cobra"
)

// Executor defines the function that requires a service description
type Executor func(config.ServiceConfig)

// Execute sets up the cmd package with the received configuration parser and executor and delegates
// the CLI execution to the cobra lib
func Execute(configParser config.Parser, f Executor) {
	DefaultRoot.Execute(configParser, f)
}

type FlagBuilder func(*cobra.Command)

func StringFlagBuilder(dst *string, long, short, defaultValue, help string) FlagBuilder {
	return func(cmd *cobra.Command) {
		cmd.PersistentFlags().StringVarP(dst, long, short, defaultValue, help)
	}
}

func BoolFlagBuilder(dst *bool, long, short string, defaultValue bool, help string) FlagBuilder {
	return func(cmd *cobra.Command) {
		cmd.PersistentFlags().BoolVarP(dst, long, short, defaultValue, help)
	}
}

func DurationFlagBuilder(dst *time.Duration, long, short string, defaultValue time.Duration, help string) FlagBuilder {
	return func(cmd *cobra.Command) {
		cmd.PersistentFlags().DurationVarP(dst, long, short, defaultValue, help)
	}
}

func Float64FlagBuilder(dst *float64, long, short string, defaultValue float64, help string) FlagBuilder {
	return func(cmd *cobra.Command) {
		cmd.PersistentFlags().Float64VarP(dst, long, short, defaultValue, help)
	}
}

func IntFlagBuilder(dst *int, long, short string, defaultValue int, help string) FlagBuilder {
	return func(cmd *cobra.Command) {
		cmd.PersistentFlags().IntVarP(dst, long, short, defaultValue, help)
	}
}

type Command struct {
	Cmd   *cobra.Command
	Flags []FlagBuilder
	once  *sync.Once
}

func NewCommand(command *cobra.Command, flags ...FlagBuilder) Command {
	return Command{Cmd: command, Flags: flags, once: new(sync.Once)}
}

func (c *Command) BuildFlags() {
	c.once.Do(func() {
		for i := range c.Flags {
			c.Flags[i](c.Cmd)
		}
	})
}

func NewRoot(root Command, subCommands ...Command) Root {
	r := Root{Command: root, SubCommands: subCommands}
	r.buildSubCommands()
	return r
}

type Root struct {
	Command
	SubCommands []Command
	once        *sync.Once
}

func (r *Root) buildSubCommands() {
	r.once.Do(func() {
		for i := range r.SubCommands {
			r.Cmd.AddCommand(r.SubCommands[i].Cmd)
		}
	})
}

func (r Root) Execute(configParser config.Parser, f Executor) {
	parser = configParser
	run = f
	if err := r.Cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
