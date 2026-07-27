; Kylix LLVM IR — module: BasicFunctions
source_filename = "BasicFunctions.klx"
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
define i64 @Max(i64 %a, i64 %b) !dbg !4 {
entry:
  %result = alloca i64, align 8
  #dbg_declare(ptr %result, !5, !DIExpression(), !6)
  %v_a_int = alloca i64, align 8, !dbg !6
  store i64 %a, ptr %v_a_int, !dbg !6
  #dbg_declare(ptr %v_a_int, !7, !DIExpression(), !6)
  %v_b_int = alloca i64, align 8, !dbg !6
  store i64 %b, ptr %v_b_int, !dbg !6
  #dbg_declare(ptr %v_b_int, !8, !DIExpression(), !6)
  %t0 = load i64, ptr %v_a_int, !dbg !9
  %t1 = load i64, ptr %v_b_int, !dbg !10
  %t2 = icmp sgt i64 %t0, %t1, !dbg !11
  br i1 %t2, label %lbl0, label %lbl2, !dbg !12
lbl0:
  %t3 = load i64, ptr %v_a_int, !dbg !13
  store i64 %t3, ptr %result, !dbg !14
  br label %lbl1, !dbg !12
lbl2:
  %t4 = load i64, ptr %v_b_int, !dbg !15
  store i64 %t4, ptr %result, !dbg !16
  br label %lbl1, !dbg !12
lbl1:
  %t5 = load i64, ptr %result, !dbg !6
  ret i64 %t5, !dbg !6
}

define i64 @Min(i64 %a, i64 %b) !dbg !17 {
entry:
  %result = alloca i64, align 8
  #dbg_declare(ptr %result, !18, !DIExpression(), !19)
  %v_a_int = alloca i64, align 8, !dbg !19
  store i64 %a, ptr %v_a_int, !dbg !19
  #dbg_declare(ptr %v_a_int, !20, !DIExpression(), !19)
  %v_b_int = alloca i64, align 8, !dbg !19
  store i64 %b, ptr %v_b_int, !dbg !19
  #dbg_declare(ptr %v_b_int, !21, !DIExpression(), !19)
  %t6 = load i64, ptr %v_a_int, !dbg !22
  %t7 = load i64, ptr %v_b_int, !dbg !23
  %t8 = icmp slt i64 %t6, %t7, !dbg !24
  br i1 %t8, label %lbl3, label %lbl5, !dbg !25
lbl3:
  %t9 = load i64, ptr %v_a_int, !dbg !26
  store i64 %t9, ptr %result, !dbg !27
  br label %lbl4, !dbg !25
lbl5:
  %t10 = load i64, ptr %v_b_int, !dbg !28
  store i64 %t10, ptr %result, !dbg !29
  br label %lbl4, !dbg !25
lbl4:
  %t11 = load i64, ptr %result, !dbg !19
  ret i64 %t11, !dbg !19
}

define i64 @Abs(i64 %n) !dbg !30 {
entry:
  %result = alloca i64, align 8
  #dbg_declare(ptr %result, !31, !DIExpression(), !32)
  %v_n_int = alloca i64, align 8, !dbg !32
  store i64 %n, ptr %v_n_int, !dbg !32
  #dbg_declare(ptr %v_n_int, !33, !DIExpression(), !32)
  %t12 = load i64, ptr %v_n_int, !dbg !34
  %t13 = add i64 0, 0, !dbg !35
  %t14 = icmp slt i64 %t12, %t13, !dbg !36
  br i1 %t14, label %lbl6, label %lbl8, !dbg !37
lbl6:
  %t15 = load i64, ptr %v_n_int, !dbg !38
  %t16 = sub i64 0, %t15, !dbg !39
  store i64 %t16, ptr %result, !dbg !40
  br label %lbl7, !dbg !37
lbl8:
  %t17 = load i64, ptr %v_n_int, !dbg !41
  store i64 %t17, ptr %result, !dbg !42
  br label %lbl7, !dbg !37
lbl7:
  %t18 = load i64, ptr %result, !dbg !32
  ret i64 %t18, !dbg !32
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
define i32 @main() !dbg !43 {
entry:
  %t19 = call ptr @malloc(i64 512), !dbg !44
  store i8 0, ptr %t19, !dbg !44
  %t20 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !44
  %t21 = getelementptr inbounds [14 x i8], ptr @.str.1, i64 0, i64 0, !dbg !45
  %t22 = call ptr @strcat(ptr %t19, ptr %t21), !dbg !44
  %t23 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !44
  %t24 = call ptr @strcat(ptr %t19, ptr %t23), !dbg !44
  %t25 = add i64 0, 10, !dbg !46
  %t26 = add i64 0, 20, !dbg !47
  %t27 = call i64 @Max(i64 %t25, i64 %t26), !dbg !48
  %t28 = call i64 @strlen(ptr %t19), !dbg !44
  %t29 = getelementptr inbounds i8, ptr %t19, i64 %t28, !dbg !44
  %t30 = sub i64 512, %t28, !dbg !44
  %t31 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t29, i64 %t30, ptr %t20, i64 %t27), !dbg !44
  %t32 = call i32 @puts(ptr noundef %t19), !dbg !44
  %t33 = call ptr @malloc(i64 512), !dbg !49
  store i8 0, ptr %t33, !dbg !49
  %t34 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !49
  %t35 = getelementptr inbounds [14 x i8], ptr @.str.3, i64 0, i64 0, !dbg !50
  %t36 = call ptr @strcat(ptr %t33, ptr %t35), !dbg !49
  %t37 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !49
  %t38 = call ptr @strcat(ptr %t33, ptr %t37), !dbg !49
  %t39 = add i64 0, 10, !dbg !51
  %t40 = add i64 0, 20, !dbg !52
  %t41 = call i64 @Min(i64 %t39, i64 %t40), !dbg !53
  %t42 = call i64 @strlen(ptr %t33), !dbg !49
  %t43 = getelementptr inbounds i8, ptr %t33, i64 %t42, !dbg !49
  %t44 = sub i64 512, %t42, !dbg !49
  %t45 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t43, i64 %t44, ptr %t34, i64 %t41), !dbg !49
  %t46 = call i32 @puts(ptr noundef %t33), !dbg !49
  %t47 = call ptr @malloc(i64 512), !dbg !54
  store i8 0, ptr %t47, !dbg !54
  %t48 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !54
  %t49 = getelementptr inbounds [11 x i8], ptr @.str.4, i64 0, i64 0, !dbg !55
  %t50 = call ptr @strcat(ptr %t47, ptr %t49), !dbg !54
  %t51 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !54
  %t52 = call ptr @strcat(ptr %t47, ptr %t51), !dbg !54
  %t53 = add i64 0, 15, !dbg !56
  %t54 = sub i64 0, %t53, !dbg !57
  %t55 = call i64 @Abs(i64 %t54), !dbg !58
  %t56 = call i64 @strlen(ptr %t47), !dbg !54
  %t57 = getelementptr inbounds i8, ptr %t47, i64 %t56, !dbg !54
  %t58 = sub i64 512, %t56, !dbg !54
  %t59 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t57, i64 %t58, ptr %t48, i64 %t55), !dbg !54
  %t60 = call i32 @puts(ptr noundef %t47), !dbg !54
  %t61 = call ptr @malloc(i64 512), !dbg !59
  store i8 0, ptr %t61, !dbg !59
  %t62 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !59
  %t63 = getelementptr inbounds [10 x i8], ptr @.str.5, i64 0, i64 0, !dbg !60
  %t64 = call ptr @strcat(ptr %t61, ptr %t63), !dbg !59
  %t65 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !59
  %t66 = call ptr @strcat(ptr %t61, ptr %t65), !dbg !59
  %t67 = add i64 0, 25, !dbg !61
  %t68 = call i64 @Abs(i64 %t67), !dbg !62
  %t69 = call i64 @strlen(ptr %t61), !dbg !59
  %t70 = getelementptr inbounds i8, ptr %t61, i64 %t69, !dbg !59
  %t71 = sub i64 512, %t69, !dbg !59
  %t72 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t70, i64 %t71, ptr %t62, i64 %t68), !dbg !59
  %t73 = call i32 @puts(ptr noundef %t61), !dbg !59
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"Max(10, 20): \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [14 x i8] c"Min(10, 20): \00", align 1
@.str.4 = private unnamed_addr constant [11 x i8] c"Abs(-15): \00", align 1
@.str.5 = private unnamed_addr constant [10 x i8] c"Abs(25): \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example29_basic_funcs.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/07_stdlib_core")
!4 = distinct !DISubprogram(name: "Max", scope: !3, file: !3, line: 4, type: !64, scopeLine: 4, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !67)
!17 = distinct !DISubprogram(name: "Min", scope: !3, file: !3, line: 12, type: !64, scopeLine: 12, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !68)
!30 = distinct !DISubprogram(name: "Abs", scope: !3, file: !3, line: 20, type: !64, scopeLine: 20, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !69)
!43 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !64, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !70)
!5 = !DILocalVariable(name: "result", scope: !4, file: !3, line: 4, type: !66)
!7 = !DILocalVariable(name: "a", scope: !4, file: !3, line: 4, type: !66)
!8 = !DILocalVariable(name: "b", scope: !4, file: !3, line: 4, type: !66)
!18 = !DILocalVariable(name: "result", scope: !17, file: !3, line: 12, type: !66)
!20 = !DILocalVariable(name: "a", scope: !17, file: !3, line: 12, type: !66)
!21 = !DILocalVariable(name: "b", scope: !17, file: !3, line: 12, type: !66)
!31 = !DILocalVariable(name: "result", scope: !30, file: !3, line: 20, type: !66)
!33 = !DILocalVariable(name: "n", scope: !30, file: !3, line: 20, type: !66)
!6 = !DILocation(line: 4, column: 1, scope: !4)
!9 = !DILocation(line: 6, column: 6, scope: !4)
!10 = !DILocation(line: 6, column: 10, scope: !4)
!11 = !DILocation(line: 6, column: 8, scope: !4)
!12 = !DILocation(line: 6, column: 3, scope: !4)
!13 = !DILocation(line: 7, column: 15, scope: !4)
!14 = !DILocation(line: 7, column: 13, scope: !4)
!15 = !DILocation(line: 9, column: 15, scope: !4)
!16 = !DILocation(line: 9, column: 13, scope: !4)
!19 = !DILocation(line: 12, column: 1, scope: !17)
!22 = !DILocation(line: 14, column: 6, scope: !17)
!23 = !DILocation(line: 14, column: 10, scope: !17)
!24 = !DILocation(line: 14, column: 8, scope: !17)
!25 = !DILocation(line: 14, column: 3, scope: !17)
!26 = !DILocation(line: 15, column: 15, scope: !17)
!27 = !DILocation(line: 15, column: 13, scope: !17)
!28 = !DILocation(line: 17, column: 15, scope: !17)
!29 = !DILocation(line: 17, column: 13, scope: !17)
!32 = !DILocation(line: 20, column: 1, scope: !30)
!34 = !DILocation(line: 22, column: 6, scope: !30)
!35 = !DILocation(line: 22, column: 10, scope: !30)
!36 = !DILocation(line: 22, column: 8, scope: !30)
!37 = !DILocation(line: 22, column: 3, scope: !30)
!38 = !DILocation(line: 23, column: 16, scope: !30)
!39 = !DILocation(line: 23, column: 15, scope: !30)
!40 = !DILocation(line: 23, column: 13, scope: !30)
!41 = !DILocation(line: 25, column: 15, scope: !30)
!42 = !DILocation(line: 25, column: 13, scope: !30)
!44 = !DILocation(line: 29, column: 10, scope: !43)
!45 = !DILocation(line: 29, column: 11, scope: !43)
!46 = !DILocation(line: 29, column: 32, scope: !43)
!47 = !DILocation(line: 29, column: 36, scope: !43)
!48 = !DILocation(line: 29, column: 31, scope: !43)
!49 = !DILocation(line: 30, column: 10, scope: !43)
!50 = !DILocation(line: 30, column: 11, scope: !43)
!51 = !DILocation(line: 30, column: 32, scope: !43)
!52 = !DILocation(line: 30, column: 36, scope: !43)
!53 = !DILocation(line: 30, column: 31, scope: !43)
!54 = !DILocation(line: 31, column: 10, scope: !43)
!55 = !DILocation(line: 31, column: 11, scope: !43)
!56 = !DILocation(line: 31, column: 30, scope: !43)
!57 = !DILocation(line: 31, column: 29, scope: !43)
!58 = !DILocation(line: 31, column: 28, scope: !43)
!59 = !DILocation(line: 32, column: 10, scope: !43)
!60 = !DILocation(line: 32, column: 11, scope: !43)
!61 = !DILocation(line: 32, column: 28, scope: !43)
!62 = !DILocation(line: 32, column: 27, scope: !43)
!63 = !{null}
!64 = !DISubroutineType(types: !63)
!65 = !{}
!66 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!70 = !{}
!67 = !{!5, !7, !8}
!68 = !{!18, !20, !21}
!69 = !{!31, !33}
