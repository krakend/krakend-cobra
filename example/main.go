package main

import (
	"os"

	"github.com/devopsfaith/krakend-cobra"
	"github.com/devopsfaith/krakend-viper"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	krakendgin "github.com/devopsfaith/krakend/router/gin"
)

func main() {

	cmd.Execute(viper.New(), func(serviceConfig config.ServiceConfig) {
		logger, _ := logging.NewLogger("DEBUG", os.Stdout, "")
		krakendgin.DefaultFactory(proxy.DefaultFactory(logger), logger).New().Run(serviceConfig)
	})

}
