package main

import (
	"os"

	cmd "github.com/krakendio/krakend-cobra/v2"
	viper "github.com/krakendio/krakend-viper/v2"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/router/gin"
)

func main() {
	cmd.Execute(viper.New(), func(serviceConfig config.ServiceConfig) {
		logger, _ := logging.NewLogger("DEBUG", os.Stdout, "")
		gin.DefaultFactory(proxy.DefaultFactory(logger), logger).New().Run(serviceConfig)
	})
}
