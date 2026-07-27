; Kylix LLVM IR — module: DatetimeDemo
source_filename = "DatetimeDemo.klx"
target datalayout = "e-m:o-i64:64-i128:128-n32:64-S128"
target triple = "arm64-apple-macosx15.0.0"

; ===== Runtime declarations (libc) =====
declare i32 @printf(ptr noundef, ...)
declare i32 @puts(ptr noundef)
declare ptr @malloc(i64 noundef)
declare void @free(ptr noundef)
declare i64 @strlen(ptr noundef)
declare ptr @strcpy(ptr noundef, ptr noundef)
declare ptr @strcat(ptr noundef, ptr noundef)
declare i32 @strcmp(ptr noundef, ptr noundef)
declare ptr @memcpy(ptr noundef, ptr noundef, i64 noundef)
declare i64 @atoll(ptr noundef)
declare i32 @snprintf(ptr noundef, i64 noundef, ptr noundef, ...)
declare double @strtod(ptr noundef, ptr noundef)
; ===== Exception handling runtime (setjmp/longjmp) =====
declare i32 @setjmp(ptr)
declare void @longjmp(ptr, i32)
declare void @exit(i32)
; ===== File I/O (libc, used by stdlib sysutil) =====
declare ptr @fopen(ptr noundef, ptr noundef)
declare i32 @fclose(ptr noundef)
declare i64 @fread(ptr noundef, i64 noundef, i64 noundef, ptr noundef)
declare i64 @fwrite(ptr noundef, i64 noundef, i64 noundef, ptr noundef)
declare i32 @fputs(ptr noundef, ptr noundef)
declare i32 @fseek(ptr noundef, i64 noundef, i32 noundef)
declare i64 @ftell(ptr noundef)
declare i32 @access(ptr noundef, i32 noundef)
; ===== BSD sockets (used by stdlib net) =====
declare i32 @socket(i32 noundef, i32 noundef, i32 noundef)
declare i32 @connect(i32 noundef, ptr noundef, i32 noundef)
declare i32 @bind(i32 noundef, ptr noundef, i32 noundef)
declare i32 @listen(i32 noundef, i32 noundef)
declare i32 @accept(i32 noundef, ptr, ptr)
declare i64 @send(i32 noundef, ptr noundef, i64 noundef, i32 noundef)
declare i64 @recv(i32 noundef, ptr noundef, i64 noundef, i32 noundef)
declare i32 @close(i32 noundef)
declare i32 @setsockopt(i32 noundef, i32 noundef, i32 noundef, ptr noundef, i32 noundef)
declare i32 @inet_pton(i32 noundef, ptr noundef, ptr noundef)
; ===== OpenSSL libcrypto (used by stdlib crypto) =====
declare ptr @SHA256(ptr noundef, i64 noundef, ptr noundef)
declare ptr @MD5(ptr noundef, i64 noundef, ptr noundef)
declare ptr @strncpy(ptr noundef, ptr noundef, i64 noundef)
; EVP_CIPHER API for AES-256-CBC (v4.5.0 stdlib Phase 3)
declare ptr @EVP_CIPHER_CTX_new()
declare void @EVP_CIPHER_CTX_free(ptr noundef)
declare ptr @EVP_aes_256_cbc()
declare i32 @EVP_EncryptInit_ex(ptr noundef, ptr noundef, ptr noundef, ptr noundef, ptr noundef)
declare i32 @EVP_EncryptUpdate(ptr noundef, ptr noundef, ptr noundef, ptr noundef, i32 noundef)
declare i32 @EVP_EncryptFinal_ex(ptr noundef, ptr noundef, ptr noundef)
declare i32 @EVP_DecryptInit_ex(ptr noundef, ptr noundef, ptr noundef, ptr noundef, ptr noundef)
declare i32 @EVP_DecryptUpdate(ptr noundef, ptr noundef, ptr noundef, ptr noundef, i32 noundef)
declare i32 @EVP_DecryptFinal_ex(ptr noundef, ptr noundef, ptr noundef)
declare i32 @EVP_CIPHER_CTX_block_size(ptr noundef)
declare i32 @RAND_bytes(ptr noundef, i32 noundef)
declare ptr @EVP_sha256()
declare i32 @PKCS5_PBKDF2_HMAC(ptr noundef, i32 noundef, ptr noundef, i32 noundef, i64 noundef, ptr noundef, i32 noundef, ptr noundef)
declare i32 @sscanf(ptr noundef, ptr noundef, ...)
; ===== SQLite (used by stdlib db) =====
declare i32 @sqlite3_open(ptr noundef, ptr noundef)
declare i32 @sqlite3_close(ptr noundef)
declare i32 @sqlite3_prepare_v2(ptr noundef, ptr noundef, i32 noundef, ptr noundef, ptr noundef)
declare i32 @sqlite3_bind_text(ptr noundef, i32 noundef, ptr noundef, i32 noundef, i64 noundef)
declare i32 @sqlite3_bind_int64(ptr noundef, i32 noundef, i64 noundef)
declare i32 @sqlite3_step(ptr noundef)
declare ptr @sqlite3_column_text(ptr noundef, i32 noundef)
declare i32 @sqlite3_finalize(ptr noundef)
; ===== libcurl (used by stdlib httpclient, v4.5.0 Phase 3) =====
declare ptr @curl_easy_init()
declare i32 @curl_easy_setopt(ptr noundef, i32 noundef, ...)
declare i32 @curl_easy_perform(ptr noundef)
declare void @curl_easy_cleanup(ptr noundef)
declare ptr @curl_slist_append(ptr noundef, ptr noundef)
declare void @curl_slist_free_all(ptr noundef)
; ===== POSIX regex (used by stdlib regex) =====
declare i32 @regcomp(ptr noundef, ptr noundef, i32 noundef)
declare i32 @regexec(ptr noundef, ptr noundef, i64 noundef, ptr, i32 noundef)
declare void @regfree(ptr noundef)
; ===== time.h (used by stdlib datetime) =====
declare i64 @time(ptr)
declare ptr @localtime(ptr)
declare ptr @localtime_r(ptr, ptr)
declare i64 @mktime(ptr)
declare i64 @strftime(ptr, i64, ptr, ptr)
; ===== LLVM intrinsics =====
declare void @llvm.memset.p0.i64(ptr nocapture writeonly, i8, i64, i1 immarg)
declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg)

; ===== datetime arena allocator =====
@__kylix_datetime_arena = internal global [1048576 x i8] zeroinitializer, align 8
@__kylix_datetime_arena_ptr = internal global ptr @__kylix_datetime_arena, align 8

%Exception = type { ptr, ptr }
; ===== Exception handling globals =====
@__kylix_exc_obj = global ptr null
@__kylix_exc_type = global i32 0
@__kylix_exc_active = global i1 false
@__kylix_jmpbuf = global ptr null

; ===== Exception subtype table =====
%__kylix_edge = type { i32, i32 }
@__kylix_exctab = constant [0 x %__kylix_edge] zeroinitializer

define i1 @__kylix_is_subtype(i32 %child, i32 %parent) {
entry:
  %eq = icmp eq i32 %child, %parent
  br i1 %eq, label %ret_true, label %loop
loop:
  %c = phi i32 [ %child, %entry ], [ %c_next, %loop_next ]
  %i = phi i64 [ 0, %entry ], [ %i_next, %loop_next ]
  %oob = icmp eq i64 %i, 0
  br i1 %oob, label %ret_false, label %body
body:
  %slot = getelementptr inbounds [0 x %__kylix_edge], ptr @__kylix_exctab, i64 0, i64 %i
  %cid_ptr = getelementptr inbounds %__kylix_edge, ptr %slot, i64 0, i32 0
  %cid = load i32, ptr %cid_ptr
  %iscur = icmp eq i32 %cid, %c
  br i1 %iscur, label %found, label %loop_next
found:
  %pid_ptr = getelementptr inbounds %__kylix_edge, ptr %slot, i64 0, i32 1
  %par = load i32, ptr %pid_ptr
  %match = icmp eq i32 %par, %parent
  br i1 %match, label %ret_true, label %update
update:
  br label %loop_next
loop_next:
  %c_next = phi i32 [ %c, %body ], [ %par, %update ]
  %i_next = add i64 %i, 1
  br label %loop
ret_true:
  ret i1 true
ret_false:
  ret i1 false
}

; ===== Entry point =====
define i32 @main() !dbg !4 {
entry:
  %t0 = call ptr @__kylix_datetime_arena_alloc(i64 8), !dbg !5
  %t1 = call i64 @time(ptr null), !dbg !5
  store i64 %t1, ptr %t0, !dbg !5
  %v_now_str = alloca ptr, align 8, !dbg !6
  store ptr %t0, ptr %v_now_str, !dbg !6
  %t2 = getelementptr inbounds [15 x i8], ptr @.str.0, i64 0, i64 0, !dbg !7
  %t3 = load ptr, ptr %v_now_str, !dbg !8
  %t4 = call i64 @__kylix_datetime_Year(ptr %t3), !dbg !9
  %t5 = alloca [24 x i8], align 1, !dbg !10
  %t6 = getelementptr inbounds [24 x i8], ptr %t5, i64 0, i64 0, !dbg !10
  %t7 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0, !dbg !10
  %t8 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t6, i64 24, ptr noundef %t7, i64 %t4), !dbg !10
  %t9 = call ptr @malloc(i64 512), !dbg !11
  call ptr @strcpy(ptr %t9, ptr %t2), !dbg !11
  call ptr @strcat(ptr %t9, ptr %t6), !dbg !11
  %t10 = call i32 @puts(ptr noundef %t9), !dbg !12
  %t11 = getelementptr inbounds [16 x i8], ptr @.str.2, i64 0, i64 0, !dbg !13
  %t12 = load ptr, ptr %v_now_str, !dbg !14
  %t13 = call i64 @__kylix_datetime_Month(ptr %t12), !dbg !15
  %t14 = alloca [24 x i8], align 1, !dbg !16
  %t15 = getelementptr inbounds [24 x i8], ptr %t14, i64 0, i64 0, !dbg !16
  %t16 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0, !dbg !16
  %t17 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t15, i64 24, ptr noundef %t16, i64 %t13), !dbg !16
  %t18 = call ptr @malloc(i64 512), !dbg !17
  call ptr @strcpy(ptr %t18, ptr %t11), !dbg !17
  call ptr @strcat(ptr %t18, ptr %t15), !dbg !17
  %t19 = call i32 @puts(ptr noundef %t18), !dbg !18
  %t20 = getelementptr inbounds [14 x i8], ptr @.str.3, i64 0, i64 0, !dbg !19
  %t21 = load ptr, ptr %v_now_str, !dbg !20
  %t22 = call i64 @__kylix_datetime_Day(ptr %t21), !dbg !21
  %t23 = alloca [24 x i8], align 1, !dbg !22
  %t24 = getelementptr inbounds [24 x i8], ptr %t23, i64 0, i64 0, !dbg !22
  %t25 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0, !dbg !22
  %t26 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t24, i64 24, ptr noundef %t25, i64 %t22), !dbg !22
  %t27 = call ptr @malloc(i64 512), !dbg !23
  call ptr @strcpy(ptr %t27, ptr %t20), !dbg !23
  call ptr @strcat(ptr %t27, ptr %t24), !dbg !23
  %t28 = call i32 @puts(ptr noundef %t27), !dbg !24
  %t29 = add i64 0, 2024, !dbg !25
  %t30 = add i64 0, 12, !dbg !26
  %t31 = add i64 0, 25, !dbg !27
  %t32 = call ptr @__kylix_datetime_MakeDate(i64 %t29, i64 %t30, i64 %t31), !dbg !28
  %v_custom_str = alloca ptr, align 8, !dbg !29
  store ptr %t32, ptr %v_custom_str, !dbg !29
  %t33 = getelementptr inbounds [17 x i8], ptr @.str.4, i64 0, i64 0, !dbg !30
  %t34 = load ptr, ptr %v_custom_str, !dbg !31
  %t35 = call ptr @__kylix_datetime_FormatDate(ptr %t34), !dbg !32
  %t36 = call ptr @malloc(i64 512), !dbg !33
  call ptr @strcpy(ptr %t36, ptr %t33), !dbg !33
  call ptr @strcat(ptr %t36, ptr %t35), !dbg !33
  %t37 = call i32 @puts(ptr noundef %t36), !dbg !34
  %t38 = load ptr, ptr %v_custom_str, !dbg !35
  %t39 = add i64 0, 7, !dbg !36
  %t40 = call ptr @__kylix_datetime_AddDays(ptr %t38, i64 %t39), !dbg !37
  %v_future_str = alloca ptr, align 8, !dbg !38
  store ptr %t40, ptr %v_future_str, !dbg !38
  %t41 = getelementptr inbounds [17 x i8], ptr @.str.5, i64 0, i64 0, !dbg !39
  %t42 = load ptr, ptr %v_future_str, !dbg !40
  %t43 = call ptr @__kylix_datetime_FormatDate(ptr %t42), !dbg !41
  %t44 = call ptr @malloc(i64 512), !dbg !42
  call ptr @strcpy(ptr %t44, ptr %t41), !dbg !42
  call ptr @strcat(ptr %t44, ptr %t43), !dbg !42
  %t45 = call i32 @puts(ptr noundef %t44), !dbg !43
  %t46 = load ptr, ptr %v_custom_str, !dbg !44
  %t47 = add i64 0, 7, !dbg !45
  %t48 = sub i64 0, %t47, !dbg !46
  %t49 = call ptr @__kylix_datetime_AddDays(ptr %t46, i64 %t48), !dbg !47
  %v_past_str = alloca ptr, align 8, !dbg !48
  store ptr %t49, ptr %v_past_str, !dbg !48
  %t50 = getelementptr inbounds [18 x i8], ptr @.str.6, i64 0, i64 0, !dbg !49
  %t51 = load ptr, ptr %v_past_str, !dbg !50
  %t52 = call ptr @__kylix_datetime_FormatDate(ptr %t51), !dbg !51
  %t53 = call ptr @malloc(i64 512), !dbg !52
  call ptr @strcpy(ptr %t53, ptr %t50), !dbg !52
  call ptr @strcat(ptr %t53, ptr %t52), !dbg !52
  %t54 = call i32 @puts(ptr noundef %t53), !dbg !53
  ret i32 0
}

define ptr @__kylix_datetime_arena_alloc(i64 %size) {
entry:
  %t55 = load ptr, ptr @__kylix_datetime_arena_ptr
  %t56 = getelementptr inbounds i8, ptr %t55, i64 %size
  %t57 = getelementptr inbounds [1048576 x i8], ptr @__kylix_datetime_arena, i64 0, i64 1048576
  %t58 = icmp ule ptr %t56, %t57
  br i1 %t58, label %ok, label %fail
ok:
  store ptr %t56, ptr @__kylix_datetime_arena_ptr
  ret ptr %t55
fail:
  ret ptr null
}

define i64 @__kylix_datetime_Year(ptr %self) {
entry:
  %t59 = alloca [56 x i8], align 8
  %t60 = call ptr @localtime_r(ptr %self, ptr %t59)
  %t61 = getelementptr inbounds [56 x i8], ptr %t59, i64 0, i64 20
  %t62 = load i32, ptr %t61
  %t63 = sext i32 %t62 to i64
  %t64 = add i64 %t63, 1900
  ret i64 %t64
}

define i64 @__kylix_datetime_Month(ptr %self) {
entry:
  %t65 = alloca [56 x i8], align 8
  %t66 = call ptr @localtime_r(ptr %self, ptr %t65)
  %t67 = getelementptr inbounds [56 x i8], ptr %t65, i64 0, i64 16
  %t68 = load i32, ptr %t67
  %t69 = sext i32 %t68 to i64
  %t70 = add i64 %t69, 1
  ret i64 %t70
}

define i64 @__kylix_datetime_Day(ptr %self) {
entry:
  %t71 = alloca [56 x i8], align 8
  %t72 = call ptr @localtime_r(ptr %self, ptr %t71)
  %t73 = getelementptr inbounds [56 x i8], ptr %t71, i64 0, i64 12
  %t74 = load i32, ptr %t73
  %t75 = sext i32 %t74 to i64
  ret i64 %t75
}

define ptr @__kylix_datetime_MakeDate(i64 %year, i64 %month, i64 %day) {
entry:
  %t76 = alloca [56 x i8], align 8
  call void @llvm.memset.p0.i64(ptr %t76, i8 0, i64 56, i1 false)
  %t77 = sub i64 %year, 1900
  %t78 = trunc i64 %t77 to i32
  %t79 = getelementptr inbounds [56 x i8], ptr %t76, i64 0, i64 20
  store i32 %t78, ptr %t79
  %t80 = sub i64 %month, 1
  %t81 = trunc i64 %t80 to i32
  %t82 = getelementptr inbounds [56 x i8], ptr %t76, i64 0, i64 16
  store i32 %t81, ptr %t82
  %t83 = trunc i64 %day to i32
  %t84 = getelementptr inbounds [56 x i8], ptr %t76, i64 0, i64 12
  store i32 %t83, ptr %t84
  %t85 = call i64 @mktime(ptr %t76)
  %t86 = call ptr @__kylix_datetime_arena_alloc(i64 8)
  store i64 %t85, ptr %t86
  ret ptr %t86
}

define ptr @__kylix_datetime_FormatDate(ptr %self) {
entry:
  %t87 = alloca [56 x i8], align 8
  %t88 = call ptr @localtime_r(ptr %self, ptr %t87)
  %t89 = call ptr @malloc(i64 20)
  %t90 = getelementptr inbounds [9 x i8], ptr @.str.7, i64 0, i64 0
  call i64 @strftime(ptr %t89, i64 20, ptr %t90, ptr %t87)
  ret ptr %t89
}

define ptr @__kylix_datetime_AddDays(ptr %self, i64 %days) {
entry:
  %t91 = load i64, ptr %self
  %t92 = mul i64 %days, 86400
  %t93 = add i64 %t91, %t92
  %t94 = call ptr @__kylix_datetime_arena_alloc(i64 8)
  store i64 %t93, ptr %t94
  ret ptr %t94
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [15 x i8] c"Current year: \00", align 1
@.str.1 = private unnamed_addr constant [5 x i8] c"%lld\00", align 1
@.str.2 = private unnamed_addr constant [16 x i8] c"Current month: \00", align 1
@.str.3 = private unnamed_addr constant [14 x i8] c"Current day: \00", align 1
@.str.4 = private unnamed_addr constant [17 x i8] c"Christmas 2024: \00", align 1
@.str.5 = private unnamed_addr constant [17 x i8] c"One week later: \00", align 1
@.str.6 = private unnamed_addr constant [18 x i8] c"One week before: \00", align 1
@.str.7 = private unnamed_addr constant [9 x i8] c"%Y-%m-%d\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example38_datetime.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/08_stdlib_utils")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 8, type: !55, scopeLine: 8, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !58)
!5 = !DILocation(line: 12, column: 17, scope: !4)
!6 = !DILocation(line: 12, column: 3, scope: !4)
!7 = !DILocation(line: 13, column: 11, scope: !4)
!8 = !DILocation(line: 13, column: 39, scope: !4)
!9 = !DILocation(line: 13, column: 47, scope: !4)
!10 = !DILocation(line: 13, column: 38, scope: !4)
!11 = !DILocation(line: 13, column: 28, scope: !4)
!12 = !DILocation(line: 13, column: 10, scope: !4)
!13 = !DILocation(line: 14, column: 11, scope: !4)
!14 = !DILocation(line: 14, column: 40, scope: !4)
!15 = !DILocation(line: 14, column: 49, scope: !4)
!16 = !DILocation(line: 14, column: 39, scope: !4)
!17 = !DILocation(line: 14, column: 29, scope: !4)
!18 = !DILocation(line: 14, column: 10, scope: !4)
!19 = !DILocation(line: 15, column: 11, scope: !4)
!20 = !DILocation(line: 15, column: 38, scope: !4)
!21 = !DILocation(line: 15, column: 45, scope: !4)
!22 = !DILocation(line: 15, column: 37, scope: !4)
!23 = !DILocation(line: 15, column: 27, scope: !4)
!24 = !DILocation(line: 15, column: 10, scope: !4)
!25 = !DILocation(line: 17, column: 26, scope: !4)
!26 = !DILocation(line: 17, column: 32, scope: !4)
!27 = !DILocation(line: 17, column: 36, scope: !4)
!28 = !DILocation(line: 17, column: 25, scope: !4)
!29 = !DILocation(line: 17, column: 3, scope: !4)
!30 = !DILocation(line: 18, column: 11, scope: !4)
!31 = !DILocation(line: 18, column: 32, scope: !4)
!32 = !DILocation(line: 18, column: 49, scope: !4)
!33 = !DILocation(line: 18, column: 30, scope: !4)
!34 = !DILocation(line: 18, column: 10, scope: !4)
!35 = !DILocation(line: 20, column: 17, scope: !4)
!36 = !DILocation(line: 20, column: 32, scope: !4)
!37 = !DILocation(line: 20, column: 31, scope: !4)
!38 = !DILocation(line: 20, column: 3, scope: !4)
!39 = !DILocation(line: 21, column: 11, scope: !4)
!40 = !DILocation(line: 21, column: 32, scope: !4)
!41 = !DILocation(line: 21, column: 49, scope: !4)
!42 = !DILocation(line: 21, column: 30, scope: !4)
!43 = !DILocation(line: 21, column: 10, scope: !4)
!44 = !DILocation(line: 23, column: 15, scope: !4)
!45 = !DILocation(line: 23, column: 31, scope: !4)
!46 = !DILocation(line: 23, column: 30, scope: !4)
!47 = !DILocation(line: 23, column: 29, scope: !4)
!48 = !DILocation(line: 23, column: 3, scope: !4)
!49 = !DILocation(line: 24, column: 11, scope: !4)
!50 = !DILocation(line: 24, column: 33, scope: !4)
!51 = !DILocation(line: 24, column: 48, scope: !4)
!52 = !DILocation(line: 24, column: 31, scope: !4)
!53 = !DILocation(line: 24, column: 10, scope: !4)
!54 = !{null}
!55 = !DISubroutineType(types: !54)
!56 = !{}
!57 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!58 = !{}
