package email

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPostmarkClientSend(t *testing.T) {
	t.Parallel()

	var got Message
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"MessageID":"msg","ErrorCode":0,"Message":"OK"}`))
	}))
	defer srv.Close()

	client, err := NewPostmarkClient(PostmarkConfig{
		ServerToken:   "token-123",
		From:          "no-reply@example.com",
		MessageStream: "outbound",
		Endpoint:      srv.URL,
		Timeout:       time.Second,
	})
	if err != nil {
		t.Fatalf("NewPostmarkClient: %v", err)
	}

	err = client.Send(context.Background(), Message{
		To:       "user@example.com",
		Subject:  "Hello",
		HtmlBody: "<p>Hi</p>",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if got.From != "no-reply@example.com" {
		t.Fatalf("expected From to be default, got %q", got.From)
	}
	if got.MessageStream != "outbound" {
		t.Fatalf("expected MessageStream to be default, got %q", got.MessageStream)
	}
	if got.To != "user@example.com" {
		t.Fatalf("expected To to be set, got %q", got.To)
	}

	if gotHeaders.Get("X-Postmark-Server-Token") != "token-123" {
		t.Fatalf("missing or wrong postmark token header")
	}
	if gotHeaders.Get("Accept") != "application/json" {
		t.Fatalf("missing accept header")
	}
	if !strings.EqualFold(gotHeaders.Get("Content-Type"), "application/json") {
		t.Fatalf("missing content-type header")
	}
}
