; Kylix LLVM IR — module: RepeatUntil
source_filename = "RepeatUntil.klx"
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
  %v_i_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_i_int, !dbg !5
  #dbg_declare(ptr %v_i_int, !6, !DIExpression(), !5)
  %v_guess_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_guess_int, !dbg !5
  #dbg_declare(ptr %v_guess_int, !7, !DIExpression(), !5)
  %t0 = add i64 0, 1, !dbg !8
  store i64 %t0, ptr %v_i_int, !dbg !9
  br label %lbl0, !dbg !10
lbl0:
  %t1 = call ptr @malloc(i64 512), !dbg !11
  store i8 0, ptr %t1, !dbg !11
  %t2 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !11
  %t3 = getelementptr inbounds [12 x i8], ptr @.str.1, i64 0, i64 0, !dbg !12
  %t4 = call ptr @strcat(ptr %t1, ptr %t3), !dbg !11
  %t5 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !11
  %t6 = call ptr @strcat(ptr %t1, ptr %t5), !dbg !11
  %t7 = load i64, ptr %v_i_int, !dbg !13
  %t8 = call i64 @strlen(ptr %t1), !dbg !11
  %t9 = getelementptr inbounds i8, ptr %t1, i64 %t8, !dbg !11
  %t10 = sub i64 512, %t8, !dbg !11
  %t11 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t9, i64 %t10, ptr %t2, i64 %t7), !dbg !11
  %t12 = call i32 @puts(ptr noundef %t1), !dbg !11
  %t13 = load i64, ptr %v_i_int, !dbg !14
  %t14 = add i64 0, 1, !dbg !15
  %t15 = add i64 %t13, %t14, !dbg !16
  store i64 %t15, ptr %v_i_int, !dbg !17
  %t16 = load i64, ptr %v_i_int, !dbg !18
  %t17 = add i64 0, 5, !dbg !19
  %t18 = icmp sgt i64 %t16, %t17, !dbg !20
  %t19 = xor i1 %t18, 1, !dbg !10
  br i1 %t19, label %lbl0, label %lbl1, !dbg !10
lbl1:
  %t20 = add i64 0, 100, !dbg !21
  store i64 %t20, ptr %v_guess_int, !dbg !22
  %t21 = getelementptr inbounds [52 x i8], ptr @.str.3, i64 0, i64 0, !dbg !23
  %t22 = call i32 @puts(ptr noundef %t21), !dbg !24
  br label %lbl2, !dbg !25
lbl2:
  %t23 = call ptr @malloc(i64 512), !dbg !26
  store i8 0, ptr %t23, !dbg !26
  %t24 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !26
  %t25 = getelementptr inbounds [14 x i8], ptr @.str.4, i64 0, i64 0, !dbg !27
  %t26 = call ptr @strcat(ptr %t23, ptr %t25), !dbg !26
  %t27 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !26
  %t28 = call ptr @strcat(ptr %t23, ptr %t27), !dbg !26
  %t29 = load i64, ptr %v_guess_int, !dbg !28
  %t30 = call i64 @strlen(ptr %t23), !dbg !26
  %t31 = getelementptr inbounds i8, ptr %t23, i64 %t30, !dbg !26
  %t32 = sub i64 512, %t30, !dbg !26
  %t33 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t31, i64 %t32, ptr %t24, i64 %t29), !dbg !26
  %t34 = call i32 @puts(ptr noundef %t23), !dbg !26
  %t35 = load i64, ptr %v_guess_int, !dbg !29
  %t36 = add i64 0, 1, !dbg !30
  %t37 = add i64 %t35, %t36, !dbg !31
  store i64 %t37, ptr %v_guess_int, !dbg !32
  %t38 = load i64, ptr %v_guess_int, !dbg !33
  %t39 = add i64 0, 50, !dbg !34
  %t40 = icmp sgt i64 %t38, %t39, !dbg !35
  %t41 = xor i1 %t40, 1, !dbg !25
  br i1 %t41, label %lbl2, label %lbl3, !dbg !25
lbl3:
  %t42 = add i64 0, 10, !dbg !36
  store i64 %t42, ptr %v_i_int, !dbg !37
  %t43 = getelementptr inbounds [23 x i8], ptr @.str.5, i64 0, i64 0, !dbg !38
  %t44 = call i32 @puts(ptr noundef %t43), !dbg !39
  br label %lbl4, !dbg !40
lbl4:
  %t45 = load i64, ptr %v_i_int, !dbg !41
  %t46 = getelementptr inbounds [6 x i8], ptr @.str.6, i64 0, i64 0, !dbg !42
  %t47 = call i32 (ptr, ...) @printf(ptr noundef %t46, i64 %t45), !dbg !42
  %t48 = load i64, ptr %v_i_int, !dbg !43
  %t49 = add i64 0, 1, !dbg !44
  %t50 = sub i64 %t48, %t49, !dbg !45
  store i64 %t50, ptr %v_i_int, !dbg !46
  %t51 = load i64, ptr %v_i_int, !dbg !47
  %t52 = add i64 0, 0, !dbg !48
  %t53 = icmp sle i64 %t51, %t52, !dbg !49
  %t54 = xor i1 %t53, 1, !dbg !40
  br i1 %t54, label %lbl4, label %lbl5, !dbg !40
lbl5:
  %t55 = getelementptr inbounds [6 x i8], ptr @.str.7, i64 0, i64 0, !dbg !50
  %t56 = call i32 @puts(ptr noundef %t55), !dbg !51
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [12 x i8] c"Iteration: \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [52 x i8] c"This will print once even though condition is true:\00", align 1
@.str.4 = private unnamed_addr constant [14 x i8] c"Guess value: \00", align 1
@.str.5 = private unnamed_addr constant [23 x i8] c"Countdown with repeat:\00", align 1
@.str.6 = private unnamed_addr constant [6 x i8] c"%lld\0A\00", align 1
@.str.7 = private unnamed_addr constant [6 x i8] c"Done!\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example10_repeat.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/02_control_flow")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !53, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !56)
!6 = !DILocalVariable(name: "i", scope: !4, file: !3, line: 4, type: !55)
!7 = !DILocalVariable(name: "guess", scope: !4, file: !3, line: 5, type: !55)
!5 = !DILocation(line: 1, column: 9, scope: !4)
!8 = !DILocation(line: 9, column: 8, scope: !4)
!9 = !DILocation(line: 9, column: 6, scope: !4)
!10 = !DILocation(line: 10, column: 3, scope: !4)
!11 = !DILocation(line: 11, column: 12, scope: !4)
!12 = !DILocation(line: 11, column: 13, scope: !4)
!13 = !DILocation(line: 11, column: 28, scope: !4)
!14 = !DILocation(line: 12, column: 10, scope: !4)
!15 = !DILocation(line: 12, column: 14, scope: !4)
!16 = !DILocation(line: 12, column: 12, scope: !4)
!17 = !DILocation(line: 12, column: 8, scope: !4)
!18 = !DILocation(line: 13, column: 9, scope: !4)
!19 = !DILocation(line: 13, column: 13, scope: !4)
!20 = !DILocation(line: 13, column: 11, scope: !4)
!21 = !DILocation(line: 17, column: 12, scope: !4)
!22 = !DILocation(line: 17, column: 10, scope: !4)
!23 = !DILocation(line: 18, column: 11, scope: !4)
!24 = !DILocation(line: 18, column: 10, scope: !4)
!25 = !DILocation(line: 19, column: 3, scope: !4)
!26 = !DILocation(line: 20, column: 12, scope: !4)
!27 = !DILocation(line: 20, column: 13, scope: !4)
!28 = !DILocation(line: 20, column: 30, scope: !4)
!29 = !DILocation(line: 21, column: 14, scope: !4)
!30 = !DILocation(line: 21, column: 22, scope: !4)
!31 = !DILocation(line: 21, column: 20, scope: !4)
!32 = !DILocation(line: 21, column: 12, scope: !4)
!33 = !DILocation(line: 22, column: 9, scope: !4)
!34 = !DILocation(line: 22, column: 17, scope: !4)
!35 = !DILocation(line: 22, column: 15, scope: !4)
!36 = !DILocation(line: 25, column: 8, scope: !4)
!37 = !DILocation(line: 25, column: 6, scope: !4)
!38 = !DILocation(line: 26, column: 11, scope: !4)
!39 = !DILocation(line: 26, column: 10, scope: !4)
!40 = !DILocation(line: 27, column: 3, scope: !4)
!41 = !DILocation(line: 28, column: 13, scope: !4)
!42 = !DILocation(line: 28, column: 12, scope: !4)
!43 = !DILocation(line: 29, column: 10, scope: !4)
!44 = !DILocation(line: 29, column: 14, scope: !4)
!45 = !DILocation(line: 29, column: 12, scope: !4)
!46 = !DILocation(line: 29, column: 8, scope: !4)
!47 = !DILocation(line: 30, column: 9, scope: !4)
!48 = !DILocation(line: 30, column: 14, scope: !4)
!49 = !DILocation(line: 30, column: 12, scope: !4)
!50 = !DILocation(line: 31, column: 11, scope: !4)
!51 = !DILocation(line: 31, column: 10, scope: !4)
!52 = !{null}
!53 = !DISubroutineType(types: !52)
!54 = !{}
!55 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!56 = !{!6, !7}
