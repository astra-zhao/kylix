; Kylix LLVM IR — module: SysutilDemo
source_filename = "SysutilDemo.klx"
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
  %v_fname_str = alloca ptr, align 8, !dbg !5
  store ptr null, ptr %v_fname_str, !dbg !5
  #dbg_declare(ptr %v_fname_str, !6, !DIExpression(), !5)
  %v_content_str = alloca ptr, align 8, !dbg !5
  store ptr null, ptr %v_content_str, !dbg !5
  #dbg_declare(ptr %v_content_str, !7, !DIExpression(), !5)
  %t0 = getelementptr inbounds [31 x i8], ptr @.str.0, i64 0, i64 0, !dbg !8
  store ptr %t0, ptr %v_fname_str, !dbg !9
  %t1 = load ptr, ptr %v_fname_str, !dbg !10
  %t2 = getelementptr inbounds [26 x i8], ptr @.str.1, i64 0, i64 0, !dbg !11
  call void @__kylix_sysutil_WriteFile(ptr %t1, ptr %t2), !dbg !12
  %t3 = getelementptr inbounds [18 x i8], ptr @.str.2, i64 0, i64 0, !dbg !13
  %t4 = load ptr, ptr %v_fname_str, !dbg !14
  %t5 = call ptr @malloc(i64 512), !dbg !15
  call ptr @strcpy(ptr %t5, ptr %t3), !dbg !15
  call ptr @strcat(ptr %t5, ptr %t4), !dbg !15
  %t6 = call i32 @puts(ptr noundef %t5), !dbg !16
  %t7 = load ptr, ptr %v_fname_str, !dbg !17
  %t8 = call i1 @__kylix_sysutil_FileExists(ptr %t7), !dbg !18
  br i1 %t8, label %lbl0, label %lbl1, !dbg !19
lbl0:
  %t9 = getelementptr inbounds [23 x i8], ptr @.str.3, i64 0, i64 0, !dbg !20
  %t10 = call i32 @puts(ptr noundef %t9), !dbg !21
  br label %lbl1, !dbg !19
lbl1:
  %t11 = load ptr, ptr %v_fname_str, !dbg !22
  %t12 = call ptr @__kylix_sysutil_ReadFile(ptr %t11), !dbg !23
  store ptr %t12, ptr %v_content_str, !dbg !24
  %t13 = getelementptr inbounds [12 x i8], ptr @.str.4, i64 0, i64 0, !dbg !25
  %t14 = load ptr, ptr %v_content_str, !dbg !26
  %t15 = call ptr @malloc(i64 512), !dbg !27
  call ptr @strcpy(ptr %t15, ptr %t13), !dbg !27
  call ptr @strcat(ptr %t15, ptr %t14), !dbg !27
  %t16 = call i32 @puts(ptr noundef %t15), !dbg !28
  %t17 = getelementptr inbounds [19 x i8], ptr @.str.5, i64 0, i64 0, !dbg !29
  %t18 = getelementptr inbounds [5 x i8], ptr @.str.6, i64 0, i64 0, !dbg !30
  %t19 = getelementptr inbounds [6 x i8], ptr @.str.7, i64 0, i64 0, !dbg !31
  %t20 = getelementptr inbounds [4 x i8], ptr @.str.8, i64 0, i64 0, !dbg !32
  %t21 = call ptr @__kylix_sysutil_PathJoin_3(ptr %t18, ptr %t19, ptr %t20), !dbg !33
  %t22 = call ptr @malloc(i64 512), !dbg !34
  call ptr @strcpy(ptr %t22, ptr %t17), !dbg !34
  call ptr @strcat(ptr %t22, ptr %t21), !dbg !34
  %t23 = call i32 @puts(ptr noundef %t22), !dbg !35
  %t24 = getelementptr inbounds [19 x i8], ptr @.str.9, i64 0, i64 0, !dbg !36
  %t25 = getelementptr inbounds [18 x i8], ptr @.str.10, i64 0, i64 0, !dbg !37
  %t26 = call ptr @__kylix_sysutil_PathBase(ptr %t25), !dbg !38
  %t27 = call ptr @malloc(i64 512), !dbg !39
  call ptr @strcpy(ptr %t27, ptr %t24), !dbg !39
  call ptr @strcat(ptr %t27, ptr %t26), !dbg !39
  %t28 = call i32 @puts(ptr noundef %t27), !dbg !40
  ret i32 0
}

define void @__kylix_sysutil_WriteFile(ptr %path, ptr %content) {
entry:
  %t29 = getelementptr inbounds [2 x i8], ptr @.str.11, i64 0, i64 0
  %t30 = call ptr @fopen(ptr %path, ptr %t29)
  %null = icmp eq ptr %t30, null
  br i1 %null, label %lbl2, label %lbl3
lbl2:
  ret void
lbl3:
  call i32 @fputs(ptr %content, ptr %t30)
  call i32 @fclose(ptr %t30)
  ret void
}

define i1 @__kylix_sysutil_FileExists(ptr %path) {
entry:
  %t31 = call i32 @access(ptr %path, i32 0)
  %ok = icmp eq i32 %t31, 0
  ret i1 %ok
}

define ptr @__kylix_sysutil_ReadFile(ptr %path) {
entry:
  %t32 = getelementptr inbounds [2 x i8], ptr @.str.12, i64 0, i64 0
  %t33 = call ptr @fopen(ptr %path, ptr %t32)
  %null = icmp eq ptr %t33, null
  br i1 %null, label %lbl4, label %lbl5
lbl4:
  ret ptr null
lbl5:
  call i32 @fseek(ptr %t33, i64 0, i32 2)
  %t34 = call i64 @ftell(ptr %t33)
  call i32 @fseek(ptr %t33, i64 0, i32 0)
  %t35 = add i64 %t34, 1
  %t36 = call ptr @malloc(i64 %t35)
  call i64 @fread(ptr %t36, i64 1, i64 %t34, ptr %t33)
  %t37 = getelementptr inbounds i8, ptr %t36, i64 %t34
  store i8 0, ptr %t37
  call i32 @fclose(ptr %t33)
  ret ptr %t36
}

define ptr @__kylix_sysutil_PathJoin_3(ptr %p0, ptr %p1, ptr %p2) {
entry:
  %t38 = call ptr @malloc(i64 4096)
  call ptr @strcpy(ptr %t38, ptr %p0)
  %t39 = getelementptr inbounds [2 x i8], ptr @.str.13, i64 0, i64 0
  call ptr @strcat(ptr %t38, ptr %t39)
  call ptr @strcat(ptr %t38, ptr %p1)
  call ptr @strcat(ptr %t38, ptr %t39)
  call ptr @strcat(ptr %t38, ptr %p2)
  ret ptr %t38
}

define ptr @__kylix_sysutil_PathBase(ptr %path) {
entry:
  %t40 = call i64 @strlen(ptr %path)
  %t41 = alloca i64
  store i64 %t40, ptr %t41
  br label %lbl6
lbl6:
  %t42 = load i64, ptr %t41
  %t43 = icmp eq i64 %t42, 0
  br i1 %t43, label %lbl9, label %lbl7
lbl7:
  %t44 = sub i64 %t42, 1
  store i64 %t44, ptr %t41
  %t45 = getelementptr inbounds i8, ptr %path, i64 %t44
  %t46 = load i8, ptr %t45
  %t47 = icmp eq i8 %t46, 47
  br i1 %t47, label %lbl8, label %lbl6
lbl8:
  %t49 = add i64 %t44, 1
  %t48 = getelementptr inbounds i8, ptr %path, i64 %t49
  %t50 = call i64 @strlen(ptr %t48)
  %t51 = add i64 %t50, 1
  %t52 = call ptr @malloc(i64 %t51)
  call ptr @strcpy(ptr %t52, ptr %t48)
  ret ptr %t52
lbl9:
  %t53 = add i64 %t40, 1
  %t54 = call ptr @malloc(i64 %t53)
  call ptr @strcpy(ptr %t54, ptr %path)
  ret ptr %t54
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [31 x i8] c"/tmp/kylix_example_sysutil.txt\00", align 1
@.str.1 = private unnamed_addr constant [26 x i8] c"Hello from Kylix sysutil!\00", align 1
@.str.2 = private unnamed_addr constant [18 x i8] c"File written to: \00", align 1
@.str.3 = private unnamed_addr constant [23 x i8] c"File exists: confirmed\00", align 1
@.str.4 = private unnamed_addr constant [12 x i8] c"Read back: \00", align 1
@.str.5 = private unnamed_addr constant [19 x i8] c"PathJoin example: \00", align 1
@.str.6 = private unnamed_addr constant [5 x i8] c"/usr\00", align 1
@.str.7 = private unnamed_addr constant [6 x i8] c"local\00", align 1
@.str.8 = private unnamed_addr constant [4 x i8] c"bin\00", align 1
@.str.9 = private unnamed_addr constant [19 x i8] c"PathBase example: \00", align 1
@.str.10 = private unnamed_addr constant [18 x i8] c"/path/to/file.txt\00", align 1
@.str.11 = private unnamed_addr constant [2 x i8] c"w\00", align 1
@.str.12 = private unnamed_addr constant [2 x i8] c"r\00", align 1
@.str.13 = private unnamed_addr constant [2 x i8] c"/\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example36_sysutil.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/08_stdlib_utils")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 7, type: !42, scopeLine: 7, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !45)
!6 = !DILocalVariable(name: "fname", scope: !4, file: !3, line: 11, type: !44)
!7 = !DILocalVariable(name: "content", scope: !4, file: !3, line: 12, type: !44)
!5 = !DILocation(line: 7, column: 9, scope: !4)
!8 = !DILocation(line: 14, column: 12, scope: !4)
!9 = !DILocation(line: 14, column: 10, scope: !4)
!10 = !DILocation(line: 16, column: 13, scope: !4)
!11 = !DILocation(line: 16, column: 20, scope: !4)
!12 = !DILocation(line: 16, column: 12, scope: !4)
!13 = !DILocation(line: 17, column: 11, scope: !4)
!14 = !DILocation(line: 17, column: 33, scope: !4)
!15 = !DILocation(line: 17, column: 31, scope: !4)
!16 = !DILocation(line: 17, column: 10, scope: !4)
!17 = !DILocation(line: 19, column: 17, scope: !4)
!18 = !DILocation(line: 19, column: 16, scope: !4)
!19 = !DILocation(line: 19, column: 3, scope: !4)
!20 = !DILocation(line: 20, column: 13, scope: !4)
!21 = !DILocation(line: 20, column: 12, scope: !4)
!22 = !DILocation(line: 22, column: 23, scope: !4)
!23 = !DILocation(line: 22, column: 22, scope: !4)
!24 = !DILocation(line: 22, column: 12, scope: !4)
!25 = !DILocation(line: 23, column: 11, scope: !4)
!26 = !DILocation(line: 23, column: 27, scope: !4)
!27 = !DILocation(line: 23, column: 25, scope: !4)
!28 = !DILocation(line: 23, column: 10, scope: !4)
!29 = !DILocation(line: 25, column: 11, scope: !4)
!30 = !DILocation(line: 25, column: 43, scope: !4)
!31 = !DILocation(line: 25, column: 51, scope: !4)
!32 = !DILocation(line: 25, column: 60, scope: !4)
!33 = !DILocation(line: 25, column: 42, scope: !4)
!34 = !DILocation(line: 25, column: 32, scope: !4)
!35 = !DILocation(line: 25, column: 10, scope: !4)
!36 = !DILocation(line: 26, column: 11, scope: !4)
!37 = !DILocation(line: 26, column: 43, scope: !4)
!38 = !DILocation(line: 26, column: 42, scope: !4)
!39 = !DILocation(line: 26, column: 32, scope: !4)
!40 = !DILocation(line: 26, column: 10, scope: !4)
!41 = !{null}
!42 = !DISubroutineType(types: !41)
!43 = !{}
!44 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!45 = !{!6, !7}
