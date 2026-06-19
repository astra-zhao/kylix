package i18n_test

import (
	"strings"
	"testing"

	"kylix/pkg/i18n"
)

func TestT_English(t *testing.T) {
	i18n.SetLang(i18n.LangEn)
	got := i18n.T("KLX201", "userName")
	if !strings.Contains(got, "undeclared") || !strings.Contains(got, "userName") {
		t.Errorf("unexpected English translation: %s", got)
	}
}

func TestT_Chinese(t *testing.T) {
	i18n.SetLang(i18n.LangZh)
	got := i18n.T("KLX201", "用户名")
	if !strings.Contains(got, "未声明") || !strings.Contains(got, "用户名") {
		t.Errorf("unexpected Chinese translation: %s", got)
	}
}

func TestT_FallbackToEnglish(t *testing.T) {
	// Set to Chinese, but use a code that has no Chinese translation
	// (in our table all codes have both — so simulate with unknown code).
	i18n.SetLang(i18n.LangZh)
	got := i18n.T("UNKNOWN_CODE", "fallback")
	// Unknown code formats args directly
	if got != "fallback" {
		t.Errorf("expected raw fallback for unknown code, got %q", got)
	}
}

func TestT_TypeMismatch_English(t *testing.T) {
	i18n.SetLang(i18n.LangEn)
	got := i18n.T("KLX101", "String", "Integer")
	if !strings.Contains(got, "String") || !strings.Contains(got, "Integer") {
		t.Errorf("expected both types in message, got %s", got)
	}
}

func TestT_TypeMismatch_Chinese(t *testing.T) {
	i18n.SetLang(i18n.LangZh)
	got := i18n.T("KLX101", "String", "Integer")
	if !strings.Contains(got, "无法将") {
		t.Errorf("expected '无法将' prefix, got %s", got)
	}
}

func TestHint_English(t *testing.T) {
	i18n.SetLang(i18n.LangEn)
	got := i18n.Hint("KLX101_str_to_int")
	if !strings.Contains(got, "StrToInt") {
		t.Errorf("expected StrToInt suggestion, got: %s", got)
	}
}

func TestHint_Chinese(t *testing.T) {
	i18n.SetLang(i18n.LangZh)
	got := i18n.Hint("KLX101_str_to_int")
	if !strings.Contains(got, "StrToInt") {
		t.Errorf("expected StrToInt name preserved in Chinese, got: %s", got)
	}
	if !strings.Contains(got, "转为") {
		t.Errorf("expected Chinese verb '转为', got: %s", got)
	}
}

func TestHint_DidYouMean_Chinese(t *testing.T) {
	i18n.SetLang(i18n.LangZh)
	got := i18n.Hint("KLX201_did_you_mean", "userName")
	if !strings.Contains(got, "userName") {
		t.Errorf("expected name preserved, got: %s", got)
	}
	if !strings.Contains(got, "你是否") {
		t.Errorf("expected Chinese phrase, got: %s", got)
	}
}

func TestHint_UnknownKey(t *testing.T) {
	i18n.SetLang(i18n.LangEn)
	if got := i18n.Hint("NO_SUCH_KEY"); got != "" {
		t.Errorf("expected empty for unknown hint, got: %s", got)
	}
}
