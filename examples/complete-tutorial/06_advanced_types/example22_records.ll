; Kylix LLVM IR — module: Records
source_filename = "Records.klx"
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
  %v_point_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_point_int, !dbg !5
  #dbg_declare(ptr %v_point_int, !6, !DIExpression(), !5)
  %v_person_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_person_int, !dbg !5
  #dbg_declare(ptr %v_person_int, !7, !DIExpression(), !5)
  %t0 = fadd double 0.0, 10.500000, !dbg !8
  ; unhandled member assignment TPoint.X
  %t1 = fadd double 0.0, 20.300000, !dbg !10
  ; unhandled member assignment TPoint.Y
  %t2 = call ptr @malloc(i64 512), !dbg !12
  store i8 0, ptr %t2, !dbg !12
  %t3 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !12
  %t4 = getelementptr inbounds [9 x i8], ptr @.str.1, i64 0, i64 0, !dbg !13
  %t5 = call ptr @strcat(ptr %t2, ptr %t4), !dbg !12
  %t6 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !12
  %t7 = call ptr @strcat(ptr %t2, ptr %t6), !dbg !12
  %t8 = add i64 0, 0 ; member access on non-class TPoint.X, !dbg !14
  %t9 = call i64 @strlen(ptr %t2), !dbg !12
  %t10 = getelementptr inbounds i8, ptr %t2, i64 %t9, !dbg !12
  %t11 = sub i64 512, %t9, !dbg !12
  %t12 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t10, i64 %t11, ptr %t3, i64 %t8), !dbg !12
  %t13 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !12
  %t14 = call ptr @strcat(ptr %t2, ptr %t13), !dbg !12
  %t15 = getelementptr inbounds [3 x i8], ptr @.str.3, i64 0, i64 0, !dbg !15
  %t16 = call ptr @strcat(ptr %t2, ptr %t15), !dbg !12
  %t17 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !12
  %t18 = call ptr @strcat(ptr %t2, ptr %t17), !dbg !12
  %t19 = add i64 0, 0 ; member access on non-class TPoint.Y, !dbg !16
  %t20 = call i64 @strlen(ptr %t2), !dbg !12
  %t21 = getelementptr inbounds i8, ptr %t2, i64 %t20, !dbg !12
  %t22 = sub i64 512, %t20, !dbg !12
  %t23 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t21, i64 %t22, ptr %t3, i64 %t19), !dbg !12
  %t24 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !12
  %t25 = call ptr @strcat(ptr %t2, ptr %t24), !dbg !12
  %t26 = getelementptr inbounds [2 x i8], ptr @.str.4, i64 0, i64 0, !dbg !17
  %t27 = call ptr @strcat(ptr %t2, ptr %t26), !dbg !12
  %t28 = call i32 @puts(ptr noundef %t2), !dbg !12
  ; unhandled member assignment TPerson.Name
  ; unhandled member assignment TPerson.Age
  ; unhandled member assignment TPerson.Email
  %t32 = call ptr @malloc(i64 512), !dbg !24
  store i8 0, ptr %t32, !dbg !24
  %t33 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !24
  %t34 = getelementptr inbounds [9 x i8], ptr @.str.7, i64 0, i64 0, !dbg !25
  %t35 = call ptr @strcat(ptr %t32, ptr %t34), !dbg !24
  %t36 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !24
  %t37 = call ptr @strcat(ptr %t32, ptr %t36), !dbg !24
  %t38 = add i64 0, 0 ; member access on non-class TPerson.Name, !dbg !26
  %t39 = call i64 @strlen(ptr %t32), !dbg !24
  %t40 = getelementptr inbounds i8, ptr %t32, i64 %t39, !dbg !24
  %t41 = sub i64 512, %t39, !dbg !24
  %t42 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t40, i64 %t41, ptr %t33, i64 %t38), !dbg !24
  %t43 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !24
  %t44 = call ptr @strcat(ptr %t32, ptr %t43), !dbg !24
  %t45 = getelementptr inbounds [8 x i8], ptr @.str.8, i64 0, i64 0, !dbg !27
  %t46 = call ptr @strcat(ptr %t32, ptr %t45), !dbg !24
  %t47 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !24
  %t48 = call ptr @strcat(ptr %t32, ptr %t47), !dbg !24
  %t49 = add i64 0, 0 ; member access on non-class TPerson.Age, !dbg !28
  %t50 = call i64 @strlen(ptr %t32), !dbg !24
  %t51 = getelementptr inbounds i8, ptr %t32, i64 %t50, !dbg !24
  %t52 = sub i64 512, %t50, !dbg !24
  %t53 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t51, i64 %t52, ptr %t33, i64 %t49), !dbg !24
  %t54 = call i32 @puts(ptr noundef %t32), !dbg !24
  %t55 = call ptr @malloc(i64 512), !dbg !29
  store i8 0, ptr %t55, !dbg !29
  %t56 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !29
  %t57 = getelementptr inbounds [8 x i8], ptr @.str.9, i64 0, i64 0, !dbg !30
  %t58 = call ptr @strcat(ptr %t55, ptr %t57), !dbg !29
  %t59 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !29
  %t60 = call ptr @strcat(ptr %t55, ptr %t59), !dbg !29
  %t61 = add i64 0, 0 ; member access on non-class TPerson.Email, !dbg !31
  %t62 = call i64 @strlen(ptr %t55), !dbg !29
  %t63 = getelementptr inbounds i8, ptr %t55, i64 %t62, !dbg !29
  %t64 = sub i64 512, %t62, !dbg !29
  %t65 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t63, i64 %t64, ptr %t56, i64 %t61), !dbg !29
  %t66 = call i32 @puts(ptr noundef %t55), !dbg !29
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [9 x i8] c"Point: (\00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [3 x i8] c", \00", align 1
@.str.4 = private unnamed_addr constant [2 x i8] c")\00", align 1
@.str.5 = private unnamed_addr constant [6 x i8] c"Alice\00", align 1
@.str.6 = private unnamed_addr constant [18 x i8] c"alice@example.com\00", align 1
@.str.7 = private unnamed_addr constant [9 x i8] c"Person: \00", align 1
@.str.8 = private unnamed_addr constant [8 x i8] c", Age: \00", align 1
@.str.9 = private unnamed_addr constant [8 x i8] c"Email: \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example22_records.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/06_advanced_types")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !33, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !36)
!6 = !DILocalVariable(name: "point", scope: !4, file: !3, line: 17, type: !35)
!7 = !DILocalVariable(name: "person", scope: !4, file: !3, line: 18, type: !35)
!5 = !DILocation(line: 1, column: 9, scope: !4)
!8 = !DILocation(line: 22, column: 14, scope: !4)
!9 = !DILocation(line: 22, column: 12, scope: !4)
!10 = !DILocation(line: 23, column: 14, scope: !4)
!11 = !DILocation(line: 23, column: 12, scope: !4)
!12 = !DILocation(line: 25, column: 10, scope: !4)
!13 = !DILocation(line: 25, column: 11, scope: !4)
!14 = !DILocation(line: 25, column: 28, scope: !4)
!15 = !DILocation(line: 25, column: 32, scope: !4)
!16 = !DILocation(line: 25, column: 43, scope: !4)
!17 = !DILocation(line: 25, column: 47, scope: !4)
!18 = !DILocation(line: 28, column: 18, scope: !4)
!19 = !DILocation(line: 28, column: 16, scope: !4)
!20 = !DILocation(line: 29, column: 17, scope: !4)
!21 = !DILocation(line: 29, column: 15, scope: !4)
!22 = !DILocation(line: 30, column: 19, scope: !4)
!23 = !DILocation(line: 30, column: 17, scope: !4)
!24 = !DILocation(line: 32, column: 10, scope: !4)
!25 = !DILocation(line: 32, column: 11, scope: !4)
!26 = !DILocation(line: 32, column: 29, scope: !4)
!27 = !DILocation(line: 32, column: 36, scope: !4)
!28 = !DILocation(line: 32, column: 53, scope: !4)
!29 = !DILocation(line: 33, column: 10, scope: !4)
!30 = !DILocation(line: 33, column: 11, scope: !4)
!31 = !DILocation(line: 33, column: 28, scope: !4)
!32 = !{null}
!33 = !DISubroutineType(types: !32)
!34 = !{}
!35 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!36 = !{!6, !7}
