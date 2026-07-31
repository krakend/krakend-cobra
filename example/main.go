package main

import (
	"os"

	cmd "github.com/krakend/krakend-cobra/v3"
	koanf "github.com/krakend/krakend-koanf/v2"
	"github.com/luraproject/lura/v3/config"
	"github.com/luraproject/lura/v3/logging"
	"github.com/luraproject/lura/v3/proxy"
	"github.com/luraproject/lura/v3/router/gin"
)

func main() {
	cmd.Execute(koanf.New(), func(serviceConfig config.ServiceConfig) {
		logger, _ := logging.NewLogger("DEBUG", os.Stdout, "")
		gin.DefaultFactory(proxy.DefaultFactory(logger), logger).New().Run(serviceConfig)
	})
}
