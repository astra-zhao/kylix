; Kylix LLVM IR — module: Variables
source_filename = "Variables.klx"
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
  %v_age_int = alloca i64, align 8
  store i64 0, ptr %v_age_int
  %v_name_str = alloca ptr, align 8
  store ptr null, ptr %v_name_str
  %v_pi_real = alloca double, align 8
  store double 0.0, ptr %v_pi_real
  %v_isActive_bool = alloca i1, align 8
  store i1 0, ptr %v_isActive_bool
  %v_count_int = alloca i64, align 8
  store i64 0, ptr %v_count_int
  %t0 = add i64 0, 42
  store i64 %t0, ptr %v_count_int
  %v_greeting_str = alloca ptr, align 8
  store ptr null, ptr %v_greeting_str
  %t1 = getelementptr inbounds [6 x i8], ptr @.str.0, i64 0, i64 0
  store ptr %t1, ptr %v_greeting_str
  %v_ratio_real = alloca double, align 8
  store double 0.0, ptr %v_ratio_real
  %t2 = fadd double 0.0, 3.141590
  store double %t2, ptr %v_ratio_real
  %v_flag_bool = alloca i1, align 8
  store i1 0, ptr %v_flag_bool
  %t3 = add i1 0, 1
  store i1 %t3, ptr %v_flag_bool
  %t4 = add i64 0, 25
  store i64 %t4, ptr %v_age_int
  %t5 = getelementptr inbounds [6 x i8], ptr @.str.1, i64 0, i64 0
  store ptr %t5, ptr %v_name_str
  %t6 = fadd double 0.0, 3.141590
  store double %t6, ptr %v_pi_real
  %t7 = add i1 0, 0
  store i1 %t7, ptr %v_isActive_bool
  %t8 = call ptr @malloc(i64 512)
  store i8 0, ptr %t8
  %t10 = getelementptr inbounds [7 x i8], ptr @.str.3, i64 0, i64 0
  %t11 = call ptr @strcat(ptr %t8, ptr %t10)
  %t12 = getelementptr inbounds [2 x i8], ptr @.str.4, i64 0, i64 0
  %t13 = call ptr @strcat(ptr %t8, ptr %t12)
  %t14 = load ptr, ptr %v_name_str
  %t15 = call ptr @strcat(ptr %t8, ptr %t14)
  %t16 = call i32 @puts(ptr noundef %t8)
  %t17 = call ptr @malloc(i64 512)
  store i8 0, ptr %t17
  %t18 = getelementptr inbounds [4 x i8], ptr @.str.2, i64 0, i64 0
  %t19 = getelementptr inbounds [6 x i8], ptr @.str.5, i64 0, i64 0
  %t20 = call ptr @strcat(ptr %t17, ptr %t19)
  %t21 = getelementptr inbounds [2 x i8], ptr @.str.4, i64 0, i64 0
  %t22 = call ptr @strcat(ptr %t17, ptr %t21)
  %t23 = load i64, ptr %v_age_int
  %t24 = call i64 @strlen(ptr %t17)
  %t25 = getelementptr inbounds i8, ptr %t17, i64 %t24
  %t26 = sub i64 512, %t24
  %t27 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t25, i64 %t26, ptr %t18, i64 %t23)
  %t28 = call i32 @puts(ptr noundef %t17)
  %t29 = call ptr @malloc(i64 512)
  store i8 0, ptr %t29
  %t31 = getelementptr inbounds [5 x i8], ptr @.str.6, i64 0, i64 0
  %t32 = call ptr @strcat(ptr %t29, ptr %t31)
  %t33 = getelementptr inbounds [2 x i8], ptr @.str.4, i64 0, i64 0
  %t34 = call ptr @strcat(ptr %t29, ptr %t33)
  %t35 = load double, ptr %v_pi_real
  %t36 = getelementptr inbounds [6 x i8], ptr @.str.7, i64 0, i64 0
  %t37 = call i64 @strlen(ptr %t29)
  %t38 = getelementptr inbounds i8, ptr %t29, i64 %t37
  %t39 = sub i64 512, %t37
  %t40 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t38, i64 %t39, ptr %t36, double %t35)
  %t41 = call i32 @puts(ptr noundef %t29)
  %t42 = call ptr @malloc(i64 512)
  store i8 0, ptr %t42
  %t44 = getelementptr inbounds [9 x i8], ptr @.str.8, i64 0, i64 0
  %t45 = call ptr @strcat(ptr %t42, ptr %t44)
  %t46 = getelementptr inbounds [2 x i8], ptr @.str.4, i64 0, i64 0
  %t47 = call ptr @strcat(ptr %t42, ptr %t46)
  %t48 = load i1, ptr %v_isActive_bool
  %t49 = getelementptr inbounds [5 x i8], ptr @.str.9, i64 0, i64 0
  %t50 = getelementptr inbounds [6 x i8], ptr @.str.10, i64 0, i64 0
  %t51 = select i1 %t48, ptr %t49, ptr %t50
  %t52 = call ptr @strcat(ptr %t42, ptr %t51)
  %t53 = call i32 @puts(ptr noundef %t42)
  %t54 = call ptr @malloc(i64 512)
  store i8 0, ptr %t54
  %t55 = getelementptr inbounds [4 x i8], ptr @.str.2, i64 0, i64 0
  %t56 = getelementptr inbounds [8 x i8], ptr @.str.11, i64 0, i64 0
  %t57 = call ptr @strcat(ptr %t54, ptr %t56)
  %t58 = getelementptr inbounds [2 x i8], ptr @.str.4, i64 0, i64 0
  %t59 = call ptr @strcat(ptr %t54, ptr %t58)
  %t60 = load i64, ptr %v_count_int
  %t61 = call i64 @strlen(ptr %t54)
  %t62 = getelementptr inbounds i8, ptr %t54, i64 %t61
  %t63 = sub i64 512, %t61
  %t64 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t62, i64 %t63, ptr %t55, i64 %t60)
  %t65 = call i32 @puts(ptr noundef %t54)
  %t66 = call ptr @malloc(i64 512)
  store i8 0, ptr %t66
  %t68 = getelementptr inbounds [11 x i8], ptr @.str.12, i64 0, i64 0
  %t69 = call ptr @strcat(ptr %t66, ptr %t68)
  %t70 = getelementptr inbounds [2 x i8], ptr @.str.4, i64 0, i64 0
  %t71 = call ptr @strcat(ptr %t66, ptr %t70)
  %t72 = load ptr, ptr %v_greeting_str
  %t73 = call ptr @strcat(ptr %t66, ptr %t72)
  %t74 = call i32 @puts(ptr noundef %t66)
  br label %lbl0
lbl0:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [6 x i8] c"Hello\00", align 1
@.str.1 = private unnamed_addr constant [6 x i8] c"Alice\00", align 1
@.str.2 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.3 = private unnamed_addr constant [7 x i8] c"Name: \00", align 1
@.str.4 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.5 = private unnamed_addr constant [6 x i8] c"Age: \00", align 1
@.str.6 = private unnamed_addr constant [5 x i8] c"Pi: \00", align 1
@.str.7 = private unnamed_addr constant [6 x i8] c"%.15g\00", align 1
@.str.8 = private unnamed_addr constant [9 x i8] c"Active: \00", align 1
@.str.9 = private unnamed_addr constant [5 x i8] c"true\00", align 1
@.str.10 = private unnamed_addr constant [6 x i8] c"false\00", align 1
@.str.11 = private unnamed_addr constant [8 x i8] c"Count: \00", align 1
@.str.12 = private unnamed_addr constant [11 x i8] c"Greeting: \00", align 1
