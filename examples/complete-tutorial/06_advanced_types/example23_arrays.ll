; Kylix LLVM IR — module: Arrays
source_filename = "Arrays.klx"
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
  %v_numbers_arr = alloca [5 x i64], align 8
  store [5 x i64] zeroinitializer, ptr %v_numbers_arr
  %v_names_arr = alloca [3 x ptr], align 8
  store [3 x ptr] zeroinitializer, ptr %v_names_arr
  %v_i_int = alloca i64, align 8
  store i64 0, ptr %v_i_int
  %t0 = add i64 0, 10
  %t1 = add i64 0, 0
  %t2 = add i64 %t1, 0
  %t3 = getelementptr inbounds [5 x i64], ptr %v_numbers_arr, i64 0, i64 %t2
  store i64 %t0, ptr %t3
  %t4 = add i64 0, 20
  %t5 = add i64 0, 1
  %t6 = add i64 %t5, 0
  %t7 = getelementptr inbounds [5 x i64], ptr %v_numbers_arr, i64 0, i64 %t6
  store i64 %t4, ptr %t7
  %t8 = add i64 0, 30
  %t9 = add i64 0, 2
  %t10 = add i64 %t9, 0
  %t11 = getelementptr inbounds [5 x i64], ptr %v_numbers_arr, i64 0, i64 %t10
  store i64 %t8, ptr %t11
  %t12 = add i64 0, 40
  %t13 = add i64 0, 3
  %t14 = add i64 %t13, 0
  %t15 = getelementptr inbounds [5 x i64], ptr %v_numbers_arr, i64 0, i64 %t14
  store i64 %t12, ptr %t15
  %t16 = add i64 0, 50
  %t17 = add i64 0, 4
  %t18 = add i64 %t17, 0
  %t19 = getelementptr inbounds [5 x i64], ptr %v_numbers_arr, i64 0, i64 %t18
  store i64 %t16, ptr %t19
  %t20 = getelementptr inbounds [15 x i8], ptr @.str.0, i64 0, i64 0
  %t21 = call i32 @puts(ptr noundef %t20)
  %t22 = add i64 0, 0
  store i64 %t22, ptr %v_i_int
  br label %lbl0
lbl0:
  %t23 = load i64, ptr %v_i_int
  %t24 = add i64 0, 4
  %t25 = icmp sle i64 %t23, %t24
  br i1 %t25, label %lbl1, label %lbl2
lbl1:
  %t26 = call ptr @malloc(i64 512)
  store i8 0, ptr %t26
  %t27 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0
  %t28 = getelementptr inbounds [9 x i8], ptr @.str.2, i64 0, i64 0
  %t29 = call ptr @strcat(ptr %t26, ptr %t28)
  %t30 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0
  %t31 = call ptr @strcat(ptr %t26, ptr %t30)
  %t32 = load i64, ptr %v_i_int
  %t33 = call i64 @strlen(ptr %t26)
  %t34 = getelementptr inbounds i8, ptr %t26, i64 %t33
  %t35 = sub i64 512, %t33
  %t36 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t34, i64 %t35, ptr %t27, i64 %t32)
  %t37 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0
  %t38 = call ptr @strcat(ptr %t26, ptr %t37)
  %t39 = getelementptr inbounds [5 x i8], ptr @.str.4, i64 0, i64 0
  %t40 = call ptr @strcat(ptr %t26, ptr %t39)
  %t41 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0
  %t42 = call ptr @strcat(ptr %t26, ptr %t41)
  %t43 = load i64, ptr %v_i_int
  %t44 = add i64 %t43, 0
  %t45 = getelementptr inbounds [5 x i64], ptr %v_numbers_arr, i64 0, i64 %t44
  %t46 = load i64, ptr %t45
  %t47 = call i64 @strlen(ptr %t26)
  %t48 = getelementptr inbounds i8, ptr %t26, i64 %t47
  %t49 = sub i64 512, %t47
  %t50 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t48, i64 %t49, ptr %t27, i64 %t46)
  %t51 = call i32 @puts(ptr noundef %t26)
  %t53 = load i64, ptr %v_i_int
  %t52 = add i64 %t53, 1
  store i64 %t52, ptr %v_i_int
  br label %lbl0
lbl2:
  %t54 = getelementptr inbounds [6 x i8], ptr @.str.5, i64 0, i64 0
  %t55 = add i64 0, 0
  %t56 = add i64 %t55, 0
  %t57 = getelementptr inbounds [3 x ptr], ptr %v_names_arr, i64 0, i64 %t56
  store ptr %t54, ptr %t57
  %t58 = getelementptr inbounds [4 x i8], ptr @.str.6, i64 0, i64 0
  %t59 = add i64 0, 1
  %t60 = add i64 %t59, 0
  %t61 = getelementptr inbounds [3 x ptr], ptr %v_names_arr, i64 0, i64 %t60
  store ptr %t58, ptr %t61
  %t62 = getelementptr inbounds [8 x i8], ptr @.str.7, i64 0, i64 0
  %t63 = add i64 0, 2
  %t64 = add i64 %t63, 0
  %t65 = getelementptr inbounds [3 x ptr], ptr %v_names_arr, i64 0, i64 %t64
  store ptr %t62, ptr %t65
  %t66 = getelementptr inbounds [7 x i8], ptr @.str.8, i64 0, i64 0
  %t67 = call i32 @puts(ptr noundef %t66)
  %t68 = add i64 0, 0
  store i64 %t68, ptr %v_i_int
  br label %lbl3
lbl3:
  %t69 = load i64, ptr %v_i_int
  %t70 = add i64 0, 2
  %t71 = icmp sle i64 %t69, %t70
  br i1 %t71, label %lbl4, label %lbl5
lbl4:
  %t72 = call ptr @malloc(i64 512)
  store i8 0, ptr %t72
  %t73 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0
  %t74 = load i64, ptr %v_i_int
  %t75 = add i64 0, 1
  %t76 = add i64 %t74, %t75
  %t77 = call i64 @strlen(ptr %t72)
  %t78 = getelementptr inbounds i8, ptr %t72, i64 %t77
  %t79 = sub i64 512, %t77
  %t80 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t78, i64 %t79, ptr %t73, i64 %t76)
  %t81 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0
  %t82 = call ptr @strcat(ptr %t72, ptr %t81)
  %t83 = getelementptr inbounds [3 x i8], ptr @.str.9, i64 0, i64 0
  %t84 = call ptr @strcat(ptr %t72, ptr %t83)
  %t85 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0
  %t86 = call ptr @strcat(ptr %t72, ptr %t85)
  %t87 = load i64, ptr %v_i_int
  %t88 = add i64 %t87, 0
  %t89 = getelementptr inbounds [3 x ptr], ptr %v_names_arr, i64 0, i64 %t88
  %t90 = load ptr, ptr %t89
  %t91 = call ptr @strcat(ptr %t72, ptr %t90)
  %t92 = call i32 @puts(ptr noundef %t72)
  %t94 = load i64, ptr %v_i_int
  %t93 = add i64 %t94, 1
  store i64 %t93, ptr %v_i_int
  br label %lbl3
lbl5:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [15 x i8] c"Numbers array:\00", align 1
@.str.1 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.2 = private unnamed_addr constant [9 x i8] c"numbers[\00", align 1
@.str.3 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.4 = private unnamed_addr constant [5 x i8] c"] = \00", align 1
@.str.5 = private unnamed_addr constant [6 x i8] c"Alice\00", align 1
@.str.6 = private unnamed_addr constant [4 x i8] c"Bob\00", align 1
@.str.7 = private unnamed_addr constant [8 x i8] c"Charlie\00", align 1
@.str.8 = private unnamed_addr constant [7 x i8] c"Names:\00", align 1
@.str.9 = private unnamed_addr constant [3 x i8] c". \00", align 1
