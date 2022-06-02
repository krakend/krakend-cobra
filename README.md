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

	"github.com/krakendio/krakend-cobra/v2"
	"github.com/krakendio/krakend-viper/v2"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	krakendgin "github.com/luraproject/lura/v2/router/gin"
)

func main() {

	cmd.Execute(viper.New(), func(serviceConfig config.ServiceConfig) {
		logger, _ := logging.NewLogger("DEBUG", os.Stdout, "")
		krakendgin.DefaultFactory(proxy.DefaultFactory(logger), logger).New().Run(serviceConfig)
	})

}
```

## Available commands

The `cmd` package includes four commands: `check`, `check-plugin`, `help` and `run`.

1. *check* validates the received config file.
2. *check-plugin* validates the dependencies shared between the binary and a plugin.
3. *help* displays details about any command.
4. *run* executes the passed executor once the received flags overwrite the parsed config.

```
$ ./krakend
 ╓▄█                          ▄▄▌                               ╓██████▄µ
▐███  ▄███╨▐███▄██H╗██████▄  ║██▌ ,▄███╨ ▄██████▄  ▓██▌█████▄  ███▀╙╙▀▀███╕
▐███▄███▀  ▐█████▀"╙▀▀"╙▀███ ║███▄███┘  ███▀""▀███ ████▀╙▀███H ███     ╙███
▐██████▌   ▐███⌐  ,▄████████M║██████▄  ║██████████M███▌   ███H ███     ,███
▐███╨▀███µ ▐███   ███▌  ,███M║███╙▀███  ███▄```▄▄` ███▌   ███H ███,,,╓▄███▀
▐███  ╙███▄▐███   ╙█████████M║██▌  ╙███▄`▀███████╨ ███▌   ███H █████████▀
                     ``                     `'`
Version: undefined

The API Gateway builder

Usage:
  krakend [command]

Available Commands:
  check        Validates that the configuration file is valid.
  check-plugin Checks your plugin dependencies are compatible.
  help         Help about any command
  run          Runs the KrakenD server.

Flags:
  -h, --help   help for krakend

Use "krakend [command] --help" for more information about a command.

```
