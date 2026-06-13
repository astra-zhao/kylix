package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, src string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.klx")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(src); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

// ── interface validation ──────────────────────────────────────────────────────

func TestInterfaceFullyImplemented(t *testing.T) {
	src := `
program Test;
type
  IAnimal = interface
    procedure Speak();
    function Name(): String;
  end;

  TDog = class implements IAnimal
    procedure Speak();
    begin end;
    function Name(): String;
    begin result := 'Dog'; end;
  end;
begin end.
`
	f := writeTemp(t, src)
	result, err := CompileFile(f, Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range result.Diagnostics {
		if d.Level == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

func TestInterfaceMissingMethod(t *testing.T) {
	src := `
program Test;
type
  IAnimal = interface
    procedure Speak();
    function Name(): String;
  end;

  TDog = class implements IAnimal
    procedure Speak();
    begin end;
  end;
begin end.
`
	f := writeTemp(t, src)
	result, err := CompileFile(f, Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("expected compilation to fail due to missing method")
	}
	found := false
	for _, d := range result.Diagnostics {
		if strings.Contains(d.Message, "Name") && strings.Contains(d.Message, "IAnimal") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected diagnostic mentioning missing method 'Name' on 'IAnimal', got: %+v", result.Diagnostics)
	}
}

func TestInterfaceMultipleMissingMethods(t *testing.T) {
	src := `
program Test;
type
  IShape = interface
    function Area(): Integer;
    function Perimeter(): Integer;
    procedure Draw();
  end;

  TCircle = class implements IShape
  end;
begin end.
`
	f := writeTemp(t, src)
	result, err := CompileFile(f, Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("expected failure for class missing all interface methods")
	}
	if len(result.Diagnostics) < 3 {
		t.Errorf("expected 3 missing-method diagnostics, got %d: %+v", len(result.Diagnostics), result.Diagnostics)
	}
}

func TestInterfaceUnknownInterfaceSkipped(t *testing.T) {
	// IExternal is defined in another unit — should not produce an error.
	src := `
unit myunit;
type
  TFoo = class implements IExternal
    procedure DoSomething();
    begin end;
  end;
`
	f := writeTemp(t, src)
	result, err := CompileFile(f, Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range result.Diagnostics {
		if d.Level == "error" && strings.Contains(d.Message, "IExternal") {
			t.Errorf("should not report error for cross-unit interface: %s", d.Message)
		}
	}
}

func TestNoImplementsNoError(t *testing.T) {
	src := `
program Test;
type
  TPlain = class
    procedure Hello();
    begin end;
  end;
begin end.
`
	f := writeTemp(t, src)
	result, err := CompileFile(f, Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range result.Diagnostics {
		if d.Level == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}
