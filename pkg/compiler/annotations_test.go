package compiler_test

import (
	"path/filepath"
	"strings"
	"testing"

	"kylix/pkg/compiler"
)

func annotationsHasCode(diags []compiler.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func writeAnnotationTemp(t *testing.T, content string) string {
	t.Helper()
	return writeTempKlx(t, t.TempDir(), "main.klx", content)
}

func TestBootAnnotationsDuplicateRoute(t *testing.T) {
	file := writeAnnotationTemp(t, `program Test;
uses boot;

[Controller('/api')]
type
  TOne = class
    [Get('/x')]
    function A(req: TRequest): TResponse;
    begin result := BootText(200, 'A'); end;

    [Get('/x')]
    function B(req: TRequest): TResponse;
    begin result := BootText(200, 'B'); end;
  end;

begin end.
`)
	result, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !annotationsHasCode(result.Diagnostics, compiler.ErrDuplicateRoute) {
		t.Fatalf("expected %s, got %+v", compiler.ErrDuplicateRoute, result.Diagnostics)
	}
}

func TestBootAnnotationsDuplicateRouteAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	unit := writeTempKlx(t, dir, "api_one.klx", `unit ApiOne;
uses boot;
interface
implementation

[Controller('/api')]
type
  TOne = class
    [Get('/x')]
    function A(req: TRequest): TResponse;
    begin result := BootText(200, 'A'); end;
  end;
end.
`)
	main := writeTempKlx(t, dir, "main.klx", `program Test;
uses boot, ApiOne;

[Controller('/api')]
type
  TTwo = class
    [Get('/x')]
    function B(req: TRequest): TResponse;
    begin result := BootText(200, 'B'); end;
  end;

begin end.
`)
	result, err := compiler.CompileProject([]string{unit, main}, compiler.Options{OutputFile: filepath.Join(dir, "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !annotationsHasCode(result.Diagnostics, compiler.ErrDuplicateRoute) {
		t.Fatalf("expected %s, got %+v", compiler.ErrDuplicateRoute, result.Diagnostics)
	}
}

func TestBootAnnotationsValidProcedureHandler(t *testing.T) {
	file := writeAnnotationTemp(t, `program Test;
uses boot;

[Controller('/api')]
type
  TOne = class
    [Get('/x')]
    procedure A(req: TRequest; res: TResponse);
    begin
      res.Send('A');
    end;
  end;

begin end.
`)
	result, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if annotationsHasCode(result.Diagnostics, compiler.ErrUnsupportedHandler) {
		t.Fatalf("did not expect %s, got %+v", compiler.ErrUnsupportedHandler, result.Diagnostics)
	}
}

func TestBootAnnotationsUnsupportedHandler(t *testing.T) {
	file := writeAnnotationTemp(t, `program Test;
uses boot;

[Controller('/api')]
type
  TOne = class
    [Get('/x')]
    procedure Bad(req: TRequest);
    begin WriteLn('bad'); end;
  end;

begin end.
`)
	result, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !annotationsHasCode(result.Diagnostics, compiler.ErrUnsupportedHandler) {
		t.Fatalf("expected %s, got %+v", compiler.ErrUnsupportedHandler, result.Diagnostics)
	}
}

func TestBootAnnotationsInvalidPathArg(t *testing.T) {
	file := writeAnnotationTemp(t, `program Test;
uses boot;

[Controller(123)]
type
  TOne = class
    [Get('/x')]
    function A(req: TRequest): TResponse;
    begin result := BootText(200, 'A'); end;
  end;

begin end.
`)
	result, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !annotationsHasCode(result.Diagnostics, compiler.ErrInvalidAnnotation) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidAnnotation, result.Diagnostics)
	}
}

func TestBootAnnotationsMissingInjectTarget(t *testing.T) {
	file := writeAnnotationTemp(t, `program Test;
uses boot;

[Controller('/api')]
type
  TOne = class
    [Inject]
    Missing: TMissingService;

    [Get('/x')]
    function A(req: TRequest): TResponse;
    begin result := BootText(200, 'A'); end;
  end;

begin end.
`)
	result, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !annotationsHasCode(result.Diagnostics, compiler.ErrMissingInjectTarget) {
		t.Fatalf("expected %s, got %+v", compiler.ErrMissingInjectTarget, result.Diagnostics)
	}
}

func TestBootAnnotationsValidProgram(t *testing.T) {
	file := writeAnnotationTemp(t, `program Test;
uses boot;

[Service]
type
  TUserService = class
    function Greeting(): String;
    begin result := 'hi'; end;
  end;

[Controller('/api')]
type
  TOne = class
    [Inject]
    UserService: TUserService;

    [Get('/x')]
    function A(req: TRequest): TResponse;
    begin result := BootText(200, self.UserService.Greeting()); end;
  end;

begin
  WriteLn('OK');
end.
`)
	result, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range result.Diagnostics {
		if strings.HasPrefix(d.Code, "KLX20") {
			t.Fatalf("unexpected semantic diagnostic: %+v", result.Diagnostics)
		}
	}
	if !result.Success {
		t.Fatalf("expected success, got %+v", result.Diagnostics)
	}
}

func TestBodyBindingMissingArg(t *testing.T) {
	file := writeAnnotationTemp(t, `program Test;
uses boot;
[Controller('/api')]
type
  TCtrl = class
    [Post('/users')]
    [Body]
    function Create(req: TRequest): TResponse;
    begin result := BootText(201, "ok"); end;
  end;
begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !annotationsHasCode(r.Diagnostics, compiler.ErrBodyBinding) {
		t.Fatalf("expected %s, got %+v", compiler.ErrBodyBinding, r.Diagnostics)
	}
}

func TestBodyBindingValid(t *testing.T) {
	file := writeAnnotationTemp(t, `program Test;
uses boot;
[Entity("users")]
type
  TUser = class
    [Required]
    Email: String;
  end;
[Controller("/api")]
type
  TCtrl = class
    [Post("/users")]
    [Body(TUser)]
    function Create(req: TRequest): TResponse;
    begin result := BootText(201, "ok"); end;
  end;
begin
  WriteLn("OK");
end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if annotationsHasCode(r.Diagnostics, compiler.ErrBodyBinding) {
		t.Fatalf("unexpected diagnostic: %+v", r.Diagnostics)
	}
	if !r.Success {
		t.Fatalf("expected success, got %+v", r.Diagnostics)
	}
}
