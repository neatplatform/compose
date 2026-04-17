package signal

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/google/uuid"
)

func runForward(args []string) {
	fs := flag.NewFlagSet("forward", flag.ExitOnError)
	jsonEnc := fs.Bool("json", false, "Use JSON encoding")

	if err := fs.Parse(args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	if *jsonEnc {
		fmt.Println("Using JSON encoding")
	} else {
		fmt.Println("Using MessagePack encoding")
	}

	addr := fs.Arg(0)
	if addr == "" {
		fmt.Println("A Forward endpoint is required.")
		os.Exit(1)
	}

	baseKV := []string{
		"name", "volt",
		"env", "local",
		"tenant", "test",
	}

	client, err := newFluentClient(addr, *jsonEnc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i, msg := range messages {
		kvs := append(baseKV,
			"uuid", uuid.NewString(),
		)

		if err := sendFluentLog(client, time.Now(), "info", msg, kvs...); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Sent log:  #%-2d  message=%s\n", i+1, msg)
		}

		time.Sleep(randMs(500, 1000))
	}
}

func newFluentClient(addr string, jsonEnc bool) (*fluent.Fluent, error) {
	config := fluent.Config{
		TagPrefix:          "volt",
		Async:              true,
		MarshalAsJSON:      jsonEnc,
		SubSecondPrecision: !jsonEnc,
	}

	if addr != "" {
		host, p, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}

		port, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}

		config.FluentHost, config.FluentPort = host, port
	}

	client, err := fluent.New(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Fluent client: %w", err)
	}

	return client, nil
}

func sendFluentLog(c *fluent.Fluent, t time.Time, level, message string, kv ...string) error {
	record := make(map[string]any, 3+len(kv)/2)

	// Add the fixed fields: timestamp, level, message
	record["timestamp"] = t.Format(time.RFC3339Nano)
	record["level"] = level
	record["message"] = message

	// Add the fields
	for i := 0; i+1 < len(kv); i += 2 {
		k, v := kv[i], kv[i+1]
		record[k] = v
	}

	// Send the record to the Forward endpoint.
	return c.PostWithTime("test", t, record)
}
