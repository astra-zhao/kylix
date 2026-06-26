package compiler_test

import (
	"path/filepath"
	"testing"

	"kylix/pkg/compiler"
)

func ormHasCode(diags []compiler.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func writeORMTemp(t *testing.T, content string) string {
	t.Helper()
	return writeTempKlx(t, t.TempDir(), "main.klx", content)
}

func TestORMEntityRequiresStringArg(t *testing.T) {
	file := writeORMTemp(t, `program Test;
[Entity(123)]
type
  TUser = class
    Id: Integer;
  end;
begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !ormHasCode(r.Diagnostics, compiler.ErrInvalidORM) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidORM, r.Diagnostics)
	}
}

func TestORMRepositoryUnknownEntity(t *testing.T) {
	file := writeORMTemp(t, `program Test;
[Repository(TMissing)]
type
  TRepo = class
  end;
begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !ormHasCode(r.Diagnostics, compiler.ErrInvalidORM) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidORM, r.Diagnostics)
	}
}

func TestORMQueryOutsideRepository(t *testing.T) {
	file := writeORMTemp(t, `program Test;
[Entity('users')]
type
  TUser = class
    Id: Integer;
  end;

type
  TStray = class
    [Query('SELECT * FROM users')]
    function Foo(): TUser;
  end;
begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !ormHasCode(r.Diagnostics, compiler.ErrInvalidORM) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidORM, r.Diagnostics)
	}
}

func TestORMQueryInvalidReturnType(t *testing.T) {
	file := writeORMTemp(t, `program Test;
[Entity('users')]
type
  TUser = class
    Id: Integer;
  end;

[Repository(TUser)]
type
  TRepo = class
    [Query('SELECT 1')]
    function Bad(): String;
  end;
begin end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if !ormHasCode(r.Diagnostics, compiler.ErrInvalidORM) {
		t.Fatalf("expected %s, got %+v", compiler.ErrInvalidORM, r.Diagnostics)
	}
}

func TestORMValidProgram(t *testing.T) {
	file := writeORMTemp(t, `program Test;
uses orm;

[Entity('users')]
type
  TUser = class
    [PrimaryKey]
    Id: Integer;
    [Column('email')]
    Email: String;
  end;

[Repository(TUser)]
type
  TUserRepository = class
    [Query('SELECT * FROM users WHERE email = ?')]
    function ByEmail(email: String): TUser;
  end;

begin
  WriteLn('OK');
end.
`)
	r, err := compiler.CompileFile(file, compiler.Options{OutputFile: filepath.Join(t.TempDir(), "out.go")})
	if err != nil {
		t.Fatal(err)
	}
	if ormHasCode(r.Diagnostics, compiler.ErrInvalidORM) {
		t.Fatalf("unexpected diagnostic: %+v", r.Diagnostics)
	}
	if !r.Success {
		t.Fatalf("expected success, got %+v", r.Diagnostics)
	}
}
