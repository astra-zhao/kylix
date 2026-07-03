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
	if !strings.Contains(ir, "call ptr @__kylix_datetime_arena_alloc(i64 8)") {
		t.Error("Now() should use arena_alloc(8) for TDateTime")
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

// Phase 2 Tests

func TestDatetimeHour(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var h := dt.Hour(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call i64 @__kylix_datetime_Hour") {
		t.Error("Hour() should call @__kylix_datetime_Hour")
	}
	if !strings.Contains(ir, "define i64 @__kylix_datetime_Hour(ptr %self)") {
		t.Error("Hour() body should be emitted")
	}
	if !strings.Contains(ir, "getelementptr inbounds [56 x i8], ptr %") && !strings.Contains(ir, "i64 8") {
		t.Error("Hour() should GEP to offset 8 (tm_hour)")
	}
}

func TestDatetimeMinute(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var m := dt.Minute(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call i64 @__kylix_datetime_Minute") {
		t.Error("Minute() should call @__kylix_datetime_Minute")
	}
	if !strings.Contains(ir, "define i64 @__kylix_datetime_Minute(ptr %self)") {
		t.Error("Minute() body should be emitted")
	}
}

func TestDatetimeSecond(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var s := dt.Second(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call i64 @__kylix_datetime_Second") {
		t.Error("Second() should call @__kylix_datetime_Second")
	}
	if !strings.Contains(ir, "define i64 @__kylix_datetime_Second(ptr %self)") {
		t.Error("Second() body should be emitted")
	}
}

func TestDatetimeDayOfWeek(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var dow := dt.DayOfWeek(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call i64 @__kylix_datetime_DayOfWeek") {
		t.Error("DayOfWeek() should call @__kylix_datetime_DayOfWeek")
	}
	if !strings.Contains(ir, "define i64 @__kylix_datetime_DayOfWeek(ptr %self)") {
		t.Error("DayOfWeek() body should be emitted")
	}
	if !strings.Contains(ir, "getelementptr inbounds [56 x i8], ptr %") && !strings.Contains(ir, "i64 24") {
		t.Error("DayOfWeek() should GEP to offset 24 (tm_wday)")
	}
}

func TestDatetimeToday(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Today(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_Today()") {
		t.Error("Today() should call @__kylix_datetime_Today()")
	}
	if !strings.Contains(ir, "define ptr @__kylix_datetime_Today()") {
		t.Error("Today() body should be emitted")
	}
	if !strings.Contains(ir, "call ptr @localtime_r") {
		t.Error("Today() should use localtime_r (thread-safe)")
	}
	if !strings.Contains(ir, "call i64 @mktime") {
		t.Error("Today() should call mktime to normalize date")
	}
}

func TestDatetimeAddHours(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var dt2 := dt.AddHours(2); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_AddHours") {
		t.Error("AddHours() should call @__kylix_datetime_AddHours")
	}
	if !strings.Contains(ir, "define ptr @__kylix_datetime_AddHours(ptr %self, i64 %hours)") {
		t.Error("AddHours() body should be emitted")
	}
	if !strings.Contains(ir, "mul i64 %hours, 3600") {
		t.Error("AddHours() should multiply hours by 3600")
	}
}

func TestDatetimeAddMinutes(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var dt2 := dt.AddMinutes(30); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_AddMinutes") {
		t.Error("AddMinutes() should call @__kylix_datetime_AddMinutes")
	}
	if !strings.Contains(ir, "mul i64 %minutes, 60") {
		t.Error("AddMinutes() should multiply minutes by 60")
	}
}

func TestDatetimeAddSeconds(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); var dt2 := dt.AddSeconds(45); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_AddSeconds") {
		t.Error("AddSeconds() should call @__kylix_datetime_AddSeconds")
	}
	if !strings.Contains(ir, "define ptr @__kylix_datetime_AddSeconds(ptr %self, i64 %seconds)") {
		t.Error("AddSeconds() body should be emitted")
	}
	// AddSeconds should NOT multiply (direct add)
	if strings.Contains(ir, "mul i64 %seconds") {
		t.Error("AddSeconds() should NOT multiply (direct add to time_t)")
	}
}

func TestDatetimeLocaltime_rDeclaration(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Today(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "declare ptr @localtime_r") {
		t.Error("Should declare localtime_r for thread-safe time conversion")
	}
}

func TestDatetimeArenaAlloc(t *testing.T) {
	src := `program Test; uses datetime; begin var dt := Now(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call ptr @__kylix_datetime_arena_alloc(i64 8)") {
		t.Error("Now() should use arena allocator instead of malloc")
	}
	if !strings.Contains(ir, "define ptr @__kylix_datetime_arena_alloc(i64 %size)") {
		t.Error("Arena allocator body should be emitted")
	}
	if !strings.Contains(ir, "@__kylix_datetime_arena = internal global [1048576 x i8]") {
		t.Error("Arena buffer (1MB) should be declared")
	}
	if !strings.Contains(ir, "@__kylix_datetime_arena_ptr = internal global ptr") {
		t.Error("Arena pointer should be declared")
	}
}

func TestDatetimeFreeArena(t *testing.T) {
	src := `program Test; uses datetime; begin FreeArena(); end.`
	ir := generateIR(t, src)
	if !strings.Contains(ir, "call void @__kylix_datetime_FreeArena()") {
		t.Error("FreeArena() should call @__kylix_datetime_FreeArena()")
	}
	if !strings.Contains(ir, "define void @__kylix_datetime_FreeArena()") {
		t.Error("FreeArena() body should be emitted")
	}
	if !strings.Contains(ir, "store ptr @__kylix_datetime_arena, ptr @__kylix_datetime_arena_ptr") {
		t.Error("FreeArena() should reset arena pointer")
	}
}
