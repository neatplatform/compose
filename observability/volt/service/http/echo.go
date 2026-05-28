package http

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func RunEcho(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	port := fs.Int("port", 8080, "Port to listen on")

	if err := fs.Parse(args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	if *port <= 0 || *port > 65535 {
		fmt.Println("Invalid port number.")
		os.Exit(1)
	}

	newEchoService().Start(*port, false, "")
}

// echoService is an HTTP echo service that responds back with the request details.
type echoService struct {
	*service
}

func newEchoService() *echoService {
	s := &echoService{
		service: newService(),
	}

	s.RegisterRoute("/", s.echo)

	return s
}

func (s *echoService) echo(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	url := r.URL.Path
	query := r.URL.RawQuery

	var b strings.Builder
	for h, vs := range r.Header {
		fmt.Fprintf(&b, "  %s: %s\n", h, strings.Join(vs, ", "))
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyLen))
	if err != nil {
		s.logger.Error("Error reading request body.", "error", err)
		http.Error(w, "Failed to read request body.", http.StatusInternalServerError)
		return
	}

	// Simulate some processing time.
	time.Sleep(randMs(10, 100))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = fmt.Fprintf(w,
		"method: %s\nurl: %s\nquery: %s\nheaders:\n%sbody:\n  %s\n",
		method, url, query, b.String(), string(body),
	)
}
