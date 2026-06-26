package compiler_test

import (
	"path/filepath"
	"strings"
	"testing"

	"kylix/pkg/compiler"
)

func validationHasCode(diags []compiler.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func writeValidationTemp(t *testing.T, content string) string {
	t.Helper()
	return writeTempKlx(t, t.TempDir(), "main.klx", content)
}

func TestValidationMinLenMissingArg(t *testing.T) {
	file := writeValidationTemp(t, `program Test;
type
  TUser = class
    [MinLen]
    Name: String;
  end;
begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !validationHasCode(r.Diagnostics, compiler.ErrInvalidValidation) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidValidation, r.Diagnostics)
	}
}

func TestValidationEmailRequiresString(t *testing.T) {
	file := writeValidationTemp(t, `program Test;
type
  TUser = class
    [Email]
    Age: Integer;
  end;
begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !validationHasCode(r.Diagnostics, compiler.ErrInvalidValidation) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidValidation, r.Diagnostics)
	}
}

func TestValidationMinRequiresInteger(t *testing.T) {
	file := writeValidationTemp(t, `program Test;
type
  TUser = class
    [Min(18)]
    Name: String;
  end;
begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !validationHasCode(r.Diagnostics, compiler.ErrInvalidValidation) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidValidation, r.Diagnostics)
	}
}

func TestValidationValidProgram(t *testing.T) {
	file := writeValidationTemp(t, `program Test;
type
  TUser = class
    [Required]
    [Email]
    Email: String;

    [MinLen(8)]
    Password: String;

    [Min(18)]
    Age: Integer;
  end;
begin
  WriteLn('OK');
end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Diagnostics {
		if strings.HasPrefix(d.Code, "KLX21") {
			t.Fatalf("unexpected diagnostic: %+v", r.Diagnostics)
		}
	}
	if !r.Success {
		t.Fatalf("expected success, got %+v", r.Diagnostics)
	}
}
