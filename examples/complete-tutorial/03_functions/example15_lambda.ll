; Kylix LLVM IR — module: LambdaDemo
source_filename = "LambdaDemo.klx"
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
  %t0 = alloca { ptr, ptr }, align 8, !dbg !5
  %t1 = getelementptr { ptr, ptr }, ptr %t0, i32 0, i32 0, !dbg !5
  store ptr @__lambda_0, ptr %t1, !dbg !5
  %t2 = getelementptr { ptr, ptr }, ptr %t0, i32 0, i32 1, !dbg !5
  store ptr null, ptr %t2, !dbg !5
  %t3 = load { ptr, ptr }, ptr %t0, !dbg !6
  %t4 = extractvalue { ptr, ptr } %t3, 0, !dbg !6
  %t5 = extractvalue { ptr, ptr } %t3, 1, !dbg !6
  %t6 = getelementptr inbounds [6 x i8], ptr @.str.0, i64 0, i64 0, !dbg !7
  call void (ptr, ptr) %t4(ptr %t5, ptr %t6), !dbg !6
  %t7 = load { ptr, ptr }, ptr %t0, !dbg !8
  %t8 = extractvalue { ptr, ptr } %t7, 0, !dbg !8
  %t9 = extractvalue { ptr, ptr } %t7, 1, !dbg !8
  %t10 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !9
  call void (ptr, ptr) %t8(ptr %t9, ptr %t10), !dbg !8
  %t11 = load { ptr, ptr }, ptr %t0, !dbg !10
  %t12 = extractvalue { ptr, ptr } %t11, 0, !dbg !10
  %t13 = extractvalue { ptr, ptr } %t11, 1, !dbg !10
  %t14 = getelementptr inbounds [6 x i8], ptr @.str.2, i64 0, i64 0, !dbg !11
  call void (ptr, ptr) %t12(ptr %t13, ptr %t14), !dbg !10
  %t15 = alloca { ptr, ptr }, align 8, !dbg !12
  %t16 = getelementptr { ptr, ptr }, ptr %t15, i32 0, i32 0, !dbg !12
  store ptr @__lambda_1, ptr %t16, !dbg !12
  %t17 = getelementptr { ptr, ptr }, ptr %t15, i32 0, i32 1, !dbg !12
  store ptr null, ptr %t17, !dbg !12
  %t18 = load { ptr, ptr }, ptr %t15, !dbg !13
  %t19 = extractvalue { ptr, ptr } %t18, 0, !dbg !13
  %t20 = extractvalue { ptr, ptr } %t18, 1, !dbg !13
  %t21 = getelementptr inbounds [7 x i8], ptr @.str.3, i64 0, i64 0, !dbg !14
  %t22 = add i64 0, 42, !dbg !15
  call void (ptr, ptr, i64) %t19(ptr %t20, ptr %t21, i64 %t22), !dbg !13
  %t23 = load { ptr, ptr }, ptr %t15, !dbg !16
  %t24 = extractvalue { ptr, ptr } %t23, 0, !dbg !16
  %t25 = extractvalue { ptr, ptr } %t23, 1, !dbg !16
  %t26 = getelementptr inbounds [5 x i8], ptr @.str.4, i64 0, i64 0, !dbg !17
  %t27 = add i64 0, 2024, !dbg !18
  call void (ptr, ptr, i64) %t24(ptr %t25, ptr %t26, i64 %t27), !dbg !16
  %t28 = load { ptr, ptr }, ptr %t15, !dbg !19
  %t29 = extractvalue { ptr, ptr } %t28, 0, !dbg !19
  %t30 = extractvalue { ptr, ptr } %t28, 1, !dbg !19
  %t31 = getelementptr inbounds [6 x i8], ptr @.str.5, i64 0, i64 0, !dbg !20
  %t32 = add i64 0, 7, !dbg !21
  call void (ptr, ptr, i64) %t29(ptr %t30, ptr %t31, i64 %t32), !dbg !19
  ret i32 0
}

define void @__lambda_0(ptr %env, ptr %name) {
entry:
  %v_name_str = alloca ptr, align 8
  store ptr %name, ptr %v_name_str
  %t33 = getelementptr inbounds [8 x i8], ptr @.str.6, i64 0, i64 0
  %t34 = load ptr, ptr %v_name_str
  %t35 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t35, ptr %t33)
  call ptr @strcat(ptr %t35, ptr %t34)
  %t36 = getelementptr inbounds [2 x i8], ptr @.str.7, i64 0, i64 0
  %t37 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t37, ptr %t35)
  call ptr @strcat(ptr %t37, ptr %t36)
  %t38 = call i32 @puts(ptr noundef %t37)
  ret void
}

define void @__lambda_1(ptr %env, ptr %label, i64 %value) {
entry:
  %v_label_str = alloca ptr, align 8
  store ptr %label, ptr %v_label_str
  %v_value_int = alloca i64, align 8
  store i64 %value, ptr %v_value_int
  %t39 = load ptr, ptr %v_label_str
  %t40 = getelementptr inbounds [3 x i8], ptr @.str.8, i64 0, i64 0
  %t41 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t41, ptr %t39)
  call ptr @strcat(ptr %t41, ptr %t40)
  %t42 = load i64, ptr %v_value_int
  %t43 = alloca [24 x i8], align 1
  %t44 = getelementptr inbounds [24 x i8], ptr %t43, i64 0, i64 0
  %t45 = getelementptr inbounds [5 x i8], ptr @.str.9, i64 0, i64 0
  %t46 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t44, i64 24, ptr noundef %t45, i64 %t42)
  %t47 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t47, ptr %t41)
  call ptr @strcat(ptr %t47, ptr %t44)
  %t48 = call i32 @puts(ptr noundef %t47)
  ret void
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [6 x i8] c"Alice\00", align 1
@.str.1 = private unnamed_addr constant [4 x i8] c"Bob\00", align 1
@.str.2 = private unnamed_addr constant [6 x i8] c"Kylix\00", align 1
@.str.3 = private unnamed_addr constant [7 x i8] c"Answer\00", align 1
@.str.4 = private unnamed_addr constant [5 x i8] c"Year\00", align 1
@.str.5 = private unnamed_addr constant [6 x i8] c"Items\00", align 1
@.str.6 = private unnamed_addr constant [8 x i8] c"Hello, \00", align 1
@.str.7 = private unnamed_addr constant [2 x i8] c"!\00", align 1
@.str.8 = private unnamed_addr constant [3 x i8] c": \00", align 1
@.str.9 = private unnamed_addr constant [5 x i8] c"%lld\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example15_lambda.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/03_functions")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 10, type: !23, scopeLine: 10, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !26)
!5 = !DILocation(line: 14, column: 3, scope: !4)
!6 = !DILocation(line: 19, column: 8, scope: !4)
!7 = !DILocation(line: 19, column: 9, scope: !4)
!8 = !DILocation(line: 20, column: 8, scope: !4)
!9 = !DILocation(line: 20, column: 9, scope: !4)
!10 = !DILocation(line: 21, column: 8, scope: !4)
!11 = !DILocation(line: 21, column: 9, scope: !4)
!12 = !DILocation(line: 23, column: 3, scope: !4)
!13 = !DILocation(line: 28, column: 12, scope: !4)
!14 = !DILocation(line: 28, column: 13, scope: !4)
!15 = !DILocation(line: 28, column: 23, scope: !4)
!16 = !DILocation(line: 29, column: 12, scope: !4)
!17 = !DILocation(line: 29, column: 13, scope: !4)
!18 = !DILocation(line: 29, column: 21, scope: !4)
!19 = !DILocation(line: 30, column: 12, scope: !4)
!20 = !DILocation(line: 30, column: 13, scope: !4)
!21 = !DILocation(line: 30, column: 22, scope: !4)
!22 = !{null}
!23 = !DISubroutineType(types: !22)
!24 = !{}
!25 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!26 = !{}
