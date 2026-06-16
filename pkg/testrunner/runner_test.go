package testrunner_test

import (
	"os"
	"path/filepath"
	"testing"

	"kylix/pkg/testrunner"
)

func TestDiscoverFile_FindsTests(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "math_test.klx")
	src := `unit math_test;

procedure TestAdd;
begin
  Assert(1 + 1 = 2, 'one plus one should be two');
end;

procedure TestSubtract;
begin
  Assert(5 - 3 = 2, 'five minus three should be two');
end;
`
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	r := testrunner.New(false)
	cases, err := r.DiscoverFile(f)
	if err != nil {
		t.Fatalf("DiscoverFile error: %v", err)
	}

	if len(cases) != 2 {
		t.Errorf("expected 2 test cases, got %d", len(cases))
	}
	if len(cases) > 0 && cases[0].Name != "TestAdd" {
		t.Errorf("expected TestAdd, got %s", cases[0].Name)
	}
	if len(cases) > 1 && cases[1].Name != "TestSubtract" {
		t.Errorf("expected TestSubtract, got %s", cases[1].Name)
	}
}

func TestDiscoverFile_NonTestFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "math.klx")
	src := `unit math;
function Add(a, b: Integer): Integer;
begin result := a + b; end;
`
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	r := testrunner.New(false)
	cases, err := r.DiscoverFile(f)
	if err != nil {
		t.Fatalf("DiscoverFile error: %v", err)
	}

	// Non-test file should return 0 cases
	if len(cases) != 0 {
		t.Errorf("expected 0 cases from non-test file, got %d", len(cases))
	}
}

func TestDiscover_FindsTestFiles(t *testing.T) {
	dir := t.TempDir()

	testSrc := `unit calc_test;
procedure TestMul;
begin
  Assert(2 * 3 = 6, 'two times three');
end;
`
	otherSrc := `unit calc;
function Mul(a, b: Integer): Integer;
begin result := a * b; end;
`
	if err := os.WriteFile(filepath.Join(dir, "calc_test.klx"), []byte(testSrc), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "calc.klx"), []byte(otherSrc), 0644); err != nil {
		t.Fatal(err)
	}

	r := testrunner.New(false)
	cases, err := r.Discover(dir)
	if err != nil {
		t.Fatalf("Discover error: %v", err)
	}

	if len(cases) != 1 {
		t.Errorf("expected 1 test case (only from _test.klx), got %d", len(cases))
	}
}

func TestSummary_Count(t *testing.T) {
	results := []testrunner.TestResult{
		{TestCase: testrunner.TestCase{Name: "TestA"}, Passed: true},
		{TestCase: testrunner.TestCase{Name: "TestB"}, Passed: false, Message: "fail"},
		{TestCase: testrunner.TestCase{Name: "TestC"}, Passed: true},
	}
	passed, failed := testrunner.Summary(results)
	if passed != 2 {
		t.Errorf("expected 2 passed, got %d", passed)
	}
	if failed != 1 {
		t.Errorf("expected 1 failed, got %d", failed)
	}
}
