package cmd

import (
	"context"
	"errors"
	"net/http"
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
		cmd.Printf("\tName: %s\n", v.Name)
		cmd.Printf("\tPort: %d\n", v.Port)
		cmd.Printf("\tDefault cache TTL: %s\n", v.CacheTTL.String())
		cmd.Printf("\tDefault timeout: %s\n", v.Timeout.String())
		cmd.Println("\tDefault backend hosts:")
		for _, h := range v.Host {
			cmd.Printf("\t\t%s\n", h)
		}
		if len(v.Host) == 0 {
			cmd.Println("\t\t-")
		}

		cmd.Printf("%s%d global component configuration(s):%s\n", colorGreen, len(v.ExtraConfig), colorReset)
		for k, e := range v.ExtraConfig {
			cmd.Printf("\t%s%s:%s\n", colorYellow, k, colorReset)
			cmd.Printf("\t\t%+v\n", e)
		}
		cmd.Println("")

		cmd.Printf("%s%d API endpoints:%s\n", colorGreen, len(v.Endpoints), colorReset)
		for _, endpoint := range v.Endpoints {
			dumpEndpoint(endpoint, cmd)
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

func dumpEndpoint(endpoint *config.EndpointConfig, cmd *cobra.Command) {
	cmd.Printf("\t%s%s%s %s%s\n", methodColor(endpoint.Method), endpoint.Method, colorCyan, endpoint.Endpoint, colorReset)
	cmd.Printf("\t\tTimeout: %s\n", endpoint.Timeout.String())
	cmd.Printf("\t\tCacheTTL: %s\n", endpoint.CacheTTL.String())
	cmd.Printf("\t\tConcurrent calls: %d\n", endpoint.ConcurrentCalls)
	cmd.Printf("\t\tQueryString: %v\n", endpoint.QueryString)

	cmd.Printf("\t\t%s%d endpoint component configuration(s):%s\n", colorGreen, len(endpoint.ExtraConfig), colorReset)
	for k, e := range endpoint.ExtraConfig {
		cmd.Printf("\t\t\t%s%s:%s\n", colorYellow, k, colorReset)
		cmd.Printf("\t\t\t\t%+v\n", e)
	}

	cmd.Printf("\t\t%sConnecting to %d backend(s):%s\n", colorGreen, len(endpoint.Backend), colorReset)
	for _, backend := range endpoint.Backend {
		dumpBackend(backend, cmd)
	}
}

func dumpBackend(backend *config.Backend, cmd *cobra.Command) {
	cmd.Printf("\t\t\t%s%s%s %s%s\n", methodColor(backend.Method), backend.Method, colorCyan, backend.URLPattern, colorReset)
	cmd.Printf("\t\t\tTimeout: %s\n", backend.Timeout.String())
	cmd.Printf("\t\t\tTarget: %s\n", backend.Target)
	cmd.Printf("\t\t\tDeny: %v, Allow: %v\n", backend.DenyList, backend.AllowList)
	cmd.Printf("\t\t\tGroup: %s\n", backend.Group)
	cmd.Printf("\t\t\tHosts: %v\n", backend.Host)
	cmd.Printf("\t\t\t%s%d backend component configuration(s):%s\n", colorGreen, len(backend.ExtraConfig), colorReset)
	for k, e := range backend.ExtraConfig {
		cmd.Printf("\t\t\t\t%s%s:%s\n", colorYellow, k, colorReset)
		cmd.Printf("\t\t\t\t\t%+v\n", e)
	}
	cmd.Println("")
}

var methodColors = map[string]string{
	http.MethodGet:     colorBlue,
	http.MethodPost:    colorCyan,
	http.MethodPut:     colorYellow,
	http.MethodDelete:  colorRed,
	http.MethodPatch:   colorGreen,
	http.MethodHead:    colorMagenta,
	http.MethodOptions: colorWhite,
}

func methodColor(method string) string {
	m, ok := methodColors[method]
	if !ok {
		return colorGreen
	}
	return m
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
