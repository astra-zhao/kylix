package compiler_test

import (
	"path/filepath"
	"testing"

	"kylix/pkg/compiler"
)

func securityHasCode(diags []compiler.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func writeSecurityTemp(t *testing.T, content string) string {
	t.Helper()
	return writeTempKlx(t, t.TempDir(), "main.klx", content)
}

func TestSecurityRoleMissingArg(t *testing.T) {
	file := writeSecurityTemp(t, `program Test;
uses boot;

[Controller('/admin')]
type
  TCtrl = class
    [Get('/users')]
    [Role]
    function Users(req: TRequest): TResponse;
    begin result := BootText(200, 'users'); end;
  end;

begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !securityHasCode(r.Diagnostics, compiler.ErrInvalidSecurity) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidSecurity, r.Diagnostics)
	}
}

func TestSecurityRoleInvalidArg(t *testing.T) {
	file := writeSecurityTemp(t, `program Test;
uses boot;

[Controller('/admin')]
type
  TCtrl = class
    [Get('/users')]
    [Role(123)]
    function Users(req: TRequest): TResponse;
    begin result := BootText(200, 'users'); end;
  end;

begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !securityHasCode(r.Diagnostics, compiler.ErrInvalidSecurity) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidSecurity, r.Diagnostics)
	}
}

func TestSecurityAnnotationOutsideController(t *testing.T) {
	file := writeSecurityTemp(t, `program Test;
uses boot;

type
  TPlain = class
    [Authenticated]
    function Foo(req: TRequest): TResponse;
    begin result := BootText(200, 'no'); end;
  end;

begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !securityHasCode(r.Diagnostics, compiler.ErrInvalidSecurity) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidSecurity, r.Diagnostics)
	}
}

func TestSecurityValidProgram(t *testing.T) {
	file := writeSecurityTemp(t, `program Test;
uses boot;

[Controller('/admin')]
type
  TCtrl = class
    [Get('/dashboard')]
    [Authenticated]
    function Dashboard(req: TRequest): TResponse;
    begin result := BootText(200, 'd'); end;

    [Get('/users')]
    [Role('admin')]
    function Users(req: TRequest): TResponse;
    begin result := BootText(200, 'u'); end;
  end;

begin
  WriteLn('OK');
end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if securityHasCode(r.Diagnostics, compiler.ErrInvalidSecurity) {
		t.Fatalf("unexpected diagnostic: %+v", r.Diagnostics)
	}
	if !r.Success {
		t.Fatalf("expected success, got %+v", r.Diagnostics)
	}
}
