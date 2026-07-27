; Kylix LLVM IR — module: Functions
source_filename = "Functions.klx"
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
define i64 @Add(i64 %a, i64 %b) !dbg !4 {
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
  %t2 = add i64 %t0, %t1, !dbg !11
  store i64 %t2, ptr %result, !dbg !12
  %t3 = load i64, ptr %result, !dbg !6
  ret i64 %t3, !dbg !6
}

define double @Average(double %x, double %y, double %z) !dbg !13 {
entry:
  %result = alloca double, align 8
  #dbg_declare(ptr %result, !14, !DIExpression(), !15)
  %v_x_real = alloca double, align 8, !dbg !15
  store double %x, ptr %v_x_real, !dbg !15
  #dbg_declare(ptr %v_x_real, !16, !DIExpression(), !15)
  %v_y_real = alloca double, align 8, !dbg !15
  store double %y, ptr %v_y_real, !dbg !15
  #dbg_declare(ptr %v_y_real, !17, !DIExpression(), !15)
  %v_z_real = alloca double, align 8, !dbg !15
  store double %z, ptr %v_z_real, !dbg !15
  #dbg_declare(ptr %v_z_real, !18, !DIExpression(), !15)
  %t4 = load double, ptr %v_x_real, !dbg !19
  %t5 = load double, ptr %v_y_real, !dbg !20
  %t6 = fadd double %t4, %t5, !dbg !21
  %t7 = load double, ptr %v_z_real, !dbg !22
  %t8 = fadd double %t6, %t7, !dbg !23
  %t9 = add i64 0, 3, !dbg !24
  %t10 = sitofp i64 %t9 to double, !dbg !25
  %t11 = fdiv double %t8, %t10, !dbg !25
  store double %t11, ptr %result, !dbg !26
  %t12 = load double, ptr %result, !dbg !15
  ret double %t12, !dbg !15
}

define double @GetPI() !dbg !27 {
entry:
  %result = alloca double, align 8
  #dbg_declare(ptr %result, !28, !DIExpression(), !29)
  %t13 = fadd double 0.0, 3.141593, !dbg !30
  store double %t13, ptr %result, !dbg !31
  %t14 = load double, ptr %result, !dbg !29
  ret double %t14, !dbg !29
}

define void @Greet(ptr %name) !dbg !32 {
entry:
  %v_name_str = alloca ptr, align 8
  store ptr %name, ptr %v_name_str
  #dbg_declare(ptr %v_name_str, !33, !DIExpression(), !34)
  %t15 = call ptr @malloc(i64 512), !dbg !35
  store i8 0, ptr %t15, !dbg !35
  %t17 = getelementptr inbounds [8 x i8], ptr @.str.1, i64 0, i64 0, !dbg !36
  %t18 = call ptr @strcat(ptr %t15, ptr %t17), !dbg !35
  %t19 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !35
  %t20 = call ptr @strcat(ptr %t15, ptr %t19), !dbg !35
  %t21 = load ptr, ptr %v_name_str, !dbg !37
  %t22 = call ptr @strcat(ptr %t15, ptr %t21), !dbg !35
  %t23 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !35
  %t24 = call ptr @strcat(ptr %t15, ptr %t23), !dbg !35
  %t25 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !38
  %t26 = call ptr @strcat(ptr %t15, ptr %t25), !dbg !35
  %t27 = call i32 @puts(ptr noundef %t15), !dbg !35
  ret void, !dbg !34
}

define void @PrintInfo(ptr %name, i64 %age) !dbg !39 {
entry:
  %v_name_str = alloca ptr, align 8
  store ptr %name, ptr %v_name_str
  #dbg_declare(ptr %v_name_str, !40, !DIExpression(), !41)
  %v_age_int = alloca i64, align 8, !dbg !41
  store i64 %age, ptr %v_age_int, !dbg !41
  #dbg_declare(ptr %v_age_int, !42, !DIExpression(), !41)
  %t28 = call ptr @malloc(i64 512), !dbg !43
  store i8 0, ptr %t28, !dbg !43
  %t30 = getelementptr inbounds [7 x i8], ptr @.str.4, i64 0, i64 0, !dbg !44
  %t31 = call ptr @strcat(ptr %t28, ptr %t30), !dbg !43
  %t32 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !43
  %t33 = call ptr @strcat(ptr %t28, ptr %t32), !dbg !43
  %t34 = load ptr, ptr %v_name_str, !dbg !45
  %t35 = call ptr @strcat(ptr %t28, ptr %t34), !dbg !43
  %t36 = call i32 @puts(ptr noundef %t28), !dbg !43
  %t37 = call ptr @malloc(i64 512), !dbg !46
  store i8 0, ptr %t37, !dbg !46
  %t38 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !46
  %t39 = getelementptr inbounds [6 x i8], ptr @.str.5, i64 0, i64 0, !dbg !47
  %t40 = call ptr @strcat(ptr %t37, ptr %t39), !dbg !46
  %t41 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !46
  %t42 = call ptr @strcat(ptr %t37, ptr %t41), !dbg !46
  %t43 = load i64, ptr %v_age_int, !dbg !48
  %t44 = call i64 @strlen(ptr %t37), !dbg !46
  %t45 = getelementptr inbounds i8, ptr %t37, i64 %t44, !dbg !46
  %t46 = sub i64 512, %t44, !dbg !46
  %t47 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t45, i64 %t46, ptr %t38, i64 %t43), !dbg !46
  %t48 = call i32 @puts(ptr noundef %t37), !dbg !46
  %t49 = getelementptr inbounds [4 x i8], ptr @.str.6, i64 0, i64 0, !dbg !49
  %t50 = call i32 @puts(ptr noundef %t49), !dbg !50
  ret void, !dbg !41
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
define i32 @main() !dbg !51 {
entry:
  %t51 = call ptr @malloc(i64 512), !dbg !52
  store i8 0, ptr %t51, !dbg !52
  %t52 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !52
  %t53 = getelementptr inbounds [9 x i8], ptr @.str.7, i64 0, i64 0, !dbg !53
  %t54 = call ptr @strcat(ptr %t51, ptr %t53), !dbg !52
  %t55 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !52
  %t56 = call ptr @strcat(ptr %t51, ptr %t55), !dbg !52
  %t57 = add i64 0, 5, !dbg !54
  %t58 = add i64 0, 3, !dbg !55
  %t59 = call i64 @Add(i64 %t57, i64 %t58), !dbg !56
  %t60 = call i64 @strlen(ptr %t51), !dbg !52
  %t61 = getelementptr inbounds i8, ptr %t51, i64 %t60, !dbg !52
  %t62 = sub i64 512, %t60, !dbg !52
  %t63 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t61, i64 %t62, ptr %t52, i64 %t59), !dbg !52
  %t64 = call i32 @puts(ptr noundef %t51), !dbg !52
  %t65 = call ptr @malloc(i64 512), !dbg !57
  store i8 0, ptr %t65, !dbg !57
  %t67 = getelementptr inbounds [25 x i8], ptr @.str.8, i64 0, i64 0, !dbg !58
  %t68 = call ptr @strcat(ptr %t65, ptr %t67), !dbg !57
  %t69 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !57
  %t70 = call ptr @strcat(ptr %t65, ptr %t69), !dbg !57
  %t71 = add i64 0, 10, !dbg !59
  %t72 = sitofp i64 %t71 to double, !dbg !60
  %t73 = add i64 0, 20, !dbg !61
  %t74 = sitofp i64 %t73 to double, !dbg !60
  %t75 = add i64 0, 30, !dbg !62
  %t76 = sitofp i64 %t75 to double, !dbg !60
  %t77 = call double @Average(double %t72, double %t74, double %t76), !dbg !60
  %t78 = getelementptr inbounds [6 x i8], ptr @.str.9, i64 0, i64 0, !dbg !57
  %t79 = call i64 @strlen(ptr %t65), !dbg !57
  %t80 = getelementptr inbounds i8, ptr %t65, i64 %t79, !dbg !57
  %t81 = sub i64 512, %t79, !dbg !57
  %t82 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t80, i64 %t81, ptr %t78, double %t77), !dbg !57
  %t83 = call i32 @puts(ptr noundef %t65), !dbg !57
  %t84 = call ptr @malloc(i64 512), !dbg !63
  store i8 0, ptr %t84, !dbg !63
  %t86 = getelementptr inbounds [6 x i8], ptr @.str.10, i64 0, i64 0, !dbg !64
  %t87 = call ptr @strcat(ptr %t84, ptr %t86), !dbg !63
  %t88 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !63
  %t89 = call ptr @strcat(ptr %t84, ptr %t88), !dbg !63
  %t90 = call double @GetPI(), !dbg !65
  %t91 = getelementptr inbounds [6 x i8], ptr @.str.9, i64 0, i64 0, !dbg !63
  %t92 = call i64 @strlen(ptr %t84), !dbg !63
  %t93 = getelementptr inbounds i8, ptr %t84, i64 %t92, !dbg !63
  %t94 = sub i64 512, %t92, !dbg !63
  %t95 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t93, i64 %t94, ptr %t91, double %t90), !dbg !63
  %t96 = call i32 @puts(ptr noundef %t84), !dbg !63
  %t97 = getelementptr inbounds [6 x i8], ptr @.str.11, i64 0, i64 0, !dbg !66
  call void @Greet(ptr %t97), !dbg !67
  %t98 = getelementptr inbounds [4 x i8], ptr @.str.12, i64 0, i64 0, !dbg !68
  call void @Greet(ptr %t98), !dbg !69
  %t99 = getelementptr inbounds [8 x i8], ptr @.str.13, i64 0, i64 0, !dbg !70
  %t100 = add i64 0, 25, !dbg !71
  call void @PrintInfo(ptr %t99, i64 %t100), !dbg !72
  %t101 = getelementptr inbounds [6 x i8], ptr @.str.14, i64 0, i64 0, !dbg !73
  %t102 = add i64 0, 30, !dbg !74
  call void @PrintInfo(ptr %t101, i64 %t102), !dbg !75
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [8 x i8] c"Hello, \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [2 x i8] c"!\00", align 1
@.str.4 = private unnamed_addr constant [7 x i8] c"Name: \00", align 1
@.str.5 = private unnamed_addr constant [6 x i8] c"Age: \00", align 1
@.str.6 = private unnamed_addr constant [4 x i8] c"---\00", align 1
@.str.7 = private unnamed_addr constant [9 x i8] c"5 + 3 = \00", align 1
@.str.8 = private unnamed_addr constant [25 x i8] c"Average of 10, 20, 30 = \00", align 1
@.str.9 = private unnamed_addr constant [6 x i8] c"%.15g\00", align 1
@.str.10 = private unnamed_addr constant [6 x i8] c"PI = \00", align 1
@.str.11 = private unnamed_addr constant [6 x i8] c"Alice\00", align 1
@.str.12 = private unnamed_addr constant [4 x i8] c"Bob\00", align 1
@.str.13 = private unnamed_addr constant [8 x i8] c"Charlie\00", align 1
@.str.14 = private unnamed_addr constant [6 x i8] c"Diana\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example13_functions.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/03_functions")
!4 = distinct !DISubprogram(name: "Add", scope: !3, file: !3, line: 4, type: !77, scopeLine: 4, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !80)
!13 = distinct !DISubprogram(name: "Average", scope: !3, file: !3, line: 10, type: !77, scopeLine: 10, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !81)
!27 = distinct !DISubprogram(name: "GetPI", scope: !3, file: !3, line: 16, type: !77, scopeLine: 16, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !82)
!32 = distinct !DISubprogram(name: "Greet", scope: !3, file: !3, line: 22, type: !77, scopeLine: 22, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !83)
!39 = distinct !DISubprogram(name: "PrintInfo", scope: !3, file: !3, line: 28, type: !77, scopeLine: 28, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !84)
!51 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !77, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !85)
!5 = !DILocalVariable(name: "result", scope: !4, file: !3, line: 4, type: !79)
!7 = !DILocalVariable(name: "a", scope: !4, file: !3, line: 4, type: !79)
!8 = !DILocalVariable(name: "b", scope: !4, file: !3, line: 4, type: !79)
!14 = !DILocalVariable(name: "result", scope: !13, file: !3, line: 10, type: !79)
!16 = !DILocalVariable(name: "x", scope: !13, file: !3, line: 10, type: !79)
!17 = !DILocalVariable(name: "y", scope: !13, file: !3, line: 10, type: !79)
!18 = !DILocalVariable(name: "z", scope: !13, file: !3, line: 10, type: !79)
!28 = !DILocalVariable(name: "result", scope: !27, file: !3, line: 16, type: !79)
!33 = !DILocalVariable(name: "name", scope: !32, file: !3, line: 22, type: !79)
!40 = !DILocalVariable(name: "name", scope: !39, file: !3, line: 28, type: !79)
!42 = !DILocalVariable(name: "age", scope: !39, file: !3, line: 28, type: !79)
!6 = !DILocation(line: 4, column: 1, scope: !4)
!9 = !DILocation(line: 6, column: 13, scope: !4)
!10 = !DILocation(line: 6, column: 17, scope: !4)
!11 = !DILocation(line: 6, column: 15, scope: !4)
!12 = !DILocation(line: 6, column: 11, scope: !4)
!15 = !DILocation(line: 10, column: 1, scope: !13)
!19 = !DILocation(line: 12, column: 14, scope: !13)
!20 = !DILocation(line: 12, column: 18, scope: !13)
!21 = !DILocation(line: 12, column: 16, scope: !13)
!22 = !DILocation(line: 12, column: 22, scope: !13)
!23 = !DILocation(line: 12, column: 20, scope: !13)
!24 = !DILocation(line: 12, column: 27, scope: !13)
!25 = !DILocation(line: 12, column: 25, scope: !13)
!26 = !DILocation(line: 12, column: 11, scope: !13)
!29 = !DILocation(line: 16, column: 1, scope: !27)
!30 = !DILocation(line: 18, column: 13, scope: !27)
!31 = !DILocation(line: 18, column: 11, scope: !27)
!34 = !DILocation(line: 22, column: 1, scope: !32)
!35 = !DILocation(line: 24, column: 10, scope: !32)
!36 = !DILocation(line: 24, column: 11, scope: !32)
!37 = !DILocation(line: 24, column: 22, scope: !32)
!38 = !DILocation(line: 24, column: 28, scope: !32)
!41 = !DILocation(line: 28, column: 1, scope: !39)
!43 = !DILocation(line: 30, column: 10, scope: !39)
!44 = !DILocation(line: 30, column: 11, scope: !39)
!45 = !DILocation(line: 30, column: 21, scope: !39)
!46 = !DILocation(line: 31, column: 10, scope: !39)
!47 = !DILocation(line: 31, column: 11, scope: !39)
!48 = !DILocation(line: 31, column: 20, scope: !39)
!49 = !DILocation(line: 32, column: 11, scope: !39)
!50 = !DILocation(line: 32, column: 10, scope: !39)
!52 = !DILocation(line: 36, column: 10, scope: !51)
!53 = !DILocation(line: 36, column: 11, scope: !51)
!54 = !DILocation(line: 36, column: 27, scope: !51)
!55 = !DILocation(line: 36, column: 30, scope: !51)
!56 = !DILocation(line: 36, column: 26, scope: !51)
!57 = !DILocation(line: 37, column: 10, scope: !51)
!58 = !DILocation(line: 37, column: 11, scope: !51)
!59 = !DILocation(line: 37, column: 47, scope: !51)
!60 = !DILocation(line: 37, column: 46, scope: !51)
!61 = !DILocation(line: 37, column: 51, scope: !51)
!62 = !DILocation(line: 37, column: 55, scope: !51)
!63 = !DILocation(line: 38, column: 10, scope: !51)
!64 = !DILocation(line: 38, column: 11, scope: !51)
!65 = !DILocation(line: 38, column: 25, scope: !51)
!66 = !DILocation(line: 40, column: 9, scope: !51)
!67 = !DILocation(line: 40, column: 8, scope: !51)
!68 = !DILocation(line: 41, column: 9, scope: !51)
!69 = !DILocation(line: 41, column: 8, scope: !51)
!70 = !DILocation(line: 43, column: 13, scope: !51)
!71 = !DILocation(line: 43, column: 24, scope: !51)
!72 = !DILocation(line: 43, column: 12, scope: !51)
!73 = !DILocation(line: 44, column: 13, scope: !51)
!74 = !DILocation(line: 44, column: 22, scope: !51)
!75 = !DILocation(line: 44, column: 12, scope: !51)
!76 = !{null}
!77 = !DISubroutineType(types: !76)
!78 = !{}
!79 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!80 = !{!5, !7, !8}
!81 = !{!14, !16, !17, !18}
!82 = !{!28}
!83 = !{!33}
!84 = !{!40, !42}
!85 = !{}
