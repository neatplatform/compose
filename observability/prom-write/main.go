package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: prom-write <remote_write_url>")
		os.Exit(1)
	}

	url := os.Args[1]

	var counter float64

	for i := range 10 {
		counter++
		now := time.Now()

		req := buildWriteRequest(
			"demo_requests_total",
			map[string]string{
				"job":      "prom-write",
				"instance": "localhost",
			},
			counter,
			now,
		)

		if err := sendWriteRequest(url, req); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Sent  sample #%-2d  value=%-2.0f  ts=%s\n", i+1, counter, now.Format(time.RFC3339))
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func buildWriteRequest(name string, labels map[string]string, value float64, timestamp time.Time) *prompb.WriteRequest {
	pbLabels := []prompb.Label{
		{
			Name:  "__name__",
			Value: name,
		},
	}

	for k, v := range labels {
		pbLabels = append(pbLabels, prompb.Label{
			Name:  k,
			Value: v,
		})
	}

	sample := prompb.Sample{
		Value:     value,
		Timestamp: timestamp.UnixMilli(),
	}

	return &prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels:  pbLabels,
				Samples: []prompb.Sample{sample},
			},
		},
	}
}

func sendWriteRequest(url string, req *prompb.WriteRequest) error {
	// Marshal the request to protobuf.
	data, err := req.Marshal()
	if err != nil {
		return fmt.Errorf("error on marshalling the request: %w", err)
	}

	// Compress the data using snappy.
	compressed := snappy.Encode(nil, data)

	// Post to the Prometheus remote write endpoint.
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("error on creating the HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("Content-Encoding", "snappy")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Error on executing the HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		body := strings.Trim(string(b), "\n")
		return fmt.Errorf("HTTP request failed: [%d] %s", resp.StatusCode, body)
	}

	return nil
}
