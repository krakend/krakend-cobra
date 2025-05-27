package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func runFunc(cmd *cobra.Command, _ []string) {
	if cfgFile == "" {
		cmd.Println("Please, provide the path to the configuration file with --config or see all the options with --help")
		return
	}
	serviceConfig, err := parser.Parse(cfgFile)
	if err != nil {
		cmd.Printf("ERROR parsing the configuration file: %s\n", err.Error())
		os.Exit(-1)
	}
	serviceConfig.Debug = serviceConfig.Debug || (debug > 0)
	if port != 0 {
		serviceConfig.Port = port
	}
	run(serviceConfig)
}
