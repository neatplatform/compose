package signal

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/snappy"
	"github.com/google/uuid"
	"github.com/grafana/loki/pkg/push"
)

type (
	LokiPushRequest struct {
		Streams []LokiStream `json:"streams"`
	}

	LokiStream struct {
		Stream map[string]string `json:"stream"`
		Values [][3]any          `json:"values"`
	}
)

func runLoki(args []string) {
	fs := flag.NewFlagSet("loki", flag.ExitOnError)
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
		fmt.Println("A Loki push endpoint is required.")
		os.Exit(1)
	}

	baseLabels := []string{
		"name", "volt",
		"env", "local",
		"tenant", "test",
	}

	for i, msg := range messages {
		now := time.Now()

		labels := append(baseLabels,
			"level", "info",
		)

		metadata := []string{
			"uuid", uuid.NewString(),
		}

		var data []byte
		var err error

		if *proto {
			data, err = buildLokiProtoPushRequest(now, labels, metadata, msg)
			if err == nil {
				err = sendLokiPushRequest(url, "application/x-protobuf", "", data)
			}
		} else {
			data, err = buildLokiJSONPushRequest(now, labels, metadata, msg)
			if err == nil {
				err = sendLokiPushRequest(url, "application/json", "", data)
			}
		}

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Sent log:  #%-2d  message=%s\n", i+1, msg)
		}

		time.Sleep(randMs(500, 1000))
	}
}

func buildLokiJSONPushRequest(t time.Time, labels, metadata []string, message string) ([]byte, error) {
	labelMap := make(map[string]string, len(labels)/2)
	for i := 0; i+1 < len(labels); i += 2 {
		labelMap[labels[i]] = labels[i+1]
	}

	req := LokiPushRequest{
		Streams: []LokiStream{
			{
				Stream: labelMap,
				Values: [][3]any{
					{strconv.FormatInt(t.UnixNano(), 10), message, metadata},
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

func buildLokiProtoPushRequest(t time.Time, labels, metadata []string, message string) ([]byte, error) {
	parts := make([]string, 0, len(labels)/2)
	for i := 0; i+1 < len(labels); i += 2 {
		parts = append(parts, fmt.Sprintf(`%s=%q`, labels[i], labels[i+1]))
	}

	labelString := fmt.Sprintf("{%s}", strings.Join(parts, ","))

	structuredMetadata := make(push.LabelsAdapter, 0, len(metadata)/2)
	for i := 0; i+1 < len(metadata); i += 2 {
		structuredMetadata = append(structuredMetadata, push.LabelAdapter{
			Name:  metadata[i],
			Value: metadata[i+1],
		})
	}

	req := push.PushRequest{
		Streams: []push.Stream{
			{
				Labels: labelString,
				Entries: []push.Entry{
					{
						Timestamp:          t,
						Line:               message,
						StructuredMetadata: structuredMetadata,
					},
				},
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

func sendLokiPushRequest(url, contentType, tenant string, data []byte) error {
	// Compress the request body using snappy.
	compressed := snappy.Encode(nil, data)

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("error creating http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Content-Encoding", "snappy")

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
