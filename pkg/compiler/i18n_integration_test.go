package compiler_test

import (
	"os"
	"strings"
	"testing"

	"kylix/pkg/compiler"
	"kylix/pkg/i18n"
)

// i18n integration tests (v2.4.0 task 1): verify error messages are localized
// when KYLIX_LANG / i18n.SetLang is set.

func i18nCompile(t *testing.T, src string) *compiler.Result {
	t.Helper()
	f := t.TempDir() + "/test.klx"
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestI18n_TypeMismatch_English(t *testing.T) {
	i18n.SetLang(i18n.LangEn)
	r := i18nCompile(t, `program Test;
var x: Integer;
begin
  x := 'hello';
end.`)
	if r.Success {
		t.Fatal("expected failure")
	}
	d := r.Diagnostics[0]
	if !strings.Contains(d.Message, "cannot assign") {
		t.Errorf("expected English message, got: %s", d.Message)
	}
	if !strings.Contains(d.Hint, "StrToInt") {
		t.Errorf("expected English hint, got: %s", d.Hint)
	}
}

func TestI18n_TypeMismatch_Chinese(t *testing.T) {
	i18n.SetLang(i18n.LangZh)
	defer i18n.SetLang(i18n.LangEn) // restore for other tests
	r := i18nCompile(t, `program Test;
var x: Integer;
begin
  x := 'hello';
end.`)
	if r.Success {
		t.Fatal("expected failure")
	}
	d := r.Diagnostics[0]
	if !strings.Contains(d.Message, "无法将") {
		t.Errorf("expected Chinese message, got: %s", d.Message)
	}
	if !strings.Contains(d.Hint, "转为") {
		t.Errorf("expected Chinese hint, got: %s", d.Hint)
	}
	// Code is language-independent
	if d.Code != compiler.ErrTypeMismatch {
		t.Errorf("code should stay KLX101 regardless of language, got %s", d.Code)
	}
}

func TestI18n_Undeclared_Chinese(t *testing.T) {
	i18n.SetLang(i18n.LangZh)
	defer i18n.SetLang(i18n.LangEn)
	r := i18nCompile(t, `program Test;
begin
  x := 42;
end.`)
	if r.Success {
		t.Fatal("expected failure")
	}
	found := false
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrUndeclared && strings.Contains(d.Message, "未声明") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Chinese undeclared message, got: %v", r.Diagnostics)
	}
}

func TestI18n_CodePreservedAcrossLanguages(t *testing.T) {
	src := `program Test;
var x: Integer;
begin
  x := 'bad';
end.`

	i18n.SetLang(i18n.LangEn)
	en := i18nCompile(t, src)

	i18n.SetLang(i18n.LangZh)
	zh := i18nCompile(t, src)
	i18n.SetLang(i18n.LangEn)

	// Same error code, different message language.
	if en.Diagnostics[0].Code != zh.Diagnostics[0].Code {
		t.Errorf("error codes should match across languages: %s vs %s",
			en.Diagnostics[0].Code, zh.Diagnostics[0].Code)
	}
	if en.Diagnostics[0].Message == zh.Diagnostics[0].Message {
		t.Error("messages should differ between English and Chinese")
	}
}
