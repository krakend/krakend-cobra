package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/krakendio/krakend-cobra/v2/dumper"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/core"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	krakendgin "github.com/luraproject/lura/v2/router/gin"

	"github.com/gin-gonic/gin"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
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

func checkFunc(cmd *cobra.Command, _ []string) {
	if cfgFile == "" {
		cmd.Println(errorMsg("Please, provide the path to the configuration file with --config or see all the options with --help"))
		os.Exit(1)
		return
	}

	cmd.Printf("Parsing configuration file: %s\n", cfgFile)

	v, err := parser.Parse(cfgFile)
	if err != nil {
		cmd.Println(errorMsg("ERROR parsing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1)
		return
	}

	if schemaValidation {
		var data []byte
		var err error
		if ls, ok := parser.(LastSourcer); ok {
			data, err = ls.LastSource()
		} else {
			data, err = os.ReadFile(cfgFile)
		}

		if err != nil {
			cmd.Println(errorMsg("ERROR loading the configuration content:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1)
			return
		}

		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			cmd.Println(errorMsg("ERROR converting configuration content to JSON:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1)
			return
		}

		var url string
		if strings.Contains(SchemaURL, "v%s") {
			url = fmt.Sprintf(SchemaURL, getVersionMinor(core.KrakendVersion))
		} else {
			// the global schema url var might have been set at build time
			// to something different
			url = SchemaURL
		}
		sch, err := jsonschema.Compile(url)
		if err != nil {
			cmd.Println(errorMsg("ERROR compiling the schema:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1)
			return
		}

		if err = sch.Validate(raw); err != nil {
			cmd.Println(errorMsg("ERROR linting the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1)
			return
		}
	}

	if debug > 0 {
		cc := dumper.NewWithColors(cmd, checkDumpPrefix, debug, IsTTY)
		if err := cc.Dump(v); err != nil {
			cmd.Println(errorMsg("ERROR checking the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1)
			return
		}
	}

	if checkGinRoutes {
		if err := RunRouterFunc(v); err != nil {
			cmd.Println(errorMsg("ERROR testing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1)
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
