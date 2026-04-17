package main

import (
	"fmt"
	"os"

	"github.com/neatplatform/compose/observability/volt/service"
	"github.com/neatplatform/compose/observability/volt/signal"
	"github.com/neatplatform/compose/observability/volt/template"
)

const helpText = `
Usage: volt <command> ...

Commands:

  service     Run mock servers for local testing and troubleshooting.
  signal      Emit observability signals (logs, metrics, traces) using various protocols.
  template    Render a Go text template from JSON input data.

`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(helpText)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "service":
		service.Run(os.Args[1:])
	case "signal":
		signal.Run(os.Args[1:])
	case "template":
		template.Run(os.Args[1:])
	default:
		fmt.Print(helpText)
		os.Exit(1)
	}
}
