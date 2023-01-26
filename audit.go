package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	audit "github.com/krakendio/krakend-audit"
	"github.com/spf13/cobra"
)

func auditFunc(cmd *cobra.Command, _ []string) {
	if cfgFile == "" {
		cmd.Println(errorMsg("Please, provide the path to your config file"))
		os.Exit(1)
		return
	}

	cfg, err := parser.Parse(cfgFile)
	if err != nil {
		cmd.Println(errorMsg("ERROR parsing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1)
		return
	}

	severitiesToInclude = strings.ReplaceAll(severitiesToInclude, " ", "")
	rules := strings.Split(strings.ReplaceAll(rulesToExclude, " ", ""), ",")

	if rulesToExcludePath != "" {
		b, err := os.ReadFile(rulesToExcludePath)
		if err != nil {
			cmd.Println(errorMsg("ERROR parsing the ignore file:") + fmt.Sprintf("\t%s\n", err.Error()))
		} else {
			for _, line := range strings.Split(strings.ReplaceAll(string(b), " ", ""), "\n") {
				if line == "" {
					continue
				}
				rules = append(rules, line)
			}
		}
	}

	result, err := audit.Audit(
		&cfg,
		rules,
		strings.Split(severitiesToInclude, ","),
	)
	if err != nil {
		cmd.Println(errorMsg("ERROR auditing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1)
		return
	}

	funcMap := template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}

	tmpl, err := template.New("audit").Funcs(funcMap).Parse(formatTmpl)
	if err != nil {
		cmd.Println(errorMsg("ERROR parsing the template:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1)
		return
	}

	if err := tmpl.Execute(os.Stdout, result); err != nil {
		cmd.Println(errorMsg("ERROR rendering the results:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1)
		return
	}
}
