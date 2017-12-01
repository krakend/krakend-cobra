package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/core"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	debug   bool
	port    int
	parser  config.Parser
	run     func(config.ServiceConfig)

	rootCmd = &cobra.Command{
		Use:   "krakend",
		Short: "The API Gateway builder",
	}

	checkCmd = &cobra.Command{
		Use:     "check",
		Short:   "Validates that the configuration file is valid.",
		Long:    "Validates that the active configuration file has a valid syntax to run the service.\nChange the configuration file by using the --config flag",
		Run:     checkFunc,
		Aliases: []string{"validate"},
		Example: "krakend check -d -c config.json",
	}

	runCmd = &cobra.Command{
		Use:     "run",
		Short:   "Run the KrakenD server.",
		Long:    "Run the KrakenD server.",
		Run:     runFunc,
		Example: "krakend run -d -c config.json",
	}
)

func init() {
	logo, err := base64.StdEncoding.DecodeString(encodedLogo)
	if err != nil {
		fmt.Println("decode error:", err)
	}
	rootCmd.SetHelpTemplate(string(logo) + "Version: " + core.KrakendVersion + "\n\n" + rootCmd.HelpTemplate())

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to the configuration filename")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable the debug")

	rootCmd.AddCommand(checkCmd)

	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "Listening port for the http service")
}

const encodedLogo = "ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCmA3TU1GJyBgWU1NJyAgICAgICAgICAgICAgICAgIGA3TU0gICAgICAgICAgICAgICAgICAgICAgICAgYDdNTSIiIlliLiAgIAogIE1NICAgLk0nICAgICAgICAgICAgICAgICAgICAgIE1NICAgICAgICAgICAgICAgICAgICAgICAgICAgTU0gICAgYFliLiAKICBNTSAuZCIgICAgIGA3TWIsb2Q4ICw2IlliLiAgICBNTSAgLE1QJy5nUCJZYSBgN01NcE1NTWIuICAgIE1NICAgICBgTWIgCiAgTU1NTU0uICAgICAgIE1NJyAiJzgpICAgTU0gICAgTU0gO1kgICxNJyAgIFliICBNTSAgICBNTSAgICBNTSAgICAgIE1NIAogIE1NICBWTUEgICAgICBNTSAgICAgLHBtOU1NICAgIE1NO01tICA4TSIiIiIiIiAgTU0gICAgTU0gICAgTU0gICAgICxNUCAKICBNTSAgIGBNTS4gICAgTU0gICAgOE0gICBNTSAgICBNTSBgTWIuWU0uICAgICwgIE1NICAgIE1NICAgIE1NICAgICxkUCcgCi5KTU1MLiAgIE1NYi4uSk1NTC4gIGBNb285XllvLi5KTU1MLiBZQS5gTWJtbWQnLkpNTUwgIEpNTUwuLkpNTW1tbWRQJyAgIApfX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fXwogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAK"
