; Kylix LLVM IR — module: TryExcept
source_filename = "TryExcept.klx"
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
define double @SafeDivide(double %a, double %b) !dbg !4 {
entry:
  %result = alloca double, align 8
  #dbg_declare(ptr %result, !5, !DIExpression(), !6)
  %v_a_real = alloca double, align 8, !dbg !6
  store double %a, ptr %v_a_real, !dbg !6
  #dbg_declare(ptr %v_a_real, !7, !DIExpression(), !6)
  %v_b_real = alloca double, align 8, !dbg !6
  store double %b, ptr %v_b_real, !dbg !6
  #dbg_declare(ptr %v_b_real, !8, !DIExpression(), !6)
  %t0 = load double, ptr %v_b_real, !dbg !9
  %t1 = add i64 0, 0, !dbg !10
  %t2 = sitofp i64 %t1 to double, !dbg !11
  %t3 = fcmp oeq double %t0, %t2, !dbg !11
  br i1 %t3, label %lbl0, label %lbl1, !dbg !12
lbl0:
  %t4 = call ptr @malloc(i64 16), !dbg !13
  %t5 = getelementptr inbounds [17 x i8], ptr @.str.0, i64 0, i64 0, !dbg !14
  %t6 = getelementptr inbounds %Exception, ptr %t4, i32 0, i32 1, !dbg !13
  store ptr %t5, ptr %t6, !dbg !13
  store ptr %t4, ptr @__kylix_exc_obj, !dbg !15
  store i32 0, ptr @__kylix_exc_type, !dbg !15
  store i1 true, ptr @__kylix_exc_active, !dbg !15
  %t7 = load ptr, ptr @__kylix_jmpbuf, !dbg !15
  %t8 = icmp ne ptr %t7, null, !dbg !15
  br i1 %t8, label %lbl2, label %lbl3, !dbg !15
lbl2:
  call void @longjmp(ptr %t7, i32 1), !dbg !15
  unreachable, !dbg !15
lbl3:
  call void @exit(i32 70), !dbg !15
  unreachable, !dbg !15
  br label %lbl1, !dbg !12
lbl1:
  %t9 = load double, ptr %v_a_real, !dbg !16
  %t10 = load double, ptr %v_b_real, !dbg !17
  %t11 = fdiv double %t9, %t10, !dbg !18
  store double %t11, ptr %result, !dbg !19
  %t12 = load double, ptr %result, !dbg !6
  ret double %t12, !dbg !6
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
define i32 @main() !dbg !20 {
entry:
  %t13 = alloca [288 x i8], align 16, !dbg !21
  %t14 = getelementptr [288 x i8], ptr %t13, i64 0, i64 0, !dbg !21
  %t15 = alloca ptr, align 8, !dbg !21
  %t16 = call i32 @setjmp(ptr %t14), !dbg !21
  %t17 = icmp ne i32 %t16, 0, !dbg !21
  br i1 %t17, label %lbl5, label %lbl4, !dbg !21
lbl4:
  %t18 = load ptr, ptr @__kylix_jmpbuf, !dbg !21
  store ptr %t18, ptr %t15, !dbg !21
  store ptr %t14, ptr @__kylix_jmpbuf, !dbg !21
  store i1 false, ptr @__kylix_exc_active, !dbg !21
  %t19 = add i64 0, 10, !dbg !22
  %t20 = sitofp i64 %t19 to double, !dbg !23
  %t21 = add i64 0, 2, !dbg !24
  %t22 = sitofp i64 %t21 to double, !dbg !23
  %t23 = call double @SafeDivide(double %t20, double %t22), !dbg !23
  %v_result_real = alloca double, align 8, !dbg !25
  store double %t23, ptr %v_result_real, !dbg !25
  %t24 = call ptr @malloc(i64 512), !dbg !26
  store i8 0, ptr %t24, !dbg !26
  %t26 = getelementptr inbounds [10 x i8], ptr @.str.2, i64 0, i64 0, !dbg !27
  %t27 = call ptr @strcat(ptr %t24, ptr %t26), !dbg !26
  %t28 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !26
  %t29 = call ptr @strcat(ptr %t24, ptr %t28), !dbg !26
  %t30 = load double, ptr %v_result_real, !dbg !28
  %t31 = getelementptr inbounds [6 x i8], ptr @.str.4, i64 0, i64 0, !dbg !26
  %t32 = call i64 @strlen(ptr %t24), !dbg !26
  %t33 = getelementptr inbounds i8, ptr %t24, i64 %t32, !dbg !26
  %t34 = sub i64 512, %t32, !dbg !26
  %t35 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t33, i64 %t34, ptr %t31, double %t30), !dbg !26
  %t36 = call i32 @puts(ptr noundef %t24), !dbg !26
  %t37 = load ptr, ptr %t15, !dbg !21
  store ptr %t37, ptr @__kylix_jmpbuf, !dbg !21
  store i1 false, ptr @__kylix_exc_active, !dbg !21
  br label %lbl6, !dbg !21
lbl5:
  %t38 = load ptr, ptr %t15, !dbg !21
  store ptr %t38, ptr @__kylix_jmpbuf, !dbg !21
  %t40 = getelementptr inbounds [18 x i8], ptr @.str.5, i64 0, i64 0, !dbg !29
  %t41 = call i32 @puts(ptr noundef %t40), !dbg !30
  store i1 false, ptr @__kylix_exc_active, !dbg !21
  br label %lbl7, !dbg !21
lbl6:
  br label %lbl9, !dbg !21
lbl7:
  br label %lbl9, !dbg !21
lbl8:
  %t42 = load ptr, ptr @__kylix_jmpbuf, !dbg !21
  call void @longjmp(ptr %t42, i32 1), !dbg !21
  unreachable, !dbg !21
lbl9:
  %t43 = alloca [288 x i8], align 16, !dbg !31
  %t44 = getelementptr [288 x i8], ptr %t43, i64 0, i64 0, !dbg !31
  %t45 = alloca ptr, align 8, !dbg !31
  %t46 = call i32 @setjmp(ptr %t44), !dbg !31
  %t47 = icmp ne i32 %t46, 0, !dbg !31
  br i1 %t47, label %lbl11, label %lbl10, !dbg !31
lbl10:
  %t48 = load ptr, ptr @__kylix_jmpbuf, !dbg !31
  store ptr %t48, ptr %t45, !dbg !31
  store ptr %t44, ptr @__kylix_jmpbuf, !dbg !31
  store i1 false, ptr @__kylix_exc_active, !dbg !31
  %t49 = add i64 0, 10, !dbg !32
  %t50 = sitofp i64 %t49 to double, !dbg !33
  %t51 = add i64 0, 0, !dbg !34
  %t52 = sitofp i64 %t51 to double, !dbg !33
  %t53 = call double @SafeDivide(double %t50, double %t52), !dbg !33
  %v_result_1_real = alloca double, align 8, !dbg !35
  store double %t53, ptr %v_result_1_real, !dbg !35
  %t54 = call ptr @malloc(i64 512), !dbg !36
  store i8 0, ptr %t54, !dbg !36
  %t56 = getelementptr inbounds [9 x i8], ptr @.str.6, i64 0, i64 0, !dbg !37
  %t57 = call ptr @strcat(ptr %t54, ptr %t56), !dbg !36
  %t58 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !36
  %t59 = call ptr @strcat(ptr %t54, ptr %t58), !dbg !36
  %t60 = load double, ptr %v_result_1_real, !dbg !38
  %t61 = getelementptr inbounds [6 x i8], ptr @.str.4, i64 0, i64 0, !dbg !36
  %t62 = call i64 @strlen(ptr %t54), !dbg !36
  %t63 = getelementptr inbounds i8, ptr %t54, i64 %t62, !dbg !36
  %t64 = sub i64 512, %t62, !dbg !36
  %t65 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t63, i64 %t64, ptr %t61, double %t60), !dbg !36
  %t66 = call i32 @puts(ptr noundef %t54), !dbg !36
  %t67 = load ptr, ptr %t45, !dbg !31
  store ptr %t67, ptr @__kylix_jmpbuf, !dbg !31
  store i1 false, ptr @__kylix_exc_active, !dbg !31
  br label %lbl12, !dbg !31
lbl11:
  %t68 = load ptr, ptr %t45, !dbg !31
  store ptr %t68, ptr @__kylix_jmpbuf, !dbg !31
  %t70 = getelementptr inbounds [23 x i8], ptr @.str.7, i64 0, i64 0, !dbg !39
  %t71 = call i32 @puts(ptr noundef %t70), !dbg !40
  store i1 false, ptr @__kylix_exc_active, !dbg !31
  br label %lbl13, !dbg !31
lbl12:
  br label %lbl15, !dbg !31
lbl13:
  br label %lbl15, !dbg !31
lbl14:
  %t72 = load ptr, ptr @__kylix_jmpbuf, !dbg !31
  call void @longjmp(ptr %t72, i32 1), !dbg !31
  unreachable, !dbg !31
lbl15:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [17 x i8] c"Division by zero\00", align 1
@.str.1 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.2 = private unnamed_addr constant [10 x i8] c"10 / 2 = \00", align 1
@.str.3 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.4 = private unnamed_addr constant [6 x i8] c"%.15g\00", align 1
@.str.5 = private unnamed_addr constant [18 x i8] c"An error occurred\00", align 1
@.str.6 = private unnamed_addr constant [9 x i8] c"Result: \00", align 1
@.str.7 = private unnamed_addr constant [23 x i8] c"Cannot divide by zero!\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example27_try_except.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/10_exceptions")
!4 = distinct !DISubprogram(name: "SafeDivide", scope: !3, file: !3, line: 4, type: !42, scopeLine: 4, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !45)
!20 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !42, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !46)
!5 = !DILocalVariable(name: "result", scope: !4, file: !3, line: 4, type: !44)
!7 = !DILocalVariable(name: "a", scope: !4, file: !3, line: 4, type: !44)
!8 = !DILocalVariable(name: "b", scope: !4, file: !3, line: 4, type: !44)
!6 = !DILocation(line: 4, column: 1, scope: !4)
!9 = !DILocation(line: 6, column: 6, scope: !4)
!10 = !DILocation(line: 6, column: 10, scope: !4)
!11 = !DILocation(line: 6, column: 8, scope: !4)
!12 = !DILocation(line: 6, column: 3, scope: !4)
!13 = !DILocation(line: 8, column: 27, scope: !4)
!14 = !DILocation(line: 8, column: 28, scope: !4)
!15 = !DILocation(line: 8, column: 5, scope: !4)
!16 = !DILocation(line: 10, column: 13, scope: !4)
!17 = !DILocation(line: 10, column: 17, scope: !4)
!18 = !DILocation(line: 10, column: 15, scope: !4)
!19 = !DILocation(line: 10, column: 11, scope: !4)
!21 = !DILocation(line: 15, column: 3, scope: !20)
!22 = !DILocation(line: 17, column: 30, scope: !20)
!23 = !DILocation(line: 17, column: 29, scope: !20)
!24 = !DILocation(line: 17, column: 34, scope: !20)
!25 = !DILocation(line: 17, column: 5, scope: !20)
!26 = !DILocation(line: 18, column: 12, scope: !20)
!27 = !DILocation(line: 18, column: 13, scope: !20)
!28 = !DILocation(line: 18, column: 26, scope: !20)
!29 = !DILocation(line: 22, column: 13, scope: !20)
!30 = !DILocation(line: 22, column: 12, scope: !20)
!31 = !DILocation(line: 27, column: 3, scope: !20)
!32 = !DILocation(line: 29, column: 30, scope: !20)
!33 = !DILocation(line: 29, column: 29, scope: !20)
!34 = !DILocation(line: 29, column: 34, scope: !20)
!35 = !DILocation(line: 29, column: 5, scope: !20)
!36 = !DILocation(line: 30, column: 12, scope: !20)
!37 = !DILocation(line: 30, column: 13, scope: !20)
!38 = !DILocation(line: 30, column: 25, scope: !20)
!39 = !DILocation(line: 34, column: 13, scope: !20)
!40 = !DILocation(line: 34, column: 12, scope: !20)
!41 = !{null}
!42 = !DISubroutineType(types: !41)
!43 = !{}
!44 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!45 = !{!5, !7, !8}
!46 = !{}
