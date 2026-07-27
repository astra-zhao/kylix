; Kylix LLVM IR — module: MapType
source_filename = "MapType.klx"
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
define i32 @main() {
entry:
  %v_scores_map = alloca ptr, align 8
  %t0 = call ptr @__kylix_htab_new()
  store ptr %t0, ptr %v_scores_map
  %v_ages_map = alloca ptr, align 8
  %t1 = call ptr @__kylix_htab_new()
  store ptr %t1, ptr %v_ages_map
  %t2 = add i64 0, 95
  %t3 = load ptr, ptr %v_scores_map
  %t4 = getelementptr inbounds [6 x i8], ptr @.str.0, i64 0, i64 0
  %t5 = call ptr @malloc(i64 32)
  %t6 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t5, i64 32, ptr %t6, i64 %t2)
  call void @__kylix_htab_put(ptr %t3, ptr %t4, ptr %t5)
  %t7 = add i64 0, 87
  %t8 = load ptr, ptr %v_scores_map
  %t9 = getelementptr inbounds [4 x i8], ptr @.str.2, i64 0, i64 0
  %t10 = call ptr @malloc(i64 32)
  %t11 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t10, i64 32, ptr %t11, i64 %t7)
  call void @__kylix_htab_put(ptr %t8, ptr %t9, ptr %t10)
  %t12 = add i64 0, 92
  %t13 = load ptr, ptr %v_scores_map
  %t14 = getelementptr inbounds [8 x i8], ptr @.str.3, i64 0, i64 0
  %t15 = call ptr @malloc(i64 32)
  %t16 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t15, i64 32, ptr %t16, i64 %t12)
  call void @__kylix_htab_put(ptr %t13, ptr %t14, ptr %t15)
  %t17 = add i64 0, 25
  %t18 = load ptr, ptr %v_ages_map
  %t19 = getelementptr inbounds [6 x i8], ptr @.str.0, i64 0, i64 0
  %t20 = call ptr @malloc(i64 32)
  %t21 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t20, i64 32, ptr %t21, i64 %t17)
  call void @__kylix_htab_put(ptr %t18, ptr %t19, ptr %t20)
  %t22 = add i64 0, 30
  %t23 = load ptr, ptr %v_ages_map
  %t24 = getelementptr inbounds [4 x i8], ptr @.str.2, i64 0, i64 0
  %t25 = call ptr @malloc(i64 32)
  %t26 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t25, i64 32, ptr %t26, i64 %t22)
  call void @__kylix_htab_put(ptr %t23, ptr %t24, ptr %t25)
  %t27 = add i64 0, 28
  %t28 = load ptr, ptr %v_ages_map
  %t29 = getelementptr inbounds [8 x i8], ptr @.str.3, i64 0, i64 0
  %t30 = call ptr @malloc(i64 32)
  %t31 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0
  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t30, i64 32, ptr %t31, i64 %t27)
  call void @__kylix_htab_put(ptr %t28, ptr %t29, ptr %t30)
  %t32 = getelementptr inbounds [8 x i8], ptr @.str.4, i64 0, i64 0
  %t33 = call i32 @puts(ptr noundef %t32)
  %t34 = call ptr @malloc(i64 512)
  store i8 0, ptr %t34
  %t36 = getelementptr inbounds [8 x i8], ptr @.str.6, i64 0, i64 0
  %t37 = call ptr @strcat(ptr %t34, ptr %t36)
  %t38 = getelementptr inbounds [2 x i8], ptr @.str.7, i64 0, i64 0
  %t39 = call ptr @strcat(ptr %t34, ptr %t38)
  %t40 = load ptr, ptr %v_scores_map
  %t41 = getelementptr inbounds [6 x i8], ptr @.str.0, i64 0, i64 0
  %t42 = call ptr @__kylix_htab_get(ptr %t40, ptr %t41)
  %t43 = call ptr @strcat(ptr %t34, ptr %t42)
  %t44 = call i32 @puts(ptr noundef %t34)
  %t45 = call ptr @malloc(i64 512)
  store i8 0, ptr %t45
  %t47 = getelementptr inbounds [6 x i8], ptr @.str.8, i64 0, i64 0
  %t48 = call ptr @strcat(ptr %t45, ptr %t47)
  %t49 = getelementptr inbounds [2 x i8], ptr @.str.7, i64 0, i64 0
  %t50 = call ptr @strcat(ptr %t45, ptr %t49)
  %t51 = load ptr, ptr %v_scores_map
  %t52 = getelementptr inbounds [4 x i8], ptr @.str.2, i64 0, i64 0
  %t53 = call ptr @__kylix_htab_get(ptr %t51, ptr %t52)
  %t54 = call ptr @strcat(ptr %t45, ptr %t53)
  %t55 = call i32 @puts(ptr noundef %t45)
  %t56 = call ptr @malloc(i64 512)
  store i8 0, ptr %t56
  %t58 = getelementptr inbounds [10 x i8], ptr @.str.9, i64 0, i64 0
  %t59 = call ptr @strcat(ptr %t56, ptr %t58)
  %t60 = getelementptr inbounds [2 x i8], ptr @.str.7, i64 0, i64 0
  %t61 = call ptr @strcat(ptr %t56, ptr %t60)
  %t62 = load ptr, ptr %v_scores_map
  %t63 = getelementptr inbounds [8 x i8], ptr @.str.3, i64 0, i64 0
  %t64 = call ptr @__kylix_htab_get(ptr %t62, ptr %t63)
  %t65 = call ptr @strcat(ptr %t56, ptr %t64)
  %t66 = call i32 @puts(ptr noundef %t56)
  %t67 = getelementptr inbounds [6 x i8], ptr @.str.10, i64 0, i64 0
  %t68 = call i32 @puts(ptr noundef %t67)
  %t69 = call ptr @malloc(i64 512)
  store i8 0, ptr %t69
  %t71 = getelementptr inbounds [8 x i8], ptr @.str.6, i64 0, i64 0
  %t72 = call ptr @strcat(ptr %t69, ptr %t71)
  %t73 = getelementptr inbounds [2 x i8], ptr @.str.7, i64 0, i64 0
  %t74 = call ptr @strcat(ptr %t69, ptr %t73)
  %t75 = load ptr, ptr %v_ages_map
  %t76 = getelementptr inbounds [6 x i8], ptr @.str.0, i64 0, i64 0
  %t77 = call ptr @__kylix_htab_get(ptr %t75, ptr %t76)
  %t78 = call ptr @strcat(ptr %t69, ptr %t77)
  %t79 = call i32 @puts(ptr noundef %t69)
  %t80 = call ptr @malloc(i64 512)
  store i8 0, ptr %t80
  %t82 = getelementptr inbounds [6 x i8], ptr @.str.8, i64 0, i64 0
  %t83 = call ptr @strcat(ptr %t80, ptr %t82)
  %t84 = getelementptr inbounds [2 x i8], ptr @.str.7, i64 0, i64 0
  %t85 = call ptr @strcat(ptr %t80, ptr %t84)
  %t86 = load ptr, ptr %v_ages_map
  %t87 = getelementptr inbounds [4 x i8], ptr @.str.2, i64 0, i64 0
  %t88 = call ptr @__kylix_htab_get(ptr %t86, ptr %t87)
  %t89 = call ptr @strcat(ptr %t80, ptr %t88)
  %t90 = call i32 @puts(ptr noundef %t80)
  %t91 = call ptr @malloc(i64 512)
  store i8 0, ptr %t91
  %t93 = getelementptr inbounds [10 x i8], ptr @.str.9, i64 0, i64 0
  %t94 = call ptr @strcat(ptr %t91, ptr %t93)
  %t95 = getelementptr inbounds [2 x i8], ptr @.str.7, i64 0, i64 0
  %t96 = call ptr @strcat(ptr %t91, ptr %t95)
  %t97 = load ptr, ptr %v_ages_map
  %t98 = getelementptr inbounds [8 x i8], ptr @.str.3, i64 0, i64 0
  %t99 = call ptr @__kylix_htab_get(ptr %t97, ptr %t98)
  %t100 = call ptr @strcat(ptr %t91, ptr %t99)
  %t101 = call i32 @puts(ptr noundef %t91)
  ret i32 0
}

define ptr @__kylix_htab_new() {
entry:
  %t102 = call ptr @malloc(i64 16)
  %t103 = call ptr @malloc(i64 2048)
  call void @llvm.memset.p0.i64(ptr %t103, i8 0, i64 2048, i1 false)
  store ptr %t103, ptr %t102
  %t104 = getelementptr inbounds i8, ptr %t102, i64 8
  store i64 0, ptr %t104
  ret ptr %t102
}

define i64 @__kylix_htab_hash(ptr %key) {
entry:
  %t105 = alloca i64, align 8
  store i64 5381, ptr %t105
  %t106 = alloca ptr, align 8
  store ptr %key, ptr %t106
  br label %lbl0
lbl0:
  %t107 = load ptr, ptr %t106
  %t108 = load i8, ptr %t107
  %t109 = icmp eq i8 %t108, 0
  br i1 %t109, label %lbl2, label %lbl1
lbl1:
  %t110 = zext i8 %t108 to i64
  %t111 = load i64, ptr %t105
  %t112 = mul i64 %t111, 33
  %t113 = add i64 %t112, %t110
  store i64 %t113, ptr %t105
  %t114 = getelementptr inbounds i8, ptr %t107, i64 1
  store ptr %t114, ptr %t106
  br label %lbl0
lbl2:
  %t115 = load i64, ptr %t105
  %t116 = and i64 %t115, 255
  ret i64 %t116
}

define ptr @__kylix_htab_find(ptr %t, ptr %key) {
entry:
  %t117 = load ptr, ptr %t
  %t118 = call i64 @__kylix_htab_hash(ptr %key)
  %t119 = getelementptr inbounds ptr, ptr %t117, i64 %t118
  %t120 = alloca ptr, align 8
  %t121 = load ptr, ptr %t119
  store ptr %t121, ptr %t120
  br label %lbl3
lbl3:
  %t122 = load ptr, ptr %t120
  %t123 = icmp eq ptr %t122, null
  br i1 %t123, label %lbl6, label %lbl4
lbl4:
  %t124 = getelementptr inbounds i8, ptr %t122, i64 0
  %t125 = load ptr, ptr %t124
  %t126 = call i32 @strcmp(ptr %t125, ptr %key)
  %t127 = icmp eq i32 %t126, 0
  br i1 %t127, label %lbl5, label %next
next:
  %t128 = getelementptr inbounds i8, ptr %t122, i64 16
  %t129 = load ptr, ptr %t128
  store ptr %t129, ptr %t120
  br label %lbl3
lbl5:
  ret ptr %t122
lbl6:
  ret ptr null
}

define void @__kylix_htab_put(ptr %t, ptr %key, ptr %val) {
entry:
  %t130 = call ptr @__kylix_htab_find(ptr %t, ptr %key)
  %t131 = icmp ne ptr %t130, null
  br i1 %t131, label %lbl7, label %lbl8
lbl7:
  %t132 = getelementptr inbounds i8, ptr %t130, i64 8
  store ptr %val, ptr %t132
  ret void
lbl8:
  %t133 = call ptr @malloc(i64 24)
  %t134 = getelementptr inbounds i8, ptr %t133, i64 0
  %t135 = call ptr @__kylix_htab_strdup(ptr %key)
  store ptr %t135, ptr %t134
  %t136 = getelementptr inbounds i8, ptr %t133, i64 8
  store ptr %val, ptr %t136
  %t137 = load ptr, ptr %t
  %t138 = call i64 @__kylix_htab_hash(ptr %key)
  %t139 = getelementptr inbounds ptr, ptr %t137, i64 %t138
  %t140 = load ptr, ptr %t139
  %t141 = getelementptr inbounds i8, ptr %t133, i64 16
  store ptr %t140, ptr %t141
  store ptr %t133, ptr %t139
  %t142 = getelementptr inbounds i8, ptr %t, i64 8
  %t143 = load i64, ptr %t142
  %t144 = add i64 %t143, 1
  store i64 %t144, ptr %t142
  ret void
}

define ptr @__kylix_htab_get(ptr %t, ptr %key) {
entry:
  %t145 = getelementptr inbounds [1 x i8], ptr @.str.11, i64 0, i64 0
  %t146 = call ptr @__kylix_htab_find(ptr %t, ptr %key)
  %t147 = icmp eq ptr %t146, null
  br i1 %t147, label %lbl9, label %lbl10
lbl9:
  ret ptr %t145
lbl10:
  %t148 = getelementptr inbounds i8, ptr %t146, i64 8
  %t149 = load ptr, ptr %t148
  ret ptr %t149
}

define i1 @__kylix_htab_has(ptr %t, ptr %key) {
entry:
  %t150 = call ptr @__kylix_htab_find(ptr %t, ptr %key)
  %t151 = icmp ne ptr %t150, null
  ret i1 %t151
}

define void @__kylix_htab_del(ptr %t, ptr %key) {
entry:
  %t152 = load ptr, ptr %t
  %t153 = call i64 @__kylix_htab_hash(ptr %key)
  %t154 = getelementptr inbounds ptr, ptr %t152, i64 %t153
  %t155 = alloca ptr, align 8
  store ptr %t154, ptr %t155
  %t156 = alloca ptr, align 8
  %t157 = load ptr, ptr %t154
  store ptr %t157, ptr %t156
  br label %lbl11
lbl11:
  %t158 = load ptr, ptr %t156
  %t159 = icmp eq ptr %t158, null
  br i1 %t159, label %lbl13, label %lbl12
lbl12:
  %t160 = getelementptr inbounds i8, ptr %t158, i64 0
  %t161 = load ptr, ptr %t160
  %t162 = call i32 @strcmp(ptr %t161, ptr %key)
  %t163 = icmp eq i32 %t162, 0
  br i1 %t163, label %lbl14, label %advance
advance:
  %t164 = getelementptr inbounds i8, ptr %t158, i64 16
  store ptr %t164, ptr %t155
  %t165 = load ptr, ptr %t164
  store ptr %t165, ptr %t156
  br label %lbl11
lbl14:
  %t166 = getelementptr inbounds i8, ptr %t158, i64 16
  %t167 = load ptr, ptr %t166
  %t168 = load ptr, ptr %t155
  store ptr %t167, ptr %t168
  call void @free(ptr %t158)
  %t169 = getelementptr inbounds i8, ptr %t, i64 8
  %t170 = load i64, ptr %t169
  %t171 = sub i64 %t170, 1
  store i64 %t171, ptr %t169
  ret void
lbl13:
  ret void
}

define i64 @__kylix_htab_size(ptr %t) {
entry:
  %t172 = getelementptr inbounds i8, ptr %t, i64 8
  %t173 = load i64, ptr %t172
  ret i64 %t173
}

define void @__kylix_htab_clear(ptr %t) {
entry:
  %t174 = load ptr, ptr %t
  %t175 = alloca i64, align 8
  store i64 0, ptr %t175
  br label %lbl15
lbl15:
  %t176 = load i64, ptr %t175
  %t177 = icmp sge i64 %t176, 256
  br i1 %t177, label %lbl17, label %lbl16
lbl16:
  %t178 = getelementptr inbounds ptr, ptr %t174, i64 %t176
  %t179 = alloca ptr, align 8
  %t180 = load ptr, ptr %t178
  store ptr %t180, ptr %t179
  br label %lbl18
lbl18:
  %t181 = load ptr, ptr %t179
  %t182 = icmp eq ptr %t181, null
  br i1 %t182, label %bucket_done, label %lbl19
lbl19:
  %t183 = getelementptr inbounds i8, ptr %t181, i64 16
  %t184 = load ptr, ptr %t183
  call void @free(ptr %t181)
  store ptr %t184, ptr %t179
  br label %lbl18
bucket_done:
  store ptr null, ptr %t178
  %t185 = add i64 %t176, 1
  store i64 %t185, ptr %t175
  br label %lbl15
lbl17:
  %t186 = getelementptr inbounds i8, ptr %t, i64 8
  store i64 0, ptr %t186
  ret void
}

define ptr @__kylix_htab_strdup(ptr %s) {
entry:
  %t187 = call i64 @strlen(ptr %s)
  %t188 = add i64 %t187, 1
  %t189 = call ptr @malloc(i64 %t188)
  call ptr @strcpy(ptr %t189, ptr %s)
  ret ptr %t189
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [6 x i8] c"Alice\00", align 1
@.str.1 = private unnamed_addr constant [5 x i8] c"%lld\00", align 1
@.str.2 = private unnamed_addr constant [4 x i8] c"Bob\00", align 1
@.str.3 = private unnamed_addr constant [8 x i8] c"Charlie\00", align 1
@.str.4 = private unnamed_addr constant [8 x i8] c"Scores:\00", align 1
@.str.5 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.6 = private unnamed_addr constant [8 x i8] c"Alice: \00", align 1
@.str.7 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.8 = private unnamed_addr constant [6 x i8] c"Bob: \00", align 1
@.str.9 = private unnamed_addr constant [10 x i8] c"Charlie: \00", align 1
@.str.10 = private unnamed_addr constant [6 x i8] c"Ages:\00", align 1
@.str.11 = private unnamed_addr constant [1 x i8] c"\00", align 1
