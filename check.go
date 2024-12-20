package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/krakendio/krakend-cobra/v2/dumper"
	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/core"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	krakendgin "github.com/luraproject/lura/v2/router/gin"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var SchemaURL = "https://www.krakend.io/schema/v%s/krakend.json"

func errorMsg(content string) string {
	if !IsTTY {
		return content
	}
	return dumper.ColorRed + content + dumper.ColorReset
}

type LastSourcer interface {
	LastSource() ([]byte, error)
}

func NewCheckCmd(rawSchema string) Command {
	rawEmbedSchema = rawSchema
	return CheckCommand
}

func checkFunc(cmd *cobra.Command, _ []string) { // skipcq: GO-R1005
	if cfgFile == "" {
		cmd.Println(errorMsg("Please, provide the path to the configuration file with --config or see all the options with --help"))
		os.Exit(1) // skipcq: RVV-A0003 // skipcq: RVV-A0003
		return
	}

	cmd.Printf("Parsing configuration file: %s\n", cfgFile)

	v, err := parser.Parse(cfgFile)
	if err != nil {
		cmd.Println(errorMsg("ERROR parsing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1) // skipcq: RVV-A0003 // skipcq: RVV-A0003
		return
	}

	shouldLint := lintCurrentSchema || lintNoNetwork || (lintCustomSchemaPath != "")

	if shouldLint {
		var data []byte
		var err error
		if ls, ok := parser.(LastSourcer); ok {
			data, err = ls.LastSource()
		} else {
			data, err = os.ReadFile(cfgFile)
		}

		if err != nil {
			cmd.Println(errorMsg("ERROR loading the configuration content:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1) // skipcq: RVV-A0003
			return
		}

		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			cmd.Println(errorMsg("ERROR converting configuration content to JSON:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1) // skipcq: RVV-A0003
			return
		}

		var sch *jsonschema.Schema
		var compilationErr error
		if lintNoNetwork {
			rawSchema, parseError := jsonschema.UnmarshalJSON(strings.NewReader(rawEmbedSchema))
			if parseError != nil {
				cmd.Println(errorMsg("ERROR parsing the embed schema:") + fmt.Sprintf("\t%s\n", parseError.Error()))
				os.Exit(1) // skipcq: RVV-A0003
				return
			}

			compiler := jsonschema.NewCompiler()
			compiler.AddResource("schema.json", rawSchema)

			sch, compilationErr = compiler.Compile("schema.json")
		} else {
			if lintCustomSchemaPath == "" {
				lintCustomSchemaPath = fmt.Sprintf(SchemaURL, getVersionMinor(core.KrakendVersion))
			}

			httpLoader := SchemaHttpLoader(http.Client{
				Timeout: 10 * time.Second,
			})

			loader := jsonschema.SchemeURLLoader{
				"file":  jsonschema.FileLoader{},
				"http":  &httpLoader,
				"https": &httpLoader,
			}
			compiler := jsonschema.NewCompiler()
			compiler.UseLoader(loader)

			sch, compilationErr = compiler.Compile(lintCustomSchemaPath)
		}

		if compilationErr != nil {
			cmd.Println(errorMsg("ERROR compiling the schema:") + fmt.Sprintf("\t%s\n", compilationErr.Error()))
			os.Exit(1) // skipcq: RVV-A0003
			return
		}

		if err = sch.Validate(raw); err != nil {
			cmd.Println(errorMsg("ERROR linting the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1) // skipcq: RVV-A0003
			return
		}
	}

	if debug > 0 {
		cc := dumper.NewWithColors(cmd, checkDumpPrefix, debug, IsTTY)
		if err := cc.Dump(v); err != nil {
			cmd.Println(errorMsg("ERROR checking the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1) // skipcq: RVV-A0003
			return
		}
	}

	if checkGinRoutes {
		if err := RunRouterFunc(v); err != nil {
			cmd.Println(errorMsg("ERROR testing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1) // skipcq: RVV-A0003
			return
		}
	}

	if IsTTY {
		cmd.Printf("%sSyntax OK!%s\n", dumper.ColorGreen, dumper.ColorReset)
		return
	}
	cmd.Println("Syntax OK!")
}

var RunRouterFunc = func(cfg config.ServiceConfig) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	cfg.Debug = cfg.Debug || debug > 0
	if port != 0 {
		cfg.Port = port
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	krakendgin.DefaultFactory(proxy.DefaultFactory(logging.NoOp), logging.NoOp).NewWithContext(ctx).Run(cfg)
	cancel()
	return nil
}

func getVersionMinor(ver string) string {
	comps := strings.Split(ver, ".")
	if len(comps) < 2 {
		return ver
	}
	return fmt.Sprintf("%s.%s", comps[0], comps[1])
}

type SchemaHttpLoader http.Client

func (l *SchemaHttpLoader) Load(url string) (interface{}, error) {
	client := (*http.Client)(l)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s returned status code %d", url, resp.StatusCode)
	}

	body, err := jsonschema.UnmarshalJSON(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	return body, resp.Body.Close()
}
