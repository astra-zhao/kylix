; Kylix LLVM IR — module: RegexDemo
source_filename = "RegexDemo.klx"
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
  %t0 = getelementptr inbounds [17 x i8], ptr @.str.0, i64 0, i64 0, !dbg !5
  %t1 = call i1 @__kylix_regex_IsEmail(ptr %t0), !dbg !6
  br i1 %t1, label %lbl0, label %lbl1, !dbg !7
lbl0:
  %t2 = getelementptr inbounds [12 x i8], ptr @.str.1, i64 0, i64 0, !dbg !8
  %t3 = call i32 @puts(ptr noundef %t2), !dbg !9
  br label %lbl1, !dbg !7
lbl1:
  %t4 = getelementptr inbounds [10 x i8], ptr @.str.2, i64 0, i64 0, !dbg !10
  %t5 = call i1 @__kylix_regex_IsEmail(ptr %t4), !dbg !11
  %t6 = xor i1 %t5, 1, !dbg !12
  br i1 %t6, label %lbl2, label %lbl3, !dbg !13
lbl2:
  %t7 = getelementptr inbounds [23 x i8], ptr @.str.3, i64 0, i64 0, !dbg !14
  %t8 = call i32 @puts(ptr noundef %t7), !dbg !15
  br label %lbl3, !dbg !13
lbl3:
  %t9 = getelementptr inbounds [18 x i8], ptr @.str.4, i64 0, i64 0, !dbg !16
  %t10 = call i1 @__kylix_regex_IsURL(ptr %t9), !dbg !17
  br i1 %t10, label %lbl4, label %lbl5, !dbg !18
lbl4:
  %t11 = getelementptr inbounds [10 x i8], ptr @.str.5, i64 0, i64 0, !dbg !19
  %t12 = call i32 @puts(ptr noundef %t11), !dbg !20
  br label %lbl5, !dbg !18
lbl5:
  %t13 = getelementptr inbounds [6 x i8], ptr @.str.6, i64 0, i64 0, !dbg !21
  %t14 = call i1 @__kylix_regex_IsNumeric(ptr %t13), !dbg !22
  br i1 %t14, label %lbl6, label %lbl7, !dbg !23
lbl6:
  %t15 = getelementptr inbounds [21 x i8], ptr @.str.7, i64 0, i64 0, !dbg !24
  %t16 = call i32 @puts(ptr noundef %t15), !dbg !25
  br label %lbl7, !dbg !23
lbl7:
  %t17 = getelementptr inbounds [7 x i8], ptr @.str.8, i64 0, i64 0, !dbg !26
  %t18 = call i1 @__kylix_regex_IsNumeric(ptr %t17), !dbg !27
  %t19 = xor i1 %t18, 1, !dbg !28
  br i1 %t19, label %lbl8, label %lbl9, !dbg !29
lbl8:
  %t20 = getelementptr inbounds [28 x i8], ptr @.str.9, i64 0, i64 0, !dbg !30
  %t21 = call i32 @puts(ptr noundef %t20), !dbg !31
  br label %lbl9, !dbg !29
lbl9:
  %t22 = getelementptr inbounds [6 x i8], ptr @.str.10, i64 0, i64 0, !dbg !32
  %t23 = call i1 @__kylix_regex_IsAlpha(ptr %t22), !dbg !33
  br i1 %t23, label %lbl10, label %lbl11, !dbg !34
lbl10:
  %t24 = getelementptr inbounds [24 x i8], ptr @.str.11, i64 0, i64 0, !dbg !35
  %t25 = call i32 @puts(ptr noundef %t24), !dbg !36
  br label %lbl11, !dbg !34
lbl11:
  ret i32 0
}

define i1 @__kylix_regex_IsEmail(ptr %str) {
entry:
  %t26 = alloca [64 x i8], align 8
  %t27 = getelementptr inbounds [51 x i8], ptr @.str.12, i64 0, i64 0
  %t28 = call i32 @regcomp(ptr %t26, ptr %t27, i32 9)
  %t29 = icmp eq i32 %t28, 0
  br i1 %t29, label %lbl12, label %lbl13
lbl13:
  ret i1 false
lbl12:
  %t30 = call i32 @regexec(ptr %t26, ptr %str, i64 0, ptr null, i32 0)
  call void @regfree(ptr %t26)
  %t31 = icmp eq i32 %t30, 0
  ret i1 %t31
}

define i1 @__kylix_regex_IsURL(ptr %str) {
entry:
  %t32 = alloca [64 x i8], align 8
  %t33 = getelementptr inbounds [43 x i8], ptr @.str.13, i64 0, i64 0
  %t34 = call i32 @regcomp(ptr %t32, ptr %t33, i32 9)
  %t35 = icmp eq i32 %t34, 0
  br i1 %t35, label %lbl14, label %lbl15
lbl15:
  ret i1 false
lbl14:
  %t36 = call i32 @regexec(ptr %t32, ptr %str, i64 0, ptr null, i32 0)
  call void @regfree(ptr %t32)
  %t37 = icmp eq i32 %t36, 0
  ret i1 %t37
}

define i1 @__kylix_regex_IsNumeric(ptr %str) {
entry:
  %t38 = alloca [64 x i8], align 8
  %t39 = getelementptr inbounds [9 x i8], ptr @.str.14, i64 0, i64 0
  %t40 = call i32 @regcomp(ptr %t38, ptr %t39, i32 9)
  %t41 = icmp eq i32 %t40, 0
  br i1 %t41, label %lbl16, label %lbl17
lbl17:
  ret i1 false
lbl16:
  %t42 = call i32 @regexec(ptr %t38, ptr %str, i64 0, ptr null, i32 0)
  call void @regfree(ptr %t38)
  %t43 = icmp eq i32 %t42, 0
  ret i1 %t43
}

define i1 @__kylix_regex_IsAlpha(ptr %str) {
entry:
  %t44 = alloca [64 x i8], align 8
  %t45 = getelementptr inbounds [12 x i8], ptr @.str.15, i64 0, i64 0
  %t46 = call i32 @regcomp(ptr %t44, ptr %t45, i32 9)
  %t47 = icmp eq i32 %t46, 0
  br i1 %t47, label %lbl18, label %lbl19
lbl19:
  ret i1 false
lbl18:
  %t48 = call i32 @regexec(ptr %t44, ptr %str, i64 0, ptr null, i32 0)
  call void @regfree(ptr %t44)
  %t49 = icmp eq i32 %t48, 0
  ret i1 %t49
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [17 x i8] c"user@example.com\00", align 1
@.str.1 = private unnamed_addr constant [12 x i8] c"Valid email\00", align 1
@.str.2 = private unnamed_addr constant [10 x i8] c"bad-email\00", align 1
@.str.3 = private unnamed_addr constant [23 x i8] c"Invalid email detected\00", align 1
@.str.4 = private unnamed_addr constant [18 x i8] c"https://kylix.top\00", align 1
@.str.5 = private unnamed_addr constant [10 x i8] c"Valid URL\00", align 1
@.str.6 = private unnamed_addr constant [6 x i8] c"12345\00", align 1
@.str.7 = private unnamed_addr constant [21 x i8] c"Valid numeric string\00", align 1
@.str.8 = private unnamed_addr constant [7 x i8] c"abc123\00", align 1
@.str.9 = private unnamed_addr constant [28 x i8] c"Non-numeric string detected\00", align 1
@.str.10 = private unnamed_addr constant [6 x i8] c"hello\00", align 1
@.str.11 = private unnamed_addr constant [24 x i8] c"Valid alphabetic string\00", align 1
@.str.12 = private unnamed_addr constant [51 x i8] c"^[a-zA-Z0-9._%+\5C-]+@[a-zA-Z0-9.\5C-]+\5C.[a-zA-Z]{2,}$\00", align 1
@.str.13 = private unnamed_addr constant [43 x i8] c"^https?://[^[:space:]/$.?#].[^[:space:]]*$\00", align 1
@.str.14 = private unnamed_addr constant [9 x i8] c"^[0-9]+$\00", align 1
@.str.15 = private unnamed_addr constant [12 x i8] c"^[a-zA-Z]+$\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example39_regex.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/08_stdlib_utils")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 6, type: !38, scopeLine: 6, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !41)
!5 = !DILocation(line: 10, column: 14, scope: !4)
!6 = !DILocation(line: 10, column: 13, scope: !4)
!7 = !DILocation(line: 10, column: 3, scope: !4)
!8 = !DILocation(line: 11, column: 13, scope: !4)
!9 = !DILocation(line: 11, column: 12, scope: !4)
!10 = !DILocation(line: 13, column: 18, scope: !4)
!11 = !DILocation(line: 13, column: 17, scope: !4)
!12 = !DILocation(line: 13, column: 6, scope: !4)
!13 = !DILocation(line: 13, column: 3, scope: !4)
!14 = !DILocation(line: 14, column: 13, scope: !4)
!15 = !DILocation(line: 14, column: 12, scope: !4)
!16 = !DILocation(line: 16, column: 12, scope: !4)
!17 = !DILocation(line: 16, column: 11, scope: !4)
!18 = !DILocation(line: 16, column: 3, scope: !4)
!19 = !DILocation(line: 17, column: 13, scope: !4)
!20 = !DILocation(line: 17, column: 12, scope: !4)
!21 = !DILocation(line: 19, column: 16, scope: !4)
!22 = !DILocation(line: 19, column: 15, scope: !4)
!23 = !DILocation(line: 19, column: 3, scope: !4)
!24 = !DILocation(line: 20, column: 13, scope: !4)
!25 = !DILocation(line: 20, column: 12, scope: !4)
!26 = !DILocation(line: 22, column: 20, scope: !4)
!27 = !DILocation(line: 22, column: 19, scope: !4)
!28 = !DILocation(line: 22, column: 6, scope: !4)
!29 = !DILocation(line: 22, column: 3, scope: !4)
!30 = !DILocation(line: 23, column: 13, scope: !4)
!31 = !DILocation(line: 23, column: 12, scope: !4)
!32 = !DILocation(line: 25, column: 14, scope: !4)
!33 = !DILocation(line: 25, column: 13, scope: !4)
!34 = !DILocation(line: 25, column: 3, scope: !4)
!35 = !DILocation(line: 26, column: 13, scope: !4)
!36 = !DILocation(line: 26, column: 12, scope: !4)
!37 = !{null}
!38 = !DISubroutineType(types: !37)
!39 = !{}
!40 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!41 = !{}
