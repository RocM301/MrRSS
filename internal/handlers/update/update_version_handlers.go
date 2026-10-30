package update

import (
	"net/http"
	"strconv"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/version"
)

// HandleVersion returns the current application version.
// @Summary      Get application version
// @Description  Get the current application version string
// @Tags         update
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Application version (version)"
// @Router       /version [get]
func HandleVersion(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	response.JSON(w, map[string]string{
		"version": version.Version,
	})
}

// leadingInt extracts the leading integer from a version component.
// This makes the comparison tolerant of suffixes glued onto a number, such as
// the fork's "-mod.N" scheme: "1.3.23-mod.2" splits on "." into
// ["1", "3", "23-mod", "2"], and leadingInt("23-mod") == 23. The trailing
// ".2" is then compared as its own numeric component, so mod.2 < mod.3, while
// a genuine upstream bump (1.3.24-...) still compares greater.
func leadingInt(s string) int {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0
	}
	n, _ := strconv.Atoi(s[:end])
	return n
}

// compareVersions compares two dotted versions (e.g., "1.1.0" vs "1.0.0", or
// "1.3.23-mod.2" vs "1.3.23-mod.3").
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1 = leadingInt(parts1[i])
		}
		if i < len(parts2) {
			p2 = leadingInt(parts2[i])
		}

		if p1 > p2 {
			return 1
		} else if p1 < p2 {
			return -1
		}
	}

	return 0
}
