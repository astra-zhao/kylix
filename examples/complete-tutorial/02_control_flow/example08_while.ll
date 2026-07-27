; Kylix LLVM IR — module: WhileLoop
source_filename = "WhileLoop.klx"
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
  %v_sum_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_sum_int, !dbg !5
  #dbg_declare(ptr %v_sum_int, !7, !DIExpression(), !5)
  %v_countdown_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_countdown_int, !dbg !5
  #dbg_declare(ptr %v_countdown_int, !8, !DIExpression(), !5)
  %t0 = add i64 0, 1, !dbg !9
  store i64 %t0, ptr %v_i_int, !dbg !10
  %t1 = add i64 0, 0, !dbg !11
  store i64 %t1, ptr %v_sum_int, !dbg !12
  br label %lbl0, !dbg !13
lbl0:
  %t2 = load i64, ptr %v_i_int, !dbg !14
  %t3 = add i64 0, 10, !dbg !15
  %t4 = icmp sle i64 %t2, %t3, !dbg !16
  br i1 %t4, label %lbl1, label %lbl2, !dbg !13
lbl1:
  %t5 = load i64, ptr %v_sum_int, !dbg !17
  %t6 = load i64, ptr %v_i_int, !dbg !18
  %t7 = add i64 %t5, %t6, !dbg !19
  store i64 %t7, ptr %v_sum_int, !dbg !20
  %t8 = load i64, ptr %v_i_int, !dbg !21
  %t9 = add i64 0, 1, !dbg !22
  %t10 = add i64 %t8, %t9, !dbg !23
  store i64 %t10, ptr %v_i_int, !dbg !24
  br label %lbl0, !dbg !13
lbl2:
  %t11 = call ptr @malloc(i64 512), !dbg !25
  store i8 0, ptr %t11, !dbg !25
  %t12 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !25
  %t13 = getelementptr inbounds [14 x i8], ptr @.str.1, i64 0, i64 0, !dbg !26
  %t14 = call ptr @strcat(ptr %t11, ptr %t13), !dbg !25
  %t15 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !25
  %t16 = call ptr @strcat(ptr %t11, ptr %t15), !dbg !25
  %t17 = load i64, ptr %v_sum_int, !dbg !27
  %t18 = call i64 @strlen(ptr %t11), !dbg !25
  %t19 = getelementptr inbounds i8, ptr %t11, i64 %t18, !dbg !25
  %t20 = sub i64 512, %t18, !dbg !25
  %t21 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t19, i64 %t20, ptr %t12, i64 %t17), !dbg !25
  %t22 = call i32 @puts(ptr noundef %t11), !dbg !25
  %t23 = add i64 0, 5, !dbg !28
  store i64 %t23, ptr %v_countdown_int, !dbg !29
  %t24 = getelementptr inbounds [11 x i8], ptr @.str.3, i64 0, i64 0, !dbg !30
  %t25 = call i32 @puts(ptr noundef %t24), !dbg !31
  br label %lbl3, !dbg !32
lbl3:
  %t26 = load i64, ptr %v_countdown_int, !dbg !33
  %t27 = add i64 0, 0, !dbg !34
  %t28 = icmp sgt i64 %t26, %t27, !dbg !35
  br i1 %t28, label %lbl4, label %lbl5, !dbg !32
lbl4:
  %t29 = load i64, ptr %v_countdown_int, !dbg !36
  %t30 = getelementptr inbounds [6 x i8], ptr @.str.4, i64 0, i64 0, !dbg !37
  %t31 = call i32 (ptr, ...) @printf(ptr noundef %t30, i64 %t29), !dbg !37
  %t32 = load i64, ptr %v_countdown_int, !dbg !38
  %t33 = add i64 0, 1, !dbg !39
  %t34 = sub i64 %t32, %t33, !dbg !40
  store i64 %t34, ptr %v_countdown_int, !dbg !41
  br label %lbl3, !dbg !32
lbl5:
  %t35 = getelementptr inbounds [11 x i8], ptr @.str.5, i64 0, i64 0, !dbg !42
  %t36 = call i32 @puts(ptr noundef %t35), !dbg !43
  %t37 = add i64 0, 100, !dbg !44
  store i64 %t37, ptr %v_i_int, !dbg !45
  br label %lbl6, !dbg !46
lbl6:
  %t38 = load i64, ptr %v_i_int, !dbg !47
  %t39 = add i64 0, 10, !dbg !48
  %t40 = icmp slt i64 %t38, %t39, !dbg !49
  br i1 %t40, label %lbl7, label %lbl8, !dbg !46
lbl7:
  %t41 = getelementptr inbounds [20 x i8], ptr @.str.6, i64 0, i64 0, !dbg !50
  %t42 = call i32 @puts(ptr noundef %t41), !dbg !51
  %t43 = load i64, ptr %v_i_int, !dbg !52
  %t44 = add i64 0, 1, !dbg !53
  %t45 = add i64 %t43, %t44, !dbg !54
  store i64 %t45, ptr %v_i_int, !dbg !55
  br label %lbl6, !dbg !46
lbl8:
  %t46 = getelementptr inbounds [13 x i8], ptr @.str.7, i64 0, i64 0, !dbg !56
  %t47 = call i32 @puts(ptr noundef %t46), !dbg !57
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"Sum of 1-10: \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [11 x i8] c"Countdown:\00", align 1
@.str.4 = private unnamed_addr constant [6 x i8] c"%lld\0A\00", align 1
@.str.5 = private unnamed_addr constant [11 x i8] c"Blast off!\00", align 1
@.str.6 = private unnamed_addr constant [20 x i8] c"This will not print\00", align 1
@.str.7 = private unnamed_addr constant [13 x i8] c"Loop skipped\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example08_while.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/02_control_flow")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !59, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !62)
!6 = !DILocalVariable(name: "i", scope: !4, file: !3, line: 4, type: !61)
!7 = !DILocalVariable(name: "sum", scope: !4, file: !3, line: 5, type: !61)
!8 = !DILocalVariable(name: "countdown", scope: !4, file: !3, line: 6, type: !61)
!5 = !DILocation(line: 1, column: 9, scope: !4)
!9 = !DILocation(line: 10, column: 8, scope: !4)
!10 = !DILocation(line: 10, column: 6, scope: !4)
!11 = !DILocation(line: 11, column: 10, scope: !4)
!12 = !DILocation(line: 11, column: 8, scope: !4)
!13 = !DILocation(line: 12, column: 3, scope: !4)
!14 = !DILocation(line: 12, column: 9, scope: !4)
!15 = !DILocation(line: 12, column: 14, scope: !4)
!16 = !DILocation(line: 12, column: 12, scope: !4)
!17 = !DILocation(line: 14, column: 12, scope: !4)
!18 = !DILocation(line: 14, column: 18, scope: !4)
!19 = !DILocation(line: 14, column: 16, scope: !4)
!20 = !DILocation(line: 14, column: 10, scope: !4)
!21 = !DILocation(line: 15, column: 10, scope: !4)
!22 = !DILocation(line: 15, column: 14, scope: !4)
!23 = !DILocation(line: 15, column: 12, scope: !4)
!24 = !DILocation(line: 15, column: 8, scope: !4)
!25 = !DILocation(line: 17, column: 10, scope: !4)
!26 = !DILocation(line: 17, column: 11, scope: !4)
!27 = !DILocation(line: 17, column: 28, scope: !4)
!28 = !DILocation(line: 20, column: 16, scope: !4)
!29 = !DILocation(line: 20, column: 14, scope: !4)
!30 = !DILocation(line: 21, column: 11, scope: !4)
!31 = !DILocation(line: 21, column: 10, scope: !4)
!32 = !DILocation(line: 22, column: 3, scope: !4)
!33 = !DILocation(line: 22, column: 9, scope: !4)
!34 = !DILocation(line: 22, column: 21, scope: !4)
!35 = !DILocation(line: 22, column: 19, scope: !4)
!36 = !DILocation(line: 24, column: 13, scope: !4)
!37 = !DILocation(line: 24, column: 12, scope: !4)
!38 = !DILocation(line: 25, column: 18, scope: !4)
!39 = !DILocation(line: 25, column: 30, scope: !4)
!40 = !DILocation(line: 25, column: 28, scope: !4)
!41 = !DILocation(line: 25, column: 16, scope: !4)
!42 = !DILocation(line: 27, column: 11, scope: !4)
!43 = !DILocation(line: 27, column: 10, scope: !4)
!44 = !DILocation(line: 30, column: 8, scope: !4)
!45 = !DILocation(line: 30, column: 6, scope: !4)
!46 = !DILocation(line: 31, column: 3, scope: !4)
!47 = !DILocation(line: 31, column: 9, scope: !4)
!48 = !DILocation(line: 31, column: 13, scope: !4)
!49 = !DILocation(line: 31, column: 11, scope: !4)
!50 = !DILocation(line: 33, column: 13, scope: !4)
!51 = !DILocation(line: 33, column: 12, scope: !4)
!52 = !DILocation(line: 34, column: 10, scope: !4)
!53 = !DILocation(line: 34, column: 14, scope: !4)
!54 = !DILocation(line: 34, column: 12, scope: !4)
!55 = !DILocation(line: 34, column: 8, scope: !4)
!56 = !DILocation(line: 36, column: 11, scope: !4)
!57 = !DILocation(line: 36, column: 10, scope: !4)
!58 = !{null}
!59 = !DISubroutineType(types: !58)
!60 = !{}
!61 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!62 = !{!6, !7, !8}
