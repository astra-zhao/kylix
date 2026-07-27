; Kylix LLVM IR — module: Constants
source_filename = "Constants.klx"
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
  %t0 = call ptr @malloc(i64 512), !dbg !5
  store i8 0, ptr %t0, !dbg !5
  %t2 = getelementptr inbounds [14 x i8], ptr @.str.1, i64 0, i64 0, !dbg !6
  %t3 = call ptr @strcat(ptr %t0, ptr %t2), !dbg !5
  %t4 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !5
  %t5 = call ptr @strcat(ptr %t0, ptr %t4), !dbg !5
  %t6 = getelementptr inbounds [11 x i8], ptr @.str.3, i64 0, i64 0, !dbg !7
  %t7 = call ptr @strcat(ptr %t0, ptr %t6), !dbg !5
  %t8 = call i32 @puts(ptr noundef %t0), !dbg !5
  %t9 = call ptr @malloc(i64 512), !dbg !8
  store i8 0, ptr %t9, !dbg !8
  %t10 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !8
  %t11 = getelementptr inbounds [11 x i8], ptr @.str.4, i64 0, i64 0, !dbg !9
  %t12 = call ptr @strcat(ptr %t9, ptr %t11), !dbg !8
  %t13 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !8
  %t14 = call ptr @strcat(ptr %t9, ptr %t13), !dbg !8
  %t15 = add i64 0, 100, !dbg !10
  %t16 = call i64 @strlen(ptr %t9), !dbg !8
  %t17 = getelementptr inbounds i8, ptr %t9, i64 %t16, !dbg !8
  %t18 = sub i64 512, %t16, !dbg !8
  %t19 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t17, i64 %t18, ptr %t10, i64 %t15), !dbg !8
  %t20 = call i32 @puts(ptr noundef %t9), !dbg !8
  %t21 = call ptr @malloc(i64 512), !dbg !11
  store i8 0, ptr %t21, !dbg !11
  %t23 = getelementptr inbounds [5 x i8], ptr @.str.5, i64 0, i64 0, !dbg !12
  %t24 = call ptr @strcat(ptr %t21, ptr %t23), !dbg !11
  %t25 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !11
  %t26 = call ptr @strcat(ptr %t21, ptr %t25), !dbg !11
  %t27 = fadd double 0.0, 3.141593, !dbg !13
  %t28 = getelementptr inbounds [6 x i8], ptr @.str.6, i64 0, i64 0, !dbg !11
  %t29 = call i64 @strlen(ptr %t21), !dbg !11
  %t30 = getelementptr inbounds i8, ptr %t21, i64 %t29, !dbg !11
  %t31 = sub i64 512, %t29, !dbg !11
  %t32 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t30, i64 %t31, ptr %t28, double %t27), !dbg !11
  %t33 = call i32 @puts(ptr noundef %t21), !dbg !11
  %t34 = call ptr @malloc(i64 512), !dbg !14
  store i8 0, ptr %t34, !dbg !14
  %t36 = getelementptr inbounds [13 x i8], ptr @.str.7, i64 0, i64 0, !dbg !15
  %t37 = call ptr @strcat(ptr %t34, ptr %t36), !dbg !14
  %t38 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !14
  %t39 = call ptr @strcat(ptr %t34, ptr %t38), !dbg !14
  %t40 = add i1 0, 1, !dbg !16
  %t41 = getelementptr inbounds [5 x i8], ptr @.str.8, i64 0, i64 0, !dbg !14
  %t42 = getelementptr inbounds [6 x i8], ptr @.str.9, i64 0, i64 0, !dbg !14
  %t43 = select i1 %t40, ptr %t41, ptr %t42, !dbg !14
  %t44 = call ptr @strcat(ptr %t34, ptr %t43), !dbg !14
  %t45 = call i32 @puts(ptr noundef %t34), !dbg !14
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"Application: \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [11 x i8] c"Kylix Demo\00", align 1
@.str.4 = private unnamed_addr constant [11 x i8] c"Max Size: \00", align 1
@.str.5 = private unnamed_addr constant [5 x i8] c"Pi: \00", align 1
@.str.6 = private unnamed_addr constant [6 x i8] c"%.15g\00", align 1
@.str.7 = private unnamed_addr constant [13 x i8] c"Debug Mode: \00", align 1
@.str.8 = private unnamed_addr constant [5 x i8] c"true\00", align 1
@.str.9 = private unnamed_addr constant [6 x i8] c"false\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example03_constants.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/01_basics")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !18, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !21)
!5 = !DILocation(line: 11, column: 10, scope: !4)
!6 = !DILocation(line: 11, column: 11, scope: !4)
!7 = !DILocation(line: 6, column: 14, scope: !4)
!8 = !DILocation(line: 12, column: 10, scope: !4)
!9 = !DILocation(line: 12, column: 11, scope: !4)
!10 = !DILocation(line: 5, column: 14, scope: !4)
!11 = !DILocation(line: 13, column: 10, scope: !4)
!12 = !DILocation(line: 13, column: 11, scope: !4)
!13 = !DILocation(line: 7, column: 8, scope: !4)
!14 = !DILocation(line: 14, column: 10, scope: !4)
!15 = !DILocation(line: 14, column: 11, scope: !4)
!16 = !DILocation(line: 8, column: 16, scope: !4)
!17 = !{null}
!18 = !DISubroutineType(types: !17)
!19 = !{}
!20 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!21 = !{}
