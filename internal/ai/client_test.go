package ai

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDetectAPIProviderTreatsChatCompletionsAsOpenAICompatible(t *testing.T) {
	endpoint := "https://api.shiyutv.cn/v1/chat/completions"

	if got := DetectAPIProvider(endpoint); got != "openai" {
		t.Fatalf("DetectAPIProvider() = %q, want openai", got)
	}
}

func TestOpenAICompatibleFailureKeepsUpstreamError(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			body := `{"error":{"message":"The model gpt-5.4-mini does not exist","type":"invalid_request_error"}}`
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
		Timeout: 2 * time.Second,
	}

	client := NewClientWithHTTPClient(ClientConfig{
		APIKey:   "key",
		Endpoint: "https://api.shiyutv.cn/v1/chat/completions",
		Model:    "gpt-5.4-mini",
		Timeout:  2 * time.Second,
	}, httpClient)

	_, err := client.RequestWithConfig(RequestConfig{
		Model: "gpt-5.4-mini",
		Messages: []map[string]string{
			{"role": "user", "content": "ping"},
		},
	})
	if err == nil {
		t.Fatalf("expected request to fail")
	}

	msg := err.Error()
	if !strings.Contains(msg, "openai-compatible request failed") ||
		!strings.Contains(msg, "The model gpt-5.4-mini does not exist") {
		t.Fatalf("expected upstream error to be preserved, got %q", msg)
	}
}

func TestTransientUpstreamErrorIsReadable(t *testing.T) {
	attempts := 0
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			attempts++
			body := `{"error":{"message":"Service temporarily unavailable","type":"api_error"}}`
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
		Timeout: 2 * time.Second,
	}

	client := NewClientWithHTTPClient(ClientConfig{
		APIKey:   "key",
		Endpoint: "https://api.pixelqd.cn/v1/chat/completions",
		Model:    "gpt-5.4-mini",
		Timeout:  2 * time.Second,
	}, httpClient)

	_, err := client.RequestWithConfig(RequestConfig{
		Model: "gpt-5.4-mini",
		Messages: []map[string]string{
			{"role": "user", "content": "ping"},
		},
	})
	if err == nil {
		t.Fatalf("expected request to fail")
	}

	msg := err.Error()
	if !strings.Contains(msg, "upstream service temporarily unavailable (HTTP 503)") ||
		!strings.Contains(msg, "Service temporarily unavailable") {
		t.Fatalf("expected readable upstream 503 message, got %q", msg)
	}
	if attempts != maxTransientRetries+1 {
		t.Fatalf("attempts = %d, want %d", attempts, maxTransientRetries+1)
	}
}

func TestParseCustomHeadersSupportsMapAndPairArray(t *testing.T) {
	mapHeaders, err := parseCustomHeaders(`{"X-Test":"one"}`)
	if err != nil {
		t.Fatalf("parseCustomHeaders(map) returned error: %v", err)
	}
	if mapHeaders["X-Test"] != "one" {
		t.Fatalf("map header = %q, want one", mapHeaders["X-Test"])
	}

	arrayHeaders, err := parseCustomHeaders(`[{"key":"X-Test","value":"two"}]`)
	if err != nil {
		t.Fatalf("parseCustomHeaders(array) returned error: %v", err)
	}
	if arrayHeaders["X-Test"] != "two" {
		t.Fatalf("array header = %q, want two", arrayHeaders["X-Test"])
	}
}
