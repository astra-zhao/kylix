; Kylix LLVM IR — module: CaseStatement
source_filename = "CaseStatement.klx"
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
  %v_day_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_day_int, !dbg !5
  #dbg_declare(ptr %v_day_int, !6, !DIExpression(), !5)
  %v_grade_str = alloca ptr, align 8, !dbg !5
  store ptr null, ptr %v_grade_str, !dbg !5
  #dbg_declare(ptr %v_grade_str, !7, !DIExpression(), !5)
  %v_month_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_month_int, !dbg !5
  #dbg_declare(ptr %v_month_int, !8, !DIExpression(), !5)
  %t0 = add i64 0, 3, !dbg !9
  store i64 %t0, ptr %v_day_int, !dbg !10
  %t1 = call ptr @malloc(i64 512), !dbg !11
  store i8 0, ptr %t1, !dbg !11
  %t2 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !11
  %t3 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0, !dbg !12
  %t4 = call ptr @strcat(ptr %t1, ptr %t3), !dbg !11
  %t5 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !11
  %t6 = call ptr @strcat(ptr %t1, ptr %t5), !dbg !11
  %t7 = load i64, ptr %v_day_int, !dbg !13
  %t8 = call i64 @strlen(ptr %t1), !dbg !11
  %t9 = getelementptr inbounds i8, ptr %t1, i64 %t8, !dbg !11
  %t10 = sub i64 512, %t8, !dbg !11
  %t11 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t9, i64 %t10, ptr %t2, i64 %t7), !dbg !11
  %t12 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !11
  %t13 = call ptr @strcat(ptr %t1, ptr %t12), !dbg !11
  %t14 = getelementptr inbounds [6 x i8], ptr @.str.3, i64 0, i64 0, !dbg !14
  %t15 = call ptr @strcat(ptr %t1, ptr %t14), !dbg !11
  %t16 = call i32 @puts(ptr noundef %t1), !dbg !11
  %t17 = load i64, ptr %v_day_int, !dbg !15
  switch i64 %t17, label %lbl0 [ i64 1, label %lbl2 i64 2, label %lbl3 i64 3, label %lbl4 i64 4, label %lbl5 i64 5, label %lbl6 i64 6, label %lbl7 i64 7, label %lbl8 ], !dbg !16
lbl2:
  %t18 = getelementptr inbounds [7 x i8], ptr @.str.4, i64 0, i64 0, !dbg !17
  %t19 = call i32 @puts(ptr noundef %t18), !dbg !18
  br label %lbl1, !dbg !16
lbl3:
  %t20 = getelementptr inbounds [8 x i8], ptr @.str.5, i64 0, i64 0, !dbg !19
  %t21 = call i32 @puts(ptr noundef %t20), !dbg !20
  br label %lbl1, !dbg !16
lbl4:
  %t22 = getelementptr inbounds [10 x i8], ptr @.str.6, i64 0, i64 0, !dbg !21
  %t23 = call i32 @puts(ptr noundef %t22), !dbg !22
  br label %lbl1, !dbg !16
lbl5:
  %t24 = getelementptr inbounds [9 x i8], ptr @.str.7, i64 0, i64 0, !dbg !23
  %t25 = call i32 @puts(ptr noundef %t24), !dbg !24
  br label %lbl1, !dbg !16
lbl6:
  %t26 = getelementptr inbounds [7 x i8], ptr @.str.8, i64 0, i64 0, !dbg !25
  %t27 = call i32 @puts(ptr noundef %t26), !dbg !26
  br label %lbl1, !dbg !16
lbl7:
  %t28 = getelementptr inbounds [9 x i8], ptr @.str.9, i64 0, i64 0, !dbg !27
  %t29 = call i32 @puts(ptr noundef %t28), !dbg !28
  br label %lbl1, !dbg !16
lbl8:
  %t30 = getelementptr inbounds [7 x i8], ptr @.str.10, i64 0, i64 0, !dbg !29
  %t31 = call i32 @puts(ptr noundef %t30), !dbg !30
  br label %lbl1, !dbg !16
lbl0:
  br label %lbl1, !dbg !16
lbl1:
  %t32 = getelementptr inbounds [11 x i8], ptr @.str.11, i64 0, i64 0, !dbg !31
  %t33 = call i32 @puts(ptr noundef %t32), !dbg !32
  %t34 = load i64, ptr %v_day_int, !dbg !33
  switch i64 %t34, label %lbl9 [ i64 1, label %lbl11 i64 2, label %lbl11 i64 3, label %lbl11 i64 4, label %lbl11 i64 5, label %lbl11 i64 6, label %lbl12 i64 7, label %lbl12 ], !dbg !34
lbl11:
  %t35 = getelementptr inbounds [8 x i8], ptr @.str.12, i64 0, i64 0, !dbg !35
  %t36 = call i32 @puts(ptr noundef %t35), !dbg !36
  br label %lbl10, !dbg !34
lbl12:
  %t37 = getelementptr inbounds [8 x i8], ptr @.str.13, i64 0, i64 0, !dbg !37
  %t38 = call i32 @puts(ptr noundef %t37), !dbg !38
  br label %lbl10, !dbg !34
lbl9:
  br label %lbl10, !dbg !34
lbl10:
  %t39 = add i64 0, 7, !dbg !39
  store i64 %t39, ptr %v_month_int, !dbg !40
  %t40 = call ptr @malloc(i64 512), !dbg !41
  store i8 0, ptr %t40, !dbg !41
  %t41 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !41
  %t42 = getelementptr inbounds [18 x i8], ptr @.str.14, i64 0, i64 0, !dbg !42
  %t43 = call ptr @strcat(ptr %t40, ptr %t42), !dbg !41
  %t44 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !41
  %t45 = call ptr @strcat(ptr %t40, ptr %t44), !dbg !41
  %t46 = load i64, ptr %v_month_int, !dbg !43
  %t47 = call i64 @strlen(ptr %t40), !dbg !41
  %t48 = getelementptr inbounds i8, ptr %t40, i64 %t47, !dbg !41
  %t49 = sub i64 512, %t47, !dbg !41
  %t50 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t48, i64 %t49, ptr %t41, i64 %t46), !dbg !41
  %t51 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !41
  %t52 = call ptr @strcat(ptr %t40, ptr %t51), !dbg !41
  %t53 = getelementptr inbounds [3 x i8], ptr @.str.15, i64 0, i64 0, !dbg !44
  %t54 = call ptr @strcat(ptr %t40, ptr %t53), !dbg !41
  %t55 = call i32 @puts(ptr noundef %t40), !dbg !41
  %t56 = load i64, ptr %v_month_int, !dbg !45
  switch i64 %t56, label %lbl13 [ i64 12, label %lbl15 i64 1, label %lbl15 i64 2, label %lbl15 i64 3, label %lbl16 i64 4, label %lbl16 i64 5, label %lbl16 i64 6, label %lbl17 i64 7, label %lbl17 i64 8, label %lbl17 i64 9, label %lbl18 i64 10, label %lbl18 i64 11, label %lbl18 ], !dbg !46
lbl15:
  %t57 = getelementptr inbounds [7 x i8], ptr @.str.16, i64 0, i64 0, !dbg !47
  %t58 = call i32 @puts(ptr noundef %t57), !dbg !48
  br label %lbl14, !dbg !46
lbl16:
  %t59 = getelementptr inbounds [7 x i8], ptr @.str.17, i64 0, i64 0, !dbg !49
  %t60 = call i32 @puts(ptr noundef %t59), !dbg !50
  br label %lbl14, !dbg !46
lbl17:
  %t61 = getelementptr inbounds [7 x i8], ptr @.str.18, i64 0, i64 0, !dbg !51
  %t62 = call i32 @puts(ptr noundef %t61), !dbg !52
  br label %lbl14, !dbg !46
lbl18:
  %t63 = getelementptr inbounds [5 x i8], ptr @.str.19, i64 0, i64 0, !dbg !53
  %t64 = call i32 @puts(ptr noundef %t63), !dbg !54
  br label %lbl14, !dbg !46
lbl13:
  br label %lbl14, !dbg !46
lbl14:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [5 x i8] c"Day \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [6 x i8] c" is: \00", align 1
@.str.4 = private unnamed_addr constant [7 x i8] c"Monday\00", align 1
@.str.5 = private unnamed_addr constant [8 x i8] c"Tuesday\00", align 1
@.str.6 = private unnamed_addr constant [10 x i8] c"Wednesday\00", align 1
@.str.7 = private unnamed_addr constant [9 x i8] c"Thursday\00", align 1
@.str.8 = private unnamed_addr constant [7 x i8] c"Friday\00", align 1
@.str.9 = private unnamed_addr constant [9 x i8] c"Saturday\00", align 1
@.str.10 = private unnamed_addr constant [7 x i8] c"Sunday\00", align 1
@.str.11 = private unnamed_addr constant [11 x i8] c"Day type: \00", align 1
@.str.12 = private unnamed_addr constant [8 x i8] c"Weekday\00", align 1
@.str.13 = private unnamed_addr constant [8 x i8] c"Weekend\00", align 1
@.str.14 = private unnamed_addr constant [18 x i8] c"Season for month \00", align 1
@.str.15 = private unnamed_addr constant [3 x i8] c": \00", align 1
@.str.16 = private unnamed_addr constant [7 x i8] c"Winter\00", align 1
@.str.17 = private unnamed_addr constant [7 x i8] c"Spring\00", align 1
@.str.18 = private unnamed_addr constant [7 x i8] c"Summer\00", align 1
@.str.19 = private unnamed_addr constant [5 x i8] c"Fall\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example11_case.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/02_control_flow")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !56, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !59)
!6 = !DILocalVariable(name: "day", scope: !4, file: !3, line: 4, type: !58)
!7 = !DILocalVariable(name: "grade", scope: !4, file: !3, line: 5, type: !58)
!8 = !DILocalVariable(name: "month", scope: !4, file: !3, line: 6, type: !58)
!5 = !DILocation(line: 1, column: 9, scope: !4)
!9 = !DILocation(line: 10, column: 10, scope: !4)
!10 = !DILocation(line: 10, column: 8, scope: !4)
!11 = !DILocation(line: 11, column: 10, scope: !4)
!12 = !DILocation(line: 11, column: 11, scope: !4)
!13 = !DILocation(line: 11, column: 19, scope: !4)
!14 = !DILocation(line: 11, column: 24, scope: !4)
!15 = !DILocation(line: 12, column: 8, scope: !4)
!16 = !DILocation(line: 12, column: 3, scope: !4)
!17 = !DILocation(line: 13, column: 16, scope: !4)
!18 = !DILocation(line: 13, column: 15, scope: !4)
!19 = !DILocation(line: 14, column: 16, scope: !4)
!20 = !DILocation(line: 14, column: 15, scope: !4)
!21 = !DILocation(line: 15, column: 16, scope: !4)
!22 = !DILocation(line: 15, column: 15, scope: !4)
!23 = !DILocation(line: 16, column: 16, scope: !4)
!24 = !DILocation(line: 16, column: 15, scope: !4)
!25 = !DILocation(line: 17, column: 16, scope: !4)
!26 = !DILocation(line: 17, column: 15, scope: !4)
!27 = !DILocation(line: 18, column: 16, scope: !4)
!28 = !DILocation(line: 18, column: 15, scope: !4)
!29 = !DILocation(line: 19, column: 16, scope: !4)
!30 = !DILocation(line: 19, column: 15, scope: !4)
!31 = !DILocation(line: 23, column: 11, scope: !4)
!32 = !DILocation(line: 23, column: 10, scope: !4)
!33 = !DILocation(line: 24, column: 8, scope: !4)
!34 = !DILocation(line: 24, column: 3, scope: !4)
!35 = !DILocation(line: 25, column: 28, scope: !4)
!36 = !DILocation(line: 25, column: 27, scope: !4)
!37 = !DILocation(line: 26, column: 19, scope: !4)
!38 = !DILocation(line: 26, column: 18, scope: !4)
!39 = !DILocation(line: 30, column: 12, scope: !4)
!40 = !DILocation(line: 30, column: 10, scope: !4)
!41 = !DILocation(line: 31, column: 10, scope: !4)
!42 = !DILocation(line: 31, column: 11, scope: !4)
!43 = !DILocation(line: 31, column: 32, scope: !4)
!44 = !DILocation(line: 31, column: 39, scope: !4)
!45 = !DILocation(line: 32, column: 8, scope: !4)
!46 = !DILocation(line: 32, column: 3, scope: !4)
!47 = !DILocation(line: 33, column: 23, scope: !4)
!48 = !DILocation(line: 33, column: 22, scope: !4)
!49 = !DILocation(line: 34, column: 22, scope: !4)
!50 = !DILocation(line: 34, column: 21, scope: !4)
!51 = !DILocation(line: 35, column: 22, scope: !4)
!52 = !DILocation(line: 35, column: 21, scope: !4)
!53 = !DILocation(line: 36, column: 24, scope: !4)
!54 = !DILocation(line: 36, column: 23, scope: !4)
!55 = !{null}
!56 = !DISubroutineType(types: !55)
!57 = !{}
!58 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!59 = !{!6, !7, !8}
