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
	if err != nil {
		cmd.Printf("%sERROR parsing the configuration file:%s\n\t%s\n", colorRed, colorReset, err.Error())
		os.Exit(1)
		return
	}

	if debug > 0 {
		cc := cfgDumper{cmd: cmd}
		if err := cc.dumpService(v); err != nil {
			cmd.Printf("%sERROR checking the configuration file:%s\n\t%s\n", colorRed, colorReset, err.Error())
			os.Exit(1)
			return
		}
	}

	if checkGinRoutes {
		if err := runRouter(v); err != nil {
			cmd.Printf("%sERROR testing the configuration file:%s\n\t%s\n", colorRed, colorReset, err.Error())
			os.Exit(1)
			return
		}
	}

	cmd.Printf("%sSyntax OK!%s\n", colorGreen, colorReset)
}

type cfgDumper struct {
	cmd *cobra.Command
}

func (c cfgDumper) dumpService(v config.ServiceConfig) error {
	c.cmd.Printf("%sGlobal settings%s\n", colorGreen, colorReset)
	c.cmd.Printf("\tName: %s\n", v.Name)
	c.cmd.Printf("\tPort: %d\n", v.Port)

	if debug > 1 {
		c.cmd.Printf("\tDefault cache TTL: %s\n", v.CacheTTL.String())
		c.cmd.Printf("\tDefault timeout: %s\n", v.Timeout.String())
	}

	if len(v.Host) > 0 || debug > 1 {
		c.cmd.Println("\tDefault backend hosts: %v\n", v.Host)
	}

	if debug > 2 {
		c.cmd.Printf("\tRead timeout: %s\n", v.ReadTimeout.String())
		c.cmd.Printf("\tWrite timeout: %s\n", v.WriteTimeout.String())
		c.cmd.Printf("\tIdle timeout: %s\n", v.IdleTimeout.String())
		c.cmd.Printf("\tRead header timeout: %s\n", v.ReadHeaderTimeout.String())
		c.cmd.Printf("\tIdle connection timeout: %s\n", v.IdleConnTimeout.String())
		c.cmd.Printf("\tResponse header timeout: %s\n", v.ResponseHeaderTimeout.String())
		c.cmd.Printf("\tExpect continue timeout: %s\n", v.ExpectContinueTimeout.String())
		c.cmd.Printf("\tDialer timeout: %s\n", v.DialerTimeout.String())
		c.cmd.Printf("\tDialer fallback delay: %s\n", v.DialerFallbackDelay.String())
		c.cmd.Printf("\tDialer keep alive: %s\n", v.DialerKeepAlive.String())
		c.cmd.Printf("\tDisable keep alives: %v\n", v.DisableKeepAlives)
		c.cmd.Printf("\tDisable compression: %v\n", v.DisableCompression)
		c.cmd.Printf("\tMax idle connections: %d\n", v.MaxIdleConns)
		c.cmd.Printf("\tMax idle connections per host: %d\n", v.MaxIdleConnsPerHost)
	}

	if v.TLS != nil {
		c.cmd.Printf("\tDisabled: %v\n", v.TLS.IsDisabled)
		c.cmd.Printf("\tPublic key: %s\n", v.TLS.PublicKey)
		c.cmd.Printf("\tPrivate key: %s\n", v.TLS.PrivateKey)
		c.cmd.Printf("\tEnable MTLS: %v\n", v.TLS.EnableMTLS)

		if debug > 1 {
			c.cmd.Printf("\tMin version: %s\n", v.TLS.MinVersion)
			c.cmd.Printf("\tMax version: %s\n", v.TLS.MaxVersion)
		}
		if debug > 2 {
			c.cmd.Printf("\tCurve preferences: %v\n", v.TLS.CurvePreferences)
			c.cmd.Printf("\tPrefer server cipher suites: %v\n", v.TLS.PreferServerCipherSuites)
			c.cmd.Printf("\tCipher suites: %v\n", v.TLS.CipherSuites)
		}
	} else if debug > 1 {
		c.cmd.Printf("\t%sNo TLS section defined%s\n", colorRed, colorReset)
	}

	if v.Plugin != nil {
		c.cmd.Printf("\tFolder: %s\n", v.Plugin.Folder)
		c.cmd.Printf("\tPattern: %s\n", v.Plugin.Pattern)
	} else if debug > 1 {
		c.cmd.Printf("\t%sNo Plugin section defined%s\n", colorRed, colorReset)
	}

	if debug > 1 || len(v.ExtraConfig) > 0 {
		c.cmd.Printf("%s%d global component configuration(s):%s\n", colorGreen, len(v.ExtraConfig), colorReset)
		for k, e := range v.ExtraConfig {
			c.cmd.Printf("\t%s%s%s\n", colorYellow, k, colorReset)
			if debug > 1 {
				c.cmd.Printf("\t\t%+v\n", e)
			}
		}
	}

	c.cmd.Println("")
	if len(v.Endpoints) == 0 {
		return errors.New("no endpoints defined")
	}

	c.cmd.Printf("%s%d API endpoints:%s\n", colorGreen, len(v.Endpoints), colorReset)
	for _, endpoint := range v.Endpoints {
		c.dumpEndpoint(endpoint)
	}
	return nil
}

func (c cfgDumper) dumpEndpoint(endpoint *config.EndpointConfig) {
	c.cmd.Printf("\t%s%s%s %s%s\n", methodColor(endpoint.Method), endpoint.Method, colorCyan, endpoint.Endpoint, colorReset)
	c.cmd.Printf("\t\tTimeout: %s\n", endpoint.Timeout.String())

	if debug > 1 || len(endpoint.QueryString) > 0 {
		c.cmd.Printf("\t\tQueryString: %v\n", endpoint.QueryString)
	}

	if debug > 1 {
		c.cmd.Printf("\t\tCacheTTL: %s\n", endpoint.CacheTTL.String())
		c.cmd.Printf("\t\tConcurrent calls: %d\n", endpoint.ConcurrentCalls)
		c.cmd.Printf("\t\tHeaders to pass: %v\n", endpoint.HeadersToPass)
		c.cmd.Printf("\t\tOutputEncoding: %s\n", endpoint.OutputEncoding)
	}

	if debug > 1 || len(endpoint.ExtraConfig) > 0 {
		c.cmd.Printf("\t\t%s%d endpoint component configuration(s):%s\n", colorGreen, len(endpoint.ExtraConfig), colorReset)
		for k, e := range endpoint.ExtraConfig {
			c.cmd.Printf("\t\t\t%s%s%s\n", colorYellow, k, colorReset)

			if debug > 1 {
				c.cmd.Printf("\t\t\t\t%+v\n", e)
			}
		}
	}

	c.cmd.Printf("\t\t%sConnecting to %d backend(s):%s\n", colorGreen, len(endpoint.Backend), colorReset)
	for _, backend := range endpoint.Backend {
		c.dumpBackend(backend)
	}
}

func (c cfgDumper) dumpBackend(backend *config.Backend) {
	c.cmd.Printf("\t\t\t%s%s%s %s%s\n", methodColor(backend.Method), backend.Method, colorCyan, backend.URLPattern, colorReset)
	c.cmd.Printf("\t\t\tTimeout: %s\n", backend.Timeout.String())
	c.cmd.Printf("\t\t\tHosts: %v\n", backend.Host)

	if debug > 1 {
		c.cmd.Printf("\t\t\tConcurrent calls: %d\n", backend.ConcurrentCalls)
		c.cmd.Printf("\t\t\tHost sanitization disabled: %v\n", backend.HostSanitizationDisabled)
		c.cmd.Printf("\t\t\tTarget: %s\n", backend.Target)
		c.cmd.Printf("\t\t\tDeny: %v, Allow: %v\n", backend.DenyList, backend.AllowList)
		c.cmd.Printf("\t\t\tMapping: %+v\n", backend.Mapping)
		c.cmd.Printf("\t\t\tGroup: %s\n", backend.Group)
		c.cmd.Printf("\t\t\tEncoding: %s\n", backend.Encoding)
		c.cmd.Printf("\t\t\tIs collection: %+v\n", backend.IsCollection)
		c.cmd.Printf("\t\t\tSD: %+v\n", backend.SD)
	}

	if debug > 1 || len(backend.ExtraConfig) > 0 {
		c.cmd.Printf("\t\t\t%s%d backend component configuration(s):%s\n", colorGreen, len(backend.ExtraConfig), colorReset)
		for k, e := range backend.ExtraConfig {
			c.cmd.Printf("\t\t\t\t%s%s%s\n", colorYellow, k, colorReset)

			if debug > 1 {
				c.cmd.Printf("\t\t\t\t\t%+v\n", e)
			}
		}
	}
	c.cmd.Println("")
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
	cfg.Debug = cfg.Debug || debug > 0
	if port != 0 {
		cfg.Port = port
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	krakendgin.DefaultFactory(proxy.DefaultFactory(logging.NoOp), logging.NoOp).NewWithContext(ctx).Run(cfg)
	return nil
}
