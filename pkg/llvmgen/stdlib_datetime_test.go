package llvmgen_test

import (
	"strings"
	"testing"
)

func TestDatetimeNow(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call i64 @time(ptr null)") {
		t.Error("Now() should call time(null)")
	}
	if !strings.Contains(ir, "call ptr @malloc(i64 8)") {
		t.Error("Now() should malloc(8) for TDateTime")
	}
}

func TestDatetimeMakeDate(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := MakeDate(2025, 1, 15); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_MakeDate") {
		t.Error("MakeDate should call @__kylix_datetime_MakeDate")
	}
	if !strings.Contains(ir, "define ptr @__kylix_datetime_MakeDate(i64 %year, i64 %month, i64 %day)") {
		t.Error("MakeDate body should be emitted")
	}
	if !strings.Contains(ir, "call i64 @mktime") {
		t.Error("MakeDate should call mktime")
	}
}

func TestDatetimeYear(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var y := dt.Year(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call i64 @__kylix_datetime_Year") {
		t.Error("Year() should call @__kylix_datetime_Year")
	}
	if !strings.Contains(ir, "define i64 @__kylix_datetime_Year(ptr %self)") {
		t.Error("Year() body should be emitted")
	}
	if !strings.Contains(ir, "call ptr @localtime") {
		t.Error("Year() should call localtime")
	}
}

func TestDatetimeMonth(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var m := dt.Month(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call i64 @__kylix_datetime_Month") {
		t.Error("Month() should call @__kylix_datetime_Month")
	}
	if !strings.Contains(ir, "define i64 @__kylix_datetime_Month(ptr %self)") {
		t.Error("Month() body should be emitted")
	}
}

func TestDatetimeDay(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var d := dt.Day(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call i64 @__kylix_datetime_Day") {
		t.Error("Day() should call @__kylix_datetime_Day")
	}
	if !strings.Contains(ir, "define i64 @__kylix_datetime_Day(ptr %self)") {
		t.Error("Day() body should be emitted")
	}
}

func TestDatetimeFormatDate(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var s := dt.FormatDate(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_FormatDate") {
		t.Error("FormatDate() should call @__kylix_datetime_FormatDate")
	}
	if !strings.Contains(ir, "define ptr @__kylix_datetime_FormatDate(ptr %self)") {
		t.Error("FormatDate() body should be emitted")
	}
	if !strings.Contains(ir, "call i64 @strftime") {
		t.Error("FormatDate() should call strftime")
	}
}

func TestDatetimeAddDays(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var dt2 := dt.AddDays(7); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_AddDays") {
		t.Error("AddDays() should call @__kylix_datetime_AddDays")
	}
	if !strings.Contains(ir, "define ptr @__kylix_datetime_AddDays(ptr %self, i64 %days)") {
		t.Error("AddDays() body should be emitted")
	}
	if !strings.Contains(ir, "mul i64 %days, 86400") {
		t.Error("AddDays() should multiply days by 86400")
	}
}

func TestDatetimeChainedCalls(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := MakeDate(2025,1,1).AddDays(10); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_MakeDate") {
		t.Error("Chained call should invoke MakeDate")
	}
	if !strings.Contains(ir, "call ptr @__kylix_datetime_AddDays") {
		t.Error("Chained call should invoke AddDays")
	}
}

func TestDatetimeLibcDeclarations(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "declare i64 @time(ptr)") {
		t.Error("Should declare time()")
	}
	if !strings.Contains(ir, "declare ptr @localtime(ptr)") {
		t.Error("Should declare localtime()")
	}
	if !strings.Contains(ir, "declare i64 @mktime(ptr)") {
		t.Error("Should declare mktime()")
	}
	if !strings.Contains(ir, "declare i64 @strftime(ptr, i64, ptr, ptr)") {
		t.Error("Should declare strftime()")
	}
	if !strings.Contains(ir, "declare void @llvm.memset.p0.i64") {
		t.Error("Should declare llvm.memset intrinsic")
	}
}
