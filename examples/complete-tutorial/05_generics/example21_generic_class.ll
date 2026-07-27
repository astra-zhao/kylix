; Kylix LLVM IR — module: GenericClass
source_filename = "GenericClass.klx"
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
%TStack_Integer = type { ptr, [100 x i64], i64 }
@TStack_Integer_vtable = constant [4 x ptr] [ ptr @TStack_Integer_Create, ptr @TStack_Integer_Push, ptr @TStack_Integer_Pop, ptr @TStack_Integer_GetCount ]
define void @TStack_Integer_Create(ptr %self) {
entry:
  %t0 = add i64 0, 0
  %t1 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  store i64 %t0, ptr %t1
  ret void
}

define void @TStack_Integer_Push(ptr %self, i64 %item) {
entry:
  %v_item_int = alloca i64, align 8
  store i64 %item, ptr %v_item_int
  %t2 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  %t3 = load i64, ptr %t2
  %t4 = add i64 0, 100
  %t5 = icmp slt i64 %t3, %t4
  br i1 %t5, label %lbl0, label %lbl1
lbl0:
  %t6 = load i64, ptr %v_item_int
  %t7 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 1
  %t8 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  %t9 = load i64, ptr %t8
  %t10 = add i64 %t9, 0
  %t11 = getelementptr inbounds [100 x i64], ptr %t7, i64 0, i64 %t10
  store i64 %t6, ptr %t11
  %t12 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  %t13 = load i64, ptr %t12
  %t14 = add i64 0, 1
  %t15 = add i64 %t13, %t14
  %t16 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  store i64 %t15, ptr %t16
  br label %lbl1
lbl1:
  ret void
}

define i64 @TStack_Integer_Pop(ptr %self) {
entry:
  %result = alloca i64, align 8
  %t17 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  %t18 = load i64, ptr %t17
  %t19 = add i64 0, 0
  %t20 = icmp sgt i64 %t18, %t19
  br i1 %t20, label %lbl2, label %lbl3
lbl2:
  %t21 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  %t22 = load i64, ptr %t21
  %t23 = add i64 0, 1
  %t24 = sub i64 %t22, %t23
  %t25 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  store i64 %t24, ptr %t25
  %t26 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 1
  %t27 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  %t28 = load i64, ptr %t27
  %t29 = add i64 %t28, 0
  %t30 = getelementptr inbounds [100 x i64], ptr %t26, i64 0, i64 %t29
  %t31 = load i64, ptr %t30
  store i64 %t31, ptr %result
  br label %lbl3
lbl3:
  %t32 = load i64, ptr %result
  ret i64 %t32
}

define i64 @TStack_Integer_GetCount(ptr %self) {
entry:
  %result = alloca i64, align 8
  %t33 = getelementptr inbounds %TStack_Integer, ptr %self, i32 0, i32 2
  %t34 = load i64, ptr %t33
  store i64 %t34, ptr %result
  %t35 = load i64, ptr %result
  ret i64 %t35
}

%TStack_String = type { ptr, [100 x ptr], i64 }
@TStack_String_vtable = constant [4 x ptr] [ ptr @TStack_String_Create, ptr @TStack_String_Push, ptr @TStack_String_Pop, ptr @TStack_String_GetCount ]
define void @TStack_String_Create(ptr %self) {
entry:
  %t36 = add i64 0, 0
  %t37 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  store i64 %t36, ptr %t37
  ret void
}

define void @TStack_String_Push(ptr %self, ptr %item) {
entry:
  %v_item_str = alloca ptr, align 8
  store ptr %item, ptr %v_item_str
  %t38 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  %t39 = load i64, ptr %t38
  %t40 = add i64 0, 100
  %t41 = icmp slt i64 %t39, %t40
  br i1 %t41, label %lbl4, label %lbl5
lbl4:
  %t42 = load ptr, ptr %v_item_str
  %t43 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 1
  %t44 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  %t45 = load i64, ptr %t44
  %t46 = add i64 %t45, 0
  %t47 = getelementptr inbounds [100 x ptr], ptr %t43, i64 0, i64 %t46
  store ptr %t42, ptr %t47
  %t48 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  %t49 = load i64, ptr %t48
  %t50 = add i64 0, 1
  %t51 = add i64 %t49, %t50
  %t52 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  store i64 %t51, ptr %t52
  br label %lbl5
lbl5:
  ret void
}

define ptr @TStack_String_Pop(ptr %self) {
entry:
  %result = alloca ptr, align 8
  %t53 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  %t54 = load i64, ptr %t53
  %t55 = add i64 0, 0
  %t56 = icmp sgt i64 %t54, %t55
  br i1 %t56, label %lbl6, label %lbl7
lbl6:
  %t57 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  %t58 = load i64, ptr %t57
  %t59 = add i64 0, 1
  %t60 = sub i64 %t58, %t59
  %t61 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  store i64 %t60, ptr %t61
  %t62 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 1
  %t63 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  %t64 = load i64, ptr %t63
  %t65 = add i64 %t64, 0
  %t66 = getelementptr inbounds [100 x ptr], ptr %t62, i64 0, i64 %t65
  %t67 = load ptr, ptr %t66
  store ptr %t67, ptr %result
  br label %lbl7
lbl7:
  %t68 = load ptr, ptr %result
  ret ptr %t68
}

define i64 @TStack_String_GetCount(ptr %self) {
entry:
  %result = alloca i64, align 8
  %t69 = getelementptr inbounds %TStack_String, ptr %self, i32 0, i32 2
  %t70 = load i64, ptr %t69
  store i64 %t70, ptr %result
  %t71 = load i64, ptr %result
  ret i64 %t71
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
define i32 @main() {
entry:
  %t72 = call ptr @malloc(i64 24)
  %t73 = getelementptr inbounds %TStack_Integer, ptr %t72, i32 0, i32 0
  store ptr @TStack_Integer_vtable, ptr %t73
  %v_intStack_str = alloca ptr, align 8
  store ptr %t72, ptr %v_intStack_str
  %t74 = add i64 0, 10
  %t75 = load ptr, ptr %v_intStack_str
  %t76 = getelementptr inbounds %TStack_Integer, ptr %t75, i32 0, i32 0
  %t77 = load ptr, ptr %t76
  %t78 = getelementptr inbounds [4 x ptr], ptr %t77, i32 0, i32 1
  %t79 = load ptr, ptr %t78
  call void (ptr, i64) %t79(ptr %t75, i64 %t74)
  %t80 = add i64 0, 20
  %t81 = load ptr, ptr %v_intStack_str
  %t82 = getelementptr inbounds %TStack_Integer, ptr %t81, i32 0, i32 0
  %t83 = load ptr, ptr %t82
  %t84 = getelementptr inbounds [4 x ptr], ptr %t83, i32 0, i32 1
  %t85 = load ptr, ptr %t84
  call void (ptr, i64) %t85(ptr %t81, i64 %t80)
  %t86 = add i64 0, 30
  %t87 = load ptr, ptr %v_intStack_str
  %t88 = getelementptr inbounds %TStack_Integer, ptr %t87, i32 0, i32 0
  %t89 = load ptr, ptr %t88
  %t90 = getelementptr inbounds [4 x ptr], ptr %t89, i32 0, i32 1
  %t91 = load ptr, ptr %t90
  call void (ptr, i64) %t91(ptr %t87, i64 %t86)
  %t92 = call ptr @malloc(i64 512)
  store i8 0, ptr %t92
  %t93 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0
  %t94 = getelementptr inbounds [14 x i8], ptr @.str.1, i64 0, i64 0
  %t95 = call ptr @strcat(ptr %t92, ptr %t94)
  %t96 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0
  %t97 = call ptr @strcat(ptr %t92, ptr %t96)
  %t98 = load ptr, ptr %v_intStack_str
  %t99 = getelementptr inbounds %TStack_Integer, ptr %t98, i32 0, i32 0
  %t100 = load ptr, ptr %t99
  %t101 = getelementptr inbounds [4 x ptr], ptr %t100, i32 0, i32 3
  %t102 = load ptr, ptr %t101
  %t103 = call i64 (ptr) %t102(ptr %t98)
  %t104 = call i64 @strlen(ptr %t92)
  %t105 = getelementptr inbounds i8, ptr %t92, i64 %t104
  %t106 = sub i64 512, %t104
  %t107 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t105, i64 %t106, ptr %t93, i64 %t103)
  %t108 = call i32 @puts(ptr noundef %t92)
  %t109 = call ptr @malloc(i64 512)
  store i8 0, ptr %t109
  %t110 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0
  %t111 = getelementptr inbounds [6 x i8], ptr @.str.3, i64 0, i64 0
  %t112 = call ptr @strcat(ptr %t109, ptr %t111)
  %t113 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0
  %t114 = call ptr @strcat(ptr %t109, ptr %t113)
  %t115 = load ptr, ptr %v_intStack_str
  %t116 = getelementptr inbounds %TStack_Integer, ptr %t115, i32 0, i32 0
  %t117 = load ptr, ptr %t116
  %t118 = getelementptr inbounds [4 x ptr], ptr %t117, i32 0, i32 2
  %t119 = load ptr, ptr %t118
  %t120 = call i64 (ptr) %t119(ptr %t115)
  %t121 = call i64 @strlen(ptr %t109)
  %t122 = getelementptr inbounds i8, ptr %t109, i64 %t121
  %t123 = sub i64 512, %t121
  %t124 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t122, i64 %t123, ptr %t110, i64 %t120)
  %t125 = call i32 @puts(ptr noundef %t109)
  %t126 = call ptr @malloc(i64 512)
  store i8 0, ptr %t126
  %t127 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0
  %t128 = getelementptr inbounds [6 x i8], ptr @.str.3, i64 0, i64 0
  %t129 = call ptr @strcat(ptr %t126, ptr %t128)
  %t130 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0
  %t131 = call ptr @strcat(ptr %t126, ptr %t130)
  %t132 = load ptr, ptr %v_intStack_str
  %t133 = getelementptr inbounds %TStack_Integer, ptr %t132, i32 0, i32 0
  %t134 = load ptr, ptr %t133
  %t135 = getelementptr inbounds [4 x ptr], ptr %t134, i32 0, i32 2
  %t136 = load ptr, ptr %t135
  %t137 = call i64 (ptr) %t136(ptr %t132)
  %t138 = call i64 @strlen(ptr %t126)
  %t139 = getelementptr inbounds i8, ptr %t126, i64 %t138
  %t140 = sub i64 512, %t138
  %t141 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t139, i64 %t140, ptr %t127, i64 %t137)
  %t142 = call i32 @puts(ptr noundef %t126)
  %t143 = call ptr @malloc(i64 512)
  store i8 0, ptr %t143
  %t144 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0
  %t145 = getelementptr inbounds [14 x i8], ptr @.str.1, i64 0, i64 0
  %t146 = call ptr @strcat(ptr %t143, ptr %t145)
  %t147 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0
  %t148 = call ptr @strcat(ptr %t143, ptr %t147)
  %t149 = load ptr, ptr %v_intStack_str
  %t150 = getelementptr inbounds %TStack_Integer, ptr %t149, i32 0, i32 0
  %t151 = load ptr, ptr %t150
  %t152 = getelementptr inbounds [4 x ptr], ptr %t151, i32 0, i32 3
  %t153 = load ptr, ptr %t152
  %t154 = call i64 (ptr) %t153(ptr %t149)
  %t155 = call i64 @strlen(ptr %t143)
  %t156 = getelementptr inbounds i8, ptr %t143, i64 %t155
  %t157 = sub i64 512, %t155
  %t158 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t156, i64 %t157, ptr %t144, i64 %t154)
  %t159 = call i32 @puts(ptr noundef %t143)
  %t160 = call ptr @malloc(i64 24)
  %t161 = getelementptr inbounds %TStack_String, ptr %t160, i32 0, i32 0
  store ptr @TStack_String_vtable, ptr %t161
  %v_strStack_str = alloca ptr, align 8
  store ptr %t160, ptr %v_strStack_str
  %t162 = getelementptr inbounds [6 x i8], ptr @.str.4, i64 0, i64 0
  %t163 = load ptr, ptr %v_strStack_str
  %t164 = getelementptr inbounds %TStack_String, ptr %t163, i32 0, i32 0
  %t165 = load ptr, ptr %t164
  %t166 = getelementptr inbounds [4 x ptr], ptr %t165, i32 0, i32 1
  %t167 = load ptr, ptr %t166
  call void (ptr, ptr) %t167(ptr %t163, ptr %t162)
  %t168 = getelementptr inbounds [6 x i8], ptr @.str.5, i64 0, i64 0
  %t169 = load ptr, ptr %v_strStack_str
  %t170 = getelementptr inbounds %TStack_String, ptr %t169, i32 0, i32 0
  %t171 = load ptr, ptr %t170
  %t172 = getelementptr inbounds [4 x ptr], ptr %t171, i32 0, i32 1
  %t173 = load ptr, ptr %t172
  call void (ptr, ptr) %t173(ptr %t169, ptr %t168)
  %t174 = call ptr @malloc(i64 512)
  store i8 0, ptr %t174
  %t176 = getelementptr inbounds [6 x i8], ptr @.str.3, i64 0, i64 0
  %t177 = call ptr @strcat(ptr %t174, ptr %t176)
  %t178 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0
  %t179 = call ptr @strcat(ptr %t174, ptr %t178)
  %t180 = load ptr, ptr %v_strStack_str
  %t181 = getelementptr inbounds %TStack_String, ptr %t180, i32 0, i32 0
  %t182 = load ptr, ptr %t181
  %t183 = getelementptr inbounds [4 x ptr], ptr %t182, i32 0, i32 2
  %t184 = load ptr, ptr %t183
  %t185 = call ptr (ptr) %t184(ptr %t180)
  %t186 = call ptr @strcat(ptr %t174, ptr %t185)
  %t187 = call i32 @puts(ptr noundef %t174)
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"Stack count: \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [6 x i8] c"Pop: \00", align 1
@.str.4 = private unnamed_addr constant [6 x i8] c"Hello\00", align 1
@.str.5 = private unnamed_addr constant [6 x i8] c"World\00", align 1
