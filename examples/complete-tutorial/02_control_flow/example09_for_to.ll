; Kylix LLVM IR — module: ForLoop
source_filename = "ForLoop.klx"
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
  %v_i_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_i_int, !dbg !5
  #dbg_declare(ptr %v_i_int, !6, !DIExpression(), !5)
  %v_sum_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_sum_int, !dbg !5
  #dbg_declare(ptr %v_sum_int, !7, !DIExpression(), !5)
  %v_j_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_j_int, !dbg !5
  #dbg_declare(ptr %v_j_int, !8, !DIExpression(), !5)
  %t0 = getelementptr inbounds [13 x i8], ptr @.str.0, i64 0, i64 0, !dbg !9
  %t1 = call i32 @puts(ptr noundef %t0), !dbg !10
  %t2 = add i64 0, 1, !dbg !11
  store i64 %t2, ptr %v_i_int, !dbg !12
  br label %lbl0, !dbg !12
lbl0:
  %t3 = load i64, ptr %v_i_int, !dbg !12
  %t4 = add i64 0, 5, !dbg !13
  %t5 = icmp sle i64 %t3, %t4, !dbg !12
  br i1 %t5, label %lbl1, label %lbl2, !dbg !12
lbl1:
  %t6 = load i64, ptr %v_i_int, !dbg !14
  %t7 = getelementptr inbounds [6 x i8], ptr @.str.1, i64 0, i64 0, !dbg !15
  %t8 = call i32 (ptr, ...) @printf(ptr noundef %t7, i64 %t6), !dbg !15
  %t10 = load i64, ptr %v_i_int, !dbg !12
  %t9 = add i64 %t10, 1, !dbg !12
  store i64 %t9, ptr %v_i_int, !dbg !12
  br label %lbl0, !dbg !12
lbl2:
  %t11 = getelementptr inbounds [15 x i8], ptr @.str.2, i64 0, i64 0, !dbg !16
  %t12 = call i32 @puts(ptr noundef %t11), !dbg !17
  %t13 = add i64 0, 5, !dbg !18
  store i64 %t13, ptr %v_i_int, !dbg !19
  br label %lbl3, !dbg !19
lbl3:
  %t14 = load i64, ptr %v_i_int, !dbg !19
  %t15 = add i64 0, 1, !dbg !20
  %t16 = icmp sge i64 %t14, %t15, !dbg !19
  br i1 %t16, label %lbl4, label %lbl5, !dbg !19
lbl4:
  %t17 = load i64, ptr %v_i_int, !dbg !21
  %t18 = getelementptr inbounds [6 x i8], ptr @.str.1, i64 0, i64 0, !dbg !22
  %t19 = call i32 (ptr, ...) @printf(ptr noundef %t18, i64 %t17), !dbg !22
  %t21 = load i64, ptr %v_i_int, !dbg !19
  %t20 = sub i64 %t21, 1, !dbg !19
  store i64 %t20, ptr %v_i_int, !dbg !19
  br label %lbl3, !dbg !19
lbl5:
  %t22 = add i64 0, 0, !dbg !23
  store i64 %t22, ptr %v_sum_int, !dbg !24
  %t23 = add i64 0, 1, !dbg !25
  store i64 %t23, ptr %v_i_int, !dbg !26
  br label %lbl6, !dbg !26
lbl6:
  %t24 = load i64, ptr %v_i_int, !dbg !26
  %t25 = add i64 0, 100, !dbg !27
  %t26 = icmp sle i64 %t24, %t25, !dbg !26
  br i1 %t26, label %lbl7, label %lbl8, !dbg !26
lbl7:
  %t27 = load i64, ptr %v_sum_int, !dbg !28
  %t28 = load i64, ptr %v_i_int, !dbg !29
  %t29 = add i64 %t27, %t28, !dbg !30
  store i64 %t29, ptr %v_sum_int, !dbg !31
  %t31 = load i64, ptr %v_i_int, !dbg !26
  %t30 = add i64 %t31, 1, !dbg !26
  store i64 %t30, ptr %v_i_int, !dbg !26
  br label %lbl6, !dbg !26
lbl8:
  %t32 = call ptr @malloc(i64 512), !dbg !32
  store i8 0, ptr %t32, !dbg !32
  %t33 = getelementptr inbounds [4 x i8], ptr @.str.3, i64 0, i64 0, !dbg !32
  %t34 = getelementptr inbounds [15 x i8], ptr @.str.4, i64 0, i64 0, !dbg !33
  %t35 = call ptr @strcat(ptr %t32, ptr %t34), !dbg !32
  %t36 = getelementptr inbounds [2 x i8], ptr @.str.5, i64 0, i64 0, !dbg !32
  %t37 = call ptr @strcat(ptr %t32, ptr %t36), !dbg !32
  %t38 = load i64, ptr %v_sum_int, !dbg !34
  %t39 = call i64 @strlen(ptr %t32), !dbg !32
  %t40 = getelementptr inbounds i8, ptr %t32, i64 %t39, !dbg !32
  %t41 = sub i64 512, %t39, !dbg !32
  %t42 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t40, i64 %t41, ptr %t33, i64 %t38), !dbg !32
  %t43 = call i32 @puts(ptr noundef %t32), !dbg !32
  %t44 = getelementptr inbounds [28 x i8], ptr @.str.6, i64 0, i64 0, !dbg !35
  %t45 = call i32 @puts(ptr noundef %t44), !dbg !36
  %t46 = add i64 0, 1, !dbg !37
  store i64 %t46, ptr %v_i_int, !dbg !38
  br label %lbl9, !dbg !38
lbl9:
  %t47 = load i64, ptr %v_i_int, !dbg !38
  %t48 = add i64 0, 3, !dbg !39
  %t49 = icmp sle i64 %t47, %t48, !dbg !38
  br i1 %t49, label %lbl10, label %lbl11, !dbg !38
lbl10:
  %t50 = add i64 0, 1, !dbg !40
  store i64 %t50, ptr %v_j_int, !dbg !41
  br label %lbl12, !dbg !41
lbl12:
  %t51 = load i64, ptr %v_j_int, !dbg !41
  %t52 = add i64 0, 3, !dbg !42
  %t53 = icmp sle i64 %t51, %t52, !dbg !41
  br i1 %t53, label %lbl13, label %lbl14, !dbg !41
lbl13:
  %t54 = call ptr @malloc(i64 512), !dbg !43
  store i8 0, ptr %t54, !dbg !43
  %t55 = getelementptr inbounds [4 x i8], ptr @.str.3, i64 0, i64 0, !dbg !43
  %t56 = load i64, ptr %v_i_int, !dbg !44
  %t57 = call i64 @strlen(ptr %t54), !dbg !43
  %t58 = getelementptr inbounds i8, ptr %t54, i64 %t57, !dbg !43
  %t59 = sub i64 512, %t57, !dbg !43
  %t60 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t58, i64 %t59, ptr %t55, i64 %t56), !dbg !43
  %t61 = getelementptr inbounds [2 x i8], ptr @.str.5, i64 0, i64 0, !dbg !43
  %t62 = call ptr @strcat(ptr %t54, ptr %t61), !dbg !43
  %t63 = getelementptr inbounds [4 x i8], ptr @.str.7, i64 0, i64 0, !dbg !45
  %t64 = call ptr @strcat(ptr %t54, ptr %t63), !dbg !43
  %t65 = getelementptr inbounds [2 x i8], ptr @.str.5, i64 0, i64 0, !dbg !43
  %t66 = call ptr @strcat(ptr %t54, ptr %t65), !dbg !43
  %t67 = load i64, ptr %v_j_int, !dbg !46
  %t68 = call i64 @strlen(ptr %t54), !dbg !43
  %t69 = getelementptr inbounds i8, ptr %t54, i64 %t68, !dbg !43
  %t70 = sub i64 512, %t68, !dbg !43
  %t71 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t69, i64 %t70, ptr %t55, i64 %t67), !dbg !43
  %t72 = getelementptr inbounds [2 x i8], ptr @.str.5, i64 0, i64 0, !dbg !43
  %t73 = call ptr @strcat(ptr %t54, ptr %t72), !dbg !43
  %t74 = getelementptr inbounds [4 x i8], ptr @.str.8, i64 0, i64 0, !dbg !47
  %t75 = call ptr @strcat(ptr %t54, ptr %t74), !dbg !43
  %t76 = getelementptr inbounds [2 x i8], ptr @.str.5, i64 0, i64 0, !dbg !43
  %t77 = call ptr @strcat(ptr %t54, ptr %t76), !dbg !43
  %t78 = load i64, ptr %v_i_int, !dbg !48
  %t79 = load i64, ptr %v_j_int, !dbg !49
  %t80 = mul i64 %t78, %t79, !dbg !50
  %t81 = call i64 @strlen(ptr %t54), !dbg !43
  %t82 = getelementptr inbounds i8, ptr %t54, i64 %t81, !dbg !43
  %t83 = sub i64 512, %t81, !dbg !43
  %t84 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t82, i64 %t83, ptr %t55, i64 %t80), !dbg !43
  %t85 = call i32 @puts(ptr noundef %t54), !dbg !43
  %t87 = load i64, ptr %v_j_int, !dbg !41
  %t86 = add i64 %t87, 1, !dbg !41
  store i64 %t86, ptr %v_j_int, !dbg !41
  br label %lbl12, !dbg !41
lbl14:
  %t89 = load i64, ptr %v_i_int, !dbg !38
  %t88 = add i64 %t89, 1, !dbg !38
  store i64 %t88, ptr %v_i_int, !dbg !38
  br label %lbl9, !dbg !38
lbl11:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [13 x i8] c"Counting up:\00", align 1
@.str.1 = private unnamed_addr constant [6 x i8] c"%lld\0A\00", align 1
@.str.2 = private unnamed_addr constant [15 x i8] c"Counting down:\00", align 1
@.str.3 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.4 = private unnamed_addr constant [15 x i8] c"Sum of 1-100: \00", align 1
@.str.5 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.6 = private unnamed_addr constant [28 x i8] c"Multiplication table (3x3):\00", align 1
@.str.7 = private unnamed_addr constant [4 x i8] c" x \00", align 1
@.str.8 = private unnamed_addr constant [4 x i8] c" = \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example09_for_to.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/02_control_flow")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !52, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !55)
!6 = !DILocalVariable(name: "i", scope: !4, file: !3, line: 4, type: !54)
!7 = !DILocalVariable(name: "sum", scope: !4, file: !3, line: 5, type: !54)
!8 = !DILocalVariable(name: "j", scope: !4, file: !3, line: 6, type: !54)
!5 = !DILocation(line: 1, column: 9, scope: !4)
!9 = !DILocation(line: 10, column: 11, scope: !4)
!10 = !DILocation(line: 10, column: 10, scope: !4)
!11 = !DILocation(line: 11, column: 12, scope: !4)
!12 = !DILocation(line: 11, column: 3, scope: !4)
!13 = !DILocation(line: 11, column: 17, scope: !4)
!14 = !DILocation(line: 13, column: 13, scope: !4)
!15 = !DILocation(line: 13, column: 12, scope: !4)
!16 = !DILocation(line: 17, column: 11, scope: !4)
!17 = !DILocation(line: 17, column: 10, scope: !4)
!18 = !DILocation(line: 18, column: 12, scope: !4)
!19 = !DILocation(line: 18, column: 3, scope: !4)
!20 = !DILocation(line: 18, column: 21, scope: !4)
!21 = !DILocation(line: 20, column: 13, scope: !4)
!22 = !DILocation(line: 20, column: 12, scope: !4)
!23 = !DILocation(line: 24, column: 10, scope: !4)
!24 = !DILocation(line: 24, column: 8, scope: !4)
!25 = !DILocation(line: 25, column: 12, scope: !4)
!26 = !DILocation(line: 25, column: 3, scope: !4)
!27 = !DILocation(line: 25, column: 17, scope: !4)
!28 = !DILocation(line: 27, column: 12, scope: !4)
!29 = !DILocation(line: 27, column: 18, scope: !4)
!30 = !DILocation(line: 27, column: 16, scope: !4)
!31 = !DILocation(line: 27, column: 10, scope: !4)
!32 = !DILocation(line: 29, column: 10, scope: !4)
!33 = !DILocation(line: 29, column: 11, scope: !4)
!34 = !DILocation(line: 29, column: 29, scope: !4)
!35 = !DILocation(line: 32, column: 11, scope: !4)
!36 = !DILocation(line: 32, column: 10, scope: !4)
!37 = !DILocation(line: 33, column: 12, scope: !4)
!38 = !DILocation(line: 33, column: 3, scope: !4)
!39 = !DILocation(line: 33, column: 17, scope: !4)
!40 = !DILocation(line: 35, column: 14, scope: !4)
!41 = !DILocation(line: 35, column: 5, scope: !4)
!42 = !DILocation(line: 35, column: 19, scope: !4)
!43 = !DILocation(line: 37, column: 14, scope: !4)
!44 = !DILocation(line: 37, column: 15, scope: !4)
!45 = !DILocation(line: 37, column: 18, scope: !4)
!46 = !DILocation(line: 37, column: 25, scope: !4)
!47 = !DILocation(line: 37, column: 28, scope: !4)
!48 = !DILocation(line: 37, column: 35, scope: !4)
!49 = !DILocation(line: 37, column: 39, scope: !4)
!50 = !DILocation(line: 37, column: 37, scope: !4)
!51 = !{null}
!52 = !DISubroutineType(types: !51)
!53 = !{}
!54 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!55 = !{!6, !7, !8}
