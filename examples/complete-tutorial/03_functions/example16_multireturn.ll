; Kylix LLVM IR — module: MultiReturn
source_filename = "MultiReturn.klx"
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
%__ret_DivMod = type { i64, i64 }
%__ret_MinMax = type { i64, i64 }
define %__ret_DivMod @DivMod(i64 %dividend, i64 %divisor) !dbg !4 {
entry:
  %result = alloca %__ret_DivMod, align 8
  #dbg_declare(ptr %result, !5, !DIExpression(), !6)
  %v_dividend_int = alloca i64, align 8, !dbg !6
  store i64 %dividend, ptr %v_dividend_int, !dbg !6
  #dbg_declare(ptr %v_dividend_int, !7, !DIExpression(), !6)
  %v_divisor_int = alloca i64, align 8, !dbg !6
  store i64 %divisor, ptr %v_divisor_int, !dbg !6
  #dbg_declare(ptr %v_divisor_int, !8, !DIExpression(), !6)
  %t0 = load i64, ptr %v_dividend_int, !dbg !9
  %t1 = load i64, ptr %v_divisor_int, !dbg !10
  %t2 = sdiv i64 %t0, %t1, !dbg !11
  %t3 = insertvalue %__ret_DivMod undef, i64 %t2, 0, !dbg !12
  %t4 = load i64, ptr %v_dividend_int, !dbg !13
  %t5 = load i64, ptr %v_divisor_int, !dbg !14
  %t6 = srem i64 %t4, %t5, !dbg !15
  %t7 = insertvalue %__ret_DivMod %t3, i64 %t6, 1, !dbg !12
  store %__ret_DivMod %t7, ptr %result, !dbg !12
  %t8 = load %__ret_DivMod, ptr %result, !dbg !6
  ret %__ret_DivMod %t8, !dbg !6
}

define %__ret_MinMax @MinMax(i64 %a, i64 %b, i64 %c) !dbg !16 {
entry:
  %result = alloca %__ret_MinMax, align 8
  #dbg_declare(ptr %result, !17, !DIExpression(), !18)
  %v_a_int = alloca i64, align 8, !dbg !18
  store i64 %a, ptr %v_a_int, !dbg !18
  #dbg_declare(ptr %v_a_int, !19, !DIExpression(), !18)
  %v_b_int = alloca i64, align 8, !dbg !18
  store i64 %b, ptr %v_b_int, !dbg !18
  #dbg_declare(ptr %v_b_int, !20, !DIExpression(), !18)
  %v_c_int = alloca i64, align 8, !dbg !18
  store i64 %c, ptr %v_c_int, !dbg !18
  #dbg_declare(ptr %v_c_int, !21, !DIExpression(), !18)
  %v_min_int = alloca i64, align 8, !dbg !18
  store i64 0, ptr %v_min_int, !dbg !18
  #dbg_declare(ptr %v_min_int, !22, !DIExpression(), !18)
  %v_max_int = alloca i64, align 8, !dbg !18
  store i64 0, ptr %v_max_int, !dbg !18
  #dbg_declare(ptr %v_max_int, !23, !DIExpression(), !18)
  %t9 = load i64, ptr %v_a_int, !dbg !24
  store i64 %t9, ptr %v_min_int, !dbg !25
  %t10 = load i64, ptr %v_a_int, !dbg !26
  store i64 %t10, ptr %v_max_int, !dbg !27
  %t11 = load i64, ptr %v_b_int, !dbg !28
  %t12 = load i64, ptr %v_min_int, !dbg !29
  %t13 = icmp slt i64 %t11, %t12, !dbg !30
  br i1 %t13, label %lbl0, label %lbl1, !dbg !31
lbl0:
  %t14 = load i64, ptr %v_b_int, !dbg !32
  store i64 %t14, ptr %v_min_int, !dbg !33
  br label %lbl1, !dbg !31
lbl1:
  %t15 = load i64, ptr %v_c_int, !dbg !34
  %t16 = load i64, ptr %v_min_int, !dbg !35
  %t17 = icmp slt i64 %t15, %t16, !dbg !36
  br i1 %t17, label %lbl2, label %lbl3, !dbg !37
lbl2:
  %t18 = load i64, ptr %v_c_int, !dbg !38
  store i64 %t18, ptr %v_min_int, !dbg !39
  br label %lbl3, !dbg !37
lbl3:
  %t19 = load i64, ptr %v_b_int, !dbg !40
  %t20 = load i64, ptr %v_max_int, !dbg !41
  %t21 = icmp sgt i64 %t19, %t20, !dbg !42
  br i1 %t21, label %lbl4, label %lbl5, !dbg !43
lbl4:
  %t22 = load i64, ptr %v_b_int, !dbg !44
  store i64 %t22, ptr %v_max_int, !dbg !45
  br label %lbl5, !dbg !43
lbl5:
  %t23 = load i64, ptr %v_c_int, !dbg !46
  %t24 = load i64, ptr %v_max_int, !dbg !47
  %t25 = icmp sgt i64 %t23, %t24, !dbg !48
  br i1 %t25, label %lbl6, label %lbl7, !dbg !49
lbl6:
  %t26 = load i64, ptr %v_c_int, !dbg !50
  store i64 %t26, ptr %v_max_int, !dbg !51
  br label %lbl7, !dbg !49
lbl7:
  %t27 = load i64, ptr %v_min_int, !dbg !52
  %t28 = insertvalue %__ret_MinMax undef, i64 %t27, 0, !dbg !53
  %t29 = load i64, ptr %v_max_int, !dbg !54
  %t30 = insertvalue %__ret_MinMax %t28, i64 %t29, 1, !dbg !53
  store %__ret_MinMax %t30, ptr %result, !dbg !53
  %t31 = load %__ret_MinMax, ptr %result, !dbg !18
  ret %__ret_MinMax %t31, !dbg !18
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
define i32 @main() !dbg !55 {
entry:
  %v_quotient_int = alloca i64, align 8, !dbg !56
  store i64 0, ptr %v_quotient_int, !dbg !56
  #dbg_declare(ptr %v_quotient_int, !57, !DIExpression(), !56)
  %v_remainder_int = alloca i64, align 8, !dbg !56
  store i64 0, ptr %v_remainder_int, !dbg !56
  #dbg_declare(ptr %v_remainder_int, !58, !DIExpression(), !56)
  %v_minVal_int = alloca i64, align 8, !dbg !56
  store i64 0, ptr %v_minVal_int, !dbg !56
  #dbg_declare(ptr %v_minVal_int, !59, !DIExpression(), !56)
  %v_maxVal_int = alloca i64, align 8, !dbg !56
  store i64 0, ptr %v_maxVal_int, !dbg !56
  #dbg_declare(ptr %v_maxVal_int, !60, !DIExpression(), !56)
  %v_q2_int = alloca i64, align 8, !dbg !56
  store i64 0, ptr %v_q2_int, !dbg !56
  #dbg_declare(ptr %v_q2_int, !61, !DIExpression(), !56)
  %v_r2_int = alloca i64, align 8, !dbg !56
  store i64 0, ptr %v_r2_int, !dbg !56
  #dbg_declare(ptr %v_r2_int, !62, !DIExpression(), !56)
  %t32 = add i64 0, 17, !dbg !63
  %t33 = add i64 0, 5, !dbg !64
  %t34 = call %__ret_DivMod @DivMod(i64 %t32, i64 %t33), !dbg !65
  %t35 = extractvalue %__ret_DivMod %t34, 0, !dbg !66
  store i64 %t35, ptr %v_quotient_int, !dbg !66
  %t36 = extractvalue %__ret_DivMod %t34, 1, !dbg !66
  store i64 %t36, ptr %v_remainder_int, !dbg !66
  %t37 = call ptr @malloc(i64 512), !dbg !67
  store i8 0, ptr %t37, !dbg !67
  %t38 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !67
  %t39 = getelementptr inbounds [12 x i8], ptr @.str.1, i64 0, i64 0, !dbg !68
  %t40 = call ptr @strcat(ptr %t37, ptr %t39), !dbg !67
  %t41 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !67
  %t42 = call ptr @strcat(ptr %t37, ptr %t41), !dbg !67
  %t43 = load i64, ptr %v_quotient_int, !dbg !69
  %t44 = call i64 @strlen(ptr %t37), !dbg !67
  %t45 = getelementptr inbounds i8, ptr %t37, i64 %t44, !dbg !67
  %t46 = sub i64 512, %t44, !dbg !67
  %t47 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t45, i64 %t46, ptr %t38, i64 %t43), !dbg !67
  %t48 = call i32 @puts(ptr noundef %t37), !dbg !67
  %t49 = call ptr @malloc(i64 512), !dbg !70
  store i8 0, ptr %t49, !dbg !70
  %t50 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !70
  %t51 = getelementptr inbounds [12 x i8], ptr @.str.3, i64 0, i64 0, !dbg !71
  %t52 = call ptr @strcat(ptr %t49, ptr %t51), !dbg !70
  %t53 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !70
  %t54 = call ptr @strcat(ptr %t49, ptr %t53), !dbg !70
  %t55 = load i64, ptr %v_remainder_int, !dbg !72
  %t56 = call i64 @strlen(ptr %t49), !dbg !70
  %t57 = getelementptr inbounds i8, ptr %t49, i64 %t56, !dbg !70
  %t58 = sub i64 512, %t56, !dbg !70
  %t59 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t57, i64 %t58, ptr %t50, i64 %t55), !dbg !70
  %t60 = call i32 @puts(ptr noundef %t49), !dbg !70
  %t61 = add i64 0, 5, !dbg !73
  %t62 = add i64 0, 12, !dbg !74
  %t63 = add i64 0, 3, !dbg !75
  %t64 = call %__ret_MinMax @MinMax(i64 %t61, i64 %t62, i64 %t63), !dbg !76
  %t65 = extractvalue %__ret_MinMax %t64, 0, !dbg !77
  store i64 %t65, ptr %v_minVal_int, !dbg !77
  %t66 = extractvalue %__ret_MinMax %t64, 1, !dbg !77
  store i64 %t66, ptr %v_maxVal_int, !dbg !77
  %t67 = call ptr @malloc(i64 512), !dbg !78
  store i8 0, ptr %t67, !dbg !78
  %t68 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !78
  %t69 = getelementptr inbounds [16 x i8], ptr @.str.4, i64 0, i64 0, !dbg !79
  %t70 = call ptr @strcat(ptr %t67, ptr %t69), !dbg !78
  %t71 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !78
  %t72 = call ptr @strcat(ptr %t67, ptr %t71), !dbg !78
  %t73 = load i64, ptr %v_minVal_int, !dbg !80
  %t74 = call i64 @strlen(ptr %t67), !dbg !78
  %t75 = getelementptr inbounds i8, ptr %t67, i64 %t74, !dbg !78
  %t76 = sub i64 512, %t74, !dbg !78
  %t77 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t75, i64 %t76, ptr %t68, i64 %t73), !dbg !78
  %t78 = call i32 @puts(ptr noundef %t67), !dbg !78
  %t79 = call ptr @malloc(i64 512), !dbg !81
  store i8 0, ptr %t79, !dbg !81
  %t80 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !81
  %t81 = getelementptr inbounds [16 x i8], ptr @.str.5, i64 0, i64 0, !dbg !82
  %t82 = call ptr @strcat(ptr %t79, ptr %t81), !dbg !81
  %t83 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !81
  %t84 = call ptr @strcat(ptr %t79, ptr %t83), !dbg !81
  %t85 = load i64, ptr %v_maxVal_int, !dbg !83
  %t86 = call i64 @strlen(ptr %t79), !dbg !81
  %t87 = getelementptr inbounds i8, ptr %t79, i64 %t86, !dbg !81
  %t88 = sub i64 512, %t86, !dbg !81
  %t89 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t87, i64 %t88, ptr %t80, i64 %t85), !dbg !81
  %t90 = call i32 @puts(ptr noundef %t79), !dbg !81
  %t91 = add i64 0, 100, !dbg !84
  %t92 = add i64 0, 7, !dbg !85
  %t93 = call %__ret_DivMod @DivMod(i64 %t91, i64 %t92), !dbg !86
  %t94 = extractvalue %__ret_DivMod %t93, 0, !dbg !87
  store i64 %t94, ptr %v_q2_int, !dbg !87
  %t95 = extractvalue %__ret_DivMod %t93, 1, !dbg !87
  store i64 %t95, ptr %v_r2_int, !dbg !87
  %t96 = call ptr @malloc(i64 512), !dbg !88
  store i8 0, ptr %t96, !dbg !88
  %t97 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !88
  %t98 = getelementptr inbounds [13 x i8], ptr @.str.6, i64 0, i64 0, !dbg !89
  %t99 = call ptr @strcat(ptr %t96, ptr %t98), !dbg !88
  %t100 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !88
  %t101 = call ptr @strcat(ptr %t96, ptr %t100), !dbg !88
  %t102 = load i64, ptr %v_q2_int, !dbg !90
  %t103 = call i64 @strlen(ptr %t96), !dbg !88
  %t104 = getelementptr inbounds i8, ptr %t96, i64 %t103, !dbg !88
  %t105 = sub i64 512, %t103, !dbg !88
  %t106 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t104, i64 %t105, ptr %t97, i64 %t102), !dbg !88
  %t107 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !88
  %t108 = call ptr @strcat(ptr %t96, ptr %t107), !dbg !88
  %t109 = getelementptr inbounds [15 x i8], ptr @.str.7, i64 0, i64 0, !dbg !91
  %t110 = call ptr @strcat(ptr %t96, ptr %t109), !dbg !88
  %t111 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !88
  %t112 = call ptr @strcat(ptr %t96, ptr %t111), !dbg !88
  %t113 = load i64, ptr %v_r2_int, !dbg !92
  %t114 = call i64 @strlen(ptr %t96), !dbg !88
  %t115 = getelementptr inbounds i8, ptr %t96, i64 %t114, !dbg !88
  %t116 = sub i64 512, %t114, !dbg !88
  %t117 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t115, i64 %t116, ptr %t97, i64 %t113), !dbg !88
  %t118 = call i32 @puts(ptr noundef %t96), !dbg !88
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [12 x i8] c"17 div 5 = \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [12 x i8] c"17 mod 5 = \00", align 1
@.str.4 = private unnamed_addr constant [16 x i8] c"Min of 5,12,3: \00", align 1
@.str.5 = private unnamed_addr constant [16 x i8] c"Max of 5,12,3: \00", align 1
@.str.6 = private unnamed_addr constant [13 x i8] c"100 div 7 = \00", align 1
@.str.7 = private unnamed_addr constant [15 x i8] c", remainder = \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example16_multireturn.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/03_functions")
!4 = distinct !DISubprogram(name: "DivMod", scope: !3, file: !3, line: 4, type: !94, scopeLine: 4, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !97)
!16 = distinct !DISubprogram(name: "MinMax", scope: !3, file: !3, line: 10, type: !94, scopeLine: 10, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !98)
!55 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !94, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !99)
!5 = !DILocalVariable(name: "result", scope: !4, file: !3, line: 4, type: !96)
!7 = !DILocalVariable(name: "dividend", scope: !4, file: !3, line: 4, type: !96)
!8 = !DILocalVariable(name: "divisor", scope: !4, file: !3, line: 4, type: !96)
!17 = !DILocalVariable(name: "result", scope: !16, file: !3, line: 10, type: !96)
!19 = !DILocalVariable(name: "a", scope: !16, file: !3, line: 10, type: !96)
!20 = !DILocalVariable(name: "b", scope: !16, file: !3, line: 10, type: !96)
!21 = !DILocalVariable(name: "c", scope: !16, file: !3, line: 10, type: !96)
!22 = !DILocalVariable(name: "min", scope: !16, file: !3, line: 12, type: !96)
!23 = !DILocalVariable(name: "max", scope: !16, file: !3, line: 12, type: !96)
!57 = !DILocalVariable(name: "quotient", scope: !55, file: !3, line: 27, type: !96)
!58 = !DILocalVariable(name: "remainder", scope: !55, file: !3, line: 27, type: !96)
!59 = !DILocalVariable(name: "minVal", scope: !55, file: !3, line: 28, type: !96)
!60 = !DILocalVariable(name: "maxVal", scope: !55, file: !3, line: 28, type: !96)
!61 = !DILocalVariable(name: "q2", scope: !55, file: !3, line: 29, type: !96)
!62 = !DILocalVariable(name: "r2", scope: !55, file: !3, line: 29, type: !96)
!6 = !DILocation(line: 4, column: 1, scope: !4)
!9 = !DILocation(line: 6, column: 14, scope: !4)
!10 = !DILocation(line: 6, column: 27, scope: !4)
!11 = !DILocation(line: 6, column: 23, scope: !4)
!12 = !DILocation(line: 6, column: 11, scope: !4)
!13 = !DILocation(line: 6, column: 36, scope: !4)
!14 = !DILocation(line: 6, column: 49, scope: !4)
!15 = !DILocation(line: 6, column: 45, scope: !4)
!18 = !DILocation(line: 10, column: 1, scope: !16)
!24 = !DILocation(line: 14, column: 10, scope: !16)
!25 = !DILocation(line: 14, column: 8, scope: !16)
!26 = !DILocation(line: 15, column: 10, scope: !16)
!27 = !DILocation(line: 15, column: 8, scope: !16)
!28 = !DILocation(line: 17, column: 6, scope: !16)
!29 = !DILocation(line: 17, column: 10, scope: !16)
!30 = !DILocation(line: 17, column: 8, scope: !16)
!31 = !DILocation(line: 17, column: 3, scope: !16)
!32 = !DILocation(line: 17, column: 26, scope: !16)
!33 = !DILocation(line: 17, column: 24, scope: !16)
!34 = !DILocation(line: 18, column: 6, scope: !16)
!35 = !DILocation(line: 18, column: 10, scope: !16)
!36 = !DILocation(line: 18, column: 8, scope: !16)
!37 = !DILocation(line: 18, column: 3, scope: !16)
!38 = !DILocation(line: 18, column: 26, scope: !16)
!39 = !DILocation(line: 18, column: 24, scope: !16)
!40 = !DILocation(line: 20, column: 6, scope: !16)
!41 = !DILocation(line: 20, column: 10, scope: !16)
!42 = !DILocation(line: 20, column: 8, scope: !16)
!43 = !DILocation(line: 20, column: 3, scope: !16)
!44 = !DILocation(line: 20, column: 26, scope: !16)
!45 = !DILocation(line: 20, column: 24, scope: !16)
!46 = !DILocation(line: 21, column: 6, scope: !16)
!47 = !DILocation(line: 21, column: 10, scope: !16)
!48 = !DILocation(line: 21, column: 8, scope: !16)
!49 = !DILocation(line: 21, column: 3, scope: !16)
!50 = !DILocation(line: 21, column: 26, scope: !16)
!51 = !DILocation(line: 21, column: 24, scope: !16)
!52 = !DILocation(line: 23, column: 14, scope: !16)
!53 = !DILocation(line: 23, column: 11, scope: !16)
!54 = !DILocation(line: 23, column: 19, scope: !16)
!56 = !DILocation(line: 1, column: 9, scope: !55)
!63 = !DILocation(line: 33, column: 35, scope: !55)
!64 = !DILocation(line: 33, column: 39, scope: !55)
!65 = !DILocation(line: 33, column: 34, scope: !55)
!66 = !DILocation(line: 33, column: 26, scope: !55)
!67 = !DILocation(line: 34, column: 10, scope: !55)
!68 = !DILocation(line: 34, column: 11, scope: !55)
!69 = !DILocation(line: 34, column: 26, scope: !55)
!70 = !DILocation(line: 35, column: 10, scope: !55)
!71 = !DILocation(line: 35, column: 11, scope: !55)
!72 = !DILocation(line: 35, column: 26, scope: !55)
!73 = !DILocation(line: 37, column: 30, scope: !55)
!74 = !DILocation(line: 37, column: 33, scope: !55)
!75 = !DILocation(line: 37, column: 37, scope: !55)
!76 = !DILocation(line: 37, column: 29, scope: !55)
!77 = !DILocation(line: 37, column: 21, scope: !55)
!78 = !DILocation(line: 38, column: 10, scope: !55)
!79 = !DILocation(line: 38, column: 11, scope: !55)
!80 = !DILocation(line: 38, column: 30, scope: !55)
!81 = !DILocation(line: 39, column: 10, scope: !55)
!82 = !DILocation(line: 39, column: 11, scope: !55)
!83 = !DILocation(line: 39, column: 30, scope: !55)
!84 = !DILocation(line: 42, column: 22, scope: !55)
!85 = !DILocation(line: 42, column: 27, scope: !55)
!86 = !DILocation(line: 42, column: 21, scope: !55)
!87 = !DILocation(line: 42, column: 13, scope: !55)
!88 = !DILocation(line: 43, column: 10, scope: !55)
!89 = !DILocation(line: 43, column: 11, scope: !55)
!90 = !DILocation(line: 43, column: 27, scope: !55)
!91 = !DILocation(line: 43, column: 31, scope: !55)
!92 = !DILocation(line: 43, column: 49, scope: !55)
!93 = !{null}
!94 = !DISubroutineType(types: !93)
!95 = !{}
!96 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!98 = !{!17, !19, !20, !21, !22, !23}
!99 = !{!57, !58, !59, !60, !61, !62}
!97 = !{!5, !7, !8}
