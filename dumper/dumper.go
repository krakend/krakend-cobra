package dumper

import (
	"errors"
	"net/http"
	"sort"

	"github.com/luraproject/lura/v2/config"
	"github.com/spf13/cobra"
)

func New(cmd *cobra.Command, prefix string, verboseLevel int) Dumper {
	return Dumper{
		cmd:             cmd,
		checkDumpPrefix: prefix,
		verboseLevel:    verboseLevel,
	}
}

type Dumper struct {
	cmd             *cobra.Command
	checkDumpPrefix string
	verboseLevel    int
}

func (c Dumper) Dump(v config.ServiceConfig) error {
	c.cmd.Printf("%sGlobal settings%s\n", ColorGreen, ColorReset)
	c.cmd.Printf("%sName: %s\n", c.checkDumpPrefix, v.Name)
	c.cmd.Printf("%sPort: %d\n", c.checkDumpPrefix, v.Port)

	if c.verboseLevel > 1 {
		c.cmd.Printf("%sDefault cache TTL: %s\n", c.checkDumpPrefix, v.CacheTTL.String())
		c.cmd.Printf("%sDefault timeout: %s\n", c.checkDumpPrefix, v.Timeout.String())
	}

	if len(v.Host) > 0 || c.verboseLevel > 1 {
		c.cmd.Printf("%sDefault backend hosts: %v\n", c.checkDumpPrefix, v.Host)
	}

	if c.verboseLevel > 2 {
		c.cmd.Printf("%sRead timeout: %s\n", c.checkDumpPrefix, v.ReadTimeout.String())
		c.cmd.Printf("%sWrite timeout: %s\n", c.checkDumpPrefix, v.WriteTimeout.String())
		c.cmd.Printf("%sIdle timeout: %s\n", c.checkDumpPrefix, v.IdleTimeout.String())
		c.cmd.Printf("%sRead header timeout: %s\n", c.checkDumpPrefix, v.ReadHeaderTimeout.String())
		c.cmd.Printf("%sIdle connection timeout: %s\n", c.checkDumpPrefix, v.IdleConnTimeout.String())
		c.cmd.Printf("%sResponse header timeout: %s\n", c.checkDumpPrefix, v.ResponseHeaderTimeout.String())
		c.cmd.Printf("%sExpect continue timeout: %s\n", c.checkDumpPrefix, v.ExpectContinueTimeout.String())
		c.cmd.Printf("%sDialer timeout: %s\n", c.checkDumpPrefix, v.DialerTimeout.String())
		c.cmd.Printf("%sDialer fallback delay: %s\n", c.checkDumpPrefix, v.DialerFallbackDelay.String())
		c.cmd.Printf("%sDialer keep alive: %s\n", c.checkDumpPrefix, v.DialerKeepAlive.String())
		c.cmd.Printf("%sDisable keep alives: %v\n", c.checkDumpPrefix, v.DisableKeepAlives)
		c.cmd.Printf("%sDisable compression: %v\n", c.checkDumpPrefix, v.DisableCompression)
		c.cmd.Printf("%sMax idle connections: %d\n", c.checkDumpPrefix, v.MaxIdleConns)
		c.cmd.Printf("%sMax idle connections per host: %d\n", c.checkDumpPrefix, v.MaxIdleConnsPerHost)
	}

	if v.TLS != nil {
		c.cmd.Printf("%sDisabled: %v\n", c.checkDumpPrefix, v.TLS.IsDisabled)
		c.cmd.Printf("%sPublic key: %s\n", c.checkDumpPrefix, v.TLS.PublicKey)
		c.cmd.Printf("%sPrivate key: %s\n", c.checkDumpPrefix, v.TLS.PrivateKey)
		c.cmd.Printf("%sEnable MTLS: %v\n", c.checkDumpPrefix, v.TLS.EnableMTLS)

		if c.verboseLevel > 1 {
			c.cmd.Printf("%sMin version: %s\n", c.checkDumpPrefix, v.TLS.MinVersion)
			c.cmd.Printf("%sMax version: %s\n", c.checkDumpPrefix, v.TLS.MaxVersion)
		}
		if c.verboseLevel > 2 {
			c.cmd.Printf("%sCurve preferences: %v\n", c.checkDumpPrefix, v.TLS.CurvePreferences)
			c.cmd.Printf("%sPrefer server cipher suites: %v\n", c.checkDumpPrefix, v.TLS.PreferServerCipherSuites)
			c.cmd.Printf("%sCipher suites: %v\n", c.checkDumpPrefix, v.TLS.CipherSuites)
		}
	} else if c.verboseLevel > 1 {
		c.cmd.Printf("%s%sNo TLS section defined%s\n", c.checkDumpPrefix, ColorRed, ColorReset)
	}

	if v.Plugin != nil {
		c.cmd.Printf("%sFolder: %s\n", c.checkDumpPrefix, v.Plugin.Folder)
		c.cmd.Printf("%sPattern: %s\n", c.checkDumpPrefix, v.Plugin.Pattern)
	} else if c.verboseLevel > 1 {
		c.cmd.Printf("%s%sNo Plugin section defined%s\n", c.checkDumpPrefix, ColorRed, ColorReset)
	}

	if c.verboseLevel > 1 || len(v.ExtraConfig) > 0 {
		c.cmd.Printf("%s%d global component configuration(s):%s\n", ColorGreen, len(v.ExtraConfig), ColorReset)
		c.dumpExtraConfig(v.ExtraConfig, "")
	}

	if len(v.Endpoints) == 0 {
		return errors.New("no endpoints defined")
	}

	c.cmd.Printf("%s%d API endpoints:%s\n", ColorGreen, len(v.Endpoints), ColorReset)
	for _, endpoint := range v.Endpoints {
		c.dumpEndpoint(endpoint)
	}
	return nil
}

func (c Dumper) dumpEndpoint(endpoint *config.EndpointConfig) {
	c.cmd.Printf("%s- %s%s%s %s%s\n", c.checkDumpPrefix, methodColor(endpoint.Method), endpoint.Method, ColorCyan, endpoint.Endpoint, ColorReset)
	c.cmd.Printf("%sTimeout: %s\n", c.checkDumpPrefix, endpoint.Timeout.String())

	if c.verboseLevel > 1 || len(endpoint.QueryString) > 0 {
		c.cmd.Printf("%sQueryString: %v\n", c.checkDumpPrefix, endpoint.QueryString)
	}

	if c.verboseLevel > 1 {
		c.cmd.Printf("%sCacheTTL: %s\n", c.checkDumpPrefix, endpoint.CacheTTL.String())
		c.cmd.Printf("%sHeaders to pass: %v\n", c.checkDumpPrefix, endpoint.HeadersToPass)
		c.cmd.Printf("%sOutputEncoding: %s\n", c.checkDumpPrefix, endpoint.OutputEncoding)
	}

	if c.verboseLevel > 1 || endpoint.ConcurrentCalls > 1 {
		c.cmd.Printf("%sConcurrent calls: %d\n", c.checkDumpPrefix, endpoint.ConcurrentCalls)
	}

	if c.verboseLevel > 1 || len(endpoint.ExtraConfig) > 0 {
		c.cmd.Printf("%s%s%d endpoint component configuration(s):%s\n", c.checkDumpPrefix, ColorGreen, len(endpoint.ExtraConfig), ColorReset)
		c.dumpExtraConfig(endpoint.ExtraConfig, c.checkDumpPrefix)
	}

	c.cmd.Printf("%s%sConnecting to %d backend(s):%s\n", c.checkDumpPrefix, ColorGreen, len(endpoint.Backend), ColorReset)
	for _, backend := range endpoint.Backend {
		c.dumpBackend(backend)
	}
}

func (c Dumper) dumpBackend(backend *config.Backend) {
	prefix := c.checkDumpPrefix + c.checkDumpPrefix
	c.cmd.Printf("%s[+] %s%s%s %s%s\n", prefix, methodColor(backend.Method), backend.Method, ColorCyan, backend.URLPattern, ColorReset)
	c.cmd.Printf("%sTimeout: %s\n", prefix, backend.Timeout.String())
	c.cmd.Printf("%sHosts: %v\n", prefix, backend.Host)

	if c.verboseLevel > 1 {
		c.cmd.Printf("%sConcurrent calls: %d\n", prefix, backend.ConcurrentCalls)
		c.cmd.Printf("%sHost sanitization disabled: %v\n", prefix, backend.HostSanitizationDisabled)
		c.cmd.Printf("%sTarget: %s\n", prefix, backend.Target)
		c.cmd.Printf("%sDeny: %v, Allow: %v\n", prefix, backend.DenyList, backend.AllowList)
		c.cmd.Printf("%sMapping: %+v\n", prefix, backend.Mapping)
		c.cmd.Printf("%sGroup: %s\n", prefix, backend.Group)
		c.cmd.Printf("%sEncoding: %s\n", prefix, backend.Encoding)
		c.cmd.Printf("%sIs collection: %+v\n", prefix, backend.IsCollection)
		c.cmd.Printf("%sSD: %+v\n", prefix, backend.SD)
	}

	if c.verboseLevel > 1 || len(backend.ExtraConfig) > 0 {
		c.cmd.Printf("%s%s%d backend component configuration(s):%s\n", prefix, ColorGreen, len(backend.ExtraConfig), ColorReset)
		c.dumpExtraConfig(backend.ExtraConfig, prefix)
	}
	c.cmd.Println("")
}

func (c Dumper) dumpExtraConfig(cfg config.ExtraConfig, prefix string) {
	keys := []string{}
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c.cmd.Printf("%s%s- %s%s\n", prefix, ColorYellow, k, ColorReset)
		if c.verboseLevel > 1 {
			switch s := cfg[k].(type) {
			case map[string]interface{}:
				for i, v := range s {
					c.cmd.Printf("\t%s%s: %+v\n", prefix, i, v)
				}
			case []interface{}:
				c.cmd.Printf("\t%s: %+v\n", prefix, s)
			default:
				c.cmd.Printf("\t%s: %+v\n", prefix, s)
			}
		}
	}
}

var methodColors = map[string]string{
	http.MethodGet:     ColorBlue,
	http.MethodPost:    ColorCyan,
	http.MethodPut:     ColorYellow,
	http.MethodDelete:  ColorRed,
	http.MethodPatch:   ColorGreen,
	http.MethodHead:    ColorMagenta,
	http.MethodOptions: ColorWhite,
}

func methodColor(method string) string {
	m, ok := methodColors[method]
	if !ok {
		return ColorGreen
	}
	return m
}

const (
	ColorGreen   = "\033[32;1m"
	ColorWhite   = "\033[37;1m"
	ColorYellow  = "\033[33;1m"
	ColorRed     = "\033[31;1m"
	ColorBlue    = "\033[34;1m"
	ColorMagenta = "\033[36;1m"
	ColorCyan    = "\033[36;1m"
	ColorReset   = "\033[0m"
)
