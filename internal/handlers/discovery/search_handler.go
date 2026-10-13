package discovery

import (
	"MrRSS/internal/handlers/core"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

func HandleSearchFeeds(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" || utf8.RuneCountInString(query) > 200 {
		http.Error(w, "Invalid search query", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	results, err := h.DiscoveryService.SearchFeeds(ctx, query)
	if err != nil {
		http.Error(w, "RSS directory is unavailable. Please retry.", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
