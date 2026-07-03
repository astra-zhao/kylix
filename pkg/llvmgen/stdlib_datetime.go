package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_datetime.go — LLVM IR implementation for datetime module
//
// Uses libc time.h to implement TDateTime class (wrapper around time_t).
// TDateTime instance: ptr → i64 (Unix timestamp in seconds since epoch).
//
// Core libc functions:
//   time_t time(time_t *) — current time
//   struct tm *localtime_r(time_t *, struct tm *) — convert time_t → tm (thread-safe)
//   time_t mktime(struct tm *) — convert tm → time_t
//   size_t strftime(char *, size_t, const char *, const struct tm *) — format tm
//
// PLATFORM COMPATIBILITY:
//   struct tm offsets are hardcoded for POSIX systems (Linux, macOS, BSD).
//   Verified on macOS Darwin 25.5.0. Windows has a different layout and is
//   NOT supported. struct tm layout (POSIX, 56 bytes):
//     offset 0:  tm_sec   (i32)    offset 16: tm_mon   (i32)
//     offset 4:  tm_min   (i32)    offset 20: tm_year  (i32)
//     offset 8:  tm_hour  (i32)    offset 24: tm_wday  (i32)
//     offset 12: tm_mday  (i32)    offset 28: tm_yday  (i32)
//                                   offset 32: tm_isdst (i32)
//
// THREAD SAFETY:
//   localtime_r is reentrant (POSIX.1-2001). Phase 3 removed non-reentrant
//   localtime() calls that used static buffers.
//
// MEMORY MANAGEMENT:
//   TDateTime instances are malloc'd without free (known leak). See Task #102
//   for planned solutions: Phase 3.5 arena allocator, Phase 4 mark-sweep GC.

// emitDatetimeCall generates a call to a datetime function (Now, MakeDate, or a
// TDateTime method) and enqueues the function body for later emission if needed.
func (g *Generator) emitDatetimeCall(funcName string, args []ast.Expression) (reg, typ string, err error) {
	switch funcName {
	case "Now":
		return g.emitDatetimeNowCall(args)
	case "Today":
		return g.emitDatetimeTodayCall(args)
	case "MakeDate":
		return g.emitDatetimeMakeDateCall(args)
	case "FreeArena":
		return g.emitDatetimeFreeArenaCall(args)
	default:
		// Unknown datetime function
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; datetime.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

// emitDatetimeNowCall emits Now() -> ptr (TDateTime instance with current time)
func (g *Generator) emitDatetimeNowCall(args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("datetime.Now expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "Now", "Now", 0)
	g.enqueueStdlib("datetime", "ArenaAlloc", "ArenaAlloc", 1)

	// Allocate TDateTime instance (8 bytes for time_t)
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_arena_alloc(i64 8)", inst))

	// Get current time: time_t now = time(NULL)
	nowVal := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @time(ptr null)", nowVal))

	// Store time_t into instance
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nowVal, inst))

	return inst, "TDateTime", nil
}

// emitDatetimeTodayCall emits Today() -> ptr (TDateTime with time set to midnight)
func (g *Generator) emitDatetimeTodayCall(args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("datetime.Today expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "Today", "Today", 0)

	// Call @__kylix_datetime_Today()
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_Today()", r))
	return r, "TDateTime", nil
}

// emitDatetimeMakeDateCall emits MakeDate(year, month, day) -> ptr
func (g *Generator) emitDatetimeMakeDateCall(args []ast.Expression) (string, string, error) {
	if len(args) != 3 {
		return "", "", fmt.Errorf("datetime.MakeDate expects 3 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "MakeDate", "MakeDate", 0)

	// Emit year, month, day arguments
	yearReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	monthReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	dayReg, _, err := g.emitExpr(args[2])
	if err != nil {
		return "", "", err
	}

	// Call @__kylix_datetime_MakeDate(i64 year, i64 month, i64 day) -> ptr
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_MakeDate(i64 %s, i64 %s, i64 %s)",
		inst, yearReg, monthReg, dayReg))

	return inst, "TDateTime", nil
}

// emitDatetimeTodayBody emits define for @__kylix_datetime_Today() -> ptr
func (g *Generator) emitDatetimeTodayBody() {
	g.line("define ptr @__kylix_datetime_Today() {")
	g.line("entry:")
	// Get current time
	nowVal := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @time(ptr null)", nowVal))

	// Allocate temporary for time_t and store it
	timeSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", timeSlot))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nowVal, timeSlot))

	// Allocate struct tm on stack (no need to copy from static buffer anymore)
	tmLocal := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmLocal))

	// Call localtime_r to populate local buffer
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %s, ptr %s)", tmPtr, timeSlot, tmLocal))

	// Zero out time fields in the local copy
	// tm_sec (offset 0)
	secPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 0", secPtr, tmLocal))
	g.line(fmt.Sprintf("  store i32 0, ptr %s", secPtr))

	// tm_min (offset 4)
	minPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 4", minPtr, tmLocal))
	g.line(fmt.Sprintf("  store i32 0, ptr %s", minPtr))

	// tm_hour (offset 8)
	hourPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 8", hourPtr, tmLocal))
	g.line(fmt.Sprintf("  store i32 0, ptr %s", hourPtr))

	// Convert back to time_t with mktime (using local copy)
	todayVal := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @mktime(ptr %s)", todayVal, tmLocal))

	// Allocate TDateTime instance and store
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_arena_alloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", todayVal, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// emitDatetimeMakeDateBody emits define for @__kylix_datetime_MakeDate(i64, i64, i64) -> ptr
// Converts (year, month, day) → time_t via mktime(struct tm{...})
func (g *Generator) emitDatetimeMakeDateBody() {
	g.line("define ptr @__kylix_datetime_MakeDate(i64 %year, i64 %month, i64 %day) {")
	g.line("entry:")

	// Allocate struct tm (56 bytes on most platforms)
	tmSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmSlot))

	// Zero-initialize tm
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 56, i1 false)", tmSlot))

	// tm.tm_year = year - 1900 (offset 20, i32)
	yearAdj := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %%year, 1900", yearAdj))
	yearI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", yearI32, yearAdj))
	yearPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 20", yearPtr, tmSlot))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", yearI32, yearPtr))

	// tm.tm_mon = month - 1 (offset 16, i32)
	monAdj := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %%month, 1", monAdj))
	monI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", monI32, monAdj))
	monPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 16", monPtr, tmSlot))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", monI32, monPtr))

	// tm.tm_mday = day (offset 12, i32)
	dayI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%day to i32", dayI32))
	dayPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 12", dayPtr, tmSlot))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", dayI32, dayPtr))

	// time_t t = mktime(&tm)
	timeVal := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @mktime(ptr %s)", timeVal, tmSlot))

	// Allocate TDateTime instance and store
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_arena_alloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", timeVal, inst))

	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// emitDatetimeMethodCall emits a call to a TDateTime instance method.
// receiver is the TDateTime instance (ptr), method is the method name.
func (g *Generator) emitDatetimeMethodCall(receiver string, method string, args []ast.Expression) (string, string, error) {
	switch method {
	case "Year":
		return g.emitDatetimeYearCall(receiver, args)
	case "Month":
		return g.emitDatetimeMonthCall(receiver, args)
	case "Day":
		return g.emitDatetimeDayCall(receiver, args)
	case "Hour":
		return g.emitDatetimeHourCall(receiver, args)
	case "Minute":
		return g.emitDatetimeMinuteCall(receiver, args)
	case "Second":
		return g.emitDatetimeSecondCall(receiver, args)
	case "DayOfWeek":
		return g.emitDatetimeDayOfWeekCall(receiver, args)
	case "FormatDate":
		return g.emitDatetimeFormatDateCall(receiver, args)
	case "AddDays":
		return g.emitDatetimeAddDaysCall(receiver, args)
	case "AddHours":
		return g.emitDatetimeAddHoursCall(receiver, args)
	case "AddMinutes":
		return g.emitDatetimeAddMinutesCall(receiver, args)
	case "AddSeconds":
		return g.emitDatetimeAddSecondsCall(receiver, args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; TDateTime.%s not implemented", r, method))
		return r, "i64", nil
	}
}

// emitDatetimeYearCall emits dt.Year() -> i64
func (g *Generator) emitDatetimeYearCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TDateTime.Year expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "Year", "Year", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_datetime_Year(ptr %s)", r, receiver))
	return r, "i64", nil
}

// emitDatetimeMonthCall emits dt.Month() -> i64
func (g *Generator) emitDatetimeMonthCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TDateTime.Month expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "Month", "Month", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_datetime_Month(ptr %s)", r, receiver))
	return r, "i64", nil
}

// emitDatetimeDayCall emits dt.Day() -> i64
func (g *Generator) emitDatetimeDayCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TDateTime.Day expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "Day", "Day", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_datetime_Day(ptr %s)", r, receiver))
	return r, "i64", nil
}

// emitDatetimeHourCall emits dt.Hour() -> i64
func (g *Generator) emitDatetimeHourCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TDateTime.Hour expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "Hour", "Hour", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_datetime_Hour(ptr %s)", r, receiver))
	return r, "i64", nil
}

// emitDatetimeMinuteCall emits dt.Minute() -> i64
func (g *Generator) emitDatetimeMinuteCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TDateTime.Minute expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "Minute", "Minute", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_datetime_Minute(ptr %s)", r, receiver))
	return r, "i64", nil
}

// emitDatetimeSecondCall emits dt.Second() -> i64
func (g *Generator) emitDatetimeSecondCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TDateTime.Second expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "Second", "Second", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_datetime_Second(ptr %s)", r, receiver))
	return r, "i64", nil
}

// emitDatetimeDayOfWeekCall emits dt.DayOfWeek() -> i64 (0=Sunday, 6=Saturday)
func (g *Generator) emitDatetimeDayOfWeekCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TDateTime.DayOfWeek expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "DayOfWeek", "DayOfWeek", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_datetime_DayOfWeek(ptr %s)", r, receiver))
	return r, "i64", nil
}

// emitDatetimeFormatDateCall emits dt.FormatDate() -> ptr (String)
func (g *Generator) emitDatetimeFormatDateCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TDateTime.FormatDate expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "FormatDate", "FormatDate", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_FormatDate(ptr %s)", r, receiver))
	return r, "ptr", nil
}

// emitDatetimeAddDaysCall emits dt.AddDays(n) -> ptr (new TDateTime)
func (g *Generator) emitDatetimeAddDaysCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TDateTime.AddDays expects 1 argument, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "AddDays", "AddDays", 0)
	daysReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_AddDays(ptr %s, i64 %s)", r, receiver, daysReg))
	return r, "TDateTime", nil
}

// emitDatetimeAddHoursCall emits dt.AddHours(n) -> ptr (new TDateTime)
func (g *Generator) emitDatetimeAddHoursCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TDateTime.AddHours expects 1 argument, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "AddHours", "AddHours", 0)
	hoursReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_AddHours(ptr %s, i64 %s)", r, receiver, hoursReg))
	return r, "TDateTime", nil
}

// emitDatetimeAddMinutesCall emits dt.AddMinutes(n) -> ptr (new TDateTime)
func (g *Generator) emitDatetimeAddMinutesCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TDateTime.AddMinutes expects 1 argument, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "AddMinutes", "AddMinutes", 0)
	minutesReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_AddMinutes(ptr %s, i64 %s)", r, receiver, minutesReg))
	return r, "TDateTime", nil
}

// emitDatetimeAddSecondsCall emits dt.AddSeconds(n) -> ptr (new TDateTime)
func (g *Generator) emitDatetimeAddSecondsCall(receiver string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TDateTime.AddSeconds expects 1 argument, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "AddSeconds", "AddSeconds", 0)
	secondsReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_AddSeconds(ptr %s, i64 %s)", r, receiver, secondsReg))
	return r, "TDateTime", nil
}

// Method body emitters (called by emitDatetimeBody via emitPendingStdlib)

// emitDatetimeYearBody emits @__kylix_datetime_Year(ptr %self) -> i64
func (g *Generator) emitDatetimeYearBody() {
	g.line("define i64 @__kylix_datetime_Year(ptr %self) {")
	g.line("entry:")
	// Allocate struct tm on stack
	tmBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmBuf))
	// localtime_r(&time_t, &tm)
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %%self, ptr %s)", tmPtr, tmBuf))
	// tm->tm_year (offset 20, i32)
	yearPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 20", yearPtr, tmBuf))
	yearI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", yearI32, yearPtr))
	// year = tm_year + 1900
	yearI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", yearI64, yearI32))
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1900", result, yearI64))
	g.line(fmt.Sprintf("  ret i64 %s", result))
	g.line("}")
	g.line("")
}

// emitDatetimeMonthBody emits @__kylix_datetime_Month(ptr %self) -> i64
func (g *Generator) emitDatetimeMonthBody() {
	g.line("define i64 @__kylix_datetime_Month(ptr %self) {")
	g.line("entry:")
	tmBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmBuf))
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %%self, ptr %s)", tmPtr, tmBuf))
	// tm->tm_mon (offset 16, i32)
	monPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 16", monPtr, tmBuf))
	monI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", monI32, monPtr))
	// month = tm_mon + 1 (tm_mon is 0-based)
	monI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", monI64, monI32))
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", result, monI64))
	g.line(fmt.Sprintf("  ret i64 %s", result))
	g.line("}")
	g.line("")
}

// emitDatetimeDayBody emits @__kylix_datetime_Day(ptr %self) -> i64
func (g *Generator) emitDatetimeDayBody() {
	g.line("define i64 @__kylix_datetime_Day(ptr %self) {")
	g.line("entry:")
	tmBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmBuf))
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %%self, ptr %s)", tmPtr, tmBuf))
	// tm->tm_mday (offset 12, i32)
	dayPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 12", dayPtr, tmBuf))
	dayI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", dayI32, dayPtr))
	dayI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", dayI64, dayI32))
	g.line(fmt.Sprintf("  ret i64 %s", dayI64))
	g.line("}")
	g.line("")
}

// emitDatetimeHourBody emits @__kylix_datetime_Hour(ptr %self) -> i64
func (g *Generator) emitDatetimeHourBody() {
	g.line("define i64 @__kylix_datetime_Hour(ptr %self) {")
	g.line("entry:")
	tmBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmBuf))
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %%self, ptr %s)", tmPtr, tmBuf))
	// tm->tm_hour (offset 8, i32)
	hourPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 8", hourPtr, tmBuf))
	hourI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", hourI32, hourPtr))
	hourI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", hourI64, hourI32))
	g.line(fmt.Sprintf("  ret i64 %s", hourI64))
	g.line("}")
	g.line("")
}

// emitDatetimeMinuteBody emits @__kylix_datetime_Minute(ptr %self) -> i64
func (g *Generator) emitDatetimeMinuteBody() {
	g.line("define i64 @__kylix_datetime_Minute(ptr %self) {")
	g.line("entry:")
	tmBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmBuf))
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %%self, ptr %s)", tmPtr, tmBuf))
	// tm->tm_min (offset 4, i32)
	minPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 4", minPtr, tmBuf))
	minI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", minI32, minPtr))
	minI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", minI64, minI32))
	g.line(fmt.Sprintf("  ret i64 %s", minI64))
	g.line("}")
	g.line("")
}

// emitDatetimeSecondBody emits @__kylix_datetime_Second(ptr %self) -> i64
func (g *Generator) emitDatetimeSecondBody() {
	g.line("define i64 @__kylix_datetime_Second(ptr %self) {")
	g.line("entry:")
	tmBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmBuf))
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %%self, ptr %s)", tmPtr, tmBuf))
	// tm->tm_sec (offset 0, i32)
	secPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 0", secPtr, tmBuf))
	secI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", secI32, secPtr))
	secI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", secI64, secI32))
	g.line(fmt.Sprintf("  ret i64 %s", secI64))
	g.line("}")
	g.line("")
}

// emitDatetimeDayOfWeekBody emits @__kylix_datetime_DayOfWeek(ptr %self) -> i64
func (g *Generator) emitDatetimeDayOfWeekBody() {
	g.line("define i64 @__kylix_datetime_DayOfWeek(ptr %self) {")
	g.line("entry:")
	tmBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmBuf))
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %%self, ptr %s)", tmPtr, tmBuf))
	// tm->tm_wday (offset 24, i32) - 0=Sunday, 6=Saturday
	wdayPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [56 x i8], ptr %s, i64 0, i64 24", wdayPtr, tmBuf))
	wdayI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", wdayI32, wdayPtr))
	wdayI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", wdayI64, wdayI32))
	g.line(fmt.Sprintf("  ret i64 %s", wdayI64))
	g.line("}")
	g.line("")
}

// emitDatetimeFormatDateBody emits @__kylix_datetime_FormatDate(ptr %self) -> ptr
func (g *Generator) emitDatetimeFormatDateBody() {
	g.line("define ptr @__kylix_datetime_FormatDate(ptr %self) {")
	g.line("entry:")
	// Allocate struct tm on stack
	tmBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [56 x i8], align 8", tmBuf))
	tmPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @localtime_r(ptr %%self, ptr %s)", tmPtr, tmBuf))
	// Allocate buffer for formatted string (20 bytes enough for "YYYY-MM-DD")
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 20)", buf))
	// Format string "%Y-%m-%d" (YYYY-MM-DD)
	fmtStr := g.addString("%Y-%m-%d")
	fmtPtr := g.ptrTo(fmtStr, 9)
	// strftime(buf, 20, "%Y-%m-%d", tm)
	g.line(fmt.Sprintf("  call i64 @strftime(ptr %s, i64 20, ptr %s, ptr %s)", buf, fmtPtr, tmBuf))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// emitDatetimeAddDaysBody emits @__kylix_datetime_AddDays(ptr %self, i64 %days) -> ptr
func (g *Generator) emitDatetimeAddDaysBody() {
	g.line("define ptr @__kylix_datetime_AddDays(ptr %self, i64 %days) {")
	g.line("entry:")
	// Load time_t from self
	timeVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%self", timeVal))
	// Add days * 86400 (seconds per day)
	daysInSec := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %%days, 86400", daysInSec))
	newTime := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", newTime, timeVal, daysInSec))
	// Allocate new TDateTime
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_arena_alloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newTime, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// emitDatetimeAddHoursBody emits @__kylix_datetime_AddHours(ptr %self, i64 %hours) -> ptr
func (g *Generator) emitDatetimeAddHoursBody() {
	g.line("define ptr @__kylix_datetime_AddHours(ptr %self, i64 %hours) {")
	g.line("entry:")
	timeVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%self", timeVal))
	// hours * 3600
	hoursInSec := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %%hours, 3600", hoursInSec))
	newTime := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", newTime, timeVal, hoursInSec))
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_arena_alloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newTime, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// emitDatetimeAddMinutesBody emits @__kylix_datetime_AddMinutes(ptr %self, i64 %minutes) -> ptr
func (g *Generator) emitDatetimeAddMinutesBody() {
	g.line("define ptr @__kylix_datetime_AddMinutes(ptr %self, i64 %minutes) {")
	g.line("entry:")
	timeVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%self", timeVal))
	// minutes * 60
	minInSec := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %%minutes, 60", minInSec))
	newTime := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", newTime, timeVal, minInSec))
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_arena_alloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newTime, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// emitDatetimeAddSecondsBody emits @__kylix_datetime_AddSeconds(ptr %self, i64 %seconds) -> ptr
func (g *Generator) emitDatetimeAddSecondsBody() {
	g.line("define ptr @__kylix_datetime_AddSeconds(ptr %self, i64 %seconds) {")
	g.line("entry:")
	timeVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%self", timeVal))
	// Direct add (no multiplication needed)
	newTime := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %%seconds", newTime, timeVal))
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_datetime_arena_alloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newTime, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// emitDatetimeBodyDispatch dispatches to the correct body emitter (called by emitPendingStdlib).
// This function is named to match the expected naming pattern in stdlib.go.
func (g *Generator) emitDatetimeBody(funcName string, argCount int) {
	switch funcName {
	case "Now":
		// Now() body inlined at call site, no separate define needed
	case "Today":
		g.emitDatetimeTodayBody()
	case "MakeDate":
		g.emitDatetimeMakeDateBody()
	case "ArenaAlloc":
		g.emitDatetimeArenaAllocBody()
	case "FreeArena":
		g.emitDatetimeFreeArenaBody()
	case "Year":
		g.emitDatetimeYearBody()
	case "Month":
		g.emitDatetimeMonthBody()
	case "Day":
		g.emitDatetimeDayBody()
	case "Hour":
		g.emitDatetimeHourBody()
	case "Minute":
		g.emitDatetimeMinuteBody()
	case "Second":
		g.emitDatetimeSecondBody()
	case "DayOfWeek":
		g.emitDatetimeDayOfWeekBody()
	case "FormatDate":
		g.emitDatetimeFormatDateBody()
	case "AddDays":
		g.emitDatetimeAddDaysBody()
	case "AddHours":
		g.emitDatetimeAddHoursBody()
	case "AddMinutes":
		g.emitDatetimeAddMinutesBody()
	case "AddSeconds":
		g.emitDatetimeAddSecondsBody()
	default:
		g.line(fmt.Sprintf("; ERROR: unsupported datetime function: %s", funcName))
	}
}

// emitDatetimeArenaAllocBody emits @__kylix_datetime_arena_alloc(i64 size) -> ptr
// Arena allocator for TDateTime instances. Allocates from 1MB global buffer.
// Returns null if out of space.
func (g *Generator) emitDatetimeArenaAllocBody() {
	g.line("define ptr @__kylix_datetime_arena_alloc(i64 %size) {")
	g.line("entry:")
	// Load current arena pointer
	currentPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_datetime_arena_ptr", currentPtr))

	// Calculate new pointer (current + size)
	newPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %%size", newPtr, currentPtr))

	// Calculate arena end (arena base + 1MB)
	arenaEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [1048576 x i8], ptr @__kylix_datetime_arena, i64 0, i64 1048576", arenaEnd))

	// Check if new pointer exceeds arena end
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ule ptr %s, %s", cmp, newPtr, arenaEnd))
	g.line(fmt.Sprintf("  br i1 %s, label %%ok, label %%fail", cmp))

	g.line("ok:")
	// Update arena pointer and return old pointer
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_datetime_arena_ptr", newPtr))
	g.line(fmt.Sprintf("  ret ptr %s", currentPtr))

	g.line("fail:")
	// Out of arena space, return null
	g.line("  ret ptr null")
	g.line("}")
	g.line("")
}

// emitDatetimeFreeArenaCall emits FreeArena() -> void (resets arena allocator)
func (g *Generator) emitDatetimeFreeArenaCall(args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("datetime.FreeArena expects 0 arguments, got %d", len(args))
	}
	g.enqueueStdlib("datetime", "FreeArena", "FreeArena", 0)

	// Call FreeArena (void return)
	g.line("  call void @__kylix_datetime_FreeArena()")

	// Return void (represented as i64 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0", r))
	return r, "i64", nil
}

// emitDatetimeFreeArenaBody emits @__kylix_datetime_FreeArena() -> void
// Resets arena pointer to arena start, effectively freeing all allocations.
func (g *Generator) emitDatetimeFreeArenaBody() {
	g.line("define void @__kylix_datetime_FreeArena() {")
	g.line("entry:")
	// Reset arena pointer to arena base
	g.line("  store ptr @__kylix_datetime_arena, ptr @__kylix_datetime_arena_ptr")
	g.line("  ret void")
	g.line("}")
	g.line("")
}

