package stdlib

import "testing"

// entitymeta_test.go — the runtime registry behind the CRUD engine (v0.11.0).
// The LLVM backend implements the same surface in IR
// (pkg/llvmgen/stdlib_entitymeta.go); these tests pin the observable behaviour
// both backends must share.

func resetEntityMeta() {
	entityMetaTables = nil
	entityMetaIndex = map[string]*entityMetaEntry{}
	entityMetaNames = ""
}

func TestEntityMeta_RegisterAndQuery(t *testing.T) {
	resetEntityMeta()
	RegisterEntity("users", "id|Users|")
	RegisterEntityField("users", "id|number|Id|")
	RegisterEntityField("users", "username|text|Username|required,searchable")
	SetEntityNames("users")

	if got := EntityMetaOf("users"); got != "id|Users|" {
		t.Errorf("EntityMetaOf = %q", got)
	}
	if got := EntityFieldCount("users"); got != 2 {
		t.Errorf("EntityFieldCount = %d, want 2", got)
	}
	if got := EntityFieldAt("users", 1); got != "username|text|Username|required,searchable" {
		t.Errorf("EntityFieldAt(1) = %q", got)
	}
	if got := EntityNames(); got != "users" {
		t.Errorf("EntityNames = %q", got)
	}
}

func TestEntityMeta_UnknownTableIsEmpty(t *testing.T) {
	resetEntityMeta()
	RegisterEntity("users", "id|Users|")

	if got := EntityMetaOf("nope"); got != "" {
		t.Errorf("EntityMetaOf(unknown) = %q, want empty", got)
	}
	if got := EntityFieldCount("nope"); got != 0 {
		t.Errorf("EntityFieldCount(unknown) = %d, want 0", got)
	}
	if got := EntityFieldAt("nope", 0); got != "" {
		t.Errorf("EntityFieldAt(unknown) = %q, want empty", got)
	}
	if got := EntityNames(); got != "" {
		t.Errorf("EntityNames before SetEntityNames = %q, want empty", got)
	}
}

func TestEntityMeta_OutOfRangeIndexIsEmpty(t *testing.T) {
	resetEntityMeta()
	RegisterEntity("users", "id|Users|")
	RegisterEntityField("users", "id|number|Id|")

	if got := EntityFieldAt("users", 1); got != "" {
		t.Errorf("EntityFieldAt(1) past the end = %q, want empty", got)
	}
	if got := EntityFieldAt("users", -1); got != "" {
		t.Errorf("EntityFieldAt(-1) = %q, want empty", got)
	}
}

// Re-registering a table replaces it rather than appending, so a second run in
// the same process (or a duplicate emission) stays deterministic.
func TestEntityMeta_ReregisterReplaces(t *testing.T) {
	resetEntityMeta()
	RegisterEntity("users", "id|Users|")
	RegisterEntityField("users", "id|number|Id|")
	RegisterEntity("users", "id|Users v2|")
	RegisterEntityField("users", "username|text|Username|")
	SetEntityNames("users")

	if got := EntityMetaOf("users"); got != "id|Users v2|" {
		t.Errorf("EntityMetaOf after re-register = %q", got)
	}
	if got := EntityFieldCount("users"); got != 1 {
		t.Errorf("EntityFieldCount after re-register = %d, want 1 (fields reset)", got)
	}
}

// Fields registered before their entity are dropped, matching the LLVM
// backend's behaviour (it looks the table up in the array).
func TestEntityMeta_FieldBeforeEntityIsDropped(t *testing.T) {
	resetEntityMeta()
	RegisterEntityField("users", "id|number|Id|")
	RegisterEntity("users", "id|Users|")

	if got := EntityFieldCount("users"); got != 0 {
		t.Errorf("EntityFieldCount = %d, want 0 (orphan field dropped)", got)
	}
}
