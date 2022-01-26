package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/core"
	"github.com/spf13/cobra"
)

var (
	cfgFile        string
	debug          int
	port           int
	checkGinRoutes bool
	parser         config.Parser
	run            func(config.ServiceConfig)

	goSum           = "./go.sum"
	goVersion       = core.GoVersion
	checkDumpPrefix = "\t"

	DefaultRoot   Root
	RootCommand   Command
	RunCommand    Command
	CheckCommand  Command
	PluginCommand Command

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
		Short:   "Runs the KrakenD server.",
		Long:    "Runs the KrakenD server.",
		Run:     runFunc,
		Example: "krakend run -d -c config.json",
	}

	pluginCmd = &cobra.Command{
		Use:     "check-plugin",
		Short:   "Checks your plugin dependencies are compatible.",
		Long:    "Checks your plugin dependencies are compatible.",
		Run:     pluginFunc,
		Example: "krakend check-plugin -v 1.17.0 -s ./go.sum",
	}
)

func init() {
	logo, err := base64.StdEncoding.DecodeString(encodedLogo)
	if err != nil {
		fmt.Println("decode error:", err)
	}
	cfgFlag := StringFlagBuilder(&cfgFile, "config", "c", "", "Path to the configuration filename")
	debugFlag := CountFlagBuilder(&debug, "debug", "d", "Enable the debug")
	RootCommand = NewCommand(rootCmd)
	RootCommand.Cmd.SetHelpTemplate(string(logo) + "Version: " + core.KrakendVersion + "\n\n" + rootCmd.HelpTemplate())

	ginRoutesFlag := BoolFlagBuilder(&checkGinRoutes, "test-gin-routes", "t", false, "Test the endpoint patterns against a real gin router on selected port")
	prefixFlag := StringFlagBuilder(&checkDumpPrefix, "indent", "i", checkDumpPrefix, "Indentation of the check dump")
	CheckCommand = NewCommand(checkCmd, cfgFlag, debugFlag, ginRoutesFlag, prefixFlag)

	portFlag := IntFlagBuilder(&port, "port", "p", 0, "Listening port for the http service")
	RunCommand = NewCommand(runCmd, cfgFlag, debugFlag, portFlag)

	goSumFlag := StringFlagBuilder(&goSum, "sum", "s", goSum, "Path to the go.sum file to analize")
	goVersionFlag := StringFlagBuilder(&goVersion, "version", "v", goVersion, "The version of the go compiler used for your plugin")
	PluginCommand = NewCommand(pluginCmd, goSumFlag, goVersionFlag)

	DefaultRoot = NewRoot(RootCommand, CheckCommand, RunCommand, PluginCommand)
}

const encodedLogo = "ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCmA3TU1GJyBgWU1NJyAgICAgICAgICAgICAgICAgIGA3TU0gICAgICAgICAgICAgICAgICAgICAgICAgYDdNTSIiIlliLiAgIAogIE1NICAgLk0nICAgICAgICAgICAgICAgICAgICAgIE1NICAgICAgICAgICAgICAgICAgICAgICAgICAgTU0gICAgYFliLiAKICBNTSAuZCIgICAgIGA3TWIsb2Q4ICw2IlliLiAgICBNTSAgLE1QJy5nUCJZYSBgN01NcE1NTWIuICAgIE1NICAgICBgTWIgCiAgTU1NTU0uICAgICAgIE1NJyAiJzgpICAgTU0gICAgTU0gO1kgICxNJyAgIFliICBNTSAgICBNTSAgICBNTSAgICAgIE1NIAogIE1NICBWTUEgICAgICBNTSAgICAgLHBtOU1NICAgIE1NO01tICA4TSIiIiIiIiAgTU0gICAgTU0gICAgTU0gICAgICxNUCAKICBNTSAgIGBNTS4gICAgTU0gICAgOE0gICBNTSAgICBNTSBgTWIuWU0uICAgICwgIE1NICAgIE1NICAgIE1NICAgICxkUCcgCi5KTU1MLiAgIE1NYi4uSk1NTC4gIGBNb285XllvLi5KTU1MLiBZQS5gTWJtbWQnLkpNTUwgIEpNTUwuLkpNTW1tbWRQJyAgIApfX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fX19fXwogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAK"
