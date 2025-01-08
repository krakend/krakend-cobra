package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/core"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var IsTTY = isatty.IsTerminal(os.Stderr.Fd())

var (
	cfgFile              string
	debug                int
	port                 int
	checkGinRoutes       bool
	checkDebug           int
	lintCurrentSchema    bool
	lintCustomSchemaPath string
	lintNoNetwork        bool
	rawEmbedSchema       string
	rulesToExclude       string
	rulesToExcludePath   string
	severitiesToInclude  = "CRITICAL,HIGH,MEDIUM,LOW"
	formatTmpl           string
	parser               config.Parser
	run                  func(config.ServiceConfig)

	goSum           = "./go.sum"
	goVersion       = core.GoVersion
	libcVersion     = core.GlibcVersion
	checkDumpPrefix = "\t"
	gogetEnabled    = false

	DefaultRoot    Root
	RootCommand    Command
	RunCommand     Command
	CheckCommand   Command
	PluginCommand  Command
	VersionCommand Command
	AuditCommand   Command

	rootCmd = &cobra.Command{
		Use:   "krakend",
		Short: "KrakenD is a high-performance API gateway that helps you publish, secure, control, and monitor your services",
	}

	checkCmd = &cobra.Command{
		Use:     "check",
		Short:   "Validates that the configuration file is valid.",
		Long:    "Validates that the active configuration file has a valid syntax to run the service.\nChange the configuration file by using the --config flag",
		Run:     checkFunc,
		Aliases: []string{"validate"},
		Example: "krakend check -d -l -c config.json",
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
		Long:    "Checks your plugin dependencies are compatible and proposes commands to update your dependencies.",
		Run:     pluginFunc,
		Example: "krakend check-plugin -g 1.19.0 -s ./go.sum -f",
	}

	versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "Shows KrakenD version.",
		Long:    "Shows KrakenD version.",
		Run:     versionFunc,
		Example: "krakend version",
	}

	auditCmd = &cobra.Command{
		Use:     "audit",
		Short:   "Audits a KrakenD configuration.",
		Long:    "Audits a KrakenD configuration.",
		Run:     auditFunc,
		Example: "krakend audit -i 1.1.1,1.1.2 -s CRITICAL -c krakend.json",
	}
)

func init() {
	logo, err := base64.StdEncoding.DecodeString(encodedLogo)
	if err != nil {
		fmt.Println("decode error:", err)
	}

	cfgFlag := StringFlagBuilder(&cfgFile, "config", "c", "", "Path to the configuration file")
	debugFlag := CountFlagBuilder(&debug, "debug", "d", "Enables the debug endpoint")
	RootCommand = NewCommand(rootCmd)
	RootCommand.Cmd.SetHelpTemplate(string(logo) + "Version: " + core.KrakendVersion + "\n\n" + rootCmd.HelpTemplate())

	ginRoutesFlag := BoolFlagBuilder(&checkGinRoutes, "test-gin-routes", "t", false, "Tests the endpoint patterns against a real gin router on the selected port")
	prefixFlag := StringFlagBuilder(&checkDumpPrefix, "indent", "i", checkDumpPrefix, "Indentation of the check dump")
	lintCurrentSchemaFlag := BoolFlagBuilder(&lintCurrentSchema, "lint", "l", lintCurrentSchema, "Enables the linting against the official KrakenD online JSON schema")
	lintCustomSchemaFlag := StringFlagBuilder(&lintCustomSchemaPath, "lint-schema", "s", lintCustomSchemaPath, "Lint against a custom schema path or URL")
	lintNoNetworkFlag := BoolFlagBuilder(&lintNoNetwork, "lint-no-network", "n", lintNoNetwork, "Lint against the builtin Krakend JSON schema, no network is required")
	checkDebugFlag := CountFlagBuilder(&checkDebug, "debug", "d", "Information about how KrakenD is interpreting your configuration file")
	CheckCommand = NewCommand(checkCmd, cfgFlag, checkDebugFlag, ginRoutesFlag, prefixFlag, lintCurrentSchemaFlag, lintCustomSchemaFlag, lintNoNetworkFlag)
	CheckCommand.AddConstraint(MutuallyExclusive("lint", "lint-no-network", "lint-schema"))

	portFlag := IntFlagBuilder(&port, "port", "p", 0, "Listening port for the http service")
	RunCommand = NewCommand(runCmd, cfgFlag, debugFlag, portFlag)

	goSumFlag := StringFlagBuilder(&goSum, "sum", "s", goSum, "Path to the go.sum file to analyze")
	goVersionFlag := StringFlagBuilder(&goVersion, "go", "g", goVersion, "The version of the go compiler used for your plugin")
	libcVersionFlag := StringFlagBuilder(&libcVersion, "libc", "l", "", "Version of the libc library used")
	gogetFlag := BoolFlagBuilder(&gogetEnabled, "format", "f", false, "Shows fix commands to update your dependencies")
	PluginCommand = NewCommand(pluginCmd, goSumFlag, goVersionFlag, libcVersionFlag, gogetFlag)

	rulesToExcludeFlag := StringFlagBuilder(&rulesToExclude, "ignore", "i", rulesToExclude, "List of rules to ignore (comma-separated, no spaces)")
	severitiesToIncludeFlag := StringFlagBuilder(&severitiesToInclude, "severity", "s", severitiesToInclude, "List of severities to include (comma-separated, no spaces)")
	pathToRulesToExcludeFlag := StringFlagBuilder(&rulesToExcludePath, "ignore-file", "I", rulesToExcludePath, "Path to a text-plain file containing the list of rules to exclude")
	formatFlag := StringFlagBuilder(&formatTmpl, "format", "f", formatTmpl, "Inline go template to render the results")
	AuditCommand = NewCommand(auditCmd, cfgFlag, rulesToExcludeFlag, severitiesToIncludeFlag, pathToRulesToExcludeFlag, formatFlag)

	VersionCommand = NewCommand(versionCmd)

	DefaultRoot = NewRoot(RootCommand, CheckCommand, RunCommand, PluginCommand, VersionCommand, AuditCommand)
}

const encodedLogo = "IOKVk+KWhOKWiCAgICAgICAgICAgICAgICAgICAgICAgICAg4paE4paE4paMICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIOKVk+KWiOKWiOKWiOKWiOKWiOKWiOKWhMK1ICAK4paQ4paI4paI4paIICDiloTilojilojilojilajilpDilojilojilojiloTilojilohI4pWX4paI4paI4paI4paI4paI4paI4paEICDilZHilojilojilowgLOKWhOKWiOKWiOKWiOKVqCDiloTilojilojilojilojilojilojiloQgIOKWk+KWiOKWiOKWjOKWiOKWiOKWiOKWiOKWiOKWhCAg4paI4paI4paI4paA4pWZ4pWZ4paA4paA4paI4paI4paI4pWVCuKWkOKWiOKWiOKWiOKWhOKWiOKWiOKWiOKWgCAg4paQ4paI4paI4paI4paI4paI4paAIuKVmeKWgOKWgCLilZniloDilojilojilogg4pWR4paI4paI4paI4paE4paI4paI4paI4pSYICDilojilojilojiloAiIuKWgOKWiOKWiOKWiCDilojilojilojilojiloDilZniloDilojilojilohIIOKWiOKWiOKWiCAgICAg4pWZ4paI4paI4paICuKWkOKWiOKWiOKWiOKWiOKWiOKWiOKWjCAgIOKWkOKWiOKWiOKWiOKMkCAgLOKWhOKWiOKWiOKWiOKWiOKWiOKWiOKWiOKWiE3ilZHilojilojilojilojilojilojiloQgIOKVkeKWiOKWiOKWiOKWiOKWiOKWiOKWiOKWiOKWiOKWiE3ilojilojilojilowgICDilojilojilohIIOKWiOKWiOKWiCAgICAgLOKWiOKWiOKWiArilpDilojilojilojilajiloDilojilojilojCtSDilpDilojilojiloggICDilojilojilojilowgICzilojilojilohN4pWR4paI4paI4paI4pWZ4paA4paI4paI4paIICDilojilojilojiloRgYGDiloTiloRgIOKWiOKWiOKWiOKWjCAgIOKWiOKWiOKWiEgg4paI4paI4paILCws4pWT4paE4paI4paI4paI4paACuKWkOKWiOKWiOKWiCAg4pWZ4paI4paI4paI4paE4paQ4paI4paI4paIICAg4pWZ4paI4paI4paI4paI4paI4paI4paI4paI4paITeKVkeKWiOKWiOKWjCAg4pWZ4paI4paI4paI4paEYOKWgOKWiOKWiOKWiOKWiOKWiOKWiOKWiOKVqCDilojilojilojilowgICDilojilojilohIIOKWiOKWiOKWiOKWiOKWiOKWiOKWiOKWiOKWiOKWgCAgCiAgICAgICAgICAgICAgICAgICAgIGBgICAgICAgICAgICAgICAgICAgICAgYCdgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAo="
