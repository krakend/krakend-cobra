package cmd

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/proxy"
	krakendgin "github.com/luraproject/lura/router/gin"
	"github.com/spf13/cobra"
)

func checkFunc(cmd *cobra.Command, args []string) {
	if cfgFile == "" {
		cmd.Println("Please, provide the path to your config file")
		return
	}

	cmd.Printf("Parsing configuration file: %s\n", cfgFile)
	v, err := parser.Parse(cfgFile)

	if debug {

		cmd.Printf("Parsed configuration: CacheTTL: %s, Port: %d\n", v.CacheTTL.String(), v.Port)
		cmd.Printf("Hosts: %v\n", v.Host)

		cmd.Printf("Extra (%d):\n", len(v.ExtraConfig))
		for k, e := range v.ExtraConfig {
			cmd.Printf("  %s: %v\n", k, e)
		}

		cmd.Printf("Endpoints (%d):\n", len(v.Endpoints))
		for _, endpoint := range v.Endpoints {
			cmd.Printf("\tEndpoint: %s, Method: %s, CacheTTL: %s, Concurrent: %d, QueryString: %v\n",
				endpoint.Endpoint, endpoint.Method, endpoint.CacheTTL.String(),
				endpoint.ConcurrentCalls, endpoint.QueryString)

			cmd.Printf("\tExtra (%d):\n", len(endpoint.ExtraConfig))
			for k, e := range endpoint.ExtraConfig {
				cmd.Printf("\t  %s: %v\n", k, e)
			}

			cmd.Printf("\tBackends (%d):\n", len(endpoint.Backend))
			for _, backend := range endpoint.Backend {
				cmd.Printf("\t\tURL: %s, Method: %s\n", backend.URLPattern, backend.Method)
				cmd.Printf("\t\t\tTimeout: %s, Target: %s, Mapping: %v, BL: %v, WL: %v, Group: %v\n",
					backend.Timeout, backend.Target, backend.Mapping, backend.Blacklist, backend.Whitelist,
					backend.Group)
				cmd.Printf("\t\t\tHosts: %v\n", backend.Host)
				cmd.Printf("\t\t\tExtra (%d):\n", len(backend.ExtraConfig))
				for k, e := range backend.ExtraConfig {
					cmd.Printf("\t\t\t  %s: %v\n", k, e)
				}
			}
		}
	}

	if err != nil {
		cmd.Println("ERROR parsing the configuration file.\n", err.Error())
		os.Exit(1)
		return
	}

	if checkGinRoutes {
		if err := runRouter(v); err != nil {
			cmd.Println("ERROR testing the configuration file.\n", err.Error())
			os.Exit(1)
			return
		}
	}

	cmd.Println("Syntax OK!")
}

func runRouter(cfg config.ServiceConfig) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	cfg.Debug = cfg.Debug || debug
	if port != 0 {
		cfg.Port = port
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	krakendgin.DefaultFactory(proxy.DefaultFactory(logging.NoOp), logging.NoOp).NewWithContext(ctx).Run(cfg)
	return nil
}
