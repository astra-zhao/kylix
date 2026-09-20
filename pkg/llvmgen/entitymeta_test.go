package llvmgen_test

import (
	"strings"
	"testing"
)

// entitymeta tests — v0.11.0 [Entity] metadata on the LLVM backend.
//
// The registration calls carry the same strings the Go backend emits
// (generator/generator_entitymeta_test.go), because the pure-Kylix CRUD engine
// renders pages from them on both backends.

const entityMetaProgram = `program p;
uses boot, entitymeta, orm;

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
  end;

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

[Controller('')]
type
  TC = class
    [Get('/meta')]
    function Show(req: TRequest): TResponse;
    begin
      result := BootText(200, EntityNames() + EntityMetaOf('users') + EntityFieldAt('users', 0) + IntToStr(EntityFieldCount('users')));
    end;
  end;

begin
  BootRun(8080);
end.`

func TestEntityMeta_RegistersAtMainTop(t *testing.T) {
	ir := generateIR(t, entityMetaProgram)
	// Globals + registration defines.
	assertIRContains(t, ir, "@__kylix_entity_tables = global [64 x { ptr, ptr }] zeroinitializer")
	assertIRContains(t, ir, "@__kylix_entity_ntables = global i64 0")
	assertIRContains(t, ir, "@__kylix_entity_fields = global [512 x { ptr, ptr }] zeroinitializer")
	assertIRContains(t, ir, "define void @__kylix_entitymeta_RegisterEntity(ptr %table, ptr %meta)")
	assertIRContains(t, ir, "define void @__kylix_entitymeta_RegisterEntityField(ptr %table, ptr %spec)")
	assertIRContains(t, ir, "define void @__kylix_entitymeta_SetEntityNames(ptr %csv)")
	// The wiring calls live in main.
	assertIRContains(t, ir, "call void @__kylix_entitymeta_RegisterEntity(ptr")
	assertIRContains(t, ir, "call void @__kylix_entitymeta_SetEntityNames(ptr")
}

// The metadata strings must be byte-identical to the Go backend's — the CRUD
// engine renders from them.
func TestEntityMeta_MetadataStringsMatchGoEncoding(t *testing.T) {
	ir := generateIR(t, entityMetaProgram)
	for _, want := range []string{
		"Id|Users|",
		"Id|number|Id|",
		"username|text|Username|required,minlen=3,searchable",
		"password|password|Password|secret",
		"Id|Operations|readonly",
		"method|text|Method|",
		"users,op_logs",
	} {
		if !strings.Contains(ir, want) {
			t.Errorf("IR is missing metadata string %q", want)
		}
	}
}

func TestEntityMeta_AccessorsLower(t *testing.T) {
	ir := generateIR(t, entityMetaProgram)
	assertIRContains(t, ir, "call ptr @__kylix_entitymeta_EntityNames()")
	assertIRContains(t, ir, "define ptr @__kylix_entitymeta_EntityMetaOf(ptr %table)")
	assertIRContains(t, ir, "define ptr @__kylix_entitymeta_EntityFieldAt(ptr %table, i64 %index)")
	assertIRContains(t, ir, "define i64 @__kylix_entitymeta_EntityFieldCount(ptr %table)")
}

// A program without [Entity] classes must not gain the globals or defines —
// this is the gate that keeps the bootstrap IR fixed point intact.
func TestEntityMeta_AbsentWithoutEntityClasses(t *testing.T) {
	ir := generateIR(t, `program p;
uses boot;
begin
  WriteLn('plain');
end.`)
	for _, unwanted := range []string{
		"@__kylix_entity_tables",
		"@__kylix_entity_fields",
		"__kylix_entitymeta_",
	} {
		if strings.Contains(ir, unwanted) {
			t.Errorf("IR of an entity-free program contains %q", unwanted)
		}
	}
}
