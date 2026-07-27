; Kylix LLVM IR — module: TryFinally
source_filename = "TryFinally.klx"
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
define void @ProcessFile(ptr %filename) !dbg !4 {
entry:
  %v_filename_str = alloca ptr, align 8
  store ptr %filename, ptr %v_filename_str
  #dbg_declare(ptr %v_filename_str, !5, !DIExpression(), !6)
  %t0 = call ptr @malloc(i64 512), !dbg !7
  store i8 0, ptr %t0, !dbg !7
  %t2 = getelementptr inbounds [15 x i8], ptr @.str.1, i64 0, i64 0, !dbg !8
  %t3 = call ptr @strcat(ptr %t0, ptr %t2), !dbg !7
  %t4 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !7
  %t5 = call ptr @strcat(ptr %t0, ptr %t4), !dbg !7
  %t6 = load ptr, ptr %v_filename_str, !dbg !9
  %t7 = call ptr @strcat(ptr %t0, ptr %t6), !dbg !7
  %t8 = call i32 @puts(ptr noundef %t0), !dbg !7
  %t9 = alloca [288 x i8], align 16, !dbg !10
  %t10 = getelementptr [288 x i8], ptr %t9, i64 0, i64 0, !dbg !10
  %t11 = alloca ptr, align 8, !dbg !10
  %t12 = call i32 @setjmp(ptr %t10), !dbg !10
  %t13 = icmp ne i32 %t12, 0, !dbg !10
  br i1 %t13, label %lbl1, label %lbl0, !dbg !10
lbl0:
  %t14 = load ptr, ptr @__kylix_jmpbuf, !dbg !10
  store ptr %t14, ptr %t11, !dbg !10
  store ptr %t10, ptr @__kylix_jmpbuf, !dbg !10
  store i1 false, ptr @__kylix_exc_active, !dbg !10
  %t15 = getelementptr inbounds [19 x i8], ptr @.str.3, i64 0, i64 0, !dbg !11
  %t16 = call i32 @puts(ptr noundef %t15), !dbg !12
  %t17 = getelementptr inbounds [28 x i8], ptr @.str.4, i64 0, i64 0, !dbg !13
  %t18 = call i32 @puts(ptr noundef %t17), !dbg !14
  %t19 = load ptr, ptr %t11, !dbg !10
  store ptr %t19, ptr @__kylix_jmpbuf, !dbg !10
  store i1 false, ptr @__kylix_exc_active, !dbg !10
  br label %lbl2, !dbg !10
lbl1:
  %t20 = load ptr, ptr %t11, !dbg !10
  store ptr %t20, ptr @__kylix_jmpbuf, !dbg !10
  br label %lbl4, !dbg !10
lbl2:
  %t22 = call ptr @malloc(i64 512), !dbg !15
  store i8 0, ptr %t22, !dbg !15
  %t24 = getelementptr inbounds [15 x i8], ptr @.str.5, i64 0, i64 0, !dbg !16
  %t25 = call ptr @strcat(ptr %t22, ptr %t24), !dbg !15
  %t26 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !15
  %t27 = call ptr @strcat(ptr %t22, ptr %t26), !dbg !15
  %t28 = load ptr, ptr %v_filename_str, !dbg !17
  %t29 = call ptr @strcat(ptr %t22, ptr %t28), !dbg !15
  %t30 = call i32 @puts(ptr noundef %t22), !dbg !15
  br label %lbl5, !dbg !10
lbl3:
  %t31 = call ptr @malloc(i64 512), !dbg !15
  store i8 0, ptr %t31, !dbg !15
  %t33 = getelementptr inbounds [15 x i8], ptr @.str.5, i64 0, i64 0, !dbg !16
  %t34 = call ptr @strcat(ptr %t31, ptr %t33), !dbg !15
  %t35 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !15
  %t36 = call ptr @strcat(ptr %t31, ptr %t35), !dbg !15
  %t37 = load ptr, ptr %v_filename_str, !dbg !17
  %t38 = call ptr @strcat(ptr %t31, ptr %t37), !dbg !15
  %t39 = call i32 @puts(ptr noundef %t31), !dbg !15
  br label %lbl5, !dbg !10
lbl4:
  %t40 = call ptr @malloc(i64 512), !dbg !15
  store i8 0, ptr %t40, !dbg !15
  %t42 = getelementptr inbounds [15 x i8], ptr @.str.5, i64 0, i64 0, !dbg !16
  %t43 = call ptr @strcat(ptr %t40, ptr %t42), !dbg !15
  %t44 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !15
  %t45 = call ptr @strcat(ptr %t40, ptr %t44), !dbg !15
  %t46 = load ptr, ptr %v_filename_str, !dbg !17
  %t47 = call ptr @strcat(ptr %t40, ptr %t46), !dbg !15
  %t48 = call i32 @puts(ptr noundef %t40), !dbg !15
  %t49 = load ptr, ptr @__kylix_jmpbuf, !dbg !10
  call void @longjmp(ptr %t49, i32 1), !dbg !10
  unreachable, !dbg !10
lbl5:
  ret void, !dbg !6
}

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
define i32 @main() !dbg !18 {
entry:
  %t50 = getelementptr inbounds [9 x i8], ptr @.str.6, i64 0, i64 0, !dbg !19
  call void @ProcessFile(ptr %t50), !dbg !20
  %t51 = getelementptr inbounds [4 x i8], ptr @.str.7, i64 0, i64 0, !dbg !21
  %t52 = call i32 @puts(ptr noundef %t51), !dbg !22
  %t53 = alloca [288 x i8], align 16, !dbg !23
  %t54 = getelementptr [288 x i8], ptr %t53, i64 0, i64 0, !dbg !23
  %t55 = alloca ptr, align 8, !dbg !23
  %t56 = call i32 @setjmp(ptr %t54), !dbg !23
  %t57 = icmp ne i32 %t56, 0, !dbg !23
  br i1 %t57, label %lbl7, label %lbl6, !dbg !23
lbl6:
  %t58 = load ptr, ptr @__kylix_jmpbuf, !dbg !23
  store ptr %t58, ptr %t55, !dbg !23
  store ptr %t54, ptr @__kylix_jmpbuf, !dbg !23
  store i1 false, ptr @__kylix_exc_active, !dbg !23
  %t59 = getelementptr inbounds [22 x i8], ptr @.str.8, i64 0, i64 0, !dbg !24
  %t60 = call i32 @puts(ptr noundef %t59), !dbg !25
  %t61 = add i64 0, 10, !dbg !26
  %t62 = add i64 0, 2, !dbg !27
  %t63 = sdiv i64 %t61, %t62, !dbg !28
  %v_x_int = alloca i64, align 8, !dbg !29
  store i64 %t63, ptr %v_x_int, !dbg !29
  %t64 = call ptr @malloc(i64 512), !dbg !30
  store i8 0, ptr %t64, !dbg !30
  %t65 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !30
  %t66 = getelementptr inbounds [9 x i8], ptr @.str.9, i64 0, i64 0, !dbg !31
  %t67 = call ptr @strcat(ptr %t64, ptr %t66), !dbg !30
  %t68 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !30
  %t69 = call ptr @strcat(ptr %t64, ptr %t68), !dbg !30
  %t70 = load i64, ptr %v_x_int, !dbg !32
  %t71 = call i64 @strlen(ptr %t64), !dbg !30
  %t72 = getelementptr inbounds i8, ptr %t64, i64 %t71, !dbg !30
  %t73 = sub i64 512, %t71, !dbg !30
  %t74 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t72, i64 %t73, ptr %t65, i64 %t70), !dbg !30
  %t75 = call i32 @puts(ptr noundef %t64), !dbg !30
  %t76 = load ptr, ptr %t55, !dbg !23
  store ptr %t76, ptr @__kylix_jmpbuf, !dbg !23
  store i1 false, ptr @__kylix_exc_active, !dbg !23
  br label %lbl8, !dbg !23
lbl7:
  %t77 = load ptr, ptr %t55, !dbg !23
  store ptr %t77, ptr @__kylix_jmpbuf, !dbg !23
  %t79 = getelementptr inbounds [15 x i8], ptr @.str.10, i64 0, i64 0, !dbg !33
  %t80 = call i32 @puts(ptr noundef %t79), !dbg !34
  store i1 false, ptr @__kylix_exc_active, !dbg !23
  br label %lbl9, !dbg !23
lbl8:
  %t81 = getelementptr inbounds [17 x i8], ptr @.str.11, i64 0, i64 0, !dbg !35
  %t82 = call i32 @puts(ptr noundef %t81), !dbg !36
  br label %lbl11, !dbg !23
lbl9:
  %t83 = getelementptr inbounds [17 x i8], ptr @.str.11, i64 0, i64 0, !dbg !35
  %t84 = call i32 @puts(ptr noundef %t83), !dbg !36
  br label %lbl11, !dbg !23
lbl10:
  %t85 = getelementptr inbounds [17 x i8], ptr @.str.11, i64 0, i64 0, !dbg !35
  %t86 = call i32 @puts(ptr noundef %t85), !dbg !36
  %t87 = load ptr, ptr @__kylix_jmpbuf, !dbg !23
  call void @longjmp(ptr %t87, i32 1), !dbg !23
  unreachable, !dbg !23
lbl11:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [15 x i8] c"Opening file: \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [19 x i8] c"Processing data...\00", align 1
@.str.4 = private unnamed_addr constant [28 x i8] c"Data processed successfully\00", align 1
@.str.5 = private unnamed_addr constant [15 x i8] c"Closing file: \00", align 1
@.str.6 = private unnamed_addr constant [9 x i8] c"data.txt\00", align 1
@.str.7 = private unnamed_addr constant [4 x i8] c"---\00", align 1
@.str.8 = private unnamed_addr constant [22 x i8] c"Starting operation...\00", align 1
@.str.9 = private unnamed_addr constant [9 x i8] c"Result: \00", align 1
@.str.10 = private unnamed_addr constant [15 x i8] c"Error occurred\00", align 1
@.str.11 = private unnamed_addr constant [17 x i8] c"Cleanup complete\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example28_finally.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/10_exceptions")
!4 = distinct !DISubprogram(name: "ProcessFile", scope: !3, file: !3, line: 3, type: !38, scopeLine: 3, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !41)
!18 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !38, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !42)
!5 = !DILocalVariable(name: "filename", scope: !4, file: !3, line: 3, type: !40)
!6 = !DILocation(line: 3, column: 1, scope: !4)
!7 = !DILocation(line: 5, column: 10, scope: !4)
!8 = !DILocation(line: 5, column: 11, scope: !4)
!9 = !DILocation(line: 5, column: 29, scope: !4)
!10 = !DILocation(line: 7, column: 3, scope: !4)
!11 = !DILocation(line: 9, column: 13, scope: !4)
!12 = !DILocation(line: 9, column: 12, scope: !4)
!13 = !DILocation(line: 11, column: 13, scope: !4)
!14 = !DILocation(line: 11, column: 12, scope: !4)
!15 = !DILocation(line: 15, column: 12, scope: !4)
!16 = !DILocation(line: 15, column: 13, scope: !4)
!17 = !DILocation(line: 15, column: 31, scope: !4)
!19 = !DILocation(line: 21, column: 15, scope: !18)
!20 = !DILocation(line: 21, column: 14, scope: !18)
!21 = !DILocation(line: 23, column: 11, scope: !18)
!22 = !DILocation(line: 23, column: 10, scope: !18)
!23 = !DILocation(line: 26, column: 3, scope: !18)
!24 = !DILocation(line: 28, column: 13, scope: !18)
!25 = !DILocation(line: 28, column: 12, scope: !18)
!26 = !DILocation(line: 29, column: 14, scope: !18)
!27 = !DILocation(line: 29, column: 21, scope: !18)
!28 = !DILocation(line: 29, column: 17, scope: !18)
!29 = !DILocation(line: 29, column: 5, scope: !18)
!30 = !DILocation(line: 30, column: 12, scope: !18)
!31 = !DILocation(line: 30, column: 13, scope: !18)
!32 = !DILocation(line: 30, column: 25, scope: !18)
!33 = !DILocation(line: 34, column: 13, scope: !18)
!34 = !DILocation(line: 34, column: 12, scope: !18)
!35 = !DILocation(line: 38, column: 13, scope: !18)
!36 = !DILocation(line: 38, column: 12, scope: !18)
!37 = !{null}
!38 = !DISubroutineType(types: !37)
!39 = !{}
!40 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!41 = !{!5}
!42 = !{}
