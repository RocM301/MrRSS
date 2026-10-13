package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Inactive    bool   `json:"inactive"`
}

// SearchFeeds searches the public Feedly directory without credentials.
func (s *Service) SearchFeeds(ctx context.Context, query string) ([]SearchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://cloud.feedly.com/v3/search/feeds?count=20&query="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search RSS directory: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS directory returned HTTP %d", res.StatusCode)
	}
	var payload struct {
		Results []struct {
			Title       string `json:"title"`
			FeedID      string `json:"feedId"`
			Description string `json:"description"`
			State       string `json:"state"`
		} `json:"results"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("read RSS directory: %w", err)
	}
	results := make([]SearchResult, 0)
	seen := make(map[string]bool)
	for _, item := range payload.Results {
		address := strings.TrimPrefix(item.FeedID, "feed/")
		u, err := url.Parse(address)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || seen[address] {
			continue
		}
		// Feedly's proxy endpoint is Cloudflare-blocked when MrRSS fetches the feed.
		if strings.EqualFold(u.Hostname(), "feedproxy.feedly.com") {
			continue
		}
		seen[address] = true
		results = append(results, SearchResult{item.Title, address, item.Description, item.State == "dormant" || item.State == "dead"})
	}
	sort.SliceStable(results, func(i, j int) bool { return !results[i].Inactive && results[j].Inactive })
	return results, nil
}
