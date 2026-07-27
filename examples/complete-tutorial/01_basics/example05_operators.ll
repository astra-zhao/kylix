; Kylix LLVM IR — module: Operators
source_filename = "Operators.klx"
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
  %v_a_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_a_int, !dbg !5
  #dbg_declare(ptr %v_a_int, !6, !DIExpression(), !5)
  %v_b_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_b_int, !dbg !5
  #dbg_declare(ptr %v_b_int, !7, !DIExpression(), !5)
  %v_x_real = alloca double, align 8, !dbg !5
  store double 0.0, ptr %v_x_real, !dbg !5
  #dbg_declare(ptr %v_x_real, !8, !DIExpression(), !5)
  %v_y_real = alloca double, align 8, !dbg !5
  store double 0.0, ptr %v_y_real, !dbg !5
  #dbg_declare(ptr %v_y_real, !9, !DIExpression(), !5)
  %v_flag1_bool = alloca i1, align 8, !dbg !5
  store i1 0, ptr %v_flag1_bool, !dbg !5
  #dbg_declare(ptr %v_flag1_bool, !10, !DIExpression(), !5)
  %v_flag2_bool = alloca i1, align 8, !dbg !5
  store i1 0, ptr %v_flag2_bool, !dbg !5
  #dbg_declare(ptr %v_flag2_bool, !11, !DIExpression(), !5)
  %t0 = add i64 0, 10, !dbg !12
  store i64 %t0, ptr %v_a_int, !dbg !13
  %t1 = add i64 0, 3, !dbg !14
  store i64 %t1, ptr %v_b_int, !dbg !15
  %t2 = getelementptr inbounds [19 x i8], ptr @.str.0, i64 0, i64 0, !dbg !16
  %t3 = call i32 @puts(ptr noundef %t2), !dbg !17
  %t4 = call ptr @malloc(i64 512), !dbg !18
  store i8 0, ptr %t4, !dbg !18
  %t5 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !18
  %t6 = getelementptr inbounds [9 x i8], ptr @.str.2, i64 0, i64 0, !dbg !19
  %t7 = call ptr @strcat(ptr %t4, ptr %t6), !dbg !18
  %t8 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !18
  %t9 = call ptr @strcat(ptr %t4, ptr %t8), !dbg !18
  %t10 = load i64, ptr %v_a_int, !dbg !20
  %t11 = load i64, ptr %v_b_int, !dbg !21
  %t12 = add i64 %t10, %t11, !dbg !22
  %t13 = call i64 @strlen(ptr %t4), !dbg !18
  %t14 = getelementptr inbounds i8, ptr %t4, i64 %t13, !dbg !18
  %t15 = sub i64 512, %t13, !dbg !18
  %t16 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t14, i64 %t15, ptr %t5, i64 %t12), !dbg !18
  %t17 = call i32 @puts(ptr noundef %t4), !dbg !18
  %t18 = call ptr @malloc(i64 512), !dbg !23
  store i8 0, ptr %t18, !dbg !23
  %t19 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !23
  %t20 = getelementptr inbounds [9 x i8], ptr @.str.4, i64 0, i64 0, !dbg !24
  %t21 = call ptr @strcat(ptr %t18, ptr %t20), !dbg !23
  %t22 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !23
  %t23 = call ptr @strcat(ptr %t18, ptr %t22), !dbg !23
  %t24 = load i64, ptr %v_a_int, !dbg !25
  %t25 = load i64, ptr %v_b_int, !dbg !26
  %t26 = sub i64 %t24, %t25, !dbg !27
  %t27 = call i64 @strlen(ptr %t18), !dbg !23
  %t28 = getelementptr inbounds i8, ptr %t18, i64 %t27, !dbg !23
  %t29 = sub i64 512, %t27, !dbg !23
  %t30 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t28, i64 %t29, ptr %t19, i64 %t26), !dbg !23
  %t31 = call i32 @puts(ptr noundef %t18), !dbg !23
  %t32 = call ptr @malloc(i64 512), !dbg !28
  store i8 0, ptr %t32, !dbg !28
  %t33 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !28
  %t34 = getelementptr inbounds [9 x i8], ptr @.str.5, i64 0, i64 0, !dbg !29
  %t35 = call ptr @strcat(ptr %t32, ptr %t34), !dbg !28
  %t36 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !28
  %t37 = call ptr @strcat(ptr %t32, ptr %t36), !dbg !28
  %t38 = load i64, ptr %v_a_int, !dbg !30
  %t39 = load i64, ptr %v_b_int, !dbg !31
  %t40 = mul i64 %t38, %t39, !dbg !32
  %t41 = call i64 @strlen(ptr %t32), !dbg !28
  %t42 = getelementptr inbounds i8, ptr %t32, i64 %t41, !dbg !28
  %t43 = sub i64 512, %t41, !dbg !28
  %t44 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t42, i64 %t43, ptr %t33, i64 %t40), !dbg !28
  %t45 = call i32 @puts(ptr noundef %t32), !dbg !28
  %t46 = call ptr @malloc(i64 512), !dbg !33
  store i8 0, ptr %t46, !dbg !33
  %t47 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !33
  %t48 = getelementptr inbounds [9 x i8], ptr @.str.6, i64 0, i64 0, !dbg !34
  %t49 = call ptr @strcat(ptr %t46, ptr %t48), !dbg !33
  %t50 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !33
  %t51 = call ptr @strcat(ptr %t46, ptr %t50), !dbg !33
  %t52 = load i64, ptr %v_a_int, !dbg !35
  %t53 = load i64, ptr %v_b_int, !dbg !36
  %t54 = sdiv i64 %t52, %t53, !dbg !37
  %t55 = call i64 @strlen(ptr %t46), !dbg !33
  %t56 = getelementptr inbounds i8, ptr %t46, i64 %t55, !dbg !33
  %t57 = sub i64 512, %t55, !dbg !33
  %t58 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t56, i64 %t57, ptr %t47, i64 %t54), !dbg !33
  %t59 = call i32 @puts(ptr noundef %t46), !dbg !33
  %t60 = call ptr @malloc(i64 512), !dbg !38
  store i8 0, ptr %t60, !dbg !38
  %t61 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !38
  %t62 = getelementptr inbounds [11 x i8], ptr @.str.7, i64 0, i64 0, !dbg !39
  %t63 = call ptr @strcat(ptr %t60, ptr %t62), !dbg !38
  %t64 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !38
  %t65 = call ptr @strcat(ptr %t60, ptr %t64), !dbg !38
  %t66 = load i64, ptr %v_a_int, !dbg !40
  %t67 = load i64, ptr %v_b_int, !dbg !41
  %t68 = srem i64 %t66, %t67, !dbg !42
  %t69 = call i64 @strlen(ptr %t60), !dbg !38
  %t70 = getelementptr inbounds i8, ptr %t60, i64 %t69, !dbg !38
  %t71 = sub i64 512, %t69, !dbg !38
  %t72 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t70, i64 %t71, ptr %t61, i64 %t68), !dbg !38
  %t73 = call i32 @puts(ptr noundef %t60), !dbg !38
  %t74 = call ptr @malloc(i64 512), !dbg !43
  store i8 0, ptr %t74, !dbg !43
  %t75 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !43
  %t76 = getelementptr inbounds [11 x i8], ptr @.str.8, i64 0, i64 0, !dbg !44
  %t77 = call ptr @strcat(ptr %t74, ptr %t76), !dbg !43
  %t78 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !43
  %t79 = call ptr @strcat(ptr %t74, ptr %t78), !dbg !43
  %t80 = load i64, ptr %v_a_int, !dbg !45
  %t81 = load i64, ptr %v_b_int, !dbg !46
  %t82 = sdiv i64 %t80, %t81, !dbg !47
  %t83 = call i64 @strlen(ptr %t74), !dbg !43
  %t84 = getelementptr inbounds i8, ptr %t74, i64 %t83, !dbg !43
  %t85 = sub i64 512, %t83, !dbg !43
  %t86 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t84, i64 %t85, ptr %t75, i64 %t82), !dbg !43
  %t87 = call i32 @puts(ptr noundef %t74), !dbg !43
  %t88 = getelementptr inbounds [1 x i8], ptr @.str.9, i64 0, i64 0, !dbg !48
  %t89 = call i32 @puts(ptr noundef %t88), !dbg !49
  %t90 = getelementptr inbounds [19 x i8], ptr @.str.10, i64 0, i64 0, !dbg !50
  %t91 = call i32 @puts(ptr noundef %t90), !dbg !51
  %t92 = call ptr @malloc(i64 512), !dbg !52
  store i8 0, ptr %t92, !dbg !52
  %t94 = getelementptr inbounds [8 x i8], ptr @.str.11, i64 0, i64 0, !dbg !53
  %t95 = call ptr @strcat(ptr %t92, ptr %t94), !dbg !52
  %t96 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !52
  %t97 = call ptr @strcat(ptr %t92, ptr %t96), !dbg !52
  %t98 = load i64, ptr %v_a_int, !dbg !54
  %t99 = load i64, ptr %v_b_int, !dbg !55
  %t100 = icmp sgt i64 %t98, %t99, !dbg !56
  %t101 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !52
  %t102 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !52
  %t103 = select i1 %t100, ptr %t101, ptr %t102, !dbg !52
  %t104 = call ptr @strcat(ptr %t92, ptr %t103), !dbg !52
  %t105 = call i32 @puts(ptr noundef %t92), !dbg !52
  %t106 = call ptr @malloc(i64 512), !dbg !57
  store i8 0, ptr %t106, !dbg !57
  %t108 = getelementptr inbounds [8 x i8], ptr @.str.14, i64 0, i64 0, !dbg !58
  %t109 = call ptr @strcat(ptr %t106, ptr %t108), !dbg !57
  %t110 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !57
  %t111 = call ptr @strcat(ptr %t106, ptr %t110), !dbg !57
  %t112 = load i64, ptr %v_a_int, !dbg !59
  %t113 = load i64, ptr %v_b_int, !dbg !60
  %t114 = icmp slt i64 %t112, %t113, !dbg !61
  %t115 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !57
  %t116 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !57
  %t117 = select i1 %t114, ptr %t115, ptr %t116, !dbg !57
  %t118 = call ptr @strcat(ptr %t106, ptr %t117), !dbg !57
  %t119 = call i32 @puts(ptr noundef %t106), !dbg !57
  %t120 = call ptr @malloc(i64 512), !dbg !62
  store i8 0, ptr %t120, !dbg !62
  %t122 = getelementptr inbounds [9 x i8], ptr @.str.15, i64 0, i64 0, !dbg !63
  %t123 = call ptr @strcat(ptr %t120, ptr %t122), !dbg !62
  %t124 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !62
  %t125 = call ptr @strcat(ptr %t120, ptr %t124), !dbg !62
  %t126 = load i64, ptr %v_a_int, !dbg !64
  %t127 = load i64, ptr %v_b_int, !dbg !65
  %t128 = icmp sge i64 %t126, %t127, !dbg !66
  %t129 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !62
  %t130 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !62
  %t131 = select i1 %t128, ptr %t129, ptr %t130, !dbg !62
  %t132 = call ptr @strcat(ptr %t120, ptr %t131), !dbg !62
  %t133 = call i32 @puts(ptr noundef %t120), !dbg !62
  %t134 = call ptr @malloc(i64 512), !dbg !67
  store i8 0, ptr %t134, !dbg !67
  %t136 = getelementptr inbounds [9 x i8], ptr @.str.16, i64 0, i64 0, !dbg !68
  %t137 = call ptr @strcat(ptr %t134, ptr %t136), !dbg !67
  %t138 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !67
  %t139 = call ptr @strcat(ptr %t134, ptr %t138), !dbg !67
  %t140 = load i64, ptr %v_a_int, !dbg !69
  %t141 = load i64, ptr %v_b_int, !dbg !70
  %t142 = icmp sle i64 %t140, %t141, !dbg !71
  %t143 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !67
  %t144 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !67
  %t145 = select i1 %t142, ptr %t143, ptr %t144, !dbg !67
  %t146 = call ptr @strcat(ptr %t134, ptr %t145), !dbg !67
  %t147 = call i32 @puts(ptr noundef %t134), !dbg !67
  %t148 = call ptr @malloc(i64 512), !dbg !72
  store i8 0, ptr %t148, !dbg !72
  %t150 = getelementptr inbounds [8 x i8], ptr @.str.17, i64 0, i64 0, !dbg !73
  %t151 = call ptr @strcat(ptr %t148, ptr %t150), !dbg !72
  %t152 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !72
  %t153 = call ptr @strcat(ptr %t148, ptr %t152), !dbg !72
  %t154 = load i64, ptr %v_a_int, !dbg !74
  %t155 = load i64, ptr %v_b_int, !dbg !75
  %t156 = icmp eq i64 %t154, %t155, !dbg !76
  %t157 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !72
  %t158 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !72
  %t159 = select i1 %t156, ptr %t157, ptr %t158, !dbg !72
  %t160 = call ptr @strcat(ptr %t148, ptr %t159), !dbg !72
  %t161 = call i32 @puts(ptr noundef %t148), !dbg !72
  %t162 = call ptr @malloc(i64 512), !dbg !77
  store i8 0, ptr %t162, !dbg !77
  %t164 = getelementptr inbounds [9 x i8], ptr @.str.18, i64 0, i64 0, !dbg !78
  %t165 = call ptr @strcat(ptr %t162, ptr %t164), !dbg !77
  %t166 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !77
  %t167 = call ptr @strcat(ptr %t162, ptr %t166), !dbg !77
  %t168 = load i64, ptr %v_a_int, !dbg !79
  %t169 = load i64, ptr %v_b_int, !dbg !80
  %t170 = icmp ne i64 %t168, %t169, !dbg !81
  %t171 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !77
  %t172 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !77
  %t173 = select i1 %t170, ptr %t171, ptr %t172, !dbg !77
  %t174 = call ptr @strcat(ptr %t162, ptr %t173), !dbg !77
  %t175 = call i32 @puts(ptr noundef %t162), !dbg !77
  %t176 = getelementptr inbounds [1 x i8], ptr @.str.9, i64 0, i64 0, !dbg !82
  %t177 = call i32 @puts(ptr noundef %t176), !dbg !83
  %t178 = getelementptr inbounds [16 x i8], ptr @.str.19, i64 0, i64 0, !dbg !84
  %t179 = call i32 @puts(ptr noundef %t178), !dbg !85
  %t180 = add i1 0, 1, !dbg !86
  store i1 %t180, ptr %v_flag1_bool, !dbg !87
  %t181 = add i1 0, 0, !dbg !88
  store i1 %t181, ptr %v_flag2_bool, !dbg !89
  %t182 = call ptr @malloc(i64 512), !dbg !90
  store i8 0, ptr %t182, !dbg !90
  %t184 = getelementptr inbounds [18 x i8], ptr @.str.20, i64 0, i64 0, !dbg !91
  %t185 = call ptr @strcat(ptr %t182, ptr %t184), !dbg !90
  %t186 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !90
  %t187 = call ptr @strcat(ptr %t182, ptr %t186), !dbg !90
  %t188 = load i1, ptr %v_flag1_bool, !dbg !92
  %t189 = load i1, ptr %v_flag2_bool, !dbg !93
  %t190 = and i1 %t188, %t189, !dbg !94
  %t191 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !90
  %t192 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !90
  %t193 = select i1 %t190, ptr %t191, ptr %t192, !dbg !90
  %t194 = call ptr @strcat(ptr %t182, ptr %t193), !dbg !90
  %t195 = call i32 @puts(ptr noundef %t182), !dbg !90
  %t196 = call ptr @malloc(i64 512), !dbg !95
  store i8 0, ptr %t196, !dbg !95
  %t198 = getelementptr inbounds [17 x i8], ptr @.str.21, i64 0, i64 0, !dbg !96
  %t199 = call ptr @strcat(ptr %t196, ptr %t198), !dbg !95
  %t200 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !95
  %t201 = call ptr @strcat(ptr %t196, ptr %t200), !dbg !95
  %t202 = load i1, ptr %v_flag1_bool, !dbg !97
  %t203 = load i1, ptr %v_flag2_bool, !dbg !98
  %t204 = or i1 %t202, %t203, !dbg !99
  %t205 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !95
  %t206 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !95
  %t207 = select i1 %t204, ptr %t205, ptr %t206, !dbg !95
  %t208 = call ptr @strcat(ptr %t196, ptr %t207), !dbg !95
  %t209 = call i32 @puts(ptr noundef %t196), !dbg !95
  %t210 = call ptr @malloc(i64 512), !dbg !100
  store i8 0, ptr %t210, !dbg !100
  %t212 = getelementptr inbounds [12 x i8], ptr @.str.22, i64 0, i64 0, !dbg !101
  %t213 = call ptr @strcat(ptr %t210, ptr %t212), !dbg !100
  %t214 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !100
  %t215 = call ptr @strcat(ptr %t210, ptr %t214), !dbg !100
  %t216 = load i1, ptr %v_flag1_bool, !dbg !102
  %t217 = xor i1 %t216, 1, !dbg !103
  %t218 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !100
  %t219 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !100
  %t220 = select i1 %t217, ptr %t218, ptr %t219, !dbg !100
  %t221 = call ptr @strcat(ptr %t210, ptr %t220), !dbg !100
  %t222 = call i32 @puts(ptr noundef %t210), !dbg !100
  %t223 = call ptr @malloc(i64 512), !dbg !104
  store i8 0, ptr %t223, !dbg !104
  %t225 = getelementptr inbounds [12 x i8], ptr @.str.23, i64 0, i64 0, !dbg !105
  %t226 = call ptr @strcat(ptr %t223, ptr %t225), !dbg !104
  %t227 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !104
  %t228 = call ptr @strcat(ptr %t223, ptr %t227), !dbg !104
  %t229 = load i1, ptr %v_flag2_bool, !dbg !106
  %t230 = xor i1 %t229, 1, !dbg !107
  %t231 = getelementptr inbounds [5 x i8], ptr @.str.12, i64 0, i64 0, !dbg !104
  %t232 = getelementptr inbounds [6 x i8], ptr @.str.13, i64 0, i64 0, !dbg !104
  %t233 = select i1 %t230, ptr %t231, ptr %t232, !dbg !104
  %t234 = call ptr @strcat(ptr %t223, ptr %t233), !dbg !104
  %t235 = call i32 @puts(ptr noundef %t223), !dbg !104
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [19 x i8] c"=== Arithmetic ===\00", align 1
@.str.1 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.2 = private unnamed_addr constant [9 x i8] c"a + b = \00", align 1
@.str.3 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.4 = private unnamed_addr constant [9 x i8] c"a - b = \00", align 1
@.str.5 = private unnamed_addr constant [9 x i8] c"a * b = \00", align 1
@.str.6 = private unnamed_addr constant [9 x i8] c"a / b = \00", align 1
@.str.7 = private unnamed_addr constant [11 x i8] c"a mod b = \00", align 1
@.str.8 = private unnamed_addr constant [11 x i8] c"a div b = \00", align 1
@.str.9 = private unnamed_addr constant [1 x i8] c"\00", align 1
@.str.10 = private unnamed_addr constant [19 x i8] c"=== Comparison ===\00", align 1
@.str.11 = private unnamed_addr constant [8 x i8] c"a > b: \00", align 1
@.str.12 = private unnamed_addr constant [5 x i8] c"true\00", align 1
@.str.13 = private unnamed_addr constant [6 x i8] c"false\00", align 1
@.str.14 = private unnamed_addr constant [8 x i8] c"a < b: \00", align 1
@.str.15 = private unnamed_addr constant [9 x i8] c"a >= b: \00", align 1
@.str.16 = private unnamed_addr constant [9 x i8] c"a <= b: \00", align 1
@.str.17 = private unnamed_addr constant [8 x i8] c"a = b: \00", align 1
@.str.18 = private unnamed_addr constant [9 x i8] c"a <> b: \00", align 1
@.str.19 = private unnamed_addr constant [16 x i8] c"=== Logical ===\00", align 1
@.str.20 = private unnamed_addr constant [18 x i8] c"flag1 and flag2: \00", align 1
@.str.21 = private unnamed_addr constant [17 x i8] c"flag1 or flag2: \00", align 1
@.str.22 = private unnamed_addr constant [12 x i8] c"not flag1: \00", align 1
@.str.23 = private unnamed_addr constant [12 x i8] c"not flag2: \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example05_operators.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/01_basics")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !109, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !112)
!6 = !DILocalVariable(name: "a", scope: !4, file: !3, line: 4, type: !111)
!7 = !DILocalVariable(name: "b", scope: !4, file: !3, line: 4, type: !111)
!8 = !DILocalVariable(name: "x", scope: !4, file: !3, line: 5, type: !111)
!9 = !DILocalVariable(name: "y", scope: !4, file: !3, line: 5, type: !111)
!10 = !DILocalVariable(name: "flag1", scope: !4, file: !3, line: 6, type: !111)
!11 = !DILocalVariable(name: "flag2", scope: !4, file: !3, line: 6, type: !111)
!5 = !DILocation(line: 1, column: 9, scope: !4)
!12 = !DILocation(line: 10, column: 8, scope: !4)
!13 = !DILocation(line: 10, column: 6, scope: !4)
!14 = !DILocation(line: 11, column: 8, scope: !4)
!15 = !DILocation(line: 11, column: 6, scope: !4)
!16 = !DILocation(line: 13, column: 11, scope: !4)
!17 = !DILocation(line: 13, column: 10, scope: !4)
!18 = !DILocation(line: 14, column: 10, scope: !4)
!19 = !DILocation(line: 14, column: 11, scope: !4)
!20 = !DILocation(line: 14, column: 23, scope: !4)
!21 = !DILocation(line: 14, column: 27, scope: !4)
!22 = !DILocation(line: 14, column: 25, scope: !4)
!23 = !DILocation(line: 15, column: 10, scope: !4)
!24 = !DILocation(line: 15, column: 11, scope: !4)
!25 = !DILocation(line: 15, column: 23, scope: !4)
!26 = !DILocation(line: 15, column: 27, scope: !4)
!27 = !DILocation(line: 15, column: 25, scope: !4)
!28 = !DILocation(line: 16, column: 10, scope: !4)
!29 = !DILocation(line: 16, column: 11, scope: !4)
!30 = !DILocation(line: 16, column: 23, scope: !4)
!31 = !DILocation(line: 16, column: 27, scope: !4)
!32 = !DILocation(line: 16, column: 25, scope: !4)
!33 = !DILocation(line: 17, column: 10, scope: !4)
!34 = !DILocation(line: 17, column: 11, scope: !4)
!35 = !DILocation(line: 17, column: 23, scope: !4)
!36 = !DILocation(line: 17, column: 27, scope: !4)
!37 = !DILocation(line: 17, column: 25, scope: !4)
!38 = !DILocation(line: 18, column: 10, scope: !4)
!39 = !DILocation(line: 18, column: 11, scope: !4)
!40 = !DILocation(line: 18, column: 25, scope: !4)
!41 = !DILocation(line: 18, column: 31, scope: !4)
!42 = !DILocation(line: 18, column: 27, scope: !4)
!43 = !DILocation(line: 19, column: 10, scope: !4)
!44 = !DILocation(line: 19, column: 11, scope: !4)
!45 = !DILocation(line: 19, column: 25, scope: !4)
!46 = !DILocation(line: 19, column: 31, scope: !4)
!47 = !DILocation(line: 19, column: 27, scope: !4)
!48 = !DILocation(line: 22, column: 11, scope: !4)
!49 = !DILocation(line: 22, column: 10, scope: !4)
!50 = !DILocation(line: 23, column: 11, scope: !4)
!51 = !DILocation(line: 23, column: 10, scope: !4)
!52 = !DILocation(line: 24, column: 10, scope: !4)
!53 = !DILocation(line: 24, column: 11, scope: !4)
!54 = !DILocation(line: 24, column: 22, scope: !4)
!55 = !DILocation(line: 24, column: 26, scope: !4)
!56 = !DILocation(line: 24, column: 24, scope: !4)
!57 = !DILocation(line: 25, column: 10, scope: !4)
!58 = !DILocation(line: 25, column: 11, scope: !4)
!59 = !DILocation(line: 25, column: 22, scope: !4)
!60 = !DILocation(line: 25, column: 26, scope: !4)
!61 = !DILocation(line: 25, column: 24, scope: !4)
!62 = !DILocation(line: 26, column: 10, scope: !4)
!63 = !DILocation(line: 26, column: 11, scope: !4)
!64 = !DILocation(line: 26, column: 23, scope: !4)
!65 = !DILocation(line: 26, column: 28, scope: !4)
!66 = !DILocation(line: 26, column: 26, scope: !4)
!67 = !DILocation(line: 27, column: 10, scope: !4)
!68 = !DILocation(line: 27, column: 11, scope: !4)
!69 = !DILocation(line: 27, column: 23, scope: !4)
!70 = !DILocation(line: 27, column: 28, scope: !4)
!71 = !DILocation(line: 27, column: 26, scope: !4)
!72 = !DILocation(line: 28, column: 10, scope: !4)
!73 = !DILocation(line: 28, column: 11, scope: !4)
!74 = !DILocation(line: 28, column: 22, scope: !4)
!75 = !DILocation(line: 28, column: 26, scope: !4)
!76 = !DILocation(line: 28, column: 24, scope: !4)
!77 = !DILocation(line: 29, column: 10, scope: !4)
!78 = !DILocation(line: 29, column: 11, scope: !4)
!79 = !DILocation(line: 29, column: 23, scope: !4)
!80 = !DILocation(line: 29, column: 28, scope: !4)
!81 = !DILocation(line: 29, column: 26, scope: !4)
!82 = !DILocation(line: 32, column: 11, scope: !4)
!83 = !DILocation(line: 32, column: 10, scope: !4)
!84 = !DILocation(line: 33, column: 11, scope: !4)
!85 = !DILocation(line: 33, column: 10, scope: !4)
!86 = !DILocation(line: 34, column: 12, scope: !4)
!87 = !DILocation(line: 34, column: 10, scope: !4)
!88 = !DILocation(line: 35, column: 12, scope: !4)
!89 = !DILocation(line: 35, column: 10, scope: !4)
!90 = !DILocation(line: 37, column: 10, scope: !4)
!91 = !DILocation(line: 37, column: 11, scope: !4)
!92 = !DILocation(line: 37, column: 32, scope: !4)
!93 = !DILocation(line: 37, column: 42, scope: !4)
!94 = !DILocation(line: 37, column: 38, scope: !4)
!95 = !DILocation(line: 38, column: 10, scope: !4)
!96 = !DILocation(line: 38, column: 11, scope: !4)
!97 = !DILocation(line: 38, column: 31, scope: !4)
!98 = !DILocation(line: 38, column: 40, scope: !4)
!99 = !DILocation(line: 38, column: 37, scope: !4)
!100 = !DILocation(line: 39, column: 10, scope: !4)
!101 = !DILocation(line: 39, column: 11, scope: !4)
!102 = !DILocation(line: 39, column: 30, scope: !4)
!103 = !DILocation(line: 39, column: 26, scope: !4)
!104 = !DILocation(line: 40, column: 10, scope: !4)
!105 = !DILocation(line: 40, column: 11, scope: !4)
!106 = !DILocation(line: 40, column: 30, scope: !4)
!107 = !DILocation(line: 40, column: 26, scope: !4)
!108 = !{null}
!109 = !DISubroutineType(types: !108)
!110 = !{}
!111 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!112 = !{!6, !7, !8, !9, !10, !11}
