package llvmgen_test

import (
	"strings"
	"testing"
)

// v0.9.0 P1.7: LLVM-side multipart upload — eager parse in BootRun, per-
// request fields/files htabs on the request handle (slots 64/72), req.File /
// req.SaveFile multi-return defines, req.MultipartField inline lookup.

const bootUploadProgram = `program p;
uses boot;
[Controller('/api')]
type
  TUploadController = class
    [Post('/avatar')]
    function Avatar(req: TRequest): TResponse;
    var
      content, filename, ok := req.File('avatar');
      path, ok2 := req.SaveFile('avatar', '/tmp/uploads');
      note := req.MultipartField('note');
    begin
      if ok then
        result := BootText(200, filename)
      else
        result := BootText(400, 'no file');
    end;
  end;
begin
  BootRun(8108);
end.`

func TestBoot_UploadIR(t *testing.T) {
	ir := generateIR(t, bootUploadProgram)
	// Eager parse: define + BootRun call right after read_body.
	assertIRContains(t, ir, "define void @__kylix_boot_multipart_parse(ptr %req)")
	assertIRContains(t, ir, "call void @__kylix_boot_multipart_parse(ptr")
	// Parse guards + split needles.
	assertIRContains(t, ir, "Content-Type: ")
	assertIRContains(t, ir, "multipart/")
	assertIRContains(t, ir, "boundary=")
	assertIRContains(t, ir, `c"name=\22\00"`)
	assertIRContains(t, ir, `c"filename=\22\00"`)
	assertIRContains(t, ir, "\\0D\\0A\\0D\\0A")
	// Size cap (8 MiB) and field cap (64).
	assertIRContains(t, ir, "8388608")
	assertIRContains(t, ir, "icmp sle i64")
	// req.File/req.SaveFile defines + the pre-scan-declared aggregates.
	assertIRContains(t, ir, "%__ret_BootRequest_File = type { ptr, ptr, i1 }")
	assertIRContains(t, ir, "%__ret_BootRequest_SaveFile = type { ptr, i1 }")
	assertIRContains(t, ir, "define %__ret_BootRequest_File @__kylix_boot_req_file(ptr %req, ptr %name)")
	assertIRContains(t, ir, "define %__ret_BootRequest_SaveFile @__kylix_boot_req_savefile(ptr %req, ptr %name, ptr %dir)")
	assertIRContains(t, ir, "call %__ret_BootRequest_File @__kylix_boot_req_file(ptr")
	assertIRContains(t, ir, "call %__ret_BootRequest_SaveFile @__kylix_boot_req_savefile(ptr")
	// Miss path returns the zero aggregate (Go (nil, "", false) parity).
	assertIRContains(t, ir, "ret %__ret_BootRequest_File zeroinitializer")
	assertIRContains(t, ir, "ret %__ret_BootRequest_SaveFile zeroinitializer")
	// SaveFile: backslash fold, base after last '/', dot-prefix rejection,
	// lowercase ext whitelist, mkdir 0755, fopen "wb".
	assertIRContains(t, ir, "select i1")
	assertIRContains(t, ir, "@strrchr(ptr")
	assertIRContains(t, ir, "@mkdir(ptr")
	assertIRContains(t, ir, "i32 493")
	assertIRContains(t, ir, "@fopen(ptr")
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".txt", ".csv", ".zip"} {
		assertIRContains(t, ir, ext)
	}
	// The handler destructures both multi-return calls via extractvalue.
	assertIRContains(t, ir, "extractvalue %__ret_BootRequest_File %")
	assertIRContains(t, ir, "extractvalue %__ret_BootRequest_SaveFile %")
	// Request handle slots 64/72 in use (fields/files maps).
	assertIRContains(t, ir, "i64 64")
	assertIRContains(t, ir, "i64 72")
	// Non-multipart early exit clears both map slots (arena-recycled handle).
	if strings.Count(ir, "store ptr null, ptr") < 2 {
		t.Fatalf("expected map-slot null stores in parse entry:\n%s", ir)
	}
	if strings.Contains(ir, "unsupported receiver") {
		t.Fatalf("upload methods collapsed to unsupported-receiver stub:\n%s", ir)
	}
}

// htab_get null-table short-circuit (Go nil-map read semantics): req_file's
// null path and MultipartField's inline lookup rely on it (a null table must
// not reach htab_check, which dereferences).
func TestBoot_HtabGetNullSafe(t *testing.T) {
	ir := generateIR(t, bootUploadProgram)
	getDefine := defineBody(ir, "define ptr @__kylix_htab_get(ptr %t, ptr %key)")
	if !strings.Contains(getDefine, "icmp eq ptr %t, null") {
		t.Fatalf("htab_get lacks null-table guard:\n%s", getDefine)
	}
	if !strings.Contains(getDefine, "ret ptr null") {
		t.Fatalf("htab_get null path must return null:\n%s", getDefine)
	}
}

// defineBody returns the body of the first define matching def.
func defineBody(ir, def string) string {
	i := strings.Index(ir, def)
	if i < 0 {
		return ""
	}
	j := strings.Index(ir[i:], "\n}")
	if j < 0 {
		return ir[i:]
	}
	return ir[i : i+j+2]
}
