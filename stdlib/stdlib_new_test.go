package stdlib

import (
	"os"
	"path/filepath"
	"testing"
)

// ===== Sysutil Tests =====

func TestFileOpenReadWrite(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "kylix_test_rw.txt")
	defer os.Remove(tmpFile)

	// Write
	f, err := FileOpen(tmpFile, FmWrite)
	if err != nil {
		t.Fatalf("FileOpen write: %v", err)
	}
	f.FileWriteLine("Hello Kylix")
	f.FileWriteLine("Second line")
	f.FileClose()

	// Read back
	f2, err := FileOpen(tmpFile, FmRead)
	if err != nil {
		t.Fatalf("FileOpen read: %v", err)
	}
	content, err := f2.FileReadAll()
	f2.FileClose()
	if err != nil {
		t.Fatalf("FileReadAll: %v", err)
	}
	if content != "Hello Kylix\nSecond line\n" {
		t.Errorf("Got: %q", content)
	}
}

func TestReadWriteFile(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "kylix_test_rwfile.txt")
	defer os.Remove(tmpFile)

	err := WriteFile(tmpFile, "test content")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	content, err := ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if content != "test content" {
		t.Errorf("Got: %q", content)
	}
}

func TestAppendFile(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "kylix_test_append.txt")
	defer os.Remove(tmpFile)

	WriteFile(tmpFile, "line1\n")
	AppendFile(tmpFile, "line2\n")

	content, _ := ReadFile(tmpFile)
	if content != "line1\nline2\n" {
		t.Errorf("Got: %q", content)
	}
}

func TestFileExists(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "kylix_test_exists.txt")
	defer os.Remove(tmpFile)

	if FileExists(tmpFile) {
		t.Error("File should not exist yet")
	}

	WriteFile(tmpFile, "x")
	if !FileExists(tmpFile) {
		t.Error("File should exist")
	}
}

func TestDirOperations(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "kylix_test_dir")
	defer os.RemoveAll(tmpDir)

	err := CreateDir(tmpDir)
	if err != nil {
		t.Fatalf("CreateDir: %v", err)
	}

	if !DirExists(tmpDir) {
		t.Error("Dir should exist")
	}

	WriteFile(filepath.Join(tmpDir, "a.txt"), "a")
	WriteFile(filepath.Join(tmpDir, "b.txt"), "b")

	files, err := ListDir(tmpDir)
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestCopyFile(t *testing.T) {
	src := filepath.Join(os.TempDir(), "kylix_test_copy_src.txt")
	dst := filepath.Join(os.TempDir(), "kylix_test_copy_dst.txt")
	defer os.Remove(src)
	defer os.Remove(dst)

	WriteFile(src, "copy me")
	err := CopyFile(src, dst)
	if err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	content, _ := ReadFile(dst)
	if content != "copy me" {
		t.Errorf("Got: %q", content)
	}
}

func TestPathOperations(t *testing.T) {
	if PathJoin("a", "b", "c.txt") != filepath.Join("a", "b", "c.txt") {
		t.Error("PathJoin failed")
	}
	if PathDir("/a/b/c.txt") != "/a/b" {
		t.Error("PathDir failed")
	}
	if PathBase("/a/b/c.txt") != "c.txt" {
		t.Error("PathBase failed")
	}
	if PathExt("file.txt") != ".txt" {
		t.Error("PathExt failed")
	}
}

func TestReadLinesWriteLines(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "kylix_test_lines.txt")
	defer os.Remove(tmpFile)

	lines := []string{"alpha", "beta", "gamma"}
	err := WriteLines(tmpFile, lines)
	if err != nil {
		t.Fatalf("WriteLines: %v", err)
	}

	result, err := ReadLines(tmpFile)
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}
	if len(result) != 3 || result[0] != "alpha" || result[2] != "gamma" {
		t.Errorf("Got: %v", result)
	}
}

func TestGetFileSize(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "kylix_test_size.txt")
	defer os.Remove(tmpFile)

	WriteFile(tmpFile, "12345")
	size, err := GetFileSize(tmpFile)
	if err != nil {
		t.Fatalf("GetFileSize: %v", err)
	}
	if size != 5 {
		t.Errorf("Expected 5, got %d", size)
	}
}

// ===== JSON Tests =====

func TestJsonEncodeDecode(t *testing.T) {
	data := map[string]interface{}{
		"name": "Kylix",
		"age":  float64(3),
	}

	jsonStr, err := JsonEncode(data)
	if err != nil {
		t.Fatalf("JsonEncode: %v", err)
	}

	decoded, err := JsonDecodeMap(jsonStr)
	if err != nil {
		t.Fatalf("JsonDecodeMap: %v", err)
	}

	if JsonGetString(decoded, "name") != "Kylix" {
		t.Error("name mismatch")
	}
	if JsonGetInt(decoded, "age") != 3 {
		t.Error("age mismatch")
	}
}

func TestJsonPretty(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	pretty, err := JsonEncodePretty(data)
	if err != nil {
		t.Fatalf("JsonEncodePretty: %v", err)
	}
	if pretty == "" {
		t.Error("Empty output")
	}
}

func TestJsonDecodeArray(t *testing.T) {
	jsonStr := `[1, 2, 3]`
	arr, err := JsonDecodeArray(jsonStr)
	if err != nil {
		t.Fatalf("JsonDecodeArray: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("Expected 3 items, got %d", len(arr))
	}
}

func TestJsonAccessors(t *testing.T) {
	data := map[string]interface{}{
		"str":   "hello",
		"num":   float64(42),
		"flag":  true,
		"pi":    3.14,
		"child": map[string]interface{}{"x": float64(1)},
		"items": []interface{}{float64(1), float64(2)},
	}

	if JsonGetString(data, "str") != "hello" {
		t.Error("GetString failed")
	}
	if JsonGetInt(data, "num") != 42 {
		t.Error("GetInt failed")
	}
	if !JsonGetBool(data, "flag") {
		t.Error("GetBool failed")
	}
	if JsonGetFloat(data, "pi") != 3.14 {
		t.Error("GetFloat failed")
	}
	if JsonGetMap(data, "child") == nil {
		t.Error("GetMap failed")
	}
	if JsonGetArray(data, "items") == nil {
		t.Error("GetArray failed")
	}
	if !JsonHasKey(data, "str") {
		t.Error("HasKey failed")
	}
	if JsonHasKey(data, "missing") {
		t.Error("HasKey false positive")
	}
}

func TestJsonIsValid(t *testing.T) {
	if !JsonIsValid(`{"a": 1}`) {
		t.Error("Should be valid")
	}
	if JsonIsValid(`{invalid}`) {
		t.Error("Should be invalid")
	}
}

func TestJsonFileRoundTrip(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "kylix_test.json")
	defer os.Remove(tmpFile)

	data := map[string]interface{}{
		"name":  "test",
		"count": float64(5),
	}

	err := JsonWriteFile(tmpFile, data)
	if err != nil {
		t.Fatalf("JsonWriteFile: %v", err)
	}

	result, err := JsonReadFile(tmpFile)
	if err != nil {
		t.Fatalf("JsonReadFile: %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map")
	}
	if m["name"] != "test" {
		t.Error("name mismatch")
	}
}

// ===== DateTime Tests =====

func TestNowToday(t *testing.T) {
	now := Now()
	if now.Year() < 2024 {
		t.Error("Year should be >= 2024")
	}

	today := Today()
	if today.Hour() != 0 || today.Minute() != 0 {
		t.Error("Today should be at midnight")
	}
}

func TestMakeDate(t *testing.T) {
	dt := MakeDate(2024, 6, 15)
	if dt.Year() != 2024 || dt.Month() != 6 || dt.Day() != 15 {
		t.Error("MakeDate failed")
	}
}

func TestMakeTime(t *testing.T) {
	dt := MakeTime(2024, 12, 25, 10, 30, 0)
	if dt.Hour() != 10 || dt.Minute() != 30 {
		t.Error("MakeTime failed")
	}
}

func TestDateFormatting(t *testing.T) {
	dt := MakeDate(2024, 1, 15)
	if dt.FormatDate() != "2024-01-15" {
		t.Errorf("FormatDate: %s", dt.FormatDate())
	}
}

func TestDateArithmetic(t *testing.T) {
	dt := MakeDate(2024, 1, 1)

	next := dt.AddDays(10)
	if next.Day() != 11 {
		t.Errorf("AddDays: expected 11, got %d", next.Day())
	}

	monthLater := dt.AddMonths(1)
	if monthLater.Month() != 2 {
		t.Error("AddMonths failed")
	}

	yearLater := dt.AddYears(1)
	if yearLater.Year() != 2025 {
		t.Error("AddYears failed")
	}
}

func TestDateDiff(t *testing.T) {
	d1 := MakeDate(2024, 1, 1)
	d2 := MakeDate(2024, 1, 11)

	if d2.DiffDays(d1) != 10 {
		t.Errorf("DiffDays: %d", d2.DiffDays(d1))
	}
}

func TestDateComparison(t *testing.T) {
	d1 := MakeDate(2024, 1, 1)
	d2 := MakeDate(2024, 6, 1)

	if !d1.Before(d2) {
		t.Error("d1 should be before d2")
	}
	if !d2.After(d1) {
		t.Error("d2 should be after d1")
	}
}

func TestDateUtilities(t *testing.T) {
	// Saturday
	dt := MakeDate(2024, 1, 6)
	if !dt.IsWeekend() {
		t.Error("Jan 6 2024 should be weekend")
	}

	// Leap year
	dt2 := MakeDate(2024, 1, 1)
	if !dt2.IsLeapYear() {
		t.Error("2024 should be leap year")
	}

	dt3 := MakeDate(2024, 2, 1)
	if dt3.DaysInMonth() != 29 {
		t.Errorf("Feb 2024 should have 29 days, got %d", dt3.DaysInMonth())
	}
}

func TestParseDate(t *testing.T) {
	dt, err := ParseDate("2024-06-15")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if dt.Year() != 2024 || dt.Month() != 6 || dt.Day() != 15 {
		t.Error("ParseDate wrong values")
	}
}

func TestParseDateTime(t *testing.T) {
	dt, err := ParseDateTime("2024-06-15 10:30:00")
	if err != nil {
		t.Fatalf("ParseDateTime: %v", err)
	}
	if dt.Hour() != 10 || dt.Minute() != 30 {
		t.Error("ParseDateTime wrong time")
	}
}

func TestTimestamp(t *testing.T) {
	ts := GetTimestamp()
	if ts < 1700000000 {
		t.Error("Timestamp too small")
	}

	tsMs := GetTimestampMs()
	if tsMs < 1700000000000 {
		t.Error("TimestampMs too small")
	}
}

// ===== Regex Tests =====

func TestRegexCompileAndMatch(t *testing.T) {
	re, err := RegexCompile(`\d+`)
	if err != nil {
		t.Fatalf("RegexCompile: %v", err)
	}

	if !re.Match("abc123def") {
		t.Error("Should match")
	}
	if re.Match("no digits here") {
		t.Error("Should not match")
	}
}

func TestRegexFind(t *testing.T) {
	re := RegexMustCompile(`\d+`)
	result := re.Find("abc123def456")
	if result != "123" {
		t.Errorf("Find: %s", result)
	}

	all := re.FindAll("abc123def456")
	if len(all) != 2 || all[0] != "123" || all[1] != "456" {
		t.Errorf("FindAll: %v", all)
	}
}

func TestRegexReplace(t *testing.T) {
	re := RegexMustCompile(`\d+`)
	result := re.Replace("abc123def456", "NUM")
	if result != "abcNUMdefNUM" {
		t.Errorf("Replace: %s", result)
	}

	first := re.ReplaceFirst("abc123def456", "NUM")
	if first != "abcNUMdef456" {
		t.Errorf("ReplaceFirst: %s", first)
	}
}

func TestRegexSplit(t *testing.T) {
	re := RegexMustCompile(`[,\s]+`)
	parts := re.Split("a, b, c")
	if len(parts) != 3 || parts[0] != "a" || parts[2] != "c" {
		t.Errorf("Split: %v", parts)
	}
}

func TestRegexGroups(t *testing.T) {
	re := RegexMustCompile(`(\w+)@(\w+)`)
	groups := re.Groups("user@host")
	if len(groups) != 3 || groups[1] != "user" || groups[2] != "host" {
		t.Errorf("Groups: %v", groups)
	}
}

func TestRegexConvenience(t *testing.T) {
	if !RegexMatch(`^\d+$`, "12345") {
		t.Error("RegexMatch failed")
	}

	result := RegexFind(`\d+`, "abc123")
	if result != "123" {
		t.Errorf("RegexFind: %s", result)
	}

	replaced := RegexReplace(`\s+`, "a  b  c", " ")
	if replaced != "a b c" {
		t.Errorf("RegexReplace: %s", replaced)
	}
}

func TestPatternHelpers(t *testing.T) {
	if !IsEmail("test@example.com") {
		t.Error("IsEmail failed")
	}
	if IsEmail("not-an-email") {
		t.Error("IsEmail false positive")
	}

	if !IsNumeric("12345") {
		t.Error("IsNumeric failed")
	}
	if IsNumeric("123abc") {
		t.Error("IsNumeric false positive")
	}

	if !IsAlpha("Hello") {
		t.Error("IsAlpha failed")
	}
	if !IsAlphaNumeric("Hello123") {
		t.Error("IsAlphaNumeric failed")
	}
}
