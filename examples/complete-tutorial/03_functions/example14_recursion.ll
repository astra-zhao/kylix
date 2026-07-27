; Kylix LLVM IR — module: Recursion
source_filename = "Recursion.klx"
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
define i64 @Factorial(i64 %n) !dbg !4 {
entry:
  %result = alloca i64, align 8
  #dbg_declare(ptr %result, !5, !DIExpression(), !6)
  %v_n_int = alloca i64, align 8, !dbg !6
  store i64 %n, ptr %v_n_int, !dbg !6
  #dbg_declare(ptr %v_n_int, !7, !DIExpression(), !6)
  %t0 = load i64, ptr %v_n_int, !dbg !8
  %t1 = add i64 0, 1, !dbg !9
  %t2 = icmp sle i64 %t0, %t1, !dbg !10
  br i1 %t2, label %lbl0, label %lbl2, !dbg !11
lbl0:
  %t3 = add i64 0, 1, !dbg !12
  store i64 %t3, ptr %result, !dbg !13
  br label %lbl1, !dbg !11
lbl2:
  %t4 = load i64, ptr %v_n_int, !dbg !14
  %t5 = load i64, ptr %v_n_int, !dbg !15
  %t6 = add i64 0, 1, !dbg !16
  %t7 = sub i64 %t5, %t6, !dbg !17
  %t8 = call i64 @Factorial(i64 %t7), !dbg !18
  %t9 = mul i64 %t4, %t8, !dbg !19
  store i64 %t9, ptr %result, !dbg !20
  br label %lbl1, !dbg !11
lbl1:
  %t10 = load i64, ptr %result, !dbg !6
  ret i64 %t10, !dbg !6
}

define i64 @Fibonacci(i64 %n) !dbg !21 {
entry:
  %result = alloca i64, align 8
  #dbg_declare(ptr %result, !22, !DIExpression(), !23)
  %v_n_int = alloca i64, align 8, !dbg !23
  store i64 %n, ptr %v_n_int, !dbg !23
  #dbg_declare(ptr %v_n_int, !24, !DIExpression(), !23)
  %t11 = load i64, ptr %v_n_int, !dbg !25
  %t12 = add i64 0, 1, !dbg !26
  %t13 = icmp sle i64 %t11, %t12, !dbg !27
  br i1 %t13, label %lbl3, label %lbl5, !dbg !28
lbl3:
  %t14 = load i64, ptr %v_n_int, !dbg !29
  store i64 %t14, ptr %result, !dbg !30
  br label %lbl4, !dbg !28
lbl5:
  %t15 = load i64, ptr %v_n_int, !dbg !31
  %t16 = add i64 0, 1, !dbg !32
  %t17 = sub i64 %t15, %t16, !dbg !33
  %t18 = call i64 @Fibonacci(i64 %t17), !dbg !34
  %t19 = load i64, ptr %v_n_int, !dbg !35
  %t20 = add i64 0, 2, !dbg !36
  %t21 = sub i64 %t19, %t20, !dbg !37
  %t22 = call i64 @Fibonacci(i64 %t21), !dbg !38
  %t23 = add i64 %t18, %t22, !dbg !39
  store i64 %t23, ptr %result, !dbg !40
  br label %lbl4, !dbg !28
lbl4:
  %t24 = load i64, ptr %result, !dbg !23
  ret i64 %t24, !dbg !23
}

define i64 @Power(i64 %base, i64 %exp) !dbg !41 {
entry:
  %result = alloca i64, align 8
  #dbg_declare(ptr %result, !42, !DIExpression(), !43)
  %v_base_int = alloca i64, align 8, !dbg !43
  store i64 %base, ptr %v_base_int, !dbg !43
  #dbg_declare(ptr %v_base_int, !44, !DIExpression(), !43)
  %v_exp_int = alloca i64, align 8, !dbg !43
  store i64 %exp, ptr %v_exp_int, !dbg !43
  #dbg_declare(ptr %v_exp_int, !45, !DIExpression(), !43)
  %t25 = load i64, ptr %v_exp_int, !dbg !46
  %t26 = add i64 0, 0, !dbg !47
  %t27 = icmp eq i64 %t25, %t26, !dbg !48
  br i1 %t27, label %lbl6, label %lbl8, !dbg !49
lbl6:
  %t28 = add i64 0, 1, !dbg !50
  store i64 %t28, ptr %result, !dbg !51
  br label %lbl7, !dbg !49
lbl8:
  %t29 = load i64, ptr %v_base_int, !dbg !52
  %t30 = load i64, ptr %v_base_int, !dbg !53
  %t31 = load i64, ptr %v_exp_int, !dbg !54
  %t32 = add i64 0, 1, !dbg !55
  %t33 = sub i64 %t31, %t32, !dbg !56
  %t34 = call i64 @Power(i64 %t30, i64 %t33), !dbg !57
  %t35 = mul i64 %t29, %t34, !dbg !58
  store i64 %t35, ptr %result, !dbg !59
  br label %lbl7, !dbg !49
lbl7:
  %t36 = load i64, ptr %result, !dbg !43
  ret i64 %t36, !dbg !43
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
define i32 @main() !dbg !60 {
entry:
  %t37 = call ptr @malloc(i64 512), !dbg !61
  store i8 0, ptr %t37, !dbg !61
  %t38 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !61
  %t39 = getelementptr inbounds [17 x i8], ptr @.str.1, i64 0, i64 0, !dbg !62
  %t40 = call ptr @strcat(ptr %t37, ptr %t39), !dbg !61
  %t41 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !61
  %t42 = call ptr @strcat(ptr %t37, ptr %t41), !dbg !61
  %t43 = add i64 0, 5, !dbg !63
  %t44 = call i64 @Factorial(i64 %t43), !dbg !64
  %t45 = call i64 @strlen(ptr %t37), !dbg !61
  %t46 = getelementptr inbounds i8, ptr %t37, i64 %t45, !dbg !61
  %t47 = sub i64 512, %t45, !dbg !61
  %t48 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t46, i64 %t47, ptr %t38, i64 %t44), !dbg !61
  %t49 = call i32 @puts(ptr noundef %t37), !dbg !61
  %t50 = call ptr @malloc(i64 512), !dbg !65
  store i8 0, ptr %t50, !dbg !65
  %t51 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !65
  %t52 = getelementptr inbounds [18 x i8], ptr @.str.3, i64 0, i64 0, !dbg !66
  %t53 = call ptr @strcat(ptr %t50, ptr %t52), !dbg !65
  %t54 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !65
  %t55 = call ptr @strcat(ptr %t50, ptr %t54), !dbg !65
  %t56 = add i64 0, 10, !dbg !67
  %t57 = call i64 @Factorial(i64 %t56), !dbg !68
  %t58 = call i64 @strlen(ptr %t50), !dbg !65
  %t59 = getelementptr inbounds i8, ptr %t50, i64 %t58, !dbg !65
  %t60 = sub i64 512, %t58, !dbg !65
  %t61 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t59, i64 %t60, ptr %t51, i64 %t57), !dbg !65
  %t62 = call i32 @puts(ptr noundef %t50), !dbg !65
  %t63 = getelementptr inbounds [20 x i8], ptr @.str.4, i64 0, i64 0, !dbg !69
  %t64 = call i32 @puts(ptr noundef %t63), !dbg !70
  %t65 = call ptr @malloc(i64 512), !dbg !71
  store i8 0, ptr %t65, !dbg !71
  %t66 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !71
  %t67 = getelementptr inbounds [8 x i8], ptr @.str.5, i64 0, i64 0, !dbg !72
  %t68 = call ptr @strcat(ptr %t65, ptr %t67), !dbg !71
  %t69 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !71
  %t70 = call ptr @strcat(ptr %t65, ptr %t69), !dbg !71
  %t71 = add i64 0, 0, !dbg !73
  %t72 = call i64 @Fibonacci(i64 %t71), !dbg !74
  %t73 = call i64 @strlen(ptr %t65), !dbg !71
  %t74 = getelementptr inbounds i8, ptr %t65, i64 %t73, !dbg !71
  %t75 = sub i64 512, %t73, !dbg !71
  %t76 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t74, i64 %t75, ptr %t66, i64 %t72), !dbg !71
  %t77 = call i32 @puts(ptr noundef %t65), !dbg !71
  %t78 = call ptr @malloc(i64 512), !dbg !75
  store i8 0, ptr %t78, !dbg !75
  %t79 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !75
  %t80 = getelementptr inbounds [8 x i8], ptr @.str.6, i64 0, i64 0, !dbg !76
  %t81 = call ptr @strcat(ptr %t78, ptr %t80), !dbg !75
  %t82 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !75
  %t83 = call ptr @strcat(ptr %t78, ptr %t82), !dbg !75
  %t84 = add i64 0, 1, !dbg !77
  %t85 = call i64 @Fibonacci(i64 %t84), !dbg !78
  %t86 = call i64 @strlen(ptr %t78), !dbg !75
  %t87 = getelementptr inbounds i8, ptr %t78, i64 %t86, !dbg !75
  %t88 = sub i64 512, %t86, !dbg !75
  %t89 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t87, i64 %t88, ptr %t79, i64 %t85), !dbg !75
  %t90 = call i32 @puts(ptr noundef %t78), !dbg !75
  %t91 = call ptr @malloc(i64 512), !dbg !79
  store i8 0, ptr %t91, !dbg !79
  %t92 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !79
  %t93 = getelementptr inbounds [8 x i8], ptr @.str.7, i64 0, i64 0, !dbg !80
  %t94 = call ptr @strcat(ptr %t91, ptr %t93), !dbg !79
  %t95 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !79
  %t96 = call ptr @strcat(ptr %t91, ptr %t95), !dbg !79
  %t97 = add i64 0, 5, !dbg !81
  %t98 = call i64 @Fibonacci(i64 %t97), !dbg !82
  %t99 = call i64 @strlen(ptr %t91), !dbg !79
  %t100 = getelementptr inbounds i8, ptr %t91, i64 %t99, !dbg !79
  %t101 = sub i64 512, %t99, !dbg !79
  %t102 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t100, i64 %t101, ptr %t92, i64 %t98), !dbg !79
  %t103 = call i32 @puts(ptr noundef %t91), !dbg !79
  %t104 = call ptr @malloc(i64 512), !dbg !83
  store i8 0, ptr %t104, !dbg !83
  %t105 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !83
  %t106 = getelementptr inbounds [9 x i8], ptr @.str.8, i64 0, i64 0, !dbg !84
  %t107 = call ptr @strcat(ptr %t104, ptr %t106), !dbg !83
  %t108 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !83
  %t109 = call ptr @strcat(ptr %t104, ptr %t108), !dbg !83
  %t110 = add i64 0, 10, !dbg !85
  %t111 = call i64 @Fibonacci(i64 %t110), !dbg !86
  %t112 = call i64 @strlen(ptr %t104), !dbg !83
  %t113 = getelementptr inbounds i8, ptr %t104, i64 %t112, !dbg !83
  %t114 = sub i64 512, %t112, !dbg !83
  %t115 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t113, i64 %t114, ptr %t105, i64 %t111), !dbg !83
  %t116 = call i32 @puts(ptr noundef %t104), !dbg !83
  %t117 = call ptr @malloc(i64 512), !dbg !87
  store i8 0, ptr %t117, !dbg !87
  %t118 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !87
  %t119 = getelementptr inbounds [7 x i8], ptr @.str.9, i64 0, i64 0, !dbg !88
  %t120 = call ptr @strcat(ptr %t117, ptr %t119), !dbg !87
  %t121 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !87
  %t122 = call ptr @strcat(ptr %t117, ptr %t121), !dbg !87
  %t123 = add i64 0, 2, !dbg !89
  %t124 = add i64 0, 3, !dbg !90
  %t125 = call i64 @Power(i64 %t123, i64 %t124), !dbg !91
  %t126 = call i64 @strlen(ptr %t117), !dbg !87
  %t127 = getelementptr inbounds i8, ptr %t117, i64 %t126, !dbg !87
  %t128 = sub i64 512, %t126, !dbg !87
  %t129 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t127, i64 %t128, ptr %t118, i64 %t125), !dbg !87
  %t130 = call i32 @puts(ptr noundef %t117), !dbg !87
  %t131 = call ptr @malloc(i64 512), !dbg !92
  store i8 0, ptr %t131, !dbg !92
  %t132 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !92
  %t133 = getelementptr inbounds [7 x i8], ptr @.str.10, i64 0, i64 0, !dbg !93
  %t134 = call ptr @strcat(ptr %t131, ptr %t133), !dbg !92
  %t135 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !92
  %t136 = call ptr @strcat(ptr %t131, ptr %t135), !dbg !92
  %t137 = add i64 0, 5, !dbg !94
  %t138 = add i64 0, 4, !dbg !95
  %t139 = call i64 @Power(i64 %t137, i64 %t138), !dbg !96
  %t140 = call i64 @strlen(ptr %t131), !dbg !92
  %t141 = getelementptr inbounds i8, ptr %t131, i64 %t140, !dbg !92
  %t142 = sub i64 512, %t140, !dbg !92
  %t143 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t141, i64 %t142, ptr %t132, i64 %t139), !dbg !92
  %t144 = call i32 @puts(ptr noundef %t131), !dbg !92
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [17 x i8] c"Factorial of 5: \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [18 x i8] c"Factorial of 10: \00", align 1
@.str.4 = private unnamed_addr constant [20 x i8] c"Fibonacci sequence:\00", align 1
@.str.5 = private unnamed_addr constant [8 x i8] c"F(0) = \00", align 1
@.str.6 = private unnamed_addr constant [8 x i8] c"F(1) = \00", align 1
@.str.7 = private unnamed_addr constant [8 x i8] c"F(5) = \00", align 1
@.str.8 = private unnamed_addr constant [9 x i8] c"F(10) = \00", align 1
@.str.9 = private unnamed_addr constant [7 x i8] c"2^3 = \00", align 1
@.str.10 = private unnamed_addr constant [7 x i8] c"5^4 = \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example14_recursion.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/03_functions")
!4 = distinct !DISubprogram(name: "Factorial", scope: !3, file: !3, line: 4, type: !98, scopeLine: 4, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !101)
!21 = distinct !DISubprogram(name: "Fibonacci", scope: !3, file: !3, line: 13, type: !98, scopeLine: 13, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !102)
!41 = distinct !DISubprogram(name: "Power", scope: !3, file: !3, line: 22, type: !98, scopeLine: 22, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !103)
!60 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !98, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !104)
!5 = !DILocalVariable(name: "result", scope: !4, file: !3, line: 4, type: !100)
!7 = !DILocalVariable(name: "n", scope: !4, file: !3, line: 4, type: !100)
!22 = !DILocalVariable(name: "result", scope: !21, file: !3, line: 13, type: !100)
!24 = !DILocalVariable(name: "n", scope: !21, file: !3, line: 13, type: !100)
!42 = !DILocalVariable(name: "result", scope: !41, file: !3, line: 22, type: !100)
!44 = !DILocalVariable(name: "base", scope: !41, file: !3, line: 22, type: !100)
!45 = !DILocalVariable(name: "exp", scope: !41, file: !3, line: 22, type: !100)
!6 = !DILocation(line: 4, column: 1, scope: !4)
!8 = !DILocation(line: 6, column: 6, scope: !4)
!9 = !DILocation(line: 6, column: 11, scope: !4)
!10 = !DILocation(line: 6, column: 9, scope: !4)
!11 = !DILocation(line: 6, column: 3, scope: !4)
!12 = !DILocation(line: 7, column: 15, scope: !4)
!13 = !DILocation(line: 7, column: 13, scope: !4)
!14 = !DILocation(line: 9, column: 15, scope: !4)
!15 = !DILocation(line: 9, column: 29, scope: !4)
!16 = !DILocation(line: 9, column: 33, scope: !4)
!17 = !DILocation(line: 9, column: 31, scope: !4)
!18 = !DILocation(line: 9, column: 28, scope: !4)
!19 = !DILocation(line: 9, column: 17, scope: !4)
!20 = !DILocation(line: 9, column: 13, scope: !4)
!23 = !DILocation(line: 13, column: 1, scope: !21)
!25 = !DILocation(line: 15, column: 6, scope: !21)
!26 = !DILocation(line: 15, column: 11, scope: !21)
!27 = !DILocation(line: 15, column: 9, scope: !21)
!28 = !DILocation(line: 15, column: 3, scope: !21)
!29 = !DILocation(line: 16, column: 15, scope: !21)
!30 = !DILocation(line: 16, column: 13, scope: !21)
!31 = !DILocation(line: 18, column: 25, scope: !21)
!32 = !DILocation(line: 18, column: 29, scope: !21)
!33 = !DILocation(line: 18, column: 27, scope: !21)
!34 = !DILocation(line: 18, column: 24, scope: !21)
!35 = !DILocation(line: 18, column: 44, scope: !21)
!36 = !DILocation(line: 18, column: 48, scope: !21)
!37 = !DILocation(line: 18, column: 46, scope: !21)
!38 = !DILocation(line: 18, column: 43, scope: !21)
!39 = !DILocation(line: 18, column: 32, scope: !21)
!40 = !DILocation(line: 18, column: 13, scope: !21)
!43 = !DILocation(line: 22, column: 1, scope: !41)
!46 = !DILocation(line: 24, column: 6, scope: !41)
!47 = !DILocation(line: 24, column: 12, scope: !41)
!48 = !DILocation(line: 24, column: 10, scope: !41)
!49 = !DILocation(line: 24, column: 3, scope: !41)
!50 = !DILocation(line: 25, column: 15, scope: !41)
!51 = !DILocation(line: 25, column: 13, scope: !41)
!52 = !DILocation(line: 27, column: 15, scope: !41)
!53 = !DILocation(line: 27, column: 28, scope: !41)
!54 = !DILocation(line: 27, column: 34, scope: !41)
!55 = !DILocation(line: 27, column: 40, scope: !41)
!56 = !DILocation(line: 27, column: 38, scope: !41)
!57 = !DILocation(line: 27, column: 27, scope: !41)
!58 = !DILocation(line: 27, column: 20, scope: !41)
!59 = !DILocation(line: 27, column: 13, scope: !41)
!61 = !DILocation(line: 31, column: 10, scope: !60)
!62 = !DILocation(line: 31, column: 11, scope: !60)
!63 = !DILocation(line: 31, column: 41, scope: !60)
!64 = !DILocation(line: 31, column: 40, scope: !60)
!65 = !DILocation(line: 32, column: 10, scope: !60)
!66 = !DILocation(line: 32, column: 11, scope: !60)
!67 = !DILocation(line: 32, column: 42, scope: !60)
!68 = !DILocation(line: 32, column: 41, scope: !60)
!69 = !DILocation(line: 34, column: 11, scope: !60)
!70 = !DILocation(line: 34, column: 10, scope: !60)
!71 = !DILocation(line: 35, column: 10, scope: !60)
!72 = !DILocation(line: 35, column: 11, scope: !60)
!73 = !DILocation(line: 35, column: 32, scope: !60)
!74 = !DILocation(line: 35, column: 31, scope: !60)
!75 = !DILocation(line: 36, column: 10, scope: !60)
!76 = !DILocation(line: 36, column: 11, scope: !60)
!77 = !DILocation(line: 36, column: 32, scope: !60)
!78 = !DILocation(line: 36, column: 31, scope: !60)
!79 = !DILocation(line: 37, column: 10, scope: !60)
!80 = !DILocation(line: 37, column: 11, scope: !60)
!81 = !DILocation(line: 37, column: 32, scope: !60)
!82 = !DILocation(line: 37, column: 31, scope: !60)
!83 = !DILocation(line: 38, column: 10, scope: !60)
!84 = !DILocation(line: 38, column: 11, scope: !60)
!85 = !DILocation(line: 38, column: 33, scope: !60)
!86 = !DILocation(line: 38, column: 32, scope: !60)
!87 = !DILocation(line: 40, column: 10, scope: !60)
!88 = !DILocation(line: 40, column: 11, scope: !60)
!89 = !DILocation(line: 40, column: 27, scope: !60)
!90 = !DILocation(line: 40, column: 30, scope: !60)
!91 = !DILocation(line: 40, column: 26, scope: !60)
!92 = !DILocation(line: 41, column: 10, scope: !60)
!93 = !DILocation(line: 41, column: 11, scope: !60)
!94 = !DILocation(line: 41, column: 27, scope: !60)
!95 = !DILocation(line: 41, column: 30, scope: !60)
!96 = !DILocation(line: 41, column: 26, scope: !60)
!97 = !{null}
!98 = !DISubroutineType(types: !97)
!99 = !{}
!100 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!101 = !{!5, !7}
!102 = !{!22, !24}
!103 = !{!42, !44, !45}
!104 = !{}
