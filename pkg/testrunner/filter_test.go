package testrunner_test

import (
	"os"
	"path/filepath"
	"testing"

	"kylix/pkg/testrunner"
)

// kylix test high-level features (Task 3): Setup/Teardown, --filter.

func TestRunner_FilterCases(t *testing.T) {
	r := testrunner.New(false)
	cases := []testrunner.TestCase{
		{Name: "TestAdd", File: "math_test.klx"},
		{Name: "TestSubtract", File: "math_test.klx"},
		{Name: "TestStringConcat", File: "string_test.klx"},
	}

	r.SetFilter("Math") // matches none
	if got := r.FilterCases(cases); len(got) != 0 {
		t.Errorf("expected 0 matches for 'Math', got %d", len(got))
	}

	r.SetFilter("Add")
	got := r.FilterCases(cases)
	if len(got) != 1 || got[0].Name != "TestAdd" {
		t.Errorf("expected only TestAdd, got: %v", got)
	}

	r.SetFilter("Test") // matches all
	if got := r.FilterCases(cases); len(got) != 3 {
		t.Errorf("expected 3 matches, got %d", len(got))
	}
}

func TestRunner_FilterEmptyReturnsAll(t *testing.T) {
	r := testrunner.New(false)
	cases := []testrunner.TestCase{
		{Name: "TestA"},
		{Name: "TestB"},
	}
	if got := r.FilterCases(cases); len(got) != 2 {
		t.Errorf("empty filter should return all, got %d", len(got))
	}
}

func TestRunner_DiscoverIgnoresSetupTeardown(t *testing.T) {
	dir := t.TempDir()
	src := `unit fix_test;

procedure Setup;
begin
end;

procedure Teardown;
begin
end;

procedure TestRealOne;
begin
  Assert(true, 'ok');
end;

procedure TestAnother;
begin
  Assert(true, 'ok');
end;
`
	f := filepath.Join(dir, "fix_test.klx")
	os.WriteFile(f, []byte(src), 0644)

	r := testrunner.New(false)
	cases, err := r.DiscoverFile(f)
	if err != nil {
		t.Fatal(err)
	}
	// Should find only Test* — Setup and Teardown excluded.
	if len(cases) != 2 {
		t.Errorf("expected 2 Test* procedures (Setup/Teardown excluded), got %d: %v", len(cases), cases)
	}
}
