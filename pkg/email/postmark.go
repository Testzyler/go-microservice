package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultPostmarkEndpoint = "https://api.postmarkapp.com/email"

type Message struct {
	From          string `json:"From,omitempty"`
	To            string `json:"To"`
	Subject       string `json:"Subject"`
	HtmlBody      string `json:"HtmlBody,omitempty"`
	TextBody      string `json:"TextBody,omitempty"`
	MessageStream string `json:"MessageStream,omitempty"`
}

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type PostmarkConfig struct {
	ServerToken   string
	From          string
	MessageStream string
	Endpoint      string
	Timeout       time.Duration
}

type PostmarkClient struct {
	client        *http.Client
	endpoint      string
	serverToken   string
	defaultFrom   string
	messageStream string
}

func NewPostmarkClient(cfg PostmarkConfig) (*PostmarkClient, error) {
	serverToken := strings.TrimSpace(cfg.ServerToken)
	if serverToken == "" {
		return nil, errors.New("postmark server token is required")
	}
	from := strings.TrimSpace(cfg.From)
	if from == "" {
		return nil, errors.New("postmark from address is required")
	}
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = defaultPostmarkEndpoint
	}
	stream := strings.TrimSpace(cfg.MessageStream)
	if stream == "" {
		stream = "outbound"
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &PostmarkClient{
		client:        &http.Client{Timeout: timeout},
		endpoint:      endpoint,
		serverToken:   serverToken,
		defaultFrom:   from,
		messageStream: stream,
	}, nil
}

func (c *PostmarkClient) Send(ctx context.Context, msg Message) error {
	if c == nil {
		return errors.New("postmark client is nil")
	}
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return errors.New("postmark recipient is required")
	}
	subject := strings.TrimSpace(msg.Subject)
	if subject == "" {
		return errors.New("postmark subject is required")
	}
	body := strings.TrimSpace(msg.HtmlBody)
	text := strings.TrimSpace(msg.TextBody)
	if body == "" && text == "" {
		return errors.New("postmark body is required")
	}
	from := strings.TrimSpace(msg.From)
	if from == "" {
		from = c.defaultFrom
	}
	stream := strings.TrimSpace(msg.MessageStream)
	if stream == "" {
		stream = c.messageStream
	}

	payload := Message{
		From:          from,
		To:            to,
		Subject:       subject,
		HtmlBody:      body,
		TextBody:      text,
		MessageStream: stream,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("postmark marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("postmark build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", c.serverToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("postmark request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("postmark error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
