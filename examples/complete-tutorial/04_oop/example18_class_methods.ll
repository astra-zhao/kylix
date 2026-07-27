; Kylix LLVM IR — module: ClassMethods
source_filename = "ClassMethods.klx"
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
%TCounter = type { ptr, i64 }
@TCounter_vtable = constant [4 x ptr] [ ptr @TCounter_Increment, ptr @TCounter_Add, ptr @TCounter_Reset, ptr @TCounter_Get ]
define void @TCounter_Increment(ptr %self) {
entry:
  %t0 = getelementptr inbounds %TCounter, ptr %self, i32 0, i32 1
  %t1 = load i64, ptr %t0
  %t2 = add i64 0, 1
  %t3 = add i64 %t1, %t2
  %t4 = getelementptr inbounds %TCounter, ptr %self, i32 0, i32 1
  store i64 %t3, ptr %t4
  ret void
}

define void @TCounter_Add(ptr %self, i64 %n) {
entry:
  %v_n_int = alloca i64, align 8
  store i64 %n, ptr %v_n_int
  %t5 = getelementptr inbounds %TCounter, ptr %self, i32 0, i32 1
  %t6 = load i64, ptr %t5
  %t7 = load i64, ptr %v_n_int
  %t8 = add i64 %t6, %t7
  %t9 = getelementptr inbounds %TCounter, ptr %self, i32 0, i32 1
  store i64 %t8, ptr %t9
  ret void
}

define void @TCounter_Reset(ptr %self) {
entry:
  %t10 = add i64 0, 0
  %t11 = getelementptr inbounds %TCounter, ptr %self, i32 0, i32 1
  store i64 %t10, ptr %t11
  ret void
}

define i64 @TCounter_Get(ptr %self) {
entry:
  %result = alloca i64, align 8
  %t12 = getelementptr inbounds %TCounter, ptr %self, i32 0, i32 1
  %t13 = load i64, ptr %t12
  store i64 %t13, ptr %result
  %t14 = load i64, ptr %result
  ret i64 %t14
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
define i32 @main() !dbg !4 {
entry:
  %t15 = call ptr @malloc(i64 16), !dbg !5
  %t16 = getelementptr inbounds %TCounter, ptr %t15, i32 0, i32 0, !dbg !5
  store ptr @TCounter_vtable, ptr %t16, !dbg !5
  %v_c_str = alloca ptr, align 8, !dbg !6
  store ptr %t15, ptr %v_c_str, !dbg !6
  %t17 = add i64 0, 0, !dbg !7
  %t18 = load ptr, ptr %v_c_str, !dbg !8
  %t19 = getelementptr inbounds %TCounter, ptr %t18, i32 0, i32 1, !dbg !8
  store i64 %t17, ptr %t19, !dbg !8
  %t20 = load ptr, ptr %v_c_str, !dbg !9
  %t21 = getelementptr inbounds %TCounter, ptr %t20, i32 0, i32 0, !dbg !9
  %t22 = load ptr, ptr %t21, !dbg !9
  %t23 = getelementptr inbounds [4 x ptr], ptr %t22, i32 0, i32 0, !dbg !9
  %t24 = load ptr, ptr %t23, !dbg !9
  call void (ptr) %t24(ptr %t20), !dbg !9
  %t25 = load ptr, ptr %v_c_str, !dbg !10
  %t26 = getelementptr inbounds %TCounter, ptr %t25, i32 0, i32 0, !dbg !10
  %t27 = load ptr, ptr %t26, !dbg !10
  %t28 = getelementptr inbounds [4 x ptr], ptr %t27, i32 0, i32 0, !dbg !10
  %t29 = load ptr, ptr %t28, !dbg !10
  call void (ptr) %t29(ptr %t25), !dbg !10
  %t30 = load ptr, ptr %v_c_str, !dbg !11
  %t31 = getelementptr inbounds %TCounter, ptr %t30, i32 0, i32 0, !dbg !11
  %t32 = load ptr, ptr %t31, !dbg !11
  %t33 = getelementptr inbounds [4 x ptr], ptr %t32, i32 0, i32 0, !dbg !11
  %t34 = load ptr, ptr %t33, !dbg !11
  call void (ptr) %t34(ptr %t30), !dbg !11
  %t35 = getelementptr inbounds [20 x i8], ptr @.str.0, i64 0, i64 0, !dbg !12
  %t36 = load ptr, ptr %v_c_str, !dbg !13
  %t37 = getelementptr inbounds %TCounter, ptr %t36, i32 0, i32 0, !dbg !13
  %t38 = load ptr, ptr %t37, !dbg !13
  %t39 = getelementptr inbounds [4 x ptr], ptr %t38, i32 0, i32 3, !dbg !13
  %t40 = load ptr, ptr %t39, !dbg !13
  %t41 = call i64 (ptr) %t40(ptr %t36), !dbg !13
  %t42 = alloca [24 x i8], align 1, !dbg !14
  %t43 = getelementptr inbounds [24 x i8], ptr %t42, i64 0, i64 0, !dbg !14
  %t44 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0, !dbg !14
  %t45 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t43, i64 24, ptr noundef %t44, i64 %t41), !dbg !14
  %t46 = call ptr @malloc(i64 512), !dbg !15
  call ptr @strcpy(ptr %t46, ptr %t35), !dbg !15
  call ptr @strcat(ptr %t46, ptr %t43), !dbg !15
  %t47 = call i32 @puts(ptr noundef %t46), !dbg !16
  %t48 = add i64 0, 10, !dbg !17
  %t49 = load ptr, ptr %v_c_str, !dbg !18
  %t50 = getelementptr inbounds %TCounter, ptr %t49, i32 0, i32 0, !dbg !18
  %t51 = load ptr, ptr %t50, !dbg !18
  %t52 = getelementptr inbounds [4 x ptr], ptr %t51, i32 0, i32 1, !dbg !18
  %t53 = load ptr, ptr %t52, !dbg !18
  call void (ptr, i64) %t53(ptr %t49, i64 %t48), !dbg !18
  %t54 = getelementptr inbounds [16 x i8], ptr @.str.2, i64 0, i64 0, !dbg !19
  %t55 = load ptr, ptr %v_c_str, !dbg !20
  %t56 = getelementptr inbounds %TCounter, ptr %t55, i32 0, i32 0, !dbg !20
  %t57 = load ptr, ptr %t56, !dbg !20
  %t58 = getelementptr inbounds [4 x ptr], ptr %t57, i32 0, i32 3, !dbg !20
  %t59 = load ptr, ptr %t58, !dbg !20
  %t60 = call i64 (ptr) %t59(ptr %t55), !dbg !20
  %t61 = alloca [24 x i8], align 1, !dbg !21
  %t62 = getelementptr inbounds [24 x i8], ptr %t61, i64 0, i64 0, !dbg !21
  %t63 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0, !dbg !21
  %t64 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t62, i64 24, ptr noundef %t63, i64 %t60), !dbg !21
  %t65 = call ptr @malloc(i64 512), !dbg !22
  call ptr @strcpy(ptr %t65, ptr %t54), !dbg !22
  call ptr @strcat(ptr %t65, ptr %t62), !dbg !22
  %t66 = call i32 @puts(ptr noundef %t65), !dbg !23
  %t67 = load ptr, ptr %v_c_str, !dbg !24
  %t68 = getelementptr inbounds %TCounter, ptr %t67, i32 0, i32 0, !dbg !24
  %t69 = load ptr, ptr %t68, !dbg !24
  %t70 = getelementptr inbounds [4 x ptr], ptr %t69, i32 0, i32 2, !dbg !24
  %t71 = load ptr, ptr %t70, !dbg !24
  call void (ptr) %t71(ptr %t67), !dbg !24
  %t72 = getelementptr inbounds [14 x i8], ptr @.str.3, i64 0, i64 0, !dbg !25
  %t73 = load ptr, ptr %v_c_str, !dbg !26
  %t74 = getelementptr inbounds %TCounter, ptr %t73, i32 0, i32 0, !dbg !26
  %t75 = load ptr, ptr %t74, !dbg !26
  %t76 = getelementptr inbounds [4 x ptr], ptr %t75, i32 0, i32 3, !dbg !26
  %t77 = load ptr, ptr %t76, !dbg !26
  %t78 = call i64 (ptr) %t77(ptr %t73), !dbg !26
  %t79 = alloca [24 x i8], align 1, !dbg !27
  %t80 = getelementptr inbounds [24 x i8], ptr %t79, i64 0, i64 0, !dbg !27
  %t81 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0, !dbg !27
  %t82 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t80, i64 24, ptr noundef %t81, i64 %t78), !dbg !27
  %t83 = call ptr @malloc(i64 512), !dbg !28
  call ptr @strcpy(ptr %t83, ptr %t72), !dbg !28
  call ptr @strcat(ptr %t83, ptr %t80), !dbg !28
  %t84 = call i32 @puts(ptr noundef %t83), !dbg !29
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [20 x i8] c"After 3 Increment: \00", align 1
@.str.1 = private unnamed_addr constant [5 x i8] c"%lld\00", align 1
@.str.2 = private unnamed_addr constant [16 x i8] c"After Add(10): \00", align 1
@.str.3 = private unnamed_addr constant [14 x i8] c"After Reset: \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example18_class_methods.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/04_oop")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 7, type: !31, scopeLine: 7, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !34)
!5 = !DILocation(line: 36, column: 20, scope: !4)
!6 = !DILocation(line: 36, column: 3, scope: !4)
!7 = !DILocation(line: 37, column: 14, scope: !4)
!8 = !DILocation(line: 37, column: 12, scope: !4)
!9 = !DILocation(line: 39, column: 14, scope: !4)
!10 = !DILocation(line: 40, column: 14, scope: !4)
!11 = !DILocation(line: 41, column: 14, scope: !4)
!12 = !DILocation(line: 42, column: 11, scope: !4)
!13 = !DILocation(line: 42, column: 49, scope: !4)
!14 = !DILocation(line: 42, column: 43, scope: !4)
!15 = !DILocation(line: 42, column: 33, scope: !4)
!16 = !DILocation(line: 42, column: 10, scope: !4)
!17 = !DILocation(line: 44, column: 9, scope: !4)
!18 = !DILocation(line: 44, column: 8, scope: !4)
!19 = !DILocation(line: 45, column: 11, scope: !4)
!20 = !DILocation(line: 45, column: 45, scope: !4)
!21 = !DILocation(line: 45, column: 39, scope: !4)
!22 = !DILocation(line: 45, column: 29, scope: !4)
!23 = !DILocation(line: 45, column: 10, scope: !4)
!24 = !DILocation(line: 47, column: 10, scope: !4)
!25 = !DILocation(line: 48, column: 11, scope: !4)
!26 = !DILocation(line: 48, column: 43, scope: !4)
!27 = !DILocation(line: 48, column: 37, scope: !4)
!28 = !DILocation(line: 48, column: 27, scope: !4)
!29 = !DILocation(line: 48, column: 10, scope: !4)
!30 = !{null}
!31 = !DISubroutineType(types: !30)
!32 = !{}
!33 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!34 = !{}
