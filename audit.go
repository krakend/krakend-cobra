package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	audit "github.com/krakend/krakend-audit"
	"github.com/spf13/cobra"
)

const (
	defaultFormatTmpl  = "{{ range .Recommendations }}{{.Rule}}\t[{{.Severity}}]   \t{{.Message}}\n{{ end }}"
	terminalFormatTmpl = "{{ range .Recommendations }}{{.Rule}}\t[{{colored .Severity}}]   \t{{.Message}}\n{{ end }}"
)

func auditFunc(cmd *cobra.Command, _ []string) {
	if cfgFile == "" {
		cmd.Println(errorMsg("Please, provide the path to the configuration file with --config or see all the options with --help"))
		os.Exit(1) // skipcq: RVV-A0003
		return
	}

	cfg, err := parser.Parse(cfgFile)
	if err != nil {
		cmd.Println(errorMsg("ERROR parsing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1) // skipcq: RVV-A0003
		return
	}
	cfg.Normalize()

	if formatTmpl == "" {
		if IsTTY {
			formatTmpl = terminalFormatTmpl
		} else {
			formatTmpl = defaultFormatTmpl
		}
	}

	severitiesToInclude = strings.ReplaceAll(severitiesToInclude, " ", "")
	rules := strings.Split(strings.ReplaceAll(rulesToExclude, " ", ""), ",")

	if rulesToExcludePath != "" {
		b, err := os.ReadFile(rulesToExcludePath)
		if err != nil {
			cmd.Println(errorMsg("ERROR accessing the ignore file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1) // skipcq: RVV-A0003
			return
		}
		for _, line := range strings.Split(strings.ReplaceAll(string(b), " ", ""), "\n") {
			if line == "" {
				continue
			}
			rules = append(rules, line)
		}
	}

	result, err := audit.Audit(
		&cfg,
		rules,
		strings.Split(severitiesToInclude, ","),
	)
	if err != nil {
		cmd.Println(errorMsg("ERROR auditing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1) // skipcq: RVV-A0003
		return
	}

	funcMap := template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"colored": func(s string) string {
			switch s {
			case audit.SeverityLow:
				return fmt.Sprintf("\033[32;1m%s\033[0m", s)
			case audit.SeverityMedium:
				return fmt.Sprintf("\033[33;1m%s\033[0m", s)
			case audit.SeverityHigh:
				return fmt.Sprintf("\033[31;1m%s\033[0m", s)
			case audit.SeverityCritical:
				return fmt.Sprintf("\033[41;1m%s\033[0m", s)
			default:
				return s
			}
		},
	}

	tmpl, err := template.New("audit").Funcs(funcMap).Parse(formatTmpl)
	if err != nil {
		cmd.Println(errorMsg("ERROR parsing the template:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1) // skipcq: RVV-A0003
		return
	}

	if err := tmpl.Execute(os.Stderr, result); err != nil {
		cmd.Println(errorMsg("ERROR rendering the results:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1) // skipcq: RVV-A0003
		return
	}

	if len(result.Recommendations) > 0 {
		os.Exit(1) // skipcq: RVV-A0003
	}
}
