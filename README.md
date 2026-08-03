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

	"github.com/krakend/krakend-cobra/v3"
	"github.com/krakend/krakend-koanf/v2"
	"github.com/luraproject/lura/v3/config"
	"github.com/luraproject/lura/v3/logging"
	"github.com/luraproject/lura/v3/proxy"
	krakendgin "github.com/luraproject/lura/v3/router/gin"
)

func main() {

	cmd.Execute(koanf.New(), func(serviceConfig config.ServiceConfig) {
		logger, _ := logging.NewLogger("DEBUG", os.Stdout, "")
		krakendgin.DefaultFactory(proxy.DefaultFactory(logger), logger).New().Run(serviceConfig)
	})

}
```

## Available commands

The `cmd` package includes four commands: `check`, `version`, `audit`, `help` and `run`.

1. *check* validates the received config file.
2. *help* displays details about any command.
3. *run* executes the passed executor once the received flags overwrite the parsed config.

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

KrakenD is a high-performance API gateway that helps you publish, secure, control, and monitor your services

Usage:
  krakend [command]

Available Commands:
  audit       Audits a KrakenD configuration.
  check       Validates that the configuration file is valid.
  help        Help about any command
  run         Runs the KrakenD server.
  version     Shows KrakenD version.

Flags:
  -h, --help   help for krakend

Use "krakend [command] --help" for more information about a command.

```
