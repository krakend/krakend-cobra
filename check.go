package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/devopsfaith/krakend-cobra/v2/dumper"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	krakendgin "github.com/luraproject/lura/v2/router/gin"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func errorMsg(content string) string {
	return dumper.ColorRed + content + dumper.ColorReset
}

func checkFunc(cmd *cobra.Command, args []string) {
	if cfgFile == "" {
		cmd.Println(errorMsg("Please, provide the path to your config file"))
		return
	}

	cmd.Printf("Parsing configuration file: %s\n", cfgFile)
	v, err := parser.Parse(cfgFile)
	if err != nil {
		cmd.Println(errorMsg("ERROR parsing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
		os.Exit(1)
		return
	}

	if debug > 0 {
		cc := dumper.New(cmd, checkDumpPrefix, debug)
		if err := cc.Dump(v); err != nil {
			cmd.Println(errorMsg("ERROR checking the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1)
			return
		}
	}

	if checkGinRoutes {
		if err := runRouter(v); err != nil {
			cmd.Println(errorMsg("ERROR testing the configuration file:") + fmt.Sprintf("\t%s\n", err.Error()))
			os.Exit(1)
			return
		}
	}

	cmd.Printf("%sSyntax OK!%s\n", dumper.ColorGreen, dumper.ColorReset)
}

func runRouter(cfg config.ServiceConfig) (err error) {
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

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	krakendgin.DefaultFactory(proxy.DefaultFactory(logging.NoOp), logging.NoOp).NewWithContext(ctx).Run(cfg)
	return nil
}
