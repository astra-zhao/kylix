; Kylix LLVM IR — module: UseModule
source_filename = "UseModule.klx"
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
define i64 @Square(i64 %x) !dbg !4 {
entry:
  %result = alloca i64, align 8
  #dbg_declare(ptr %result, !5, !DIExpression(), !6)
  %v_x_int = alloca i64, align 8, !dbg !6
  store i64 %x, ptr %v_x_int, !dbg !6
  #dbg_declare(ptr %v_x_int, !7, !DIExpression(), !6)
  %t0 = load i64, ptr %v_x_int, !dbg !8
  %t1 = load i64, ptr %v_x_int, !dbg !9
  %t2 = mul i64 %t0, %t1, !dbg !10
  store i64 %t2, ptr %result, !dbg !11
  %t3 = load i64, ptr %result, !dbg !6
  ret i64 %t3, !dbg !6
}

define i64 @Cube(i64 %x) !dbg !12 {
entry:
  %result = alloca i64, align 8
  #dbg_declare(ptr %result, !13, !DIExpression(), !14)
  %v_x_int = alloca i64, align 8, !dbg !14
  store i64 %x, ptr %v_x_int, !dbg !14
  #dbg_declare(ptr %v_x_int, !15, !DIExpression(), !14)
  %t4 = load i64, ptr %v_x_int, !dbg !16
  %t5 = load i64, ptr %v_x_int, !dbg !17
  %t6 = mul i64 %t4, %t5, !dbg !18
  %t7 = load i64, ptr %v_x_int, !dbg !19
  %t8 = mul i64 %t6, %t7, !dbg !20
  store i64 %t8, ptr %result, !dbg !21
  %t9 = load i64, ptr %result, !dbg !14
  ret i64 %t9, !dbg !14
}

define i1 @IsEven(i64 %n) !dbg !22 {
entry:
  %result = alloca i1, align 8
  #dbg_declare(ptr %result, !23, !DIExpression(), !24)
  %v_n_int = alloca i64, align 8, !dbg !24
  store i64 %n, ptr %v_n_int, !dbg !24
  #dbg_declare(ptr %v_n_int, !25, !DIExpression(), !24)
  %t10 = load i64, ptr %v_n_int, !dbg !26
  %t11 = add i64 0, 2, !dbg !27
  %t12 = srem i64 %t10, %t11, !dbg !28
  %t13 = add i64 0, 0, !dbg !29
  %t14 = icmp eq i64 %t12, %t13, !dbg !30
  store i1 %t14, ptr %result, !dbg !31
  %t15 = load i1, ptr %result, !dbg !24
  ret i1 %t15, !dbg !24
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
define i32 @main() !dbg !32 {
entry:
  %t16 = call ptr @malloc(i64 512), !dbg !33
  store i8 0, ptr %t16, !dbg !33
  %t17 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !33
  %t18 = getelementptr inbounds [14 x i8], ptr @.str.1, i64 0, i64 0, !dbg !34
  %t19 = call ptr @strcat(ptr %t16, ptr %t18), !dbg !33
  %t20 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !33
  %t21 = call ptr @strcat(ptr %t16, ptr %t20), !dbg !33
  %t22 = add i64 0, 5, !dbg !35
  %t23 = call i64 @Square(i64 %t22), !dbg !36
  %t24 = call i64 @strlen(ptr %t16), !dbg !33
  %t25 = getelementptr inbounds i8, ptr %t16, i64 %t24, !dbg !33
  %t26 = sub i64 512, %t24, !dbg !33
  %t27 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t25, i64 %t26, ptr %t17, i64 %t23), !dbg !33
  %t28 = call i32 @puts(ptr noundef %t16), !dbg !33
  %t29 = call ptr @malloc(i64 512), !dbg !37
  store i8 0, ptr %t29, !dbg !37
  %t30 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !37
  %t31 = getelementptr inbounds [12 x i8], ptr @.str.3, i64 0, i64 0, !dbg !38
  %t32 = call ptr @strcat(ptr %t29, ptr %t31), !dbg !37
  %t33 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !37
  %t34 = call ptr @strcat(ptr %t29, ptr %t33), !dbg !37
  %t35 = add i64 0, 3, !dbg !39
  %t36 = call i64 @Cube(i64 %t35), !dbg !40
  %t37 = call i64 @strlen(ptr %t29), !dbg !37
  %t38 = getelementptr inbounds i8, ptr %t29, i64 %t37, !dbg !37
  %t39 = sub i64 512, %t37, !dbg !37
  %t40 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t38, i64 %t39, ptr %t30, i64 %t36), !dbg !37
  %t41 = call i32 @puts(ptr noundef %t29), !dbg !37
  %t42 = call ptr @malloc(i64 512), !dbg !41
  store i8 0, ptr %t42, !dbg !41
  %t44 = getelementptr inbounds [12 x i8], ptr @.str.4, i64 0, i64 0, !dbg !42
  %t45 = call ptr @strcat(ptr %t42, ptr %t44), !dbg !41
  %t46 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !41
  %t47 = call ptr @strcat(ptr %t42, ptr %t46), !dbg !41
  %t48 = add i64 0, 4, !dbg !43
  %t49 = call i1 @IsEven(i64 %t48), !dbg !44
  %t50 = getelementptr inbounds [5 x i8], ptr @.str.5, i64 0, i64 0, !dbg !41
  %t51 = getelementptr inbounds [6 x i8], ptr @.str.6, i64 0, i64 0, !dbg !41
  %t52 = select i1 %t49, ptr %t50, ptr %t51, !dbg !41
  %t53 = call ptr @strcat(ptr %t42, ptr %t52), !dbg !41
  %t54 = call i32 @puts(ptr noundef %t42), !dbg !41
  %t55 = call ptr @malloc(i64 512), !dbg !45
  store i8 0, ptr %t55, !dbg !45
  %t57 = getelementptr inbounds [12 x i8], ptr @.str.7, i64 0, i64 0, !dbg !46
  %t58 = call ptr @strcat(ptr %t55, ptr %t57), !dbg !45
  %t59 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !45
  %t60 = call ptr @strcat(ptr %t55, ptr %t59), !dbg !45
  %t61 = add i64 0, 7, !dbg !47
  %t62 = call i1 @IsEven(i64 %t61), !dbg !48
  %t63 = getelementptr inbounds [5 x i8], ptr @.str.5, i64 0, i64 0, !dbg !45
  %t64 = getelementptr inbounds [6 x i8], ptr @.str.6, i64 0, i64 0, !dbg !45
  %t65 = select i1 %t62, ptr %t63, ptr %t64, !dbg !45
  %t66 = call ptr @strcat(ptr %t55, ptr %t65), !dbg !45
  %t67 = call i32 @puts(ptr noundef %t55), !dbg !45
  %t68 = call ptr @malloc(i64 512), !dbg !49
  store i8 0, ptr %t68, !dbg !49
  %t69 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !49
  %t70 = getelementptr inbounds [15 x i8], ptr @.str.8, i64 0, i64 0, !dbg !50
  %t71 = call ptr @strcat(ptr %t68, ptr %t70), !dbg !49
  %t72 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !49
  %t73 = call ptr @strcat(ptr %t68, ptr %t72), !dbg !49
  %t74 = add i64 0, 10, !dbg !51
  %t75 = call i64 @Square(i64 %t74), !dbg !52
  %t76 = call i64 @strlen(ptr %t68), !dbg !49
  %t77 = getelementptr inbounds i8, ptr %t68, i64 %t76, !dbg !49
  %t78 = sub i64 512, %t76, !dbg !49
  %t79 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t77, i64 %t78, ptr %t69, i64 %t75), !dbg !49
  %t80 = call i32 @puts(ptr noundef %t68), !dbg !49
  %t81 = call ptr @malloc(i64 512), !dbg !53
  store i8 0, ptr %t81, !dbg !53
  %t82 = getelementptr inbounds [4 x i8], ptr @.str.0, i64 0, i64 0, !dbg !53
  %t83 = getelementptr inbounds [12 x i8], ptr @.str.9, i64 0, i64 0, !dbg !54
  %t84 = call ptr @strcat(ptr %t81, ptr %t83), !dbg !53
  %t85 = getelementptr inbounds [2 x i8], ptr @.str.2, i64 0, i64 0, !dbg !53
  %t86 = call ptr @strcat(ptr %t81, ptr %t85), !dbg !53
  %t87 = add i64 0, 4, !dbg !55
  %t88 = call i64 @Cube(i64 %t87), !dbg !56
  %t89 = call i64 @strlen(ptr %t81), !dbg !53
  %t90 = getelementptr inbounds i8, ptr %t81, i64 %t89, !dbg !53
  %t91 = sub i64 512, %t89, !dbg !53
  %t92 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %t90, i64 %t91, ptr %t82, i64 %t88), !dbg !53
  %t93 = call i32 @puts(ptr noundef %t81), !dbg !53
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [4 x i8] c"%ld\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"Square of 5: \00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.3 = private unnamed_addr constant [12 x i8] c"Cube of 3: \00", align 1
@.str.4 = private unnamed_addr constant [12 x i8] c"Is 4 even? \00", align 1
@.str.5 = private unnamed_addr constant [5 x i8] c"true\00", align 1
@.str.6 = private unnamed_addr constant [6 x i8] c"false\00", align 1
@.str.7 = private unnamed_addr constant [12 x i8] c"Is 7 even? \00", align 1
@.str.8 = private unnamed_addr constant [15 x i8] c"Square of 10: \00", align 1
@.str.9 = private unnamed_addr constant [12 x i8] c"Cube of 4: \00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example33_use_module.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/11_modules")
!4 = distinct !DISubprogram(name: "Square", scope: !3, file: !3, line: 11, type: !58, scopeLine: 11, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !61)
!12 = distinct !DISubprogram(name: "Cube", scope: !3, file: !3, line: 16, type: !58, scopeLine: 16, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !62)
!22 = distinct !DISubprogram(name: "IsEven", scope: !3, file: !3, line: 21, type: !58, scopeLine: 21, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !63)
!32 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 1, type: !58, scopeLine: 1, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !64)
!5 = !DILocalVariable(name: "result", scope: !4, file: !3, line: 11, type: !60)
!7 = !DILocalVariable(name: "x", scope: !4, file: !3, line: 11, type: !60)
!13 = !DILocalVariable(name: "result", scope: !12, file: !3, line: 16, type: !60)
!15 = !DILocalVariable(name: "x", scope: !12, file: !3, line: 16, type: !60)
!23 = !DILocalVariable(name: "result", scope: !22, file: !3, line: 21, type: !60)
!25 = !DILocalVariable(name: "n", scope: !22, file: !3, line: 21, type: !60)
!6 = !DILocation(line: 11, column: 1, scope: !4)
!8 = !DILocation(line: 13, column: 13, scope: !4)
!9 = !DILocation(line: 13, column: 17, scope: !4)
!10 = !DILocation(line: 13, column: 15, scope: !4)
!11 = !DILocation(line: 13, column: 11, scope: !4)
!14 = !DILocation(line: 16, column: 1, scope: !12)
!16 = !DILocation(line: 18, column: 13, scope: !12)
!17 = !DILocation(line: 18, column: 17, scope: !12)
!18 = !DILocation(line: 18, column: 15, scope: !12)
!19 = !DILocation(line: 18, column: 21, scope: !12)
!20 = !DILocation(line: 18, column: 19, scope: !12)
!21 = !DILocation(line: 18, column: 11, scope: !12)
!24 = !DILocation(line: 21, column: 1, scope: !22)
!26 = !DILocation(line: 23, column: 14, scope: !22)
!27 = !DILocation(line: 23, column: 20, scope: !22)
!28 = !DILocation(line: 23, column: 16, scope: !22)
!29 = !DILocation(line: 23, column: 25, scope: !22)
!30 = !DILocation(line: 23, column: 23, scope: !22)
!31 = !DILocation(line: 23, column: 11, scope: !22)
!33 = !DILocation(line: 5, column: 10, scope: !32)
!34 = !DILocation(line: 5, column: 11, scope: !32)
!35 = !DILocation(line: 5, column: 35, scope: !32)
!36 = !DILocation(line: 5, column: 34, scope: !32)
!37 = !DILocation(line: 6, column: 10, scope: !32)
!38 = !DILocation(line: 6, column: 11, scope: !32)
!39 = !DILocation(line: 6, column: 31, scope: !32)
!40 = !DILocation(line: 6, column: 30, scope: !32)
!41 = !DILocation(line: 8, column: 10, scope: !32)
!42 = !DILocation(line: 8, column: 11, scope: !32)
!43 = !DILocation(line: 8, column: 33, scope: !32)
!44 = !DILocation(line: 8, column: 32, scope: !32)
!45 = !DILocation(line: 9, column: 10, scope: !32)
!46 = !DILocation(line: 9, column: 11, scope: !32)
!47 = !DILocation(line: 9, column: 33, scope: !32)
!48 = !DILocation(line: 9, column: 32, scope: !32)
!49 = !DILocation(line: 11, column: 10, scope: !32)
!50 = !DILocation(line: 11, column: 11, scope: !32)
!51 = !DILocation(line: 11, column: 36, scope: !32)
!52 = !DILocation(line: 11, column: 35, scope: !32)
!53 = !DILocation(line: 12, column: 10, scope: !32)
!54 = !DILocation(line: 12, column: 11, scope: !32)
!55 = !DILocation(line: 12, column: 31, scope: !32)
!56 = !DILocation(line: 12, column: 30, scope: !32)
!57 = !{null}
!58 = !DISubroutineType(types: !57)
!59 = !{}
!60 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!61 = !{!5, !7}
!62 = !{!13, !15}
!63 = !{!23, !25}
!64 = !{}
