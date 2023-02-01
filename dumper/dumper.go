package dumper

import (
	"net/http"
	"sort"

	"github.com/luraproject/lura/v2/config"
	"github.com/spf13/cobra"
)

func New(cmd *cobra.Command, prefix string, verboseLevel int) Dumper {
	return NewWithColors(cmd, prefix, verboseLevel, true)
}

func NewWithColors(cmd *cobra.Command, prefix string, verboseLevel int, enableColors bool) Dumper {
	d := Dumper{
		cmd:             cmd,
		checkDumpPrefix: prefix,
		verboseLevel:    verboseLevel,
	}
	if !enableColors {
		return d
	}
	d.colorRed = ColorRed
	d.colorGreen = ColorGreen
	d.colorReset = ColorReset
	d.colorBlue = ColorBlue
	d.colorCyan = ColorCyan
	d.colorYellow = ColorYellow
	d.colorMagenta = ColorMagenta
	d.colorWhite = ColorWhite
	return d
}

type Dumper struct {
	cmd             *cobra.Command
	checkDumpPrefix string
	verboseLevel    int
	colorRed        string
	colorGreen      string
	colorReset      string
	colorBlue       string
	colorCyan       string
	colorYellow     string
	colorMagenta    string
	colorWhite      string
}

func (c Dumper) Dump(v config.ServiceConfig) error {
	c.cmd.Printf("%sGlobal settings%s\n", c.colorGreen, c.colorReset)
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
		c.cmd.Printf("%sSequential start: %t\n", c.checkDumpPrefix, v.SequentialStart)
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
		c.cmd.Printf("%s%sNo TLS section defined%s\n", c.checkDumpPrefix, c.colorRed, c.colorReset)
	}

	if v.Plugin != nil {
		c.cmd.Printf("%sFolder: %s\n", c.checkDumpPrefix, v.Plugin.Folder)
		c.cmd.Printf("%sPattern: %s\n", c.checkDumpPrefix, v.Plugin.Pattern)
	} else if c.verboseLevel > 1 {
		c.cmd.Printf("%s%sNo Plugin section defined%s\n", c.checkDumpPrefix, c.colorRed, c.colorReset)
	}

	if c.verboseLevel > 1 || len(v.ExtraConfig) > 0 {
		c.cmd.Printf("%s%d global component configuration(s):%s\n", c.colorGreen, len(v.ExtraConfig), c.colorReset)
		c.dumpExtraConfig(v.ExtraConfig, "")
	}

	c.cmd.Printf("%s%d API endpoint(s):%s\n", c.colorGreen, len(v.Endpoints), c.colorReset)
	for _, endpoint := range v.Endpoints {
		c.dumpEndpoint(endpoint)
	}

	c.cmd.Printf("%s%d async agent(s):%s\n", c.colorGreen, len(v.AsyncAgents), c.colorReset)
	for _, agent := range v.AsyncAgents {
		c.dumpAgent(agent)
	}
	return nil
}

func (c Dumper) dumpAgent(agent *config.AsyncAgent) {
	c.cmd.Printf("%s- %s%s%s\n", c.checkDumpPrefix, c.colorCyan, agent.Name, c.colorReset)

	if c.verboseLevel > 1 {
		c.cmd.Printf("%sEncoding: %s\n", c.checkDumpPrefix, agent.Encoding)

		c.cmd.Printf("%sConsumer Timeout: %s\n", c.checkDumpPrefix, agent.Consumer.Timeout.String())
		c.cmd.Printf("%sConsumer Workers: %d\n", c.checkDumpPrefix, agent.Consumer.Workers)
		c.cmd.Printf("%sConsumer Topic: %s\n", c.checkDumpPrefix, agent.Consumer.Topic)
		c.cmd.Printf("%sConsumer Max Rate: %f\n", c.checkDumpPrefix, agent.Consumer.MaxRate)

		c.cmd.Printf("%sConnection Max Retries: %d\n", c.checkDumpPrefix, agent.Connection.MaxRetries)
		c.cmd.Printf("%sConnection Backoff Strategy: %s\n", c.checkDumpPrefix, agent.Connection.BackoffStrategy)
		c.cmd.Printf("%sConnection Health Interval: %s\n", c.checkDumpPrefix, agent.Connection.HealthInterval.String())
	}

	if c.verboseLevel > 1 || len(agent.ExtraConfig) > 0 {
		c.cmd.Printf("%s%s%d agent component configuration(s):%s\n", c.checkDumpPrefix, c.colorGreen, len(agent.ExtraConfig), c.colorReset)
		c.dumpExtraConfig(agent.ExtraConfig, c.checkDumpPrefix)
	}

	c.cmd.Printf("%s%sConnecting to %d backend(s):%s\n", c.checkDumpPrefix, c.colorGreen, len(agent.Backend), c.colorReset)
	for _, backend := range agent.Backend {
		c.dumpBackend(backend)
	}
}

func (c Dumper) dumpEndpoint(endpoint *config.EndpointConfig) {
	c.cmd.Printf("%s- %s%s%s %s%s\n", c.checkDumpPrefix, c.methodColor(endpoint.Method), endpoint.Method, c.colorCyan, endpoint.Endpoint, c.colorReset)
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
		c.cmd.Printf("%s%s%d endpoint component configuration(s):%s\n", c.checkDumpPrefix, c.colorGreen, len(endpoint.ExtraConfig), c.colorReset)
		c.dumpExtraConfig(endpoint.ExtraConfig, c.checkDumpPrefix)
	}

	c.cmd.Printf("%s%sConnecting to %d backend(s):%s\n", c.checkDumpPrefix, c.colorGreen, len(endpoint.Backend), c.colorReset)
	for _, backend := range endpoint.Backend {
		c.dumpBackend(backend)
	}
}

func (c Dumper) dumpBackend(backend *config.Backend) {
	prefix := c.checkDumpPrefix + c.checkDumpPrefix
	c.cmd.Printf("%s[+] %s%s%s %s%s\n", prefix, c.methodColor(backend.Method), backend.Method, c.colorCyan, backend.URLPattern, c.colorReset)
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
		c.cmd.Printf("%s%s%d backend component configuration(s):%s\n", prefix, c.colorGreen, len(backend.ExtraConfig), c.colorReset)
		c.dumpExtraConfig(backend.ExtraConfig, prefix)
	}
	c.cmd.Println("")
}

func (c Dumper) dumpExtraConfig(cfg config.ExtraConfig, prefix string) {
	var keys []string

	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c.cmd.Printf("%s%s- %s%s\n", prefix, c.colorYellow, k, c.colorReset)
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

func (c Dumper) methodColor(s string) string {
	switch s {
	case http.MethodGet:
		return c.colorBlue
	case http.MethodPost:
		return c.colorCyan
	case http.MethodPut:
		return c.colorYellow
	case http.MethodDelete:
		return c.colorRed
	case http.MethodPatch:
		return c.colorGreen
	case http.MethodHead:
		return c.colorMagenta
	case http.MethodOptions:
		return c.colorWhite
	}
	return ""
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
