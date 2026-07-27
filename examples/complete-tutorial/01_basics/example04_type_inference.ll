; Kylix LLVM IR — module: TypeInference
source_filename = "TypeInference.klx"
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
  %t0 = add i64 0, 42, !dbg !5
  %v_count_int = alloca i64, align 8, !dbg !6
  store i64 %t0, ptr %v_count_int, !dbg !6
  %t1 = getelementptr inbounds [6 x i8], ptr @.str.0, i64 0, i64 0, !dbg !7
  %v_message_str = alloca ptr, align 8, !dbg !8
  store ptr %t1, ptr %v_message_str, !dbg !8
  %t2 = fadd double 0.0, 3.141590, !dbg !9
  %v_ratio_real = alloca double, align 8, !dbg !10
  store double %t2, ptr %v_ratio_real, !dbg !10
  %t3 = add i1 0, 1, !dbg !11
  %v_active_bool = alloca i1, align 8, !dbg !12
  store i1 %t3, ptr %v_active_bool, !dbg !12
  %t4 = call ptr @malloc(i64 512), !dbg !13
  store i8 0, ptr %t4, !dbg !13
  %t5 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !13
  %t6 = getelementptr inbounds [8 x i8], ptr @.str.2, i64 0, i64 0, !dbg !14
  %t7 = call ptr @strcat(ptr %t4, ptr %t6), !dbg !13
  %t8 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !13
  %t9 = call ptr @strcat(ptr %t4, ptr %t8), !dbg !13
  %t10 = load i64, ptr %v_count_int, !dbg !15
  %t11 = call i64 @strlen(ptr %t4), !dbg !13
  %t12 = getelementptr inbounds i8, ptr %t4, i64 %t11, !dbg !13
  %t13 = sub i64 512, %t11, !dbg !13
  %t14 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t12, i64 %t13, ptr %t5, i64 %t10), !dbg !13
  %t15 = call i32 @puts(ptr noundef %t4), !dbg !13
  %t16 = call ptr @malloc(i64 512), !dbg !16
  store i8 0, ptr %t16, !dbg !16
  %t18 = getelementptr inbounds [10 x i8], ptr @.str.4, i64 0, i64 0, !dbg !17
  %t19 = call ptr @strcat(ptr %t16, ptr %t18), !dbg !16
  %t20 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !16
  %t21 = call ptr @strcat(ptr %t16, ptr %t20), !dbg !16
  %t22 = load ptr, ptr %v_message_str, !dbg !18
  %t23 = call ptr @strcat(ptr %t16, ptr %t22), !dbg !16
  %t24 = call i32 @puts(ptr noundef %t16), !dbg !16
  %t25 = call ptr @malloc(i64 512), !dbg !19
  store i8 0, ptr %t25, !dbg !19
  %t27 = getelementptr inbounds [8 x i8], ptr @.str.5, i64 0, i64 0, !dbg !20
  %t28 = call ptr @strcat(ptr %t25, ptr %t27), !dbg !19
  %t29 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !19
  %t30 = call ptr @strcat(ptr %t25, ptr %t29), !dbg !19
  %t31 = load double, ptr %v_ratio_real, !dbg !21
  %t32 = getelementptr inbounds [6 x i8], ptr @.str.6, i64 0, i64 0, !dbg !19
  %t33 = call i64 @strlen(ptr %t25), !dbg !19
  %t34 = getelementptr inbounds i8, ptr %t25, i64 %t33, !dbg !19
  %t35 = sub i64 512, %t33, !dbg !19
  %t36 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t34, i64 %t35, ptr %t32, double %t31), !dbg !19
  %t37 = call i32 @puts(ptr noundef %t25), !dbg !19
  %t38 = call ptr @malloc(i64 512), !dbg !22
  store i8 0, ptr %t38, !dbg !22
  %t40 = getelementptr inbounds [9 x i8], ptr @.str.7, i64 0, i64 0, !dbg !23
  %t41 = call ptr @strcat(ptr %t38, ptr %t40), !dbg !22
  %t42 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !22
  %t43 = call ptr @strcat(ptr %t38, ptr %t42), !dbg !22
  %t44 = load i1, ptr %v_active_bool, !dbg !24
  %t45 = getelementptr inbounds [5 x i8], ptr @.str.8, i64 0, i64 0, !dbg !22
  %t46 = getelementptr inbounds [6 x i8], ptr @.str.9, i64 0, i64 0, !dbg !22
  %t47 = select i1 %t44, ptr %t45, ptr %t46, !dbg !22
  %t48 = call ptr @strcat(ptr %t38, ptr %t47), !dbg !22
  %t49 = call i32 @puts(ptr noundef %t38), !dbg !22
  %t50 = add i64 0, 10, !dbg !25
  %v_result_int = alloca i64, align 8, !dbg !26
  store i64 %t50, ptr %v_result_int, !dbg !26
  %t51 = load i64, ptr %v_result_int, !dbg !27
  %t52 = add i64 0, 5, !dbg !28
  %t53 = add i64 %t51, %t52, !dbg !29
  store i64 %t53, ptr %v_result_int, !dbg !30
  %t54 = call ptr @malloc(i64 512), !dbg !31
  store i8 0, ptr %t54, !dbg !31
  %t55 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !31
  %t56 = getelementptr inbounds [9 x i8], ptr @.str.10, i64 0, i64 0, !dbg !32
  %t57 = call ptr @strcat(ptr %t54, ptr %t56), !dbg !31
  %t58 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !31
  %t59 = call ptr @strcat(ptr %t54, ptr %t58), !dbg !31
  %t60 = load i64, ptr %v_result_int, !dbg !33
  %t61 = call i64 @strlen(ptr %t54), !dbg !31
  %t62 = getelementptr inbounds i8, ptr %t54, i64 %t61, !dbg !31
  %t63 = sub i64 512, %t61, !dbg !31
  %t64 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t62, i64 %t63, ptr %t55, i64 %t60), !dbg !31
  %t65 = call i32 @puts(ptr noundef %t54), !dbg !31
  %t66 = add i64 0, 10, !dbg !34
  %t67 = add i64 0, 20, !dbg !35
  %t68 = add i64 %t66, %t67, !dbg !36
  %t69 = add i64 0, 30, !dbg !37
  %t70 = add i64 %t68, %t69, !dbg !38
  %v_sum_int = alloca i64, align 8, !dbg !39
  store i64 %t70, ptr %v_sum_int, !dbg !39
  %t71 = add i64 0, 5, !dbg !40
  %t72 = add i64 0, 6, !dbg !41
  %t73 = mul i64 %t71, %t72, !dbg !42
  %v_product_int = alloca i64, align 8, !dbg !43
  store i64 %t73, ptr %v_product_int, !dbg !43
  %t74 = add i64 0, 100, !dbg !44
  %t75 = add i64 0, 4, !dbg !45
  %t76 = sdiv i64 %t74, %t75, !dbg !46
  %v_division_int = alloca i64, align 8, !dbg !47
  store i64 %t76, ptr %v_division_int, !dbg !47
  %t77 = call ptr @malloc(i64 512), !dbg !48
  store i8 0, ptr %t77, !dbg !48
  %t78 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !48
  %t79 = getelementptr inbounds [6 x i8], ptr @.str.11, i64 0, i64 0, !dbg !49
  %t80 = call ptr @strcat(ptr %t77, ptr %t79), !dbg !48
  %t81 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !48
  %t82 = call ptr @strcat(ptr %t77, ptr %t81), !dbg !48
  %t83 = load i64, ptr %v_sum_int, !dbg !50
  %t84 = call i64 @strlen(ptr %t77), !dbg !48
  %t85 = getelementptr inbounds i8, ptr %t77, i64 %t84, !dbg !48
  %t86 = sub i64 512, %t84, !dbg !48
  %t87 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t85, i64 %t86, ptr %t78, i64 %t83), !dbg !48
  %t88 = call i32 @puts(ptr noundef %t77), !dbg !48
  %t89 = call ptr @malloc(i64 512), !dbg !51
  store i8 0, ptr %t89, !dbg !51
  %t90 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !51
  %t91 = getelementptr inbounds [10 x i8], ptr @.str.12, i64 0, i64 0, !dbg !52
  %t92 = call ptr @strcat(ptr %t89, ptr %t91), !dbg !51
  %t93 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !51
  %t94 = call ptr @strcat(ptr %t89, ptr %t93), !dbg !51
  %t95 = load i64, ptr %v_product_int, !dbg !53
  %t96 = call i64 @strlen(ptr %t89), !dbg !51
  %t97 = getelementptr inbounds i8, ptr %t89, i64 %t96, !dbg !51
  %t98 = sub i64 512, %t96, !dbg !51
  %t99 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t97, i64 %t98, ptr %t90, i64 %t95), !dbg !51
  %t100 = call i32 @puts(ptr noundef %t89), !dbg !51
  %t101 = call ptr @malloc(i64 512), !dbg !54
  store i8 0, ptr %t101, !dbg !54
  %t102 = getelementptr inbounds [4 x i8], ptr @.str.1, i64 0, i64 0, !dbg !54
  %t103 = getelementptr inbounds [11 x i8], ptr @.str.13, i64 0, i64 0, !dbg !55
  %t104 = call ptr @strcat(ptr %t101, ptr %t103), !dbg !54
  %t105 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0, !dbg !54
  %t106 = call ptr @strcat(ptr %t101, ptr %t105), !dbg !54
  %t107 = load i64, ptr %v_division_int, !dbg !56
  %t108 = call i64 @strlen(ptr %t101), !dbg !54
  %t109 = getelementptr inbounds i8, ptr %t101, i64 %t108, !dbg !54
  %t110 = sub i64 512, %t108, !dbg !54
  %t111 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t109, i64 %t110, ptr %t102, i64 %t107), !dbg !54
  %t112 = call i32 @puts(ptr noundef %t101), !dbg !54
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [6 x i8] c"Hello\00", align 1
@.str.1 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.2 = private unnamed_addr constant [8 x i8] c"Count: \00", align 1
@.str.3 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.4 = private unnamed_addr constant [10 x i8] c"Message: \00", align 1
@.str.5 = private unnamed_addr constant [8 x i8] c"Ratio: \00", align 1
@.str.6 = private unnamed_addr constant [6 x i8] c"%.15g\00", align 1
@.str.7 = private unnamed_addr constant [9 x i8] c"Active: \00", align 1
@.str.8 = private unnamed_addr constant [5 x i8] c"true\00", align 1
@.str.9 = private unnamed_addr constant [6 x i8] c"false\00", align 1
@.str.10 = private unnamed_addr constant [9 x i8] c"Result: \00", align 1
@.str.11 = private unnamed_addr constant [6 x i8] c"Sum: \00", align 1
@.str.12 = private unnamed_addr constant [10 x i8] c"Product: \00", align 1
@.str.13 = private unnamed_addr constant [11 x i8] c"Division: \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example04_type_inference.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/01_basics")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !58, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !61)
!5 = !DILocation(line: 6, column: 16, scope: !4)
!6 = !DILocation(line: 6, column: 3, scope: !4)
!7 = !DILocation(line: 7, column: 18, scope: !4)
!8 = !DILocation(line: 7, column: 3, scope: !4)
!9 = !DILocation(line: 8, column: 16, scope: !4)
!10 = !DILocation(line: 8, column: 3, scope: !4)
!11 = !DILocation(line: 9, column: 17, scope: !4)
!12 = !DILocation(line: 9, column: 3, scope: !4)
!13 = !DILocation(line: 11, column: 10, scope: !4)
!14 = !DILocation(line: 11, column: 11, scope: !4)
!15 = !DILocation(line: 11, column: 22, scope: !4)
!16 = !DILocation(line: 12, column: 10, scope: !4)
!17 = !DILocation(line: 12, column: 11, scope: !4)
!18 = !DILocation(line: 12, column: 24, scope: !4)
!19 = !DILocation(line: 13, column: 10, scope: !4)
!20 = !DILocation(line: 13, column: 11, scope: !4)
!21 = !DILocation(line: 13, column: 22, scope: !4)
!22 = !DILocation(line: 14, column: 10, scope: !4)
!23 = !DILocation(line: 14, column: 11, scope: !4)
!24 = !DILocation(line: 14, column: 23, scope: !4)
!25 = !DILocation(line: 17, column: 17, scope: !4)
!26 = !DILocation(line: 17, column: 3, scope: !4)
!27 = !DILocation(line: 18, column: 13, scope: !4)
!28 = !DILocation(line: 18, column: 22, scope: !4)
!29 = !DILocation(line: 18, column: 20, scope: !4)
!30 = !DILocation(line: 18, column: 11, scope: !4)
!31 = !DILocation(line: 19, column: 10, scope: !4)
!32 = !DILocation(line: 19, column: 11, scope: !4)
!33 = !DILocation(line: 19, column: 23, scope: !4)
!34 = !DILocation(line: 22, column: 14, scope: !4)
!35 = !DILocation(line: 22, column: 19, scope: !4)
!36 = !DILocation(line: 22, column: 17, scope: !4)
!37 = !DILocation(line: 22, column: 24, scope: !4)
!38 = !DILocation(line: 22, column: 22, scope: !4)
!39 = !DILocation(line: 22, column: 3, scope: !4)
!40 = !DILocation(line: 23, column: 18, scope: !4)
!41 = !DILocation(line: 23, column: 22, scope: !4)
!42 = !DILocation(line: 23, column: 20, scope: !4)
!43 = !DILocation(line: 23, column: 3, scope: !4)
!44 = !DILocation(line: 24, column: 19, scope: !4)
!45 = !DILocation(line: 24, column: 25, scope: !4)
!46 = !DILocation(line: 24, column: 23, scope: !4)
!47 = !DILocation(line: 24, column: 3, scope: !4)
!48 = !DILocation(line: 26, column: 10, scope: !4)
!49 = !DILocation(line: 26, column: 11, scope: !4)
!50 = !DILocation(line: 26, column: 20, scope: !4)
!51 = !DILocation(line: 27, column: 10, scope: !4)
!52 = !DILocation(line: 27, column: 11, scope: !4)
!53 = !DILocation(line: 27, column: 24, scope: !4)
!54 = !DILocation(line: 28, column: 10, scope: !4)
!55 = !DILocation(line: 28, column: 11, scope: !4)
!56 = !DILocation(line: 28, column: 25, scope: !4)
!57 = !{null}
!58 = !DISubroutineType(types: !57)
!59 = !{}
!60 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!61 = !{}
