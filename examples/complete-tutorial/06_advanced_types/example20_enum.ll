; Kylix LLVM IR — module: EnumTypes
source_filename = "EnumTypes.klx"
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
  %v_dir_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_dir_int, !dbg !5
  #dbg_declare(ptr %v_dir_int, !6, !DIExpression(), !5)
  %v_color_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_color_int, !dbg !5
  #dbg_declare(ptr %v_color_int, !7, !DIExpression(), !5)
  %v_status_int = alloca i64, align 8, !dbg !5
  store i64 0, ptr %v_status_int, !dbg !5
  #dbg_declare(ptr %v_status_int, !8, !DIExpression(), !5)
  %t0 = add i64 0, 0 ; undefined var North, !dbg !9
  store i64 %t0, ptr %v_dir_int, !dbg !10
  br label %lbl0, !dbg !12
lbl2:
  %t2 = getelementptr inbounds [17 x i8], ptr @.str.0, i64 0, i64 0, !dbg !13
  %t3 = call i32 @puts(ptr noundef %t2), !dbg !14
  br label %lbl1, !dbg !12
lbl3:
  %t4 = getelementptr inbounds [17 x i8], ptr @.str.1, i64 0, i64 0, !dbg !15
  %t5 = call i32 @puts(ptr noundef %t4), !dbg !16
  br label %lbl1, !dbg !12
lbl4:
  %t6 = getelementptr inbounds [16 x i8], ptr @.str.2, i64 0, i64 0, !dbg !17
  %t7 = call i32 @puts(ptr noundef %t6), !dbg !18
  br label %lbl1, !dbg !12
lbl5:
  %t8 = getelementptr inbounds [16 x i8], ptr @.str.3, i64 0, i64 0, !dbg !19
  %t9 = call i32 @puts(ptr noundef %t8), !dbg !20
  br label %lbl1, !dbg !12
lbl0:
  br label %lbl1, !dbg !12
lbl1:
  %t10 = add i64 0, 0 ; undefined var Green, !dbg !21
  store i64 %t10, ptr %v_color_int, !dbg !22
  %t11 = load i64, ptr %v_color_int, !dbg !23
  %t12 = add i64 0, 0 ; undefined var Green, !dbg !24
  %t13 = icmp eq i64 %t11, %t12, !dbg !25
  br i1 %t13, label %lbl6, label %lbl7, !dbg !26
lbl6:
  %t14 = getelementptr inbounds [15 x i8], ptr @.str.4, i64 0, i64 0, !dbg !27
  %t15 = call i32 @puts(ptr noundef %t14), !dbg !28
  br label %lbl7, !dbg !26
lbl7:
  %t16 = add i64 0, 0 ; undefined var Active, !dbg !29
  store i64 %t16, ptr %v_status_int, !dbg !30
  br label %lbl8, !dbg !32
lbl10:
  %t18 = getelementptr inbounds [16 x i8], ptr @.str.5, i64 0, i64 0, !dbg !33
  %t19 = call i32 @puts(ptr noundef %t18), !dbg !34
  br label %lbl9, !dbg !32
lbl11:
  %t20 = getelementptr inbounds [15 x i8], ptr @.str.6, i64 0, i64 0, !dbg !35
  %t21 = call i32 @puts(ptr noundef %t20), !dbg !36
  br label %lbl9, !dbg !32
lbl12:
  %t22 = getelementptr inbounds [17 x i8], ptr @.str.7, i64 0, i64 0, !dbg !37
  %t23 = call i32 @puts(ptr noundef %t22), !dbg !38
  br label %lbl9, !dbg !32
lbl13:
  %t24 = getelementptr inbounds [16 x i8], ptr @.str.8, i64 0, i64 0, !dbg !39
  %t25 = call i32 @puts(ptr noundef %t24), !dbg !40
  br label %lbl9, !dbg !32
lbl8:
  br label %lbl9, !dbg !32
lbl9:
  %t26 = add i64 0, 0 ; undefined var West, !dbg !41
  store i64 %t26, ptr %v_dir_int, !dbg !42
  %t27 = load i64, ptr %v_dir_int, !dbg !43
  %t28 = add i64 0, 0 ; undefined var North, !dbg !44
  %t29 = icmp ne i64 %t27, %t28, !dbg !45
  br i1 %t29, label %lbl14, label %lbl15, !dbg !46
lbl14:
  %t30 = getelementptr inbounds [16 x i8], ptr @.str.9, i64 0, i64 0, !dbg !47
  %t31 = call i32 @puts(ptr noundef %t30), !dbg !48
  br label %lbl15, !dbg !46
lbl15:
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [17 x i8] c"Direction: North\00", align 1
@.str.1 = private unnamed_addr constant [17 x i8] c"Direction: South\00", align 1
@.str.2 = private unnamed_addr constant [16 x i8] c"Direction: East\00", align 1
@.str.3 = private unnamed_addr constant [16 x i8] c"Direction: West\00", align 1
@.str.4 = private unnamed_addr constant [15 x i8] c"Color is green\00", align 1
@.str.5 = private unnamed_addr constant [16 x i8] c"Status: Pending\00", align 1
@.str.6 = private unnamed_addr constant [15 x i8] c"Status: Active\00", align 1
@.str.7 = private unnamed_addr constant [17 x i8] c"Status: Inactive\00", align 1
@.str.8 = private unnamed_addr constant [16 x i8] c"Status: Deleted\00", align 1
@.str.9 = private unnamed_addr constant [16 x i8] c"Not going North\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example20_enum.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/06_advanced_types")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 7, type: !50, scopeLine: 7, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !53)
!6 = !DILocalVariable(name: "dir", scope: !4, file: !3, line: 15, type: !52)
!7 = !DILocalVariable(name: "color", scope: !4, file: !3, line: 16, type: !52)
!8 = !DILocalVariable(name: "status", scope: !4, file: !3, line: 17, type: !52)
!5 = !DILocation(line: 7, column: 9, scope: !4)
!9 = !DILocation(line: 19, column: 10, scope: !4)
!10 = !DILocation(line: 19, column: 8, scope: !4)
!11 = !DILocation(line: 20, column: 8, scope: !4)
!12 = !DILocation(line: 20, column: 3, scope: !4)
!13 = !DILocation(line: 21, column: 20, scope: !4)
!14 = !DILocation(line: 21, column: 19, scope: !4)
!15 = !DILocation(line: 22, column: 20, scope: !4)
!16 = !DILocation(line: 22, column: 19, scope: !4)
!17 = !DILocation(line: 23, column: 20, scope: !4)
!18 = !DILocation(line: 23, column: 19, scope: !4)
!19 = !DILocation(line: 24, column: 20, scope: !4)
!20 = !DILocation(line: 24, column: 19, scope: !4)
!21 = !DILocation(line: 27, column: 12, scope: !4)
!22 = !DILocation(line: 27, column: 10, scope: !4)
!23 = !DILocation(line: 28, column: 6, scope: !4)
!24 = !DILocation(line: 28, column: 14, scope: !4)
!25 = !DILocation(line: 28, column: 12, scope: !4)
!26 = !DILocation(line: 28, column: 3, scope: !4)
!27 = !DILocation(line: 29, column: 13, scope: !4)
!28 = !DILocation(line: 29, column: 12, scope: !4)
!29 = !DILocation(line: 31, column: 13, scope: !4)
!30 = !DILocation(line: 31, column: 11, scope: !4)
!31 = !DILocation(line: 32, column: 8, scope: !4)
!32 = !DILocation(line: 32, column: 3, scope: !4)
!33 = !DILocation(line: 33, column: 23, scope: !4)
!34 = !DILocation(line: 33, column: 22, scope: !4)
!35 = !DILocation(line: 34, column: 23, scope: !4)
!36 = !DILocation(line: 34, column: 22, scope: !4)
!37 = !DILocation(line: 35, column: 23, scope: !4)
!38 = !DILocation(line: 35, column: 22, scope: !4)
!39 = !DILocation(line: 36, column: 23, scope: !4)
!40 = !DILocation(line: 36, column: 22, scope: !4)
!41 = !DILocation(line: 39, column: 10, scope: !4)
!42 = !DILocation(line: 39, column: 8, scope: !4)
!43 = !DILocation(line: 40, column: 6, scope: !4)
!44 = !DILocation(line: 40, column: 13, scope: !4)
!45 = !DILocation(line: 40, column: 11, scope: !4)
!46 = !DILocation(line: 40, column: 3, scope: !4)
!47 = !DILocation(line: 41, column: 13, scope: !4)
!48 = !DILocation(line: 41, column: 12, scope: !4)
!49 = !{null}
!50 = !DISubroutineType(types: !49)
!51 = !{}
!52 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!53 = !{!6, !7, !8}
