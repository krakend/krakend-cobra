package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func runFunc(cmd *cobra.Command, args []string) {
	if cfgFile == "" {
		cmd.Println("Please, provide the path to your config file")
		return
	}
	serviceConfig, err := parser.Parse(cfgFile)
	if err != nil {
		cmd.Printf("ERROR parsing the configuration file: %s\n", err.Error())
		os.Exit(-1)
	}
	serviceConfig.Debug = serviceConfig.Debug || debug
	if port != 0 {
		serviceConfig.Port = port
	}
	cmd.Printf("Parsing configuration file: %s\n", cfgFile)
	run(serviceConfig)
}
