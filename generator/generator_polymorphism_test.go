package generator

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// v5.2.0: polymorphic base classes → Go interfaces (opt-in via is/as)
// ---------------------------------------------------------------------------

// TestPolymorphism_BaseClassBecomesInterface: a program that uses `is`/`as`
// must emit inheritance-participating base classes as empty Go interfaces so
// heterogeneous collections and type assertions compile.
func TestPolymorphism_BaseClassBecomesInterface(t *testing.T) {
	input := `
program Poly;

type
  TNode = class end;
  TVar = class(TNode) Name: String; end;
  TList = class Items: array of TNode; end;

var
  n: TNode;
  v: TVar;
  v2: TVar;
  li: TList;
begin
  v := TVar.Create;
  v.Name := 'x';
  li := TList.Create;
  append(li.Items, v);
  n := li.Items[0];
  if n is TVar then
  begin
    v2 := n as TVar;
    WriteLn(v2.Name);
  end;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)

	// Base class with children → empty interface.
	assertContains(t, out, "type TNode interface {")
	assertNotContains(t, out, "type TNode struct {")
	// Concrete subclass → struct.
	assertContains(t, out, "type TVar struct {")
	// Base-typed variable → interface (no pointer).
	assertContains(t, out, "var n TNode")
	// Concrete-typed variable → pointer.
	assertContains(t, out, "var v *TVar")
	// Heterogeneous array field → []TBase (interface slice).
	assertContains(t, out, "Items []TNode")
	// is/as → Go type assertion on the interface.
	assertContains(t, out, "n.(*TVar)")
}

// TestPolymorphism_NoIsAs_KeepsStructInheritance: programs WITHOUT `is`/`as`
// must keep the v3.1.0 behavior — base classes as structs with embedded parent
// (field inheritance). Regression guard for example19/example40.
func TestPolymorphism_NoIsAs_KeepsStructInheritance(t *testing.T) {
	input := `
program Inherit;

type
  TAnimal = class Name: String; end;
  TDog = class(TAnimal) Breed: String; end;
  TShape = class Color: String; end;
  TRectangle = class(TShape) Width: Integer; end;

var
  cat: TAnimal;
  rect: TRectangle;
begin
  cat := TAnimal.Create;
  cat.Name := 'Whiskers';
  rect := TRectangle.Create;
  rect.Color := 'blue';
  rect.Width := 8;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)

	// No is/as → base classes stay structs (field inheritance via embedding).
	assertContains(t, out, "type TAnimal struct {")
	assertNotContains(t, out, "type TAnimal interface {")
	assertContains(t, out, "type TShape struct {")
	// Concrete subclass embeds the parent struct (Color inherited by TRectangle).
	assertContains(t, out, "TRectangle struct {")
	assertContains(t, out, "TShape")
	// Base-typed variable → pointer (field access works directly).
	assertContains(t, out, "var cat *TAnimal")
	// Field access through base-typed var and through embedded parent preserved.
	assertContains(t, out, "cat.Name")
	assertContains(t, out, "rect.Color")
}

// TestPolymorphism_AsBuiltin: `Args` maps to os.Args[1:] and registers the os import.
func TestPolymorphism_AsBuiltinArgs(t *testing.T) {
	input := `
program UseArgs;
begin
  if Length(Args) > 0 then
    WriteLn(Args[0]);
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "os.Args[1:]")
	if !strings.Contains(out, `"os"`) {
		t.Errorf("expected os import to be registered; got:\n%s", out)
	}
}
