// Package cmd defines the cobra command structs and an execution method for adding an improved CLI to
// KrakenD based api gateways
package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/spf13/cobra"
)

// Executor defines the function that requires a service description
type Executor func(config.ServiceConfig)

// Execute sets up the cmd package with the received configuration parser and executor and delegates
// the CLI execution to the cobra lib
func Execute(configParser config.Parser, f Executor) {
	ExecuteRoot(configParser, f, DefaultRoot)
}

func ExecuteRoot(configParser config.Parser, f Executor, root Root) {
	root.Build()
	root.Execute(configParser, f)
}

func GetConfigFlag() string {
	return cfgFile
}

func GetDebugFlag() bool {
	return debug > 0
}

func GetConfigParser() config.Parser {
	return parser
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

func CountFlagBuilder(dst *int, long, short, help string) FlagBuilder {
	return func(cmd *cobra.Command) {
		cmd.PersistentFlags().CountVarP(dst, long, short, help)
	}
}

type ContraintBuilder func(*cobra.Command)

func OneRequired(flags ...string) ContraintBuilder {
	return func(cmd *cobra.Command) {
		cmd.MarkFlagsOneRequired(flags...)
	}
}

func RequiredTogether(flags ...string) ContraintBuilder {
	return func(cmd *cobra.Command) {
		cmd.MarkFlagsRequiredTogether(flags...)
	}
}

func MutuallyExclusive(flags ...string) ContraintBuilder {
	return func(cmd *cobra.Command) {
		cmd.MarkFlagsMutuallyExclusive(flags...)
	}
}

type Command struct {
	Cmd         *cobra.Command
	Flags       []FlagBuilder
	once        *sync.Once
	Constraints []ContraintBuilder
}

func NewCommand(command *cobra.Command, flags ...FlagBuilder) Command {
	return Command{Cmd: command, Flags: flags, once: new(sync.Once)}
}

func (c *Command) AddConstraint(r ContraintBuilder) {
	c.Constraints = append(c.Constraints, r)
}

func (c *Command) AddFlag(f FlagBuilder) {
	c.Flags = append(c.Flags, f)
}

func (c *Command) BuildFlags() {
	c.once.Do(func() {
		for i := range c.Flags {
			c.Flags[i](c.Cmd)
		}
	})
}

func (c *Command) AddSubCommand(cmd *cobra.Command) {
	c.Cmd.AddCommand(cmd)
}

func NewRoot(root Command, subCommands ...Command) Root {
	r := Root{Command: root, SubCommands: subCommands, once: new(sync.Once)}
	return r
}

type Root struct {
	Command
	SubCommands []Command
	once        *sync.Once
}

func (r *Root) Build() {
	r.once.Do(func() {
		r.BuildFlags()
		for i := range r.Constraints {
			r.Constraints[i](r.Cmd)
		}
		for i := range r.SubCommands {
			s := r.SubCommands[i]
			s.BuildFlags()
			for j := range s.Constraints {
				s.Constraints[j](s.Cmd)
			}
			r.Cmd.AddCommand(s.Cmd)
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
