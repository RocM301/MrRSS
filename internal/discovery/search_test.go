package discovery

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type searchTransport func(*http.Request) (*http.Response, error)

func (f searchTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSearchFeeds(t *testing.T) {
	s := NewService()
	s.client.Transport = searchTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("query") != "网易 & news" {
			t.Fatalf("query not encoded correctly: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"results":[{"title":"Proxy","feedId":"feed/https://feedproxy.feedly.com/protected"},{"title":"Old","feedId":"feed/https://example.com/old","state":"dormant"},{"title":"News","feedId":"feed/https://example.com/rss"},{"feedId":"feed/javascript:alert(1)"},{"feedId":"feed/https://example.com/rss"}]}`))}, nil
	})
	results, err := s.SearchFeeds(context.Background(), "网易 & news")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || results[0].Title != "News" || !results[1].Inactive {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestSearchFeedsUpstreamFailure(t *testing.T) {
	s := NewService()
	s.client.Transport = searchTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
	})
	if _, err := s.SearchFeeds(context.Background(), "news"); err == nil {
		t.Fatal("expected upstream error")
	}
}
