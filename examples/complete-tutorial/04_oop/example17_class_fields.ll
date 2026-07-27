; Kylix LLVM IR — module: ClassFields
source_filename = "ClassFields.klx"
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
%TPerson = type { ptr, ptr, i64, ptr }
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
  %t0 = call ptr @malloc(i64 32), !dbg !5
  %v_p_str = alloca ptr, align 8, !dbg !6
  store ptr %t0, ptr %v_p_str, !dbg !6
  %t1 = getelementptr inbounds [6 x i8], ptr @.str.0, i64 0, i64 0, !dbg !7
  %t2 = load ptr, ptr %v_p_str, !dbg !8
  %t3 = getelementptr inbounds %TPerson, ptr %t2, i32 0, i32 1, !dbg !8
  store ptr %t1, ptr %t3, !dbg !8
  %t4 = add i64 0, 30, !dbg !9
  %t5 = load ptr, ptr %v_p_str, !dbg !10
  %t6 = getelementptr inbounds %TPerson, ptr %t5, i32 0, i32 2, !dbg !10
  store i64 %t4, ptr %t6, !dbg !10
  %t7 = getelementptr inbounds [18 x i8], ptr @.str.1, i64 0, i64 0, !dbg !11
  %t8 = load ptr, ptr %v_p_str, !dbg !12
  %t9 = getelementptr inbounds %TPerson, ptr %t8, i32 0, i32 3, !dbg !12
  store ptr %t7, ptr %t9, !dbg !12
  %t10 = getelementptr inbounds [7 x i8], ptr @.str.2, i64 0, i64 0, !dbg !13
  %t11 = load ptr, ptr %v_p_str, !dbg !14
  %t12 = getelementptr inbounds %TPerson, ptr %t11, i32 0, i32 1, !dbg !14
  %t13 = load ptr, ptr %t12, !dbg !14
  %t14 = call ptr @malloc(i64 512), !dbg !15
  call ptr @strcpy(ptr %t14, ptr %t10), !dbg !15
  call ptr @strcat(ptr %t14, ptr %t13), !dbg !15
  %t15 = call i32 @puts(ptr noundef %t14), !dbg !16
  %t16 = getelementptr inbounds [6 x i8], ptr @.str.3, i64 0, i64 0, !dbg !17
  %t17 = load ptr, ptr %v_p_str, !dbg !18
  %t18 = getelementptr inbounds %TPerson, ptr %t17, i32 0, i32 2, !dbg !18
  %t19 = load i64, ptr %t18, !dbg !18
  %t20 = alloca [24 x i8], align 1, !dbg !19
  %t21 = getelementptr inbounds [24 x i8], ptr %t20, i64 0, i64 0, !dbg !19
  %t22 = getelementptr inbounds [5 x i8], ptr @.str.4, i64 0, i64 0, !dbg !19
  %t23 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t21, i64 24, ptr noundef %t22, i64 %t19), !dbg !19
  %t24 = call ptr @malloc(i64 512), !dbg !20
  call ptr @strcpy(ptr %t24, ptr %t16), !dbg !20
  call ptr @strcat(ptr %t24, ptr %t21), !dbg !20
  %t25 = call i32 @puts(ptr noundef %t24), !dbg !21
  %t26 = getelementptr inbounds [8 x i8], ptr @.str.5, i64 0, i64 0, !dbg !22
  %t27 = load ptr, ptr %v_p_str, !dbg !23
  %t28 = getelementptr inbounds %TPerson, ptr %t27, i32 0, i32 3, !dbg !23
  %t29 = load ptr, ptr %t28, !dbg !23
  %t30 = call ptr @malloc(i64 512), !dbg !24
  call ptr @strcpy(ptr %t30, ptr %t26), !dbg !24
  call ptr @strcat(ptr %t30, ptr %t29), !dbg !24
  %t31 = call i32 @puts(ptr noundef %t30), !dbg !25
  %t32 = call ptr @malloc(i64 32), !dbg !26
  %v_p2_str = alloca ptr, align 8, !dbg !27
  store ptr %t32, ptr %v_p2_str, !dbg !27
  %t33 = getelementptr inbounds [4 x i8], ptr @.str.6, i64 0, i64 0, !dbg !28
  %t34 = load ptr, ptr %v_p2_str, !dbg !29
  %t35 = getelementptr inbounds %TPerson, ptr %t34, i32 0, i32 1, !dbg !29
  store ptr %t33, ptr %t35, !dbg !29
  %t36 = add i64 0, 25, !dbg !30
  %t37 = load ptr, ptr %v_p2_str, !dbg !31
  %t38 = getelementptr inbounds %TPerson, ptr %t37, i32 0, i32 2, !dbg !31
  store i64 %t36, ptr %t38, !dbg !31
  %t39 = getelementptr inbounds [16 x i8], ptr @.str.7, i64 0, i64 0, !dbg !32
  %t40 = load ptr, ptr %v_p2_str, !dbg !33
  %t41 = getelementptr inbounds %TPerson, ptr %t40, i32 0, i32 3, !dbg !33
  store ptr %t39, ptr %t41, !dbg !33
  %t42 = getelementptr inbounds [22 x i8], ptr @.str.8, i64 0, i64 0, !dbg !34
  %t43 = call i32 @puts(ptr noundef %t42), !dbg !35
  %t44 = getelementptr inbounds [7 x i8], ptr @.str.2, i64 0, i64 0, !dbg !36
  %t45 = load ptr, ptr %v_p2_str, !dbg !37
  %t46 = getelementptr inbounds %TPerson, ptr %t45, i32 0, i32 1, !dbg !37
  %t47 = load ptr, ptr %t46, !dbg !37
  %t48 = call ptr @malloc(i64 512), !dbg !38
  call ptr @strcpy(ptr %t48, ptr %t44), !dbg !38
  call ptr @strcat(ptr %t48, ptr %t47), !dbg !38
  %t49 = call i32 @puts(ptr noundef %t48), !dbg !39
  %t50 = getelementptr inbounds [6 x i8], ptr @.str.3, i64 0, i64 0, !dbg !40
  %t51 = load ptr, ptr %v_p2_str, !dbg !41
  %t52 = getelementptr inbounds %TPerson, ptr %t51, i32 0, i32 2, !dbg !41
  %t53 = load i64, ptr %t52, !dbg !41
  %t54 = alloca [24 x i8], align 1, !dbg !42
  %t55 = getelementptr inbounds [24 x i8], ptr %t54, i64 0, i64 0, !dbg !42
  %t56 = getelementptr inbounds [5 x i8], ptr @.str.4, i64 0, i64 0, !dbg !42
  %t57 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t55, i64 24, ptr noundef %t56, i64 %t53), !dbg !42
  %t58 = call ptr @malloc(i64 512), !dbg !43
  call ptr @strcpy(ptr %t58, ptr %t50), !dbg !43
  call ptr @strcat(ptr %t58, ptr %t55), !dbg !43
  %t59 = call i32 @puts(ptr noundef %t58), !dbg !44
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [6 x i8] c"Alice\00", align 1
@.str.1 = private unnamed_addr constant [18 x i8] c"alice@example.com\00", align 1
@.str.2 = private unnamed_addr constant [7 x i8] c"Name: \00", align 1
@.str.3 = private unnamed_addr constant [6 x i8] c"Age: \00", align 1
@.str.4 = private unnamed_addr constant [5 x i8] c"%lld\00", align 1
@.str.5 = private unnamed_addr constant [8 x i8] c"Email: \00", align 1
@.str.6 = private unnamed_addr constant [4 x i8] c"Bob\00", align 1
@.str.7 = private unnamed_addr constant [16 x i8] c"bob@example.com\00", align 1
@.str.8 = private unnamed_addr constant [22 x i8] c"--- Second person ---\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example17_class_fields.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/04_oop")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 7, type: !46, scopeLine: 7, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !49)
!5 = !DILocation(line: 19, column: 19, scope: !4)
!6 = !DILocation(line: 19, column: 3, scope: !4)
!7 = !DILocation(line: 20, column: 13, scope: !4)
!8 = !DILocation(line: 20, column: 11, scope: !4)
!9 = !DILocation(line: 21, column: 12, scope: !4)
!10 = !DILocation(line: 21, column: 10, scope: !4)
!11 = !DILocation(line: 22, column: 14, scope: !4)
!12 = !DILocation(line: 22, column: 12, scope: !4)
!13 = !DILocation(line: 24, column: 11, scope: !4)
!14 = !DILocation(line: 24, column: 23, scope: !4)
!15 = !DILocation(line: 24, column: 20, scope: !4)
!16 = !DILocation(line: 24, column: 10, scope: !4)
!17 = !DILocation(line: 25, column: 11, scope: !4)
!18 = !DILocation(line: 25, column: 31, scope: !4)
!19 = !DILocation(line: 25, column: 29, scope: !4)
!20 = !DILocation(line: 25, column: 19, scope: !4)
!21 = !DILocation(line: 25, column: 10, scope: !4)
!22 = !DILocation(line: 26, column: 11, scope: !4)
!23 = !DILocation(line: 26, column: 24, scope: !4)
!24 = !DILocation(line: 26, column: 21, scope: !4)
!25 = !DILocation(line: 26, column: 10, scope: !4)
!26 = !DILocation(line: 29, column: 20, scope: !4)
!27 = !DILocation(line: 29, column: 3, scope: !4)
!28 = !DILocation(line: 30, column: 14, scope: !4)
!29 = !DILocation(line: 30, column: 12, scope: !4)
!30 = !DILocation(line: 31, column: 13, scope: !4)
!31 = !DILocation(line: 31, column: 11, scope: !4)
!32 = !DILocation(line: 32, column: 15, scope: !4)
!33 = !DILocation(line: 32, column: 13, scope: !4)
!34 = !DILocation(line: 34, column: 11, scope: !4)
!35 = !DILocation(line: 34, column: 10, scope: !4)
!36 = !DILocation(line: 35, column: 11, scope: !4)
!37 = !DILocation(line: 35, column: 24, scope: !4)
!38 = !DILocation(line: 35, column: 20, scope: !4)
!39 = !DILocation(line: 35, column: 10, scope: !4)
!40 = !DILocation(line: 36, column: 11, scope: !4)
!41 = !DILocation(line: 36, column: 32, scope: !4)
!42 = !DILocation(line: 36, column: 29, scope: !4)
!43 = !DILocation(line: 36, column: 19, scope: !4)
!44 = !DILocation(line: 36, column: 10, scope: !4)
!45 = !{null}
!46 = !DISubroutineType(types: !45)
!47 = !{}
!48 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!49 = !{}
