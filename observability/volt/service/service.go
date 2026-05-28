package service

import (
	"fmt"
	"os"

	"github.com/neatplatform/compose/observability/volt/service/http"
)

const helpText = `
Usage: volt service <command> [options]

All services emit logs, expose metrics, and serve pprof profiles.

If a service can be exposed to the public internet with the -ngrok flag,
set an ngrok authtoken using the NGROK_AUTHTOKEN environment variable.

Commands:

  echo       Run an HTTP echo service that responds back with the request details.
  webhook    Run an HTTP webhook service that receives and logs incoming requests.

Examples:

  volt service echo
  volt service echo -port 9999

  volt service webhook
  volt service webhook -port 9999
  volt service webhook -endpoint "/callback"
  volt service webhook -ngrok

`

func Run(args []string) {
	if len(args) < 2 {
		fmt.Print(helpText)
		os.Exit(1)
	}

	switch args[1] {
	case "echo":
		http.RunEcho(args[1:])
	case "webhook":
		http.RunWebhook(args[1:])
	default:
		fmt.Print(helpText)
		os.Exit(1)
	}
}
