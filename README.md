KrakenD Cobra
====

An adapter of the [cobra](http://github.com/spf13/cobra) lib for the [KrakenD](http://www.krakend.io) framework

Package cmd defines the cobra command structs and an execution method for adding an improved CLI to
KrakenD based api gateways

## Basic example

```
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
```

## Available commands

The `cmd` package includes two commands: `check` and `run`. 

1. *check* validates the received config file.
2. *run* executes the passed executor once the received flags overwrite the parsed config.

```
$ ./krakend

`7MMF' `YMM'                  `7MM                         `7MM"""Yb.
  MM   .M'                      MM                           MM    `Yb.
  MM .d"     `7Mb,od8 ,6"Yb.    MM  ,MP'.gP"Ya `7MMpMMMb.    MM     `Mb
  MMMMM.       MM' "'8)   MM    MM ;Y  ,M'   Yb  MM    MM    MM      MM
  MM  VMA      MM     ,pm9MM    MM;Mm  8M""""""  MM    MM    MM     ,MP
  MM   `MM.    MM    8M   MM    MM `Mb.YM.    ,  MM    MM    MM    ,dP'
.JMML.   MMb..JMML.  `Moo9^Yo..JMML. YA.`Mbmmd'.JMML  JMML..JMMmmmdP'
_______________________________________________________________________

Version: undefined

The API Gateway builder

Usage:
  krakend [command]

Available Commands:
  check       Validates that the configuration file is valid.
  run         Run the KrakenD server.

Flags:
  -c, --config string   Path to the configuration filename
  -d, --debug           Enable the debug

Use "krakend [command] --help" for more information about a command.
```
