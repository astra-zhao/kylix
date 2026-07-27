; Kylix LLVM IR — module: Inheritance
source_filename = "Inheritance.klx"
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
%TShape = type { ptr, ptr }
@TShape_vtable = constant [1 x ptr] [ ptr @TShape_Describe ]
define void @TShape_Describe(ptr %self) {
entry:
  %t0 = getelementptr inbounds [15 x i8], ptr @.str.0, i64 0, i64 0
  %t1 = getelementptr inbounds %TShape, ptr %self, i32 0, i32 1
  %t2 = load ptr, ptr %t1
  %t3 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t3, ptr %t0)
  call ptr @strcat(ptr %t3, ptr %t2)
  %t4 = call i32 @puts(ptr noundef %t3)
  ret void
}

%TRectangle = type { ptr, ptr, i64, i64 }
@TRectangle_vtable = constant [3 x ptr] [ ptr @TShape_Describe, ptr @TRectangle_Area, ptr @TRectangle_Perimeter ]
define i64 @TRectangle_Area(ptr %self) {
entry:
  %result = alloca i64, align 8
  %t5 = getelementptr inbounds %TRectangle, ptr %self, i32 0, i32 2
  %t6 = load i64, ptr %t5
  %t7 = getelementptr inbounds %TRectangle, ptr %self, i32 0, i32 3
  %t8 = load i64, ptr %t7
  %t9 = mul i64 %t6, %t8
  store i64 %t9, ptr %result
  %t10 = load i64, ptr %result
  ret i64 %t10
}

define i64 @TRectangle_Perimeter(ptr %self) {
entry:
  %result = alloca i64, align 8
  %t11 = add i64 0, 2
  %t12 = getelementptr inbounds %TRectangle, ptr %self, i32 0, i32 2
  %t13 = load i64, ptr %t12
  %t14 = getelementptr inbounds %TRectangle, ptr %self, i32 0, i32 3
  %t15 = load i64, ptr %t14
  %t16 = add i64 %t13, %t15
  %t17 = mul i64 %t11, %t16
  store i64 %t17, ptr %result
  %t18 = load i64, ptr %result
  ret i64 %t18
}

%TSquare = type { ptr, ptr, i64, i64 }
@TSquare_vtable = constant [4 x ptr] [ ptr @TShape_Describe, ptr @TRectangle_Area, ptr @TRectangle_Perimeter, ptr @TSquare_SetSide ]
define void @TSquare_SetSide(ptr %self, i64 %n) {
entry:
  %v_n_int = alloca i64, align 8
  store i64 %n, ptr %v_n_int
  %t19 = load i64, ptr %v_n_int
  %t20 = getelementptr inbounds %TSquare, ptr %self, i32 0, i32 2
  store i64 %t19, ptr %t20
  %t21 = load i64, ptr %v_n_int
  %t22 = getelementptr inbounds %TSquare, ptr %self, i32 0, i32 3
  store i64 %t21, ptr %t22
  ret void
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
define i32 @main() !dbg !4 {
entry:
  %t23 = call ptr @malloc(i64 32), !dbg !5
  %t24 = getelementptr inbounds %TRectangle, ptr %t23, i32 0, i32 0, !dbg !5
  store ptr @TRectangle_vtable, ptr %t24, !dbg !5
  %v_rect_str = alloca ptr, align 8, !dbg !6
  store ptr %t23, ptr %v_rect_str, !dbg !6
  %t25 = getelementptr inbounds [5 x i8], ptr @.str.1, i64 0, i64 0, !dbg !7
  %t26 = load ptr, ptr %v_rect_str, !dbg !8
  %t27 = getelementptr inbounds %TRectangle, ptr %t26, i32 0, i32 1, !dbg !8
  store ptr %t25, ptr %t27, !dbg !8
  %t28 = add i64 0, 8, !dbg !9
  %t29 = load ptr, ptr %v_rect_str, !dbg !10
  %t30 = getelementptr inbounds %TRectangle, ptr %t29, i32 0, i32 2, !dbg !10
  store i64 %t28, ptr %t30, !dbg !10
  %t31 = add i64 0, 5, !dbg !11
  %t32 = load ptr, ptr %v_rect_str, !dbg !12
  %t33 = getelementptr inbounds %TRectangle, ptr %t32, i32 0, i32 3, !dbg !12
  store i64 %t31, ptr %t33, !dbg !12
  %t34 = load ptr, ptr %v_rect_str, !dbg !13
  %t35 = getelementptr inbounds %TRectangle, ptr %t34, i32 0, i32 0, !dbg !13
  %t36 = load ptr, ptr %t35, !dbg !13
  %t37 = getelementptr inbounds [3 x ptr], ptr %t36, i32 0, i32 0, !dbg !13
  %t38 = load ptr, ptr %t37, !dbg !13
  call void (ptr) %t38(ptr %t34), !dbg !13
  %t39 = getelementptr inbounds [7 x i8], ptr @.str.2, i64 0, i64 0, !dbg !14
  %t40 = load ptr, ptr %v_rect_str, !dbg !15
  %t41 = getelementptr inbounds %TRectangle, ptr %t40, i32 0, i32 0, !dbg !15
  %t42 = load ptr, ptr %t41, !dbg !15
  %t43 = getelementptr inbounds [3 x ptr], ptr %t42, i32 0, i32 1, !dbg !15
  %t44 = load ptr, ptr %t43, !dbg !15
  %t45 = call i64 (ptr) %t44(ptr %t40), !dbg !15
  %t46 = alloca [24 x i8], align 1, !dbg !16
  %t47 = getelementptr inbounds [24 x i8], ptr %t46, i64 0, i64 0, !dbg !16
  %t48 = getelementptr inbounds [5 x i8], ptr @.str.3, i64 0, i64 0, !dbg !16
  %t49 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t47, i64 24, ptr noundef %t48, i64 %t45), !dbg !16
  %t50 = call ptr @malloc(i64 512), !dbg !17
  call ptr @strcpy(ptr %t50, ptr %t39), !dbg !17
  call ptr @strcat(ptr %t50, ptr %t47), !dbg !17
  %t51 = call i32 @puts(ptr noundef %t50), !dbg !18
  %t52 = getelementptr inbounds [12 x i8], ptr @.str.4, i64 0, i64 0, !dbg !19
  %t53 = load ptr, ptr %v_rect_str, !dbg !20
  %t54 = getelementptr inbounds %TRectangle, ptr %t53, i32 0, i32 0, !dbg !20
  %t55 = load ptr, ptr %t54, !dbg !20
  %t56 = getelementptr inbounds [3 x ptr], ptr %t55, i32 0, i32 2, !dbg !20
  %t57 = load ptr, ptr %t56, !dbg !20
  %t58 = call i64 (ptr) %t57(ptr %t53), !dbg !20
  %t59 = alloca [24 x i8], align 1, !dbg !21
  %t60 = getelementptr inbounds [24 x i8], ptr %t59, i64 0, i64 0, !dbg !21
  %t61 = getelementptr inbounds [5 x i8], ptr @.str.3, i64 0, i64 0, !dbg !21
  %t62 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t60, i64 24, ptr noundef %t61, i64 %t58), !dbg !21
  %t63 = call ptr @malloc(i64 512), !dbg !22
  call ptr @strcpy(ptr %t63, ptr %t52), !dbg !22
  call ptr @strcat(ptr %t63, ptr %t60), !dbg !22
  %t64 = call i32 @puts(ptr noundef %t63), !dbg !23
  %t65 = getelementptr inbounds [4 x i8], ptr @.str.5, i64 0, i64 0, !dbg !24
  %t66 = call i32 @puts(ptr noundef %t65), !dbg !25
  %t67 = call ptr @malloc(i64 32), !dbg !26
  %t68 = getelementptr inbounds %TSquare, ptr %t67, i32 0, i32 0, !dbg !26
  store ptr @TSquare_vtable, ptr %t68, !dbg !26
  %v_sq_str = alloca ptr, align 8, !dbg !27
  store ptr %t67, ptr %v_sq_str, !dbg !27
  %t69 = getelementptr inbounds [4 x i8], ptr @.str.6, i64 0, i64 0, !dbg !28
  %t70 = load ptr, ptr %v_sq_str, !dbg !29
  %t71 = getelementptr inbounds %TSquare, ptr %t70, i32 0, i32 1, !dbg !29
  store ptr %t69, ptr %t71, !dbg !29
  %t72 = add i64 0, 6, !dbg !30
  %t73 = load ptr, ptr %v_sq_str, !dbg !31
  %t74 = getelementptr inbounds %TSquare, ptr %t73, i32 0, i32 0, !dbg !31
  %t75 = load ptr, ptr %t74, !dbg !31
  %t76 = getelementptr inbounds [4 x ptr], ptr %t75, i32 0, i32 3, !dbg !31
  %t77 = load ptr, ptr %t76, !dbg !31
  call void (ptr, i64) %t77(ptr %t73, i64 %t72), !dbg !31
  %t78 = load ptr, ptr %v_sq_str, !dbg !32
  %t79 = getelementptr inbounds %TSquare, ptr %t78, i32 0, i32 0, !dbg !32
  %t80 = load ptr, ptr %t79, !dbg !32
  %t81 = getelementptr inbounds [4 x ptr], ptr %t80, i32 0, i32 0, !dbg !32
  %t82 = load ptr, ptr %t81, !dbg !32
  call void (ptr) %t82(ptr %t78), !dbg !32
  %t83 = getelementptr inbounds [7 x i8], ptr @.str.2, i64 0, i64 0, !dbg !33
  %t84 = load ptr, ptr %v_sq_str, !dbg !34
  %t85 = getelementptr inbounds %TSquare, ptr %t84, i32 0, i32 0, !dbg !34
  %t86 = load ptr, ptr %t85, !dbg !34
  %t87 = getelementptr inbounds [4 x ptr], ptr %t86, i32 0, i32 1, !dbg !34
  %t88 = load ptr, ptr %t87, !dbg !34
  %t89 = call i64 (ptr) %t88(ptr %t84), !dbg !34
  %t90 = alloca [24 x i8], align 1, !dbg !35
  %t91 = getelementptr inbounds [24 x i8], ptr %t90, i64 0, i64 0, !dbg !35
  %t92 = getelementptr inbounds [5 x i8], ptr @.str.3, i64 0, i64 0, !dbg !35
  %t93 = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %t91, i64 24, ptr noundef %t92, i64 %t89), !dbg !35
  %t94 = call ptr @malloc(i64 512), !dbg !36
  call ptr @strcpy(ptr %t94, ptr %t83), !dbg !36
  call ptr @strcat(ptr %t94, ptr %t91), !dbg !36
  %t95 = call i32 @puts(ptr noundef %t94), !dbg !37
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [15 x i8] c"Shape, color: \00", align 1
@.str.1 = private unnamed_addr constant [5 x i8] c"blue\00", align 1
@.str.2 = private unnamed_addr constant [7 x i8] c"Area: \00", align 1
@.str.3 = private unnamed_addr constant [5 x i8] c"%lld\00", align 1
@.str.4 = private unnamed_addr constant [12 x i8] c"Perimeter: \00", align 1
@.str.5 = private unnamed_addr constant [4 x i8] c"---\00", align 1
@.str.6 = private unnamed_addr constant [4 x i8] c"red\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example19_inheritance.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/04_oop")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 8, type: !39, scopeLine: 8, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !42)
!5 = !DILocation(line: 47, column: 25, scope: !4)
!6 = !DILocation(line: 47, column: 3, scope: !4)
!7 = !DILocation(line: 48, column: 17, scope: !4)
!8 = !DILocation(line: 48, column: 15, scope: !4)
!9 = !DILocation(line: 49, column: 17, scope: !4)
!10 = !DILocation(line: 49, column: 15, scope: !4)
!11 = !DILocation(line: 50, column: 18, scope: !4)
!12 = !DILocation(line: 50, column: 16, scope: !4)
!13 = !DILocation(line: 51, column: 16, scope: !4)
!14 = !DILocation(line: 52, column: 11, scope: !4)
!15 = !DILocation(line: 52, column: 40, scope: !4)
!16 = !DILocation(line: 52, column: 30, scope: !4)
!17 = !DILocation(line: 52, column: 20, scope: !4)
!18 = !DILocation(line: 52, column: 10, scope: !4)
!19 = !DILocation(line: 53, column: 11, scope: !4)
!20 = !DILocation(line: 53, column: 50, scope: !4)
!21 = !DILocation(line: 53, column: 35, scope: !4)
!22 = !DILocation(line: 53, column: 25, scope: !4)
!23 = !DILocation(line: 53, column: 10, scope: !4)
!24 = !DILocation(line: 55, column: 11, scope: !4)
!25 = !DILocation(line: 55, column: 10, scope: !4)
!26 = !DILocation(line: 57, column: 20, scope: !4)
!27 = !DILocation(line: 57, column: 3, scope: !4)
!28 = !DILocation(line: 58, column: 15, scope: !4)
!29 = !DILocation(line: 58, column: 13, scope: !4)
!30 = !DILocation(line: 59, column: 14, scope: !4)
!31 = !DILocation(line: 59, column: 13, scope: !4)
!32 = !DILocation(line: 60, column: 14, scope: !4)
!33 = !DILocation(line: 61, column: 11, scope: !4)
!34 = !DILocation(line: 61, column: 38, scope: !4)
!35 = !DILocation(line: 61, column: 30, scope: !4)
!36 = !DILocation(line: 61, column: 20, scope: !4)
!37 = !DILocation(line: 61, column: 10, scope: !4)
!38 = !{null}
!39 = !DISubroutineType(types: !38)
!40 = !{}
!41 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!42 = !{}
