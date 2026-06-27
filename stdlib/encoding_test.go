package stdlib

import "testing"

func TestEncoding_Base64RoundTrip(t *testing.T) {
	enc := Base64Encode("hello")
	if enc != "aGVsbG8=" {
		t.Fatalf("Base64Encode(hello)=%q", enc)
	}
	dec, err := Base64Decode(enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != "hello" {
		t.Fatalf("round trip: %q", dec)
	}
}

func TestEncoding_Base64URL(t *testing.T) {
	enc := Base64URLEncode("abc 123")
	dec, err := Base64URLDecode(enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != "abc 123" {
		t.Fatalf("url-safe round trip: %q", dec)
	}
}

func TestEncoding_HexRoundTrip(t *testing.T) {
	enc := HexEncode("\x00\xff\xab")
	dec, err := HexDecode(enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != "\x00\xff\xab" {
		t.Fatalf("hex round trip: %q", dec)
	}
}

func TestEncoding_UrlEncode(t *testing.T) {
	got := UrlEncode("a b=c")
	if got != "a+b%3Dc" {
		t.Fatalf("UrlEncode=%q", got)
	}
	dec, err := UrlDecode(got)
	if err != nil {
		t.Fatal(err)
	}
	if dec != "a b=c" {
		t.Fatalf("UrlDecode round trip: %q", dec)
	}
}

func TestEncoding_CsvRoundTrip(t *testing.T) {
	rows := [][]string{{"Name", "Email"}, {"Alice", "alice@example.com"}}
	csv, err := CsvEncode(rows)
	if err != nil {
		t.Fatal(err)
	}
	got, err := CsvDecode(csv)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0][0] != "Name" || got[1][1] != "alice@example.com" {
		t.Fatalf("CSV round trip: %v", got)
	}
}

func TestEncoding_JsonLinesRoundTrip(t *testing.T) {
	rows := []map[string]interface{}{{"name": "Alice", "age": float64(30)}}
	jl, err := JsonLinesEncode(rows)
	if err != nil {
		t.Fatal(err)
	}
	got, err := JsonLinesDecode(jl)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0]["name"] != "Alice" {
		t.Fatalf("JSON-Lines round trip: %v", got)
	}
}
