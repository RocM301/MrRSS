package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// MTranTranslator implements translation via a self-hosted MTranServer instance
// (https://github.com/xxnuo/MTranServer).
//
// MTranServer uses language-pair NMT models, so each request must specify an
// explicit source language ("from"). MrRSS's built-in language detector is used
// to determine the source language of each text segment automatically.
//
// Native API:
//
//	POST {endpoint}/translate
//	  request : {"from":"en","to":"zh","text":"..."}
//	  response: {"result":"..."}
type MTranTranslator struct {
	Endpoint string // Base URL, e.g. http://192.168.5.88:8989
	Token    string // Optional API token (MT_API_TOKEN); empty if not set
	client   *http.Client
}

// NewMTranTranslator creates a new MTranServer translator.
func NewMTranTranslator(endpoint, token string) *MTranTranslator {
	return &MTranTranslator{
		Endpoint: strings.TrimSuffix(endpoint, "/"),
		Token:    token,
		// MTranServer is a local/LAN service and responds in tens of ms, so a
		// short timeout is plenty and keeps the UI snappy. Note: not routed
		// through the app's HTTP proxy on purpose — it's a local service.
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// mtranLangCode normalizes an ISO code to what MTranServer expects (base
// language, lowercase). e.g. "zh-tw" / "zh-Hant" -> "zh", "EN" -> "en".
func mtranLangCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return ""
	}
	// Chinese variants collapse to "zh"
	if strings.HasPrefix(code, "zh") {
		return "zh"
	}
	if len(code) > 2 {
		code = code[:2]
	}
	return code
}

// Translate translates text into targetLang using MTranServer.
func (t *MTranTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}
	if t.Endpoint == "" {
		return "", fmt.Errorf("MTranServer endpoint is not configured")
	}

	to := mtranLangCode(targetLang)
	if to == "" {
		to = "zh"
	}

	// This integration is scoped to English -> Chinese translation only, so the
	// source language is fixed to English.
	from := "en"

	// If the target is English there is nothing to do for an en->en pair.
	if from == to {
		return text, nil
	}

	// Only translate text that is purely English. If it contains any Chinese
	// (Han) characters it's treated as already-Chinese content (e.g. feeds like
	// 逛逛GitHub) and returned untouched — feeding such text to the en->zh model
	// would mangle it into garbage ("腾讯开源了…" -> "门 腾讯开源纬 特…").
	if containsChinese(text) {
		return text, nil
	}

	// MTranServer's NMT model often rewrites compact price-comparison snippets
	// like "$1,299, up from $1,099" into "1299美元，高于1099美元". Handle that
	// common RSS-list pattern deterministically so prices stay readable.
	if translated, ok := translateMTranPriceChange(text, to); ok {
		return polishMTranTranslation(text, translated, to), nil
	}

	// Protect brand names so the model can't translate them (e.g. "Apple" ->
	// "苹果"). Each brand is swapped for an opaque placeholder before
	// translation and restored afterwards.
	protectedText, restore := protectBrands(text)

	reqBody := map[string]string{
		"from": from,
		"to":   to,
		"text": protectedText,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal MTranServer request: %w", err)
	}

	req, err := http.NewRequest("POST", t.Endpoint+"/translate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create MTranServer request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.Token)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("MTranServer request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("MTranServer returned status %d: %s", resp.StatusCode, truncateBody(string(body), 200))
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode MTranServer response: %w", err)
	}

	// Restore the original brand names in place of the placeholders, then apply
	// a small news-title polish layer for the awkward phrases MTran commonly
	// emits around protected product names and tech terms.
	return polishMTranTranslation(text, restore(result.Result), to), nil
}

var priceChangeRegexp = regexp.MustCompile(`^\s*([^:\n]+?)\s*:\s*(\$[0-9]+(?:,[0-9]{3})*(?:\.[0-9]+)?)\s*,?\s*(?i:(up|down)\s+from)\s*(\$[0-9]+(?:,[0-9]{3})*(?:\.[0-9]+)?)(?:\s*\([+-]?\$[0-9]+(?:,[0-9]{3})*(?:\.[0-9]+)?\))?\s*$`)

func translateMTranPriceChange(text, targetLang string) (string, bool) {
	if targetLang != "zh" {
		return "", false
	}

	matches := priceChangeRegexp.FindStringSubmatch(text)
	if matches == nil {
		return "", false
	}

	product := strings.TrimSpace(matches[1])
	currentPrice := matches[2]
	direction := strings.ToLower(matches[3])
	previousPrice := matches[4]
	if product == "" || currentPrice == "" || previousPrice == "" {
		return "", false
	}

	verb := "上调至"
	if direction == "down" {
		verb = "下调至"
	}

	return fmt.Sprintf("%s：由 %s %s %s", product, previousPrice, verb, currentPrice), true
}

// protectedBrands lists English brand/product names that should always be kept
// as-is and never translated to Chinese. Matched case-insensitively on whole
// words. Order matters: longer, more specific names come before shorter ones
// (e.g. "Apple Watch" before "Apple") so the more specific match wins.
var protectedBrands = []string{
	// Multi-word product names MUST come before their shorter prefixes so the
	// whole name is protected as a single unit. Otherwise only the prefix is
	// protected and the trailing word gets translated on its own
	// ("MacBook Air" -> "MacBook 空气", "Mac Studio" -> "Mac 工作室").
	"Apple Vision Pro", "Apple Watch Ultra", "Apple Watch Series", "Apple Watch SE", "Apple Watch",
	"Apple TV", "Apple Music", "Apple Intelligence", "Apple Podcasts", "Apple News", "Apple Store",
	"Apple Pencil", "Apple Pay", "Apple Card", "Apple Arcade",
	"MacBook Air", "MacBook Pro", "MacBook Neo", "MacBook", "Mac Studio", "Mac Pro", "Mac mini", "Mac",
	"iPad Air", "iPad Pro", "iPad mini", "iPad", "iMac",
	"Vision Pro", "HomePod mini", "HomePod",
	"AirPods Max", "AirPods Pro", "AirPods", "AirTag", "Magic Keyboard", "Magic Mouse",
	"iPhone", "Apple",
	"Prime Day", "App Store", "Play Store", "Google TV", "Google Pixel", "Pixel", "Google",
	"Microsoft", "Amazon", "Meta", "Tesla",
	"OpenAI", "ChatGPT", "Claude", "Anthropic", "Gemini", "Nvidia", "AMD", "Intel",
	"iPadOS", "iOS", "macOS", "watchOS", "tvOS", "visionOS", "Siri",
	"Android", "Chrome", "ChromeOS", "Windows", "Surface", "Xbox", "Copilot",
	"Shortcuts", "Touch ID", "Face ID", "Final Cut Pro", "Logic Pro", "TestFlight",
	"Thunderbolt 5", "Thunderbolt", "USB-C", "Wi-Fi", "OLED", "SSD", "MSRP", "VPN", "GTA VI",
	"Macworld", "SanDisk", "LaCie", "Anker", "Belkin", "UGREEN", "Riot", "flowkey",
	"YouTube Shorts", "Shorts", "9to5Mac Overtime", "9to5Mac", "GitHub", "GitLab", "Slack", "Notion", "Figma", "Spotify", "Netflix", "YouTube", "TikTok",
	"DeepSeek", "Qwen", "Llama", "Mistral", "Nintendo", "Switch", "PlayStation", "Sony", "Samsung", "Galaxy", "Huawei", "Xiaomi",
}

var (
	mtranSpacesRegexp                 = regexp.MustCompile(`[ \t]{2,}`)
	mtranSpaceBeforeCJKPunctRegexp    = regexp.MustCompile(`[ \t]+([，。！？；：）】])`)
	mtranSpaceAfterOpeningPunctRegexp = regexp.MustCompile(`([（【])[ \t]+`)
	mtranStandaloneIARegexp           = regexp.MustCompile(`(^|[^A-Za-z0-9])IA([^A-Za-z0-9]|$)`)
	protectedVersionedTermRegexps     = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(?:iOS|iPadOS|macOS|watchOS|tvOS|visionOS)\s+[A-Za-z0-9.]+(?:['’]s)?\b`),
		regexp.MustCompile(`(?i)\b(?:Apple Watch Series|Apple Watch SE|Apple Watch Ultra|AirPods Max|MacBook Neo|HomePod mini)\s+[0-9]+(?:['’]s)?\b`),
	}
)

func polishMTranTranslation(source, translated, targetLang string) string {
	polished := strings.TrimSpace(translated)
	if targetLang != "zh" || polished == "" {
		return polished
	}

	polished = normalizeMTranTitlePunctuation(polished)
	polished = fixMTranCommonTerms(source, polished)
	polished = normalizeMTranTitlePunctuation(polished)
	return polished
}

func normalizeMTranTitlePunctuation(text string) string {
	text = strings.TrimSpace(text)
	if !strings.Contains(text, "://") {
		text = strings.NewReplacer(
			":", "：",
			"?", "？",
			"!", "！",
		).Replace(text)
	}
	text = mtranSpaceBeforeCJKPunctRegexp.ReplaceAllString(text, "$1")
	text = mtranSpaceAfterOpeningPunctRegexp.ReplaceAllString(text, "$1")
	text = mtranSpacesRegexp.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func fixMTranCommonTerms(source, translated string) string {
	out := mtranStandaloneIARegexp.ReplaceAllString(translated, "${1}AI${2}")

	if strings.Contains(source, "Prime Day") {
		out = strings.NewReplacer(
			"Prime会员日", "Prime Day",
			"Prime 会员日", "Prime Day",
			"黄金会员日", "Prime Day",
		).Replace(out)
	}

	if strings.Contains(source, "Shortcuts") {
		out = strings.ReplaceAll(out, "捷径", "Shortcuts")
	}

	if strings.Contains(source, "App Store") {
		out = strings.ReplaceAll(out, "应用商店", "App Store")
	}
	if strings.Contains(source, "Play Store") {
		out = strings.ReplaceAll(out, "Play 商店", "Play Store")
	}
	if strings.Contains(source, "Apple Watch Series") {
		out = regexp.MustCompile(`Apple Watch 第([0-9]+)季`).ReplaceAllString(out, "Apple Watch Series $1")
	}

	if strings.Contains(source, "Save") || strings.Contains(source, "save") {
		out = strings.ReplaceAll(out, "保存", "省下")
	}

	return out
}

// brandRegexpCache memoizes the compiled whole-word, case-insensitive regexp
// for each brand so repeated translations don't recompile them.
var (
	brandRegexpCache = map[string]*regexp.Regexp{}
	brandRegexpMu    sync.Mutex
)

// brandWordRegexp returns a cached case-insensitive, whole-word regexp for the
// given brand name.
func brandWordRegexp(brand string) *regexp.Regexp {
	brandRegexpMu.Lock()
	defer brandRegexpMu.Unlock()
	if re, ok := brandRegexpCache[brand]; ok {
		return re
	}
	// (?i) case-insensitive; \b word boundaries keep it whole-word so "Meta"
	// doesn't match "Metaphor". The optional (?:es|s)? suffix matches plural
	// product names so "Apple Watch" also catches "Apple Watches" and "iPad"
	// catches "iPads" — otherwise the plural's "...es"/"...s" tail gets
	// translated on its own ("Apple Watches" -> "Apple 观看"). The matched text
	// (including any plural or possessive suffix) is preserved verbatim on
	// restore.
	// Brand names are ASCII and only pure-English text reaches this point, so
	// ASCII word boundaries are sufficient.
	re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(brand) + `(?:es|s)?(?:['’]s)?\b`)
	brandRegexpCache[brand] = re
	return re
}

// brandPlaceholder builds the opaque token used to stand in for a brand during
// translation. Brace-wrapped placeholders are preserved more reliably by
// MTranServer than punctuation-heavy tokens such as "@@BRND0@@".
func brandPlaceholder(i int) string {
	return fmt.Sprintf("{BRAND%d}", i)
}

// restoreBrandsRegexp matches a (possibly model-mangled) brand placeholder and
// captures its index. MTranServer may alter wrappers ("{BRAND0}" -> "#BRAND0}",
// "[BRAND0]", "(BRAND0}", or legacy "@@BRND0@@"), so the wrapper is treated as
// disposable while the BRAND/BRND core and numeric index are restored.
var restoreBrandsRegexp = regexp.MustCompile(`(?i)[#@\{\[\(（"“'‘]*[ ]?BR(?:AND|ND)[ ]*(\d+)[ ]*[#@\}\]\)）"”'’]*`)

// protectBrands replaces known brand names in text with placeholders and
// returns the protected text plus a restore function that swaps the original
// brand names back in. Matching is case-insensitive and whole-word.
func protectBrands(text string) (string, func(string) string) {
	originalByIdx := map[int]string{}

	protected := text
	idx := 0

	for _, re := range protectedVersionedTermRegexps {
		protected = protectRegexpMatches(protected, re, originalByIdx, &idx)
	}

	for _, brand := range protectedBrands {
		re := brandWordRegexp(brand)
		loc := re.FindStringIndex(protected)
		if loc == nil {
			continue
		}
		// Preserve the original (first-matched) spelling, then replace every
		// occurrence of this brand with the same placeholder.
		original := protected[loc[0]:loc[1]]
		protected = re.ReplaceAllString(protected, brandPlaceholder(idx))
		originalByIdx[idx] = original
		idx++
	}

	restore := func(s string) string {
		s = restoreBrandsRegexp.ReplaceAllStringFunc(s, func(m string) string {
			sub := restoreBrandsRegexp.FindStringSubmatch(m)
			if sub == nil {
				return m
			}
			i, err := strconv.Atoi(sub[1])
			if err != nil {
				return m
			}
			if orig, ok := originalByIdx[i]; ok {
				return orig
			}
			// Unknown index (shouldn't happen): drop the stray placeholder
			// rather than leaking it into the output.
			return ""
		})

		// The NMT model occasionally emits a placeholder twice in a row, which
		// restores to a duplicated brand ("SamsungSamsung"). Collapse any
		// brand immediately repeated (with or without a space) back to one.
		for _, orig := range originalByIdx {
			for strings.Contains(s, orig+orig) {
				s = strings.ReplaceAll(s, orig+orig, orig)
			}
			s = strings.ReplaceAll(s, orig+" "+orig, orig)
		}
		return s
	}
	return protected, restore
}

func protectRegexpMatches(text string, re *regexp.Regexp, originalByIdx map[int]string, idx *int) string {
	return re.ReplaceAllStringFunc(text, func(match string) string {
		i := *idx
		*idx = i + 1
		originalByIdx[i] = match
		return brandPlaceholder(i)
	})
}

// containsChinese reports whether text contains any Chinese (Han) character.
// The MTranServer integration only translates purely-English text, so any Han
// character means the content is already Chinese and must be left untouched.
func containsChinese(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// truncateBody shortens a response body for safe error messages.
func truncateBody(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
