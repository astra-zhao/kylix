// encoding.go — Kylix stdlib encoding module: Base64, Hex, URL, CSV, JSON Lines.
package stdlib

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func Base64Decode(data string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func Base64URLEncode(data string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(data))
}

func Base64URLDecode(data string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func HexEncode(data string) string {
	return hex.EncodeToString([]byte(data))
}

func HexDecode(data string) (string, error) {
	raw, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func UrlEncode(s string) string {
	return url.QueryEscape(s)
}

func UrlDecode(s string) (string, error) {
	return url.QueryUnescape(s)
}

func CsvEncode(rows [][]string) (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)
	if err := w.WriteAll(rows); err != nil {
		return "", err
	}
	w.Flush()
	return b.String(), nil
}

func CsvDecode(s string) ([][]string, error) {
	return csv.NewReader(strings.NewReader(s)).ReadAll()
}

func JsonLinesEncode(rows []map[string]interface{}) (string, error) {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return "", fmt.Errorf("encoding: JsonLinesEncode: %w", err)
		}
	}
	return b.String(), nil
}

func JsonLinesDecode(s string) ([]map[string]interface{}, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	var out []map[string]interface{}
	for {
		var m map[string]interface{}
		if err := dec.Decode(&m); err != nil {
			break
		}
		if m != nil {
			out = append(out, m)
		}
	}
	return out, nil
}
