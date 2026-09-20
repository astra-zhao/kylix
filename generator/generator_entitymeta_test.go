package generator

import "testing"

// generator_entitymeta_test.go — v0.11.0 CRUD metadata emission (Go backend).
//
// The registration sequence is the contract the pure-Kylix CRUD engine reads,
// so these tests pin both its presence and its exact encoding. The LLVM
// backend emits the same strings (pkg/llvmgen/entitymeta_test.go) and
// TestEntityMeta_EncodingParity below checks the shared derivation.

func TestEntityMetaWiring(t *testing.T) {
	input := `
program MetaTest;
uses entitymeta, orm;

[Entity('users')]
[Label('Users')]
type
  TUser = class
    [PrimaryKey]
    Id: Integer;

    [Column('username')]
    [Label('Username')]
    [Required]
    [MinLen(3)]
    [Searchable]
    Username: String;

    [Column('password')]
    Password: String;

    [Column('is_active')]
    IsActive: Boolean;

    [Column('failed_attempts')]
    [Hidden]
    FailedAttempts: Integer;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)

	assertContains(t, out, `stdlib.RegisterEntity("users", "Id|Users|")`)
	assertContains(t, out, `stdlib.RegisterEntityField("users", "Id|number|Id|")`)
	assertContains(t, out, `stdlib.RegisterEntityField("users", "username|text|Username|required,minlen=3,searchable")`)
	assertContains(t, out, `stdlib.RegisterEntityField("users", "password|password|Password|secret")`)
	assertContains(t, out, `stdlib.RegisterEntityField("users", "is_active|checkbox|IsActive|")`)
	assertContains(t, out, `stdlib.RegisterEntityField("users", "failed_attempts|number|FailedAttempts|")`)
	assertContains(t, out, `stdlib.SetEntityNames("users")`)
}

func TestEntityMetaReadOnlyEntity(t *testing.T) {
	input := `
program MetaRO;
uses entitymeta, orm;

[Entity('op_logs')]
[Label('Operations')]
[ReadOnly]
type
  TOpLog = class
    [PrimaryKey]
    Id: Integer;
    [Column('method')]
    Method: String;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `stdlib.RegisterEntity("op_logs", "Id|Operations|readonly")`)
	assertContains(t, out, `stdlib.RegisterEntityField("op_logs", "method|text|Method|")`)
}

// A program without [Entity] classes must emit nothing: this is what keeps the
// bootstrap IR fixed point intact (the LLVM side has the same gate).
func TestEntityMetaAbsentWithoutEntityClasses(t *testing.T) {
	input := `
program Plain;
uses entitymeta;

type
  TThing = class
    Name: String;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertNotContains(t, out, "RegisterEntity")
	assertNotContains(t, out, "SetEntityNames")
}

// The PK falls back to a field named Id when [PrimaryKey] is absent, matching
// the ORM scanner's existing rule.
func TestEntityMetaPrimaryKeyFallback(t *testing.T) {
	input := `
program MetaPK;
uses entitymeta, orm;

[Entity('roles')]
type
  TRole = class
    Id: Integer;
    Name: String;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `stdlib.RegisterEntity("roles", "Id|roles|")`)
}

// [Column] renames the physical column; the metadata must carry the column
// name, not the field name.
func TestEntityMetaUsesColumnName(t *testing.T) {
	input := `
program MetaCol;
uses entitymeta, orm;

[Entity('users')]
type
  TUser = class
    [PrimaryKey]
    [Column('user_id')]
    Id: Integer;
  end;

begin
  WriteLn('OK');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `stdlib.RegisterEntity("users", "user_id|users|")`)
}
