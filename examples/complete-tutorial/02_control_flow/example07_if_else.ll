; Kylix LLVM IR — module: IfElse
source_filename = "IfElse.klx"
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
  %v_x_int = alloca i64, align 8
  store i64 0, ptr %v_x_int
  %v_age_int = alloca i64, align 8
  store i64 0, ptr %v_age_int
  %v_score_int = alloca i64, align 8
  store i64 0, ptr %v_score_int
  %t0 = add i64 0, 10
  store i64 %t0, ptr %v_x_int
  %t1 = load i64, ptr %v_x_int
  %t2 = add i64 0, 5
  %t3 = icmp sgt i64 %t1, %t2
  br i1 %t3, label %lbl1, label %lbl2
lbl1:
  %t4 = getelementptr inbounds [20 x i8], ptr @.str.0, i64 0, i64 0
  %t5 = call i32 @puts(ptr noundef %t4)
  br label %lbl2
lbl2:
  %t6 = add i64 0, 18
  store i64 %t6, ptr %v_age_int
  %t7 = load i64, ptr %v_age_int
  %t8 = add i64 0, 18
  %t9 = icmp sge i64 %t7, %t8
  br i1 %t9, label %lbl3, label %lbl5
lbl3:
  %t10 = getelementptr inbounds [17 x i8], ptr @.str.1, i64 0, i64 0
  %t11 = call i32 @puts(ptr noundef %t10)
  br label %lbl4
lbl5:
  %t12 = getelementptr inbounds [16 x i8], ptr @.str.2, i64 0, i64 0
  %t13 = call i32 @puts(ptr noundef %t12)
  br label %lbl4
lbl4:
  %t14 = add i64 0, 85
  store i64 %t14, ptr %v_score_int
  %t15 = load i64, ptr %v_score_int
  %t16 = add i64 0, 90
  %t17 = icmp sge i64 %t15, %t16
  br i1 %t17, label %lbl6, label %lbl8
lbl6:
  %t18 = getelementptr inbounds [9 x i8], ptr @.str.3, i64 0, i64 0
  %t19 = call i32 @puts(ptr noundef %t18)
  br label %lbl7
lbl8:
  %t20 = load i64, ptr %v_score_int
  %t21 = add i64 0, 80
  %t22 = icmp sge i64 %t20, %t21
  br i1 %t22, label %lbl9, label %lbl11
lbl9:
  %t23 = getelementptr inbounds [9 x i8], ptr @.str.4, i64 0, i64 0
  %t24 = call i32 @puts(ptr noundef %t23)
  br label %lbl10
lbl11:
  %t25 = load i64, ptr %v_score_int
  %t26 = add i64 0, 70
  %t27 = icmp sge i64 %t25, %t26
  br i1 %t27, label %lbl12, label %lbl14
lbl12:
  %t28 = getelementptr inbounds [9 x i8], ptr @.str.5, i64 0, i64 0
  %t29 = call i32 @puts(ptr noundef %t28)
  br label %lbl13
lbl14:
  %t30 = getelementptr inbounds [9 x i8], ptr @.str.6, i64 0, i64 0
  %t31 = call i32 @puts(ptr noundef %t30)
  br label %lbl13
lbl13:
  br label %lbl10
lbl10:
  br label %lbl7
lbl7:
  br label %lbl0
lbl0:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [20 x i8] c"x is greater than 5\00", align 1
@.str.1 = private unnamed_addr constant [17 x i8] c"You are an adult\00", align 1
@.str.2 = private unnamed_addr constant [16 x i8] c"You are a minor\00", align 1
@.str.3 = private unnamed_addr constant [9 x i8] c"Grade: A\00", align 1
@.str.4 = private unnamed_addr constant [9 x i8] c"Grade: B\00", align 1
@.str.5 = private unnamed_addr constant [9 x i8] c"Grade: C\00", align 1
@.str.6 = private unnamed_addr constant [9 x i8] c"Grade: F\00", align 1
