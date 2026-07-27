; Kylix LLVM IR — module: Phase6Demo
source_filename = "Phase6Demo.klx"
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
  %t0 = getelementptr inbounds [9 x i8], ptr @.str.0, i64 0, i64 0, !dbg !5
  %t1 = getelementptr inbounds [6 x i8], ptr @.str.1, i64 0, i64 0, !dbg !6
  %t2 = call ptr @__kylix_crypto_Sha256(ptr %t1), !dbg !7
  %t3 = call ptr @malloc(i64 512), !dbg !8
  call ptr @strcpy(ptr %t3, ptr %t0), !dbg !8
  call ptr @strcat(ptr %t3, ptr %t2), !dbg !8
  %t4 = call i32 @puts(ptr noundef %t3), !dbg !9
  %t5 = getelementptr inbounds [9 x i8], ptr @.str.2, i64 0, i64 0, !dbg !10
  %t6 = getelementptr inbounds [6 x i8], ptr @.str.1, i64 0, i64 0, !dbg !11
  %t7 = call ptr @__kylix_encoding_Base64Encode(ptr %t6), !dbg !12
  %t8 = call ptr @malloc(i64 512), !dbg !13
  call ptr @strcpy(ptr %t8, ptr %t5), !dbg !13
  call ptr @strcat(ptr %t8, ptr %t7), !dbg !13
  %t9 = call i32 @puts(ptr noundef %t8), !dbg !14
  %t10 = getelementptr inbounds [6 x i8], ptr @.str.3, i64 0, i64 0, !dbg !15
  %t11 = getelementptr inbounds [6 x i8], ptr @.str.1, i64 0, i64 0, !dbg !16
  %t12 = call ptr @__kylix_encoding_HexEncode(ptr %t11), !dbg !17
  %t13 = call ptr @malloc(i64 512), !dbg !18
  call ptr @strcpy(ptr %t13, ptr %t10), !dbg !18
  call ptr @strcat(ptr %t13, ptr %t12), !dbg !18
  %t14 = call i32 @puts(ptr noundef %t13), !dbg !19
  %t15 = getelementptr inbounds [6 x i8], ptr @.str.4, i64 0, i64 0, !dbg !20
  %t16 = getelementptr inbounds [4 x i8], ptr @.str.5, i64 0, i64 0, !dbg !21
  %t17 = call ptr @__kylix_crypto_Md5(ptr %t16), !dbg !22
  %t18 = call ptr @malloc(i64 512), !dbg !23
  call ptr @strcpy(ptr %t18, ptr %t15), !dbg !23
  call ptr @strcat(ptr %t18, ptr %t17), !dbg !23
  %t19 = call i32 @puts(ptr noundef %t18), !dbg !24
  %t20 = getelementptr inbounds [7 x i8], ptr @.str.6, i64 0, i64 0, !dbg !25
  %t21 = getelementptr inbounds [4 x i8], ptr @.str.7, i64 0, i64 0, !dbg !26
  %t22 = getelementptr inbounds [5 x i8], ptr @.str.8, i64 0, i64 0, !dbg !27
  %t23 = call ptr @__kylix_crypto_HmacSha256(ptr %t21, ptr %t22), !dbg !28
  %t24 = call ptr @malloc(i64 512), !dbg !29
  call ptr @strcpy(ptr %t24, ptr %t20), !dbg !29
  call ptr @strcat(ptr %t24, ptr %t23), !dbg !29
  %t25 = call i32 @puts(ptr noundef %t24), !dbg !30
  %t26 = getelementptr inbounds [5 x i8], ptr @.str.9, i64 0, i64 0, !dbg !31
  %t27 = add i64 0, 4, !dbg !32
  %t28 = call ptr @__kylix_crypto_BCryptHash(ptr %t26, i64 %t27), !dbg !33
  %v_hash_str = alloca ptr, align 8, !dbg !34
  store ptr %t28, ptr %v_hash_str, !dbg !34
  %t29 = getelementptr inbounds [5 x i8], ptr @.str.9, i64 0, i64 0, !dbg !35
  %t30 = load ptr, ptr %v_hash_str, !dbg !36
  %t31 = call i1 @__kylix_crypto_BCryptCompare(ptr %t29, ptr %t30), !dbg !37
  br i1 %t31, label %lbl0, label %lbl2, !dbg !38
lbl0:
  %t32 = getelementptr inbounds [11 x i8], ptr @.str.10, i64 0, i64 0, !dbg !39
  %t33 = call i32 @puts(ptr noundef %t32), !dbg !40
  br label %lbl1, !dbg !38
lbl2:
  %t34 = getelementptr inbounds [13 x i8], ptr @.str.11, i64 0, i64 0, !dbg !41
  %t35 = call i32 @puts(ptr noundef %t34), !dbg !42
  br label %lbl1, !dbg !38
lbl1:
  %t36 = getelementptr inbounds [6 x i8], ptr @.str.12, i64 0, i64 0, !dbg !43
  %t37 = getelementptr inbounds [12 x i8], ptr @.str.13, i64 0, i64 0, !dbg !44
  %t38 = call ptr @__kylix_crypto_AesEncrypt(ptr %t36, ptr %t37), !dbg !45
  %v_enc_str = alloca ptr, align 8, !dbg !46
  store ptr %t38, ptr %v_enc_str, !dbg !46
  %t39 = getelementptr inbounds [6 x i8], ptr @.str.12, i64 0, i64 0, !dbg !47
  %t40 = load ptr, ptr %v_enc_str, !dbg !48
  %t41 = call ptr @__kylix_crypto_AesDecrypt(ptr %t39, ptr %t40), !dbg !49
  %v_dec_str = alloca ptr, align 8, !dbg !50
  store ptr %t41, ptr %v_dec_str, !dbg !50
  %t42 = load ptr, ptr %v_dec_str, !dbg !51
  %t43 = getelementptr inbounds [12 x i8], ptr @.str.13, i64 0, i64 0, !dbg !52
  %t44 = call i32 @strcmp(ptr %t42, ptr %t43), !dbg !53
  %t45 = icmp eq i32 %t44, 0, !dbg !53
  br i1 %t45, label %lbl3, label %lbl5, !dbg !54
lbl3:
  %t46 = getelementptr inbounds [19 x i8], ptr @.str.14, i64 0, i64 0, !dbg !55
  %t47 = call i32 @puts(ptr noundef %t46), !dbg !56
  br label %lbl4, !dbg !54
lbl5:
  %t48 = getelementptr inbounds [21 x i8], ptr @.str.15, i64 0, i64 0, !dbg !57
  %t49 = call i32 @puts(ptr noundef %t48), !dbg !58
  br label %lbl4, !dbg !54
lbl4:
  %t50 = getelementptr inbounds [18 x i8], ptr @.str.16, i64 0, i64 0, !dbg !59
  %t51 = call i32 @puts(ptr noundef %t50), !dbg !60
  ret i32 0
}

define ptr @__kylix_crypto_Sha256(ptr %data) {
entry:
  %t52 = call i64 @strlen(ptr %data)
  %t53 = alloca [32 x i8], align 1
  call ptr @SHA256(ptr %data, i64 %t52, ptr %t53)
  %t54 = call ptr @__kylix_crypto_hexbytes(ptr %t53, i64 32)
  ret ptr %t54
}

@__kylix_b64_table = private unnamed_addr constant [64 x i8] c"\41\42\43\44\45\46\47\48\49\4A\4B\4C\4D\4E\4F\50\51\52\53\54\55\56\57\58\59\5A\61\62\63\64\65\66\67\68\69\6A\6B\6C\6D\6E\6F\70\71\72\73\74\75\76\77\78\79\7A\30\31\32\33\34\35\36\37\38\39\2B\2F", align 1
define ptr @__kylix_encoding_Base64Encode(ptr %str) {
entry:
  %t55 = call i64 @strlen(ptr %str)
  %t56 = add i64 %t55, 2
  %t57 = udiv i64 %t56, 3
  %t58 = shl i64 %t57, 2
  %t59 = add i64 %t58, 1
  %t60 = call ptr @malloc(i64 %t59)
  store i8 0, ptr %t60
  %t61 = alloca i64, align 8
  store i64 0, ptr %t61
  %t62 = alloca i64, align 8
  store i64 0, ptr %t62
  br label %lbl6
lbl6:
  %t63 = load i64, ptr %t61
  %t64 = icmp slt i64 %t63, %t55
  br i1 %t64, label %lbl7, label %lbl8
lbl7:
  %t65 = getelementptr inbounds i8, ptr %str, i64 %t63
  %t66 = load i8, ptr %t65
  %t67 = zext i8 %t66 to i64
  %t68 = add i64 %t63, 1
  %t69 = add i64 %t68, 1
  %t70 = icmp slt i64 %t68, %t55
  %t71 = icmp slt i64 %t69, %t55
  %t72 = getelementptr inbounds i8, ptr %str, i64 %t68
  %t73 = load i8, ptr %t72
  %t74 = zext i8 %t73 to i64
  %t75 = select i1 %t70, i64 %t74, i64 0
  %t76 = getelementptr inbounds i8, ptr %str, i64 %t69
  %t77 = load i8, ptr %t76
  %t78 = zext i8 %t77 to i64
  %t79 = select i1 %t71, i64 %t78, i64 0
  %t80 = shl i64 %t67, 16
  %t81 = shl i64 %t75, 8
  %t82 = or i64 %t81, %t79
  %t83 = or i64 %t80, %t82
  %t84 = lshr i64 %t83, 18
  %t85 = and i64 %t84, 63
  %t86 = lshr i64 %t83, 12
  %t87 = and i64 %t86, 63
  %t88 = lshr i64 %t83, 6
  %t89 = and i64 %t88, 63
  %t90 = and i64 %t83, 63
  %t91 = load i64, ptr %t62
  %t92 = getelementptr inbounds [64 x i8], ptr @__kylix_b64_table, i64 0, i64 %t85
  %t93 = load i8, ptr %t92
  %t94 = getelementptr inbounds i8, ptr %t60, i64 %t91
  store i8 %t93, ptr %t94
  %t95 = getelementptr inbounds [64 x i8], ptr @__kylix_b64_table, i64 0, i64 %t87
  %t96 = load i8, ptr %t95
  %t97 = add i64 %t91, 1
  %t98 = getelementptr inbounds i8, ptr %t60, i64 %t97
  store i8 %t96, ptr %t98
  %t99 = getelementptr inbounds [64 x i8], ptr @__kylix_b64_table, i64 0, i64 %t89
  %t100 = load i8, ptr %t99
  %t101 = select i1 %t70, i8 %t100, i8 61
  %t102 = add i64 %t91, 2
  %t103 = getelementptr inbounds i8, ptr %t60, i64 %t102
  store i8 %t101, ptr %t103
  %t104 = getelementptr inbounds [64 x i8], ptr @__kylix_b64_table, i64 0, i64 %t90
  %t105 = load i8, ptr %t104
  %t106 = select i1 %t71, i8 %t105, i8 61
  %t107 = add i64 %t91, 3
  %t108 = getelementptr inbounds i8, ptr %t60, i64 %t107
  store i8 %t106, ptr %t108
  %t109 = add i64 %t91, 4
  store i64 %t109, ptr %t62
  %t110 = add i64 %t63, 3
  store i64 %t110, ptr %t61
  br label %lbl6
lbl8:
  %t111 = load i64, ptr %t62
  %t112 = getelementptr inbounds i8, ptr %t60, i64 %t111
  store i8 0, ptr %t112
  ret ptr %t60
}

define ptr @__kylix_encoding_HexEncode(ptr %str) {
entry:
  %t113 = call i64 @strlen(ptr %str)
  %t114 = shl i64 %t113, 1
  %t115 = add i64 %t114, 1
  %t116 = call ptr @malloc(i64 %t115)
  store i8 0, ptr %t116
  %t117 = getelementptr inbounds [5 x i8], ptr @.str.17, i64 0, i64 0
  %t118 = alloca i64, align 8
  store i64 0, ptr %t118
  br label %lbl9
lbl9:
  %t119 = load i64, ptr %t118
  %t120 = icmp sge i64 %t119, %t113
  br i1 %t120, label %lbl11, label %lbl10
lbl10:
  %t121 = getelementptr inbounds i8, ptr %str, i64 %t119
  %t122 = load i8, ptr %t121
  %t123 = zext i8 %t122 to i64
  %t124 = shl i64 %t119, 1
  %t125 = getelementptr inbounds i8, ptr %t116, i64 %t124
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t125, i64 3, ptr %t117, i64 %t123)
  %t126 = add i64 %t119, 1
  store i64 %t126, ptr %t118
  br label %lbl9
lbl11:
  ret ptr %t116
}

define ptr @__kylix_crypto_Md5(ptr %data) {
entry:
  %t127 = call i64 @strlen(ptr %data)
  %t128 = alloca [16 x i8], align 1
  call ptr @MD5(ptr %data, i64 %t127, ptr %t128)
  %t129 = call ptr @__kylix_crypto_hexbytes(ptr %t128, i64 16)
  ret ptr %t129
}

define ptr @__kylix_crypto_HmacSha256(ptr %key, ptr %data) {
entry:
  %t130 = alloca [64 x i8], align 1
  %t131 = alloca [64 x i8], align 1
  call void @llvm.memset.p0.i64(ptr %t130, i8 0, i64 64, i1 false)
  call void @llvm.memset.p0.i64(ptr %t131, i8 0, i64 64, i1 false)
  call ptr @strncpy(ptr %t130, ptr %key, i64 64)
  call ptr @strncpy(ptr %t131, ptr %key, i64 64)
  %t132 = alloca i64, align 8
  store i64 0, ptr %t132
  br label %lbl12
lbl12:
  %t133 = load i64, ptr %t132
  %t134 = icmp slt i64 %t133, 64
  br i1 %t134, label %lbl13, label %lbl14
lbl13:
  %t135 = getelementptr inbounds i8, ptr %t130, i64 %t133
  %t136 = load i8, ptr %t135
  %t137 = xor i8 %t136, 54
  store i8 %t137, ptr %t135
  %t138 = getelementptr inbounds i8, ptr %t131, i64 %t133
  %t139 = load i8, ptr %t138
  %t140 = xor i8 %t139, 92
  store i8 %t140, ptr %t138
  %t141 = add i64 %t133, 1
  store i64 %t141, ptr %t132
  br label %lbl12
lbl14:
  %t142 = call i64 @strlen(ptr %data)
  %t143 = add i64 %t142, 64
  %t144 = call ptr @malloc(i64 %t143)
  call ptr @memcpy(ptr %t144, ptr %t130, i64 64)
  %t145 = getelementptr inbounds i8, ptr %t144, i64 64
  call ptr @memcpy(ptr %t145, ptr %data, i64 %t142)
  %t146 = alloca [32 x i8], align 1
  call ptr @SHA256(ptr %t144, i64 %t143, ptr %t146)
  %t147 = alloca [96 x i8], align 1
  call ptr @memcpy(ptr %t147, ptr %t131, i64 64)
  %t148 = getelementptr inbounds i8, ptr %t147, i64 64
  call ptr @memcpy(ptr %t148, ptr %t146, i64 32)
  %t149 = alloca [32 x i8], align 1
  call ptr @SHA256(ptr %t147, i64 96, ptr %t149)
  %t150 = call ptr @__kylix_crypto_hexbytes(ptr %t149, i64 32)
  ret ptr %t150
}

define ptr @__kylix_crypto_BCryptHash(ptr %password, i64 %cost) {
entry:
  %t151 = alloca [16 x i8], align 1
  call i32 @RAND_bytes(ptr %t151, i32 16)
  %t152 = alloca [32 x i8], align 1
  %t153 = shl i64 1, %cost
  %t154 = call i64 @strlen(ptr %password)
  %t155 = trunc i64 %t154 to i32
  %t156 = call ptr @EVP_sha256()
  call i32 @PKCS5_PBKDF2_HMAC(ptr %password, i32 %t155, ptr %t151, i32 16, i64 %t153, ptr %t156, i32 32, ptr %t152)
  %t157 = call ptr @__kylix_crypto_hexbytes(ptr %t151, i64 16)
  %t158 = call ptr @__kylix_crypto_hexbytes(ptr %t152, i64 32)
  %t159 = call ptr @malloc(i64 256)
  %t160 = getelementptr inbounds [25 x i8], ptr @.str.18, i64 0, i64 0
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t159, i64 256, ptr %t160, i64 %cost, ptr %t157, ptr %t158)
  ret ptr %t159
}

define ptr @__kylix_crypto_hexdecode(ptr %hex) {
entry:
  %t161 = call i64 @strlen(ptr %hex)
  %t162 = lshr i64 %t161, 1
  %t163 = add i64 %t162, 1
  %t164 = call ptr @malloc(i64 %t163)
  store i8 0, ptr %t164
  %t165 = alloca i64, align 8
  store i64 0, ptr %t165
  %t166 = alloca i64, align 8
  store i64 0, ptr %t166
  br label %lbl15
lbl15:
  %t167 = load i64, ptr %t165
  %t168 = add i64 %t167, 1
  %t169 = icmp slt i64 %t168, %t161
  br i1 %t169, label %lbl16, label %lbl17
lbl16:
  %t170 = getelementptr inbounds i8, ptr %hex, i64 %t167
  %t171 = load i8, ptr %t170
  %t172 = getelementptr inbounds i8, ptr %hex, i64 %t168
  %t173 = load i8, ptr %t172
  %t174 = call i64 @__kylix_crypto_hexval(i8 %t171)
  %t175 = call i64 @__kylix_crypto_hexval(i8 %t173)
  %t176 = shl i64 %t174, 4
  %t177 = or i64 %t176, %t175
  %t178 = trunc i64 %t177 to i8
  %t179 = load i64, ptr %t166
  %t180 = getelementptr inbounds i8, ptr %t164, i64 %t179
  store i8 %t178, ptr %t180
  %t181 = add i64 %t179, 1
  store i64 %t181, ptr %t166
  %t182 = add i64 %t167, 2
  store i64 %t182, ptr %t165
  br label %lbl15
lbl17:
  %t183 = load i64, ptr %t166
  %t184 = getelementptr inbounds i8, ptr %t164, i64 %t183
  store i8 0, ptr %t184
  ret ptr %t164
}

define i64 @__kylix_crypto_hexval(i8 %c) {
entry:
  %t185 = zext i8 %c to i64
  %t186 = sub i64 %t185, 48
  %t187 = icmp ult i64 %t186, 10
  br i1 %t187, label %ret_digit, label %check_upper
ret_digit:
  ret i64 %t186
check_upper:
  %t188 = sub i64 %t185, 65
  %t189 = icmp ult i64 %t188, 6
  br i1 %t189, label %ret_upper, label %check_lower
ret_upper:
  %t190 = add i64 %t188, 10
  ret i64 %t190
check_lower:
  %t191 = sub i64 %t185, 97
  %t192 = icmp ult i64 %t191, 6
  br i1 %t192, label %ret_lower, label %ret_zero
ret_lower:
  %t193 = add i64 %t191, 10
  ret i64 %t193
ret_zero:
  ret i64 0
}

define i1 @__kylix_crypto_BCryptCompare(ptr %password, ptr %hash) {
entry:
  %t194 = alloca i64, align 8
  %t195 = alloca [33 x i8], align 1
  %t196 = alloca [65 x i8], align 1
  %t197 = getelementptr inbounds [35 x i8], ptr @.str.19, i64 0, i64 0
  %t198 = call i32 (ptr, ptr, ...) @sscanf(ptr %hash, ptr %t197, ptr %t194, ptr %t195, ptr %t196)
  %t199 = icmp eq i32 %t198, 3
  br i1 %t199, label %lbl18, label %lbl19
lbl18:
  %t200 = call ptr @__kylix_crypto_hexdecode(ptr %t195)
  %t201 = alloca [32 x i8], align 1
  %t202 = load i64, ptr %t194
  %t203 = shl i64 1, %t202
  %t204 = call i64 @strlen(ptr %password)
  %t205 = trunc i64 %t204 to i32
  %t206 = call ptr @EVP_sha256()
  call i32 @PKCS5_PBKDF2_HMAC(ptr %password, i32 %t205, ptr %t200, i32 16, i64 %t203, ptr %t206, i32 32, ptr %t201)
  %t207 = call ptr @__kylix_crypto_hexbytes(ptr %t201, i64 32)
  %t208 = call i32 @strcmp(ptr %t207, ptr %t196)
  %t209 = icmp eq i32 %t208, 0
  ret i1 %t209
lbl19:
  ret i1 false
}

define ptr @__kylix_crypto_AesEncrypt(ptr %key, ptr %plaintext) {
entry:
  %t210 = alloca [32 x i8], align 1
  call void @llvm.memset.p0.i64(ptr %t210, i8 0, i64 32, i1 false)
  call ptr @strncpy(ptr %t210, ptr %key, i64 32)
  %t211 = call i64 @strlen(ptr %plaintext)
  %t212 = trunc i64 %t211 to i32
  %t213 = add i64 %t211, 48
  %t214 = call ptr @malloc(i64 %t213)
  call i32 @RAND_bytes(ptr %t214, i32 16)
  %t215 = call ptr @EVP_CIPHER_CTX_new()
  %t216 = call ptr @EVP_aes_256_cbc()
  call i32 @EVP_EncryptInit_ex(ptr %t215, ptr %t216, ptr null, ptr %t210, ptr %t214)
  %t217 = alloca i32, align 4
  %t218 = getelementptr inbounds i8, ptr %t214, i64 16
  call i32 @EVP_EncryptUpdate(ptr %t215, ptr %t218, ptr %t217, ptr %plaintext, i32 %t212)
  %t219 = load i32, ptr %t217
  %t220 = zext i32 %t219 to i64
  %t221 = alloca i32, align 4
  %t222 = add i64 16, %t220
  %t223 = getelementptr inbounds i8, ptr %t214, i64 %t222
  call i32 @EVP_EncryptFinal_ex(ptr %t215, ptr %t223, ptr %t221)
  %t224 = load i32, ptr %t221
  %t225 = zext i32 %t224 to i64
  call void @EVP_CIPHER_CTX_free(ptr %t215)
  %t226 = add i64 %t220, %t225
  %t227 = add i64 16, %t226
  %t228 = call ptr @__kylix_crypto_hexbytes(ptr %t214, i64 %t227)
  ret ptr %t228
}

define ptr @__kylix_crypto_AesDecrypt(ptr %key, ptr %ciphertext) {
entry:
  %t229 = alloca [32 x i8], align 1
  call void @llvm.memset.p0.i64(ptr %t229, i8 0, i64 32, i1 false)
  call ptr @strncpy(ptr %t229, ptr %key, i64 32)
  %t230 = call ptr @__kylix_crypto_hexdecode(ptr %ciphertext)
  %t231 = call i64 @strlen(ptr %ciphertext)
  %t232 = lshr i64 %t231, 1
  %t233 = sub i64 %t232, 16
  %t234 = trunc i64 %t233 to i32
  %t235 = add i64 %t232, 16
  %t236 = call ptr @malloc(i64 %t235)
  %t237 = call ptr @EVP_CIPHER_CTX_new()
  %t238 = call ptr @EVP_aes_256_cbc()
  call i32 @EVP_DecryptInit_ex(ptr %t237, ptr %t238, ptr null, ptr %t229, ptr %t230)
  %t239 = alloca i32, align 4
  %t240 = getelementptr inbounds i8, ptr %t230, i64 16
  call i32 @EVP_DecryptUpdate(ptr %t237, ptr %t236, ptr %t239, ptr %t240, i32 %t234)
  %t241 = load i32, ptr %t239
  %t242 = zext i32 %t241 to i64
  %t243 = alloca i32, align 4
  %t244 = getelementptr inbounds i8, ptr %t236, i64 %t242
  call i32 @EVP_DecryptFinal_ex(ptr %t237, ptr %t244, ptr %t243)
  %t245 = load i32, ptr %t243
  %t246 = zext i32 %t245 to i64
  call void @EVP_CIPHER_CTX_free(ptr %t237)
  %t247 = add i64 %t242, %t246
  %t248 = getelementptr inbounds i8, ptr %t236, i64 %t247
  store i8 0, ptr %t248
  ret ptr %t236
}

define ptr @__kylix_crypto_hexbytes(ptr %bytes, i64 %n) {
entry:
  %t249 = shl i64 %n, 1
  %t250 = add i64 %t249, 1
  %t251 = call ptr @malloc(i64 %t250)
  store i8 0, ptr %t251
  %t252 = getelementptr inbounds [5 x i8], ptr @.str.17, i64 0, i64 0
  %t253 = alloca i64, align 8
  store i64 0, ptr %t253
  br label %lbl20
lbl20:
  %t254 = load i64, ptr %t253
  %t255 = icmp slt i64 %t254, %n
  br i1 %t255, label %lbl21, label %lbl22
lbl21:
  %t256 = getelementptr inbounds i8, ptr %bytes, i64 %t254
  %t257 = load i8, ptr %t256
  %t258 = zext i8 %t257 to i64
  %t259 = shl i64 %t254, 1
  %t260 = getelementptr inbounds i8, ptr %t251, i64 %t259
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t260, i64 3, ptr %t252, i64 %t258)
  %t261 = add i64 %t254, 1
  store i64 %t261, ptr %t253
  br label %lbl20
lbl22:
  ret ptr %t251
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [9 x i8] c"SHA256: \00", align 1
@.str.1 = private unnamed_addr constant [6 x i8] c"hello\00", align 1
@.str.2 = private unnamed_addr constant [9 x i8] c"Base64: \00", align 1
@.str.3 = private unnamed_addr constant [6 x i8] c"Hex: \00", align 1
@.str.4 = private unnamed_addr constant [6 x i8] c"MD5: \00", align 1
@.str.5 = private unnamed_addr constant [4 x i8] c"abc\00", align 1
@.str.6 = private unnamed_addr constant [7 x i8] c"HMAC: \00", align 1
@.str.7 = private unnamed_addr constant [4 x i8] c"key\00", align 1
@.str.8 = private unnamed_addr constant [5 x i8] c"data\00", align 1
@.str.9 = private unnamed_addr constant [5 x i8] c"test\00", align 1
@.str.10 = private unnamed_addr constant [11 x i8] c"BCrypt: OK\00", align 1
@.str.11 = private unnamed_addr constant [13 x i8] c"BCrypt: FAIL\00", align 1
@.str.12 = private unnamed_addr constant [6 x i8] c"mykey\00", align 1
@.str.13 = private unnamed_addr constant [12 x i8] c"hello world\00", align 1
@.str.14 = private unnamed_addr constant [19 x i8] c"AES round-trip: OK\00", align 1
@.str.15 = private unnamed_addr constant [21 x i8] c"AES round-trip: FAIL\00", align 1
@.str.16 = private unnamed_addr constant [18 x i8] c"Phase 6 stdlib OK\00", align 1
@.str.17 = private unnamed_addr constant [5 x i8] c"%02x\00", align 1
@.str.18 = private unnamed_addr constant [25 x i8] c"pbkdf2$sha256$%lld$%s$%s\00", align 1
@.str.19 = private unnamed_addr constant [35 x i8] c"pbkdf2$sha256$%lld$%32[^$]$%64[^$]\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example48_phase6_net_crypto_encoding.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/13_stdlib_phase6")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !62, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !65)
!5 = !DILocation(line: 5, column: 11, scope: !4)
!6 = !DILocation(line: 5, column: 31, scope: !4)
!7 = !DILocation(line: 5, column: 30, scope: !4)
!8 = !DILocation(line: 5, column: 22, scope: !4)
!9 = !DILocation(line: 5, column: 10, scope: !4)
!10 = !DILocation(line: 6, column: 11, scope: !4)
!11 = !DILocation(line: 6, column: 37, scope: !4)
!12 = !DILocation(line: 6, column: 36, scope: !4)
!13 = !DILocation(line: 6, column: 22, scope: !4)
!14 = !DILocation(line: 6, column: 10, scope: !4)
!15 = !DILocation(line: 7, column: 11, scope: !4)
!16 = !DILocation(line: 7, column: 31, scope: !4)
!17 = !DILocation(line: 7, column: 30, scope: !4)
!18 = !DILocation(line: 7, column: 19, scope: !4)
!19 = !DILocation(line: 7, column: 10, scope: !4)
!20 = !DILocation(line: 8, column: 11, scope: !4)
!21 = !DILocation(line: 8, column: 25, scope: !4)
!22 = !DILocation(line: 8, column: 24, scope: !4)
!23 = !DILocation(line: 8, column: 19, scope: !4)
!24 = !DILocation(line: 8, column: 10, scope: !4)
!25 = !DILocation(line: 9, column: 11, scope: !4)
!26 = !DILocation(line: 9, column: 33, scope: !4)
!27 = !DILocation(line: 9, column: 40, scope: !4)
!28 = !DILocation(line: 9, column: 32, scope: !4)
!29 = !DILocation(line: 9, column: 20, scope: !4)
!30 = !DILocation(line: 9, column: 10, scope: !4)
!31 = !DILocation(line: 11, column: 26, scope: !4)
!32 = !DILocation(line: 11, column: 34, scope: !4)
!33 = !DILocation(line: 11, column: 25, scope: !4)
!34 = !DILocation(line: 11, column: 3, scope: !4)
!35 = !DILocation(line: 12, column: 20, scope: !4)
!36 = !DILocation(line: 12, column: 28, scope: !4)
!37 = !DILocation(line: 12, column: 19, scope: !4)
!38 = !DILocation(line: 12, column: 3, scope: !4)
!39 = !DILocation(line: 13, column: 13, scope: !4)
!40 = !DILocation(line: 13, column: 12, scope: !4)
!41 = !DILocation(line: 15, column: 13, scope: !4)
!42 = !DILocation(line: 15, column: 12, scope: !4)
!43 = !DILocation(line: 17, column: 25, scope: !4)
!44 = !DILocation(line: 17, column: 34, scope: !4)
!45 = !DILocation(line: 17, column: 24, scope: !4)
!46 = !DILocation(line: 17, column: 3, scope: !4)
!47 = !DILocation(line: 18, column: 25, scope: !4)
!48 = !DILocation(line: 18, column: 34, scope: !4)
!49 = !DILocation(line: 18, column: 24, scope: !4)
!50 = !DILocation(line: 18, column: 3, scope: !4)
!51 = !DILocation(line: 19, column: 6, scope: !4)
!52 = !DILocation(line: 19, column: 12, scope: !4)
!53 = !DILocation(line: 19, column: 10, scope: !4)
!54 = !DILocation(line: 19, column: 3, scope: !4)
!55 = !DILocation(line: 20, column: 13, scope: !4)
!56 = !DILocation(line: 20, column: 12, scope: !4)
!57 = !DILocation(line: 22, column: 13, scope: !4)
!58 = !DILocation(line: 22, column: 12, scope: !4)
!59 = !DILocation(line: 24, column: 11, scope: !4)
!60 = !DILocation(line: 24, column: 10, scope: !4)
!61 = !{null}
!62 = !DISubroutineType(types: !61)
!63 = !{}
!64 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!65 = !{}
