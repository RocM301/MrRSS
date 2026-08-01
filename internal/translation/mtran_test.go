package translation

import (
	"strings"
	"testing"
)

func TestTranslateMTranPriceChange(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "Apple product with comma prices",
			text: "MacBook Air: $1,299, up from $1,099",
			want: "MacBook Air：由 $1,099 上调至 $1,299",
		},
		{
			name: "Product with parenthetical model",
			text: "Mac Studio (M4 Max): $2,499, up from $1,999",
			want: "Mac Studio (M4 Max)：由 $1,999 上调至 $2,499",
		},
		{
			name: "Product with no comma before up from",
			text: "Vision Pro: $3,699 up from $3,499",
			want: "Vision Pro：由 $3,499 上调至 $3,699",
		},
		{
			name: "Two digit prices",
			text: "HomePod mini: $129, up from $99",
			want: "HomePod mini：由 $99 上调至 $129",
		},
		{
			name: "Price increase with delta suffix",
			text: "Apple TV: $199, up from $129 (+$70)",
			want: "Apple TV：由 $129 上调至 $199",
		},
		{
			name: "Large delta suffix",
			text: "Mac Studio (M3 Ultra): $5,299, up from $3,999 (+$1,300)",
			want: "Mac Studio (M3 Ultra)：由 $3,999 上调至 $5,299",
		},
		{
			name: "Down from uses stable decrease wording",
			text: "Example Widget: $79, down from $99",
			want: "Example Widget：由 $99 下调至 $79",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := translateMTranPriceChange(tt.text, "zh")
			if !ok {
				t.Fatalf("expected price-change text to be handled")
			}
			if got != tt.want {
				t.Fatalf("translateMTranPriceChange() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTranslateMTranPriceChangeIgnoresOtherText(t *testing.T) {
	tests := []string{
		"MacBook Air starts at $1,299.",
		"MacBook Air: available today",
		"MacBook Air: $1,299",
	}

	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			if got, ok := translateMTranPriceChange(text, "zh"); ok {
				t.Fatalf("expected text to pass through MTranServer, got %q", got)
			}
		})
	}
}

func TestTranslateMTranPriceChangeOnlyForChineseTarget(t *testing.T) {
	if got, ok := translateMTranPriceChange("MacBook Air: $1,299, up from $1,099", "ja"); ok {
		t.Fatalf("expected non-Chinese target to pass through MTranServer, got %q", got)
	}
}

func TestMTranTranslateShortCircuitsPriceChangeBeforeRequest(t *testing.T) {
	translator := NewMTranTranslator("http://127.0.0.1:1", "")

	got, err := translator.Translate("MacBook Air: $1,299, up from $1,099", "zh")
	if err != nil {
		t.Fatalf("Translate() returned unexpected error: %v", err)
	}

	want := "MacBook Air：由 $1,099 上调至 $1,299"
	if got != want {
		t.Fatalf("Translate() = %q, want %q", got, want)
	}
}

func TestProtectBrandsKeepsMultiWordProductNames(t *testing.T) {
	text := "MacBook Air and Mac Studio are different from Apple Watches."
	protected, restore := protectBrands(text)

	if protected == text {
		t.Fatalf("expected product names to be protected")
	}

	got := restore(protected)
	if got != text {
		t.Fatalf("restore(protectBrands()) = %q, want %q", got, text)
	}
}

func TestProtectBrandsKeepsYouTubeShortsTerm(t *testing.T) {
	text := "YouTube updates Shorts to make it even more like TikTok"
	protected, restore := protectBrands(text)

	if protected == text {
		t.Fatalf("expected YouTube, Shorts, and TikTok to be protected")
	}

	translatedByMTran := strings.Replace(protected, "updates", "更新", 1)
	translatedByMTran = strings.Replace(translatedByMTran, "to make it even more like", "使其更像", 1)
	got := restore(translatedByMTran)
	if strings.Contains(got, "短裤") {
		t.Fatalf("expected Shorts to stay protected, got %q", got)
	}
	if !strings.Contains(got, "Shorts") {
		t.Fatalf("expected restored text to contain Shorts, got %q", got)
	}
}

func TestProtectBrandsKeeps9to5MacOvertimeAndSiri(t *testing.T) {
	text := "9to5Mac Overtime 070: The Siri Redemption Tour"
	protected, restore := protectBrands(text)

	if protected == text {
		t.Fatalf("expected 9to5Mac Overtime and Siri to be protected")
	}

	translatedByMTran := strings.Replace(protected, "The", "", 1)
	translatedByMTran = strings.Replace(translatedByMTran, "Redemption Tour", "救赎之旅", 1)
	got := restore(translatedByMTran)
	if strings.Contains(got, "BRAND") || strings.Contains(got, "BRND") {
		t.Fatalf("expected no leaked placeholders, got %q", got)
	}
	if !strings.Contains(got, "9to5Mac Overtime") || !strings.Contains(got, "Siri") {
		t.Fatalf("expected protected terms to be restored, got %q", got)
	}
}

func TestProtectBrandsKeepsAppleNewsDealTerms(t *testing.T) {
	text := "Best MacBook Prime Day Deals 2026: Save $$$ and avoid Apple’s price hikes"
	protected, restore := protectBrands(text)

	if protected == text {
		t.Fatalf("expected MacBook, Prime Day, and Apple to be protected")
	}

	got := restore(protected)
	if got != text {
		t.Fatalf("restore(protectBrands()) = %q, want %q", got, text)
	}
}

func TestProtectBrandsKeepsAdditionalProductTerms(t *testing.T) {
	text := "Apple Watch Series 12, MacBook Neo, AirPods Max, SanDisk SSD, and Shortcuts"
	protected, restore := protectBrands(text)

	if protected == text {
		t.Fatalf("expected product terms to be protected")
	}

	got := restore(protected)
	if got != text {
		t.Fatalf("restore(protectBrands()) = %q, want %q", got, text)
	}
}

func TestProtectBrandsKeepsVersionedProductTerms(t *testing.T) {
	text := "iOS 27’s Shortcuts is AI at its best, and Apple Watch Series 12 is coming"
	protected, restore := protectBrands(text)

	if protected == text {
		t.Fatalf("expected versioned product terms to be protected")
	}

	got := restore(protected)
	if got != text {
		t.Fatalf("restore(protectBrands()) = %q, want %q", got, text)
	}
}

func TestPolishMTranTranslationNormalizesTitlePunctuationAndAI(t *testing.T) {
	got := polishMTranTranslation(
		"Macworld Podcast: Our experience with the Siri AI beta",
		"Macworld 播客:我们使用 IA 测试版的经验",
		"zh",
	)
	want := "Macworld 播客：我们使用 AI 测试版的经验"
	if got != want {
		t.Fatalf("polishMTranTranslation() = %q, want %q", got, want)
	}
}

func TestPolishMTranTranslationFixesCommonNewsTerms(t *testing.T) {
	got := polishMTranTranslation(
		"Best Prime Day iPad deals 2026: Grab these discounts while you still can",
		"2026年最佳Prime会员日优惠:趁您仍可享受时享受这些优惠",
		"zh",
	)
	want := "2026年最佳Prime Day优惠：趁您仍可享受时享受这些优惠"
	if got != want {
		t.Fatalf("polishMTranTranslation() = %q, want %q", got, want)
	}
}

func TestPolishMTranTranslationOnlyForChineseTarget(t *testing.T) {
	text := "Example: still ascii?"
	if got := polishMTranTranslation(text, text, "en"); got != text {
		t.Fatalf("expected non-Chinese target to skip polish, got %q", got)
	}
}

func TestRestoreBrandsHandlesMangledPlaceholders(t *testing.T) {
	_, restore := protectBrands("Apple Siri Amazon")

	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "hash and closing brace",
			text: "PSA:#BRAND0} 的价格上涨尚未生效",
			want: "PSA:Apple 的价格上涨尚未生效",
		},
		{
			name: "square bracket",
			text: "这次[BRAND2]真的好吗?",
			want: "这次Siri真的好吗?",
		},
		{
			name: "curly quote and brace",
			text: "以下是“BRAND0}对其翻新店铺的描述",
			want: "以下是Apple对其翻新店铺的描述",
		},
		{
			name: "paren and brace",
			text: "很难知道何时(BRAND1})会涨价",
			want: "很难知道何时Amazon会涨价",
		},
		{
			name: "legacy BRND wrapper",
			text: "@@BRND2@@ 救赎之旅",
			want: "Siri 救赎之旅",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := restore(tt.text)
			if got != tt.want {
				t.Fatalf("restore() = %q, want %q", got, tt.want)
			}
		})
	}
}
