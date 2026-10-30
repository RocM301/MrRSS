package ai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type failoverTransport func(*http.Request) (*http.Response, error)

func (f failoverTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFailoverRetriesPrimaryAndRecoversOnNextRequest(t *testing.T) {
	var calls []string
	primaryHealthy := false
	transport := failoverTransport(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, r.URL.Host)
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > RequestTimeout || time.Until(deadline) < RequestTimeout-time.Second {
			t.Fatal("each AI attempt must have a 15-second deadline")
		}
		status, body := 200, `{"choices":[{"message":{"content":"ok"}}]}`
		if r.URL.Host == "primary.example" && !primaryHealthy {
			status, body = 503, `{"error":"unavailable"}`
		}
		return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	client := NewClientWithHTTPClient(ClientConfig{Endpoint: "https://primary.example/v1/chat/completions", Model: "primary", FallbackConfigs: []ClientConfig{{Endpoint: "https://backup.example/v1/chat/completions", Model: "backup"}}}, &http.Client{Transport: transport})
	if _, err := client.Request("", "hello"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls, ",") != "primary.example,primary.example,backup.example" {
		t.Fatalf("unexpected attempt order: %v", calls)
	}
	primaryHealthy = true
	calls = nil
	if _, err := client.Request("", "hello"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls, ",") != "primary.example" {
		t.Fatalf("did not recover to primary: %v", calls)
	}
}

func TestFailoverStopsWhenCallerCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	client := NewClientWithHTTPClient(ClientConfig{Endpoint: "https://primary.example/v1/chat/completions", FallbackConfigs: []ClientConfig{{Endpoint: "https://backup.example/v1/chat/completions"}}}, &http.Client{Transport: failoverTransport(func(r *http.Request) (*http.Response, error) { calls++; cancel(); return nil, context.Canceled })})
	_, err := client.RequestWithConfigContext(ctx, RequestConfig{UserPrompt: "hello"})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("cancel should stop all attempts: calls=%d err=%v", calls, err)
	}
}

func TestCustomHeadersAcceptLegacyPairsAndObjects(t *testing.T) {
	for _, value := range []string{`{"X-Relay":"value"}`, `[{"key":" X-Relay ","value":"value"},{"key":" ","value":"ignored"}]`} {
		headers, err := parseCustomHeaders(value)
		if err != nil || headers["X-Relay"] != "value" || len(headers) != 1 {
			t.Fatalf("legacy headers lost: %v %v", headers, err)
		}
	}
	if _, err := parseCustomHeaders(`[{"key":123}]`); err == nil {
		t.Fatal("accepted malformed headers")
	}
}
