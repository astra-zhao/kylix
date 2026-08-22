package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_jwt tests — verify the v6.3.0 (signature) + v6.6.0 (claims) jwt
// module lowers to real IR: JwtVerify parses the payload into a Variant map
// with exp expiry; JwtSign supports extraClaims; accessors read claims.

func TestJwt_VerifyReturnsClaimsMap(t *testing.T) {
	ir := generateIR(t, `program p;
uses jwt;
begin
  var t := JwtSign('s', 'alice', 3600);
  var c := JwtVerify('s', t);
  WriteLn(JwtSubject(c));
end.`)
	// Verify decodes the payload (b64urldecode) and parses it into a claims
	// htab, then box_map's it into a Variant map (v6.6.0).
	assertIRContains(t, ir, "define ptr @__kylix_jwt_b64url_decode(ptr %str, i64 %n)")
	assertIRContains(t, ir, "call ptr @__kylix_json_parse_flat(ptr")
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_map(ptr")
	// exp expiry check consults the claims' exp via variant_as_int.
	assertIRContains(t, ir, "call i64 @__kylix_variant_as_int(ptr")
	assertIRContains(t, ir, "call i64 @time(ptr null)")
}

func TestJwt_FailureReturnsNilbox(t *testing.T) {
	ir := generateIR(t, `program p;
uses jwt;
begin
  var c := JwtVerify('s', 'bad.token.here');
  if c = nil then WriteLn('bad');
end.`)
	// A rejected/tampered token must return the nilbox Variant (not a null
	// ptr), so `c = nil` compares correctly without crashing. v6.6.0.
	assertIRContains(t, ir, "ret ptr @__kylix_variant_nilbox")
}

func TestJwt_SubjectAndClaimAccessors(t *testing.T) {
	ir := generateIR(t, `program p;
uses jwt;
begin
  var t := JwtSign('s', 'alice', 3600);
  var c := JwtVerify('s', t);
  WriteLn(JwtSubject(c));
  WriteLn(JwtGetString(c, 'role'));
  WriteLn(JwtGetInt(c, 'level'));
end.`)
	// Accessors read claims through the Variant map runtime.
	assertIRContains(t, ir, "call ptr @__kylix_variant_map_get(ptr")
	assertIRContains(t, ir, "call ptr @__kylix_variant_as_str(ptr")
	assertIRContains(t, ir, "call i64 @__kylix_variant_as_int(ptr")
}

func TestJwt_SignMergesExtraClaims(t *testing.T) {
	ir := generateIR(t, `program p;
uses jwt;
var extra: map[String]Variant;
begin
  extra['role'] := 'admin';
  var t := JwtSign('s', 'alice', 3600, extra);
end.`)
	// extraClaims (a map[String]Variant) is boxed into a Variant map and
	// walked: htab_keys + htab_get_variant, with the key serialized into the
	// payload JSON. A comma is emitted before every claim except the first
	// (the first-claim no-comma branch guards `{,"...`).
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_map(ptr")
	assertIRContains(t, ir, "call { ptr, i64 } @__kylix_htab_keys(ptr")
	assertIRContains(t, ir, "call ptr @__kylix_htab_get_variant(ptr")
	if strings.Contains(ir, "{\",\"") {
		t.Errorf("payload starts with a stray comma (first extra claim)\nIR:\n%s", ir)
	}
}
