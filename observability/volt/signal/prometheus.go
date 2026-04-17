package signal

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

type (
	PrometheusWriteRequest struct {
		Timeseries []PrometheusTimeseries `json:"timeseries"`
	}

	PrometheusTimeseries struct {
		Labels  []PrometheusLabel  `json:"labels"`
		Samples []PrometheusSample `json:"samples"`
	}

	PrometheusLabel struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	PrometheusSample struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"` // milliseconds
	}
)

func runPrometheus(args []string) {
	fs := flag.NewFlagSet("prometheus", flag.ExitOnError)
	proto := fs.Bool("proto", false, "Use proto encoding")

	if err := fs.Parse(args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	if *proto {
		fmt.Println("Using proto encoding")
	} else {
		fmt.Println("Using JSON encoding")
	}

	url := fs.Arg(0)
	if url == "" {
		fmt.Println("A Prometheus remote write endpoint is required.")
		os.Exit(1)
	}

	baseLabels := []string{
		"job", "volt",
		"instance", "localhost",
		"env", "local",
		"tenant", "test",
	}

	var counter float64

	for i := range 10 {
		counter++
		now := time.Now()

		labels := append(baseLabels,
			"req_method", "GET",
			"req_path", "/",
		)

		var data []byte
		var err error

		if *proto {
			data, err = buildPrometheusProtoWriteRequest("demo_prometheus_requests_total", labels, counter, now)
			if err == nil {
				err = sendPrometheusWriteRequest(url, "application/x-protobuf", "", data)
			}
		} else {
			data, err = buildPrometheusJSONWriteRequest("demo_prometheus_requests_total", labels, counter, now)
			if err == nil {
				err = sendPrometheusWriteRequest(url, "application/json", "", data)
			}
		}

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Sent metric:  #%-2d  value=%-2.0f\n", i+1, counter)
		}

		time.Sleep(randMs(500, 1000))
	}
}

func buildPrometheusJSONWriteRequest(name string, labels []string, value float64, t time.Time) ([]byte, error) {
	promLabels := make([]PrometheusLabel, 0, 1+len(labels)/2)

	promLabels = append(promLabels, PrometheusLabel{
		Name:  "__name__",
		Value: name,
	})

	for i := 0; i+1 < len(labels); i += 2 {
		promLabels = append(promLabels, PrometheusLabel{
			Name:  labels[i],
			Value: labels[i+1],
		})
	}

	req := PrometheusWriteRequest{
		Timeseries: []PrometheusTimeseries{
			{
				Labels: promLabels,
				Samples: []PrometheusSample{
					{
						Value:     value,
						Timestamp: t.UnixMilli()},
				},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error encoding the request: %w", err)
	}

	return data, nil
}

func buildPrometheusProtoWriteRequest(name string, labels []string, value float64, t time.Time) ([]byte, error) {
	pbLabels := []prompb.Label{
		{
			Name:  "__name__",
			Value: name,
		},
	}

	for i := 0; i+1 < len(labels); i += 2 {
		pbLabels = append(pbLabels, prompb.Label{
			Name:  labels[i],
			Value: labels[i+1],
		})
	}

	sample := prompb.Sample{
		Value:     value,
		Timestamp: t.UnixMilli(),
	}

	req := &prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels:  pbLabels,
				Samples: []prompb.Sample{sample},
			},
		},
	}

	// Marshal the request to protobuf.
	data, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("error encoding the request: %w", err)
	}

	return data, nil
}

func sendPrometheusWriteRequest(url, contentType, tenant string, body []byte) error {
	// Compress the request body using snappy.
	compressed := snappy.Encode(nil, body)

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("error creating http request: %w", err)
	}

	httpReq.Header.Set("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("User-Agent", filepath.Base(os.Args[0]))
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	if tenant != "" {
		httpReq.Header.Set("X-Scope-OrgID", tenant)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending http request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		body := strings.Trim(string(b), "\n")
		return fmt.Errorf("unexpected http status: [%d] %s", resp.StatusCode, body)
	}

	return nil
}
