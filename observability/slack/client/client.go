package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	_ "embed"

	"github.com/neatplatform/compose/observability/slack/schema"
)

const baseURL = "https://slack.com/api"

var (
	//go:embed message-fixture.json
	MessageFixture string

	//go:embed modal-fixture.json
	ModalFixture string
)

type Client struct {
	logger    *slog.Logger
	authToken string
	client    *http.Client
}

func New(logger *slog.Logger, authToken string) *Client {
	c := &Client{
		logger:    logger,
		authToken: authToken,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	return c
}

func (c *Client) post(url string, req, resp any) error {
	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error encoding request: %w", err)
	}

	respBody, err := c.postRaw(url, b)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(respBody, resp); err != nil {
		return fmt.Errorf("error decoding response body: %w", err)
	}

	return nil
}

func (c *Client) postRaw(url string, body []byte) ([]byte, error) {
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Authorization", "Bearer "+c.authToken)

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending http request: %w", err)
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("unexpected http status code: [%d]", httpResp.StatusCode)
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading http response body: %w", err)
	}

	return respBody, nil
}

func (c *Client) PostMessageFixture() error {
	c.logger.Info("[client] Posting message fixture ...")

	url, _ := url.JoinPath(baseURL, "chat.postMessage")

	respBody, err := c.postRaw(url, []byte(MessageFixture))
	if err != nil {
		return err
	}

	var resp schema.PostMessageResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("error decoding response body: %w", err)
	}

	c.logger.Info("[client] Posted message fixture.",
		slog.Any("resp", resp),
	)

	return nil
}

func (c *Client) PostMessageResponse(responseURL string, message schema.MessageResponse) (string, error) {
	c.logger.Info("[client] Posting message response ...")

	var respBody string

	if err := c.post(responseURL, message, &respBody); err != nil {
		return "", fmt.Errorf("error posting message: %w", err)
	}

	c.logger.Info("[client] Posted message response.",
		slog.String("resp", respBody),
	)

	return respBody, nil
}

func (c *Client) PostMessage(req *schema.PostMessageRequest) (*schema.PostMessageResponse, error) {
	c.logger.Info("[client] Posting message ...")

	url, _ := url.JoinPath(baseURL, "chat.postMessage")
	resp := &schema.PostMessageResponse{}

	if err := c.post(url, req, resp); err != nil {
		return nil, fmt.Errorf("error posting message: %w", err)
	}

	c.logger.Info("[client] Posted message.",
		slog.Any("resp", resp),
	)

	return resp, nil
}

func (c *Client) AddReaction(req *schema.AddReactionRequest) (*schema.AddReactionResponse, error) {
	c.logger.Info("[client] Adding reaction ...")

	url, _ := url.JoinPath(baseURL, "reactions.add")
	resp := &schema.AddReactionResponse{}

	if err := c.post(url, req, resp); err != nil {
		return nil, fmt.Errorf("error adding reaction: %w", err)
	}

	c.logger.Info("[client] Added reaction.",
		slog.Any("resp", resp),
	)

	return resp, nil
}

func (c *Client) OpenView(req *schema.OpenViewRequest) (*schema.OpenViewResponse, error) {
	c.logger.Info("[client] Opening view ...")

	url, _ := url.JoinPath(baseURL, "views.open")
	resp := &schema.OpenViewResponse{}

	if err := c.post(url, req, resp); err != nil {
		return nil, fmt.Errorf("error opening view: %w", err)
	}

	c.logger.Info("[client] Opened view.",
		slog.Any("resp", resp),
	)

	return resp, nil
}
