package cmd

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	krakendgin "github.com/luraproject/lura/v2/router/gin"
	"github.com/spf13/cobra"
)

const (
	colorGreen   = "\033[32m"
	colorWhite   = "\033[37m"
	colorYellow  = "\033[33m"
	colorRed     = "\033[31m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[36m"
	colorCyan    = "\033[36m"
	colorReset   = "\033[0m"
)

func checkFunc(cmd *cobra.Command, args []string) {
	if cfgFile == "" {
		cmd.Printf("%sPlease, provide the path to your config file%s", colorRed, colorReset)
		return
	}

	cmd.Printf("Parsing configuration file: %s\n", cfgFile)
	v, err := parser.Parse(cfgFile)

	if debug {

		cmd.Printf("%sGlobal settings%s\n", colorGreen, colorReset)
		cmd.Printf("\tCacheTTL: %s\n", v.CacheTTL.String())
		cmd.Printf("\tPort: %d\n", v.Port)
		cmd.Printf("\tDefault backends: %v\n", v.Host)

		cmd.Printf("%sExtra configurations (%d components):%s\n", colorGreen, len(v.ExtraConfig), colorReset)
		for k, e := range v.ExtraConfig {
			cmd.Printf("  %s: %v\n", k, e)
		}

		cmd.Printf("%sConfigured endpoints (%d):%s\n", colorGreen, len(v.Endpoints), colorReset)
		for _, endpoint := range v.Endpoints {
			cmd.Printf("\tEndpoint: %s%s %s%s, CacheTTL: %s, Concurrent: %d, QueryString: %v\n",
				colorCyan, endpoint.Method, endpoint.Endpoint, colorReset, endpoint.CacheTTL.String(),
				endpoint.ConcurrentCalls, endpoint.QueryString)

			cmd.Printf("\tExtra configurations (%d components):\n", len(endpoint.ExtraConfig))
			for k, e := range endpoint.ExtraConfig {
				cmd.Printf("\t  %s: %v\n", k, e)
			}

			cmd.Printf("\tConnecting to %d backends:\n", len(endpoint.Backend))
			for _, backend := range endpoint.Backend {
				cmd.Printf("\t\tURL: %s%s%s, Method: %s\n", colorMagenta, backend.URLPattern, colorReset, backend.Method)
				cmd.Printf("\t\t\tTimeout: %s, Target: %s, Mapping: %v, Deny: %v, Allow: %v, Group: %v\n",
					backend.Timeout, backend.Target, backend.Mapping, backend.DenyList, backend.AllowList,
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
		cmd.Printf("%sERROR parsing the configuration file.%s\n%s\n", colorRed, colorReset, err.Error())
		os.Exit(1)
		return
	}

	if checkGinRoutes {
		if err := runRouter(v); err != nil {
			cmd.Printf("%sERROR testing the configuration file.%s\n%s\n", colorRed, colorReset, err.Error())
			os.Exit(1)
			return
		}
	}

	cmd.Printf("%sSyntax OK!%s\n", colorGreen, colorReset)
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
