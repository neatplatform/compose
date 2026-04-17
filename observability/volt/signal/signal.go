package signal

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"
)

const helpText = `
Usage: volt signal <command> [options] <args>

Commands:

  forward        Send a few logs via the Fluentd Forward protocol.
  loki           Send a few logs via the Loki Push API.
  prometheus     Push a few Prometheus metrics via the Prometheus Remote Write protocol.
  opentelemetry  Send a few OpenTelemetry logs, metrics, and/or traces via the OTLP HTTP or gRPC.

Examples:

  volt signal forward localhost:24224
  volt signal -json forward localhost:24224

  volt signal loki http://localhost:3100/loki/api/v1/push
  volt signal loki -proto http://localhost:3100/loki/api/v1/push

  volt signal prometheus http://localhost:9090/api/v1/write
  volt signal prometheus -proto http://localhost:9090/api/v1/write

  volt signal opentelemetry localhost:4318
  volt signal opentelemetry -grpc localhost:4317

`

func Run(args []string) {
	if len(args) < 2 {
		fmt.Print(helpText)
		os.Exit(1)
	}

	switch args[1] {
	case "forward":
		runForward(args[1:])
	case "loki":
		runLoki(args[1:])
	case "prometheus":
		runPrometheus(args[1:])
	case "opentelemetry":
		runOpenTelemetry(args[1:])
	default:
		fmt.Print(helpText)
		os.Exit(1)
	}
}

func randMs(min, max int) time.Duration {
	i := min + rand.IntN(max-min)
	return time.Duration(i) * time.Millisecond
}

var messages = []string{
	"Hello, World!",
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
	"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
	"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
	"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
	"Goodbye, World!",
}
