; Kylix LLVM IR — module: StringOps
source_filename = "StringOps.klx"
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
  %t0 = getelementptr inbounds [14 x i8], ptr @.str.0, i64 0, i64 0, !dbg !5
  %v_s_str = alloca ptr, align 8, !dbg !6
  store ptr %t0, ptr %v_s_str, !dbg !6
  %t1 = getelementptr inbounds [9 x i8], ptr @.str.1, i64 0, i64 0, !dbg !7
  %t2 = load ptr, ptr %v_s_str, !dbg !8
  %t3 = call ptr @malloc(i64 512), !dbg !9
  call ptr @strcpy(ptr %t3, ptr %t1), !dbg !9
  call ptr @strcat(ptr %t3, ptr %t2), !dbg !9
  %t4 = call i32 @puts(ptr noundef %t3), !dbg !10
  %t5 = getelementptr inbounds [9 x i8], ptr @.str.2, i64 0, i64 0, !dbg !11
  %t6 = load ptr, ptr %v_s_str, !dbg !12
  %t7 = call i64 @strlen(ptr noundef %t6), !dbg !13
  %t8 = alloca [24 x i8], align 1, !dbg !14
  %t9 = getelementptr inbounds [24 x i8], ptr %t8, i64 0, i64 0, !dbg !14
  %t10 = getelementptr inbounds [5 x i8], ptr @.str.3, i64 0, i64 0, !dbg !14
  %t11 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t9, i64 24, ptr noundef %t10, i64 %t7), !dbg !14
  %t12 = call ptr @malloc(i64 512), !dbg !15
  call ptr @strcpy(ptr %t12, ptr %t5), !dbg !15
  call ptr @strcat(ptr %t12, ptr %t9), !dbg !15
  %t13 = call i32 @puts(ptr noundef %t12), !dbg !16
  %t14 = getelementptr inbounds [16 x i8], ptr @.str.4, i64 0, i64 0, !dbg !17
  %t15 = load ptr, ptr %v_s_str, !dbg !18
  %t16 = add i64 0, 0, !dbg !19
  %t17 = add i64 0, 5, !dbg !20
  %t18 = sub i64 %t17, %t16, !dbg !21
  %t19 = add i64 %t18, 1, !dbg !21
  %t20 = call ptr @malloc(i64 %t19), !dbg !21
  %t21 = getelementptr inbounds i8, ptr %t15, i64 %t16, !dbg !21
  call ptr @memcpy(ptr %t20, ptr %t21, i64 %t18), !dbg !21
  %t22 = getelementptr inbounds i8, ptr %t20, i64 %t18, !dbg !21
  store i8 0, ptr %t22, !dbg !21
  %t23 = call ptr @malloc(i64 512), !dbg !22
  call ptr @strcpy(ptr %t23, ptr %t14), !dbg !22
  call ptr @strcat(ptr %t23, ptr %t20), !dbg !22
  %t24 = call i32 @puts(ptr noundef %t23), !dbg !23
  %t25 = getelementptr inbounds [15 x i8], ptr @.str.5, i64 0, i64 0, !dbg !24
  %t26 = load ptr, ptr %v_s_str, !dbg !25
  %t27 = add i64 0, 7, !dbg !26
  %t28 = load ptr, ptr %v_s_str, !dbg !27
  %t29 = call i64 @strlen(ptr noundef %t28), !dbg !28
  %t30 = sub i64 %t29, %t27, !dbg !29
  %t31 = add i64 %t30, 1, !dbg !29
  %t32 = call ptr @malloc(i64 %t31), !dbg !29
  %t33 = getelementptr inbounds i8, ptr %t26, i64 %t27, !dbg !29
  call ptr @memcpy(ptr %t32, ptr %t33, i64 %t30), !dbg !29
  %t34 = getelementptr inbounds i8, ptr %t32, i64 %t30, !dbg !29
  store i8 0, ptr %t34, !dbg !29
  %t35 = call ptr @malloc(i64 512), !dbg !30
  call ptr @strcpy(ptr %t35, ptr %t25), !dbg !30
  call ptr @strcat(ptr %t35, ptr %t32), !dbg !30
  %t36 = call i32 @puts(ptr noundef %t35), !dbg !31
  %t37 = getelementptr inbounds [4 x i8], ptr @.str.6, i64 0, i64 0, !dbg !32
  %v_a_str = alloca ptr, align 8, !dbg !33
  store ptr %t37, ptr %v_a_str, !dbg !33
  %t38 = getelementptr inbounds [4 x i8], ptr @.str.7, i64 0, i64 0, !dbg !34
  %v_b_str = alloca ptr, align 8, !dbg !35
  store ptr %t38, ptr %v_b_str, !dbg !35
  %t39 = getelementptr inbounds [9 x i8], ptr @.str.8, i64 0, i64 0, !dbg !36
  %t40 = load ptr, ptr %v_a_str, !dbg !37
  %t41 = call ptr @malloc(i64 512), !dbg !38
  call ptr @strcpy(ptr %t41, ptr %t39), !dbg !38
  call ptr @strcat(ptr %t41, ptr %t40), !dbg !38
  %t42 = load ptr, ptr %v_b_str, !dbg !39
  %t43 = call ptr @malloc(i64 512), !dbg !40
  call ptr @strcpy(ptr %t43, ptr %t41), !dbg !40
  call ptr @strcat(ptr %t43, ptr %t42), !dbg !40
  %t44 = call i32 @puts(ptr noundef %t43), !dbg !41
  %t45 = load ptr, ptr %v_a_str, !dbg !42
  %t46 = getelementptr inbounds [4 x i8], ptr @.str.6, i64 0, i64 0, !dbg !43
  %t47 = call i32 @strcmp(ptr %t45, ptr %t46), !dbg !44
  %t48 = icmp eq i32 %t47, 0, !dbg !44
  br i1 %t48, label %lbl0, label %lbl1, !dbg !45
lbl0:
  %t49 = getelementptr inbounds [13 x i8], ptr @.str.9, i64 0, i64 0, !dbg !46
  %t50 = call i32 @puts(ptr noundef %t49), !dbg !47
  br label %lbl1, !dbg !45
lbl1:
  %t51 = load ptr, ptr %v_a_str, !dbg !48
  %t52 = load ptr, ptr %v_b_str, !dbg !49
  %t53 = call i32 @strcmp(ptr %t51, ptr %t52), !dbg !50
  %t54 = icmp ne i32 %t53, 0, !dbg !50
  br i1 %t54, label %lbl2, label %lbl3, !dbg !51
lbl2:
  %t55 = getelementptr inbounds [15 x i8], ptr @.str.10, i64 0, i64 0, !dbg !52
  %t56 = call i32 @puts(ptr noundef %t55), !dbg !53
  br label %lbl3, !dbg !51
lbl3:
  %t57 = add i64 0, 2024, !dbg !54
  %v_n_int = alloca i64, align 8, !dbg !55
  store i64 %t57, ptr %v_n_int, !dbg !55
  %t58 = getelementptr inbounds [7 x i8], ptr @.str.11, i64 0, i64 0, !dbg !56
  %t59 = load i64, ptr %v_n_int, !dbg !57
  %t60 = alloca [24 x i8], align 1, !dbg !58
  %t61 = getelementptr inbounds [24 x i8], ptr %t60, i64 0, i64 0, !dbg !58
  %t62 = getelementptr inbounds [5 x i8], ptr @.str.3, i64 0, i64 0, !dbg !58
  %t63 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t61, i64 24, ptr noundef %t62, i64 %t59), !dbg !58
  %t64 = call ptr @malloc(i64 512), !dbg !59
  call ptr @strcpy(ptr %t64, ptr %t58), !dbg !59
  call ptr @strcat(ptr %t64, ptr %t61), !dbg !59
  %t65 = call i32 @puts(ptr noundef %t64), !dbg !60
  %t66 = getelementptr inbounds [1 x i8], ptr @.str.12, i64 0, i64 0, !dbg !61
  %v_empty_str = alloca ptr, align 8, !dbg !62
  store ptr %t66, ptr %v_empty_str, !dbg !62
  %t67 = load ptr, ptr %v_empty_str, !dbg !63
  %t68 = call i64 @strlen(ptr noundef %t67), !dbg !64
  %t69 = add i64 0, 0, !dbg !65
  %t70 = icmp eq i64 %t68, %t69, !dbg !66
  br i1 %t70, label %lbl4, label %lbl5, !dbg !67
lbl4:
  %t71 = getelementptr inbounds [13 x i8], ptr @.str.13, i64 0, i64 0, !dbg !68
  %t72 = call i32 @puts(ptr noundef %t71), !dbg !69
  br label %lbl5, !dbg !67
lbl5:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [14 x i8] c"Hello, Kylix!\00", align 1
@.str.1 = private unnamed_addr constant [9 x i8] c"String: \00", align 1
@.str.2 = private unnamed_addr constant [9 x i8] c"Length: \00", align 1
@.str.3 = private unnamed_addr constant [5 x i8] c"%lld\00", align 1
@.str.4 = private unnamed_addr constant [16 x i8] c"First 5 chars: \00", align 1
@.str.5 = private unnamed_addr constant [15 x i8] c"From index 7: \00", align 1
@.str.6 = private unnamed_addr constant [4 x i8] c"foo\00", align 1
@.str.7 = private unnamed_addr constant [4 x i8] c"bar\00", align 1
@.str.8 = private unnamed_addr constant [9 x i8] c"Concat: \00", align 1
@.str.9 = private unnamed_addr constant [13 x i8] c"a equals foo\00", align 1
@.str.10 = private unnamed_addr constant [15 x i8] c"a and b differ\00", align 1
@.str.11 = private unnamed_addr constant [7 x i8] c"Year: \00", align 1
@.str.12 = private unnamed_addr constant [1 x i8] c"\00", align 1
@.str.13 = private unnamed_addr constant [13 x i8] c"empty string\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example25_string_ops.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/06_advanced_types")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 7, type: !71, scopeLine: 7, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !74)
!5 = !DILocation(line: 10, column: 12, scope: !4)
!6 = !DILocation(line: 10, column: 3, scope: !4)
!7 = !DILocation(line: 13, column: 11, scope: !4)
!8 = !DILocation(line: 13, column: 24, scope: !4)
!9 = !DILocation(line: 13, column: 22, scope: !4)
!10 = !DILocation(line: 13, column: 10, scope: !4)
!11 = !DILocation(line: 14, column: 11, scope: !4)
!12 = !DILocation(line: 14, column: 40, scope: !4)
!13 = !DILocation(line: 14, column: 39, scope: !4)
!14 = !DILocation(line: 14, column: 32, scope: !4)
!15 = !DILocation(line: 14, column: 22, scope: !4)
!16 = !DILocation(line: 14, column: 10, scope: !4)
!17 = !DILocation(line: 17, column: 11, scope: !4)
!18 = !DILocation(line: 17, column: 31, scope: !4)
!19 = !DILocation(line: 17, column: 33, scope: !4)
!20 = !DILocation(line: 17, column: 35, scope: !4)
!21 = !DILocation(line: 17, column: 32, scope: !4)
!22 = !DILocation(line: 17, column: 29, scope: !4)
!23 = !DILocation(line: 17, column: 10, scope: !4)
!24 = !DILocation(line: 18, column: 11, scope: !4)
!25 = !DILocation(line: 18, column: 30, scope: !4)
!26 = !DILocation(line: 18, column: 32, scope: !4)
!27 = !DILocation(line: 18, column: 41, scope: !4)
!28 = !DILocation(line: 18, column: 40, scope: !4)
!29 = !DILocation(line: 18, column: 31, scope: !4)
!30 = !DILocation(line: 18, column: 28, scope: !4)
!31 = !DILocation(line: 18, column: 10, scope: !4)
!32 = !DILocation(line: 21, column: 12, scope: !4)
!33 = !DILocation(line: 21, column: 3, scope: !4)
!34 = !DILocation(line: 22, column: 12, scope: !4)
!35 = !DILocation(line: 22, column: 3, scope: !4)
!36 = !DILocation(line: 23, column: 11, scope: !4)
!37 = !DILocation(line: 23, column: 24, scope: !4)
!38 = !DILocation(line: 23, column: 22, scope: !4)
!39 = !DILocation(line: 23, column: 28, scope: !4)
!40 = !DILocation(line: 23, column: 26, scope: !4)
!41 = !DILocation(line: 23, column: 10, scope: !4)
!42 = !DILocation(line: 26, column: 6, scope: !4)
!43 = !DILocation(line: 26, column: 10, scope: !4)
!44 = !DILocation(line: 26, column: 8, scope: !4)
!45 = !DILocation(line: 26, column: 3, scope: !4)
!46 = !DILocation(line: 27, column: 13, scope: !4)
!47 = !DILocation(line: 27, column: 12, scope: !4)
!48 = !DILocation(line: 28, column: 6, scope: !4)
!49 = !DILocation(line: 28, column: 11, scope: !4)
!50 = !DILocation(line: 28, column: 9, scope: !4)
!51 = !DILocation(line: 28, column: 3, scope: !4)
!52 = !DILocation(line: 29, column: 13, scope: !4)
!53 = !DILocation(line: 29, column: 12, scope: !4)
!54 = !DILocation(line: 32, column: 12, scope: !4)
!55 = !DILocation(line: 32, column: 3, scope: !4)
!56 = !DILocation(line: 33, column: 11, scope: !4)
!57 = !DILocation(line: 33, column: 31, scope: !4)
!58 = !DILocation(line: 33, column: 30, scope: !4)
!59 = !DILocation(line: 33, column: 20, scope: !4)
!60 = !DILocation(line: 33, column: 10, scope: !4)
!61 = !DILocation(line: 36, column: 16, scope: !4)
!62 = !DILocation(line: 36, column: 3, scope: !4)
!63 = !DILocation(line: 37, column: 13, scope: !4)
!64 = !DILocation(line: 37, column: 12, scope: !4)
!65 = !DILocation(line: 37, column: 22, scope: !4)
!66 = !DILocation(line: 37, column: 20, scope: !4)
!67 = !DILocation(line: 37, column: 3, scope: !4)
!68 = !DILocation(line: 38, column: 13, scope: !4)
!69 = !DILocation(line: 38, column: 12, scope: !4)
!70 = !{null}
!71 = !DISubroutineType(types: !70)
!72 = !{}
!73 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!74 = !{}
