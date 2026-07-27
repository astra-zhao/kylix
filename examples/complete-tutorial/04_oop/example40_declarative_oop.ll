; Kylix LLVM IR — module: DeclarativeOOP
source_filename = "DeclarativeOOP.klx"
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
%TAnimal = type { ptr, ptr, ptr }
@TAnimal_vtable = constant [1 x ptr] [ ptr @TAnimal_Speak ]
define void @TAnimal_Speak(ptr %self) {
entry:
  %t0 = getelementptr inbounds %TAnimal, ptr %self, i32 0, i32 1
  %t1 = load ptr, ptr %t0
  %t2 = getelementptr inbounds [7 x i8], ptr @.str.0, i64 0, i64 0
  %t3 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t3, ptr %t1)
  call ptr @strcat(ptr %t3, ptr %t2)
  %t4 = getelementptr inbounds %TAnimal, ptr %self, i32 0, i32 2
  %t5 = load ptr, ptr %t4
  %t6 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t6, ptr %t3)
  call ptr @strcat(ptr %t6, ptr %t5)
  %t7 = call i32 @puts(ptr noundef %t6)
  ret void
}

%TDog = type { ptr, ptr, ptr, ptr }
@TDog_vtable = constant [2 x ptr] [ ptr @TAnimal_Speak, ptr @TDog_Describe ]
define void @TDog_Describe(ptr %self) {
entry:
  %t8 = getelementptr inbounds [6 x i8], ptr @.str.1, i64 0, i64 0
  %t9 = getelementptr inbounds %TDog, ptr %self, i32 0, i32 1
  %t10 = load ptr, ptr %t9
  %t11 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t11, ptr %t8)
  call ptr @strcat(ptr %t11, ptr %t10)
  %t12 = getelementptr inbounds [3 x i8], ptr @.str.2, i64 0, i64 0
  %t13 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t13, ptr %t11)
  call ptr @strcat(ptr %t13, ptr %t12)
  %t14 = getelementptr inbounds %TDog, ptr %self, i32 0, i32 3
  %t15 = load ptr, ptr %t14
  %t16 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t16, ptr %t13)
  call ptr @strcat(ptr %t16, ptr %t15)
  %t17 = getelementptr inbounds [2 x i8], ptr @.str.3, i64 0, i64 0
  %t18 = call ptr @malloc(i64 512)
  call ptr @strcpy(ptr %t18, ptr %t16)
  call ptr @strcat(ptr %t18, ptr %t17)
  %t19 = call i32 @puts(ptr noundef %t18)
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
  %v_cat = alloca ptr, align 8, !dbg !5
  store ptr null, ptr %v_cat, !dbg !5
  %v_dog = alloca ptr, align 8, !dbg !5
  store ptr null, ptr %v_dog, !dbg !5
  %t20 = call ptr @malloc(i64 24), !dbg !6
  %t21 = getelementptr inbounds %TAnimal, ptr %t20, i32 0, i32 0, !dbg !6
  store ptr @TAnimal_vtable, ptr %t21, !dbg !6
  store ptr %t20, ptr %v_cat, !dbg !7
  %t22 = getelementptr inbounds [9 x i8], ptr @.str.4, i64 0, i64 0, !dbg !8
  %t23 = load ptr, ptr %v_cat, !dbg !9
  %t24 = getelementptr inbounds %TAnimal, ptr %t23, i32 0, i32 1, !dbg !9
  store ptr %t22, ptr %t24, !dbg !9
  %t25 = getelementptr inbounds [5 x i8], ptr @.str.5, i64 0, i64 0, !dbg !10
  %t26 = load ptr, ptr %v_cat, !dbg !11
  %t27 = getelementptr inbounds %TAnimal, ptr %t26, i32 0, i32 2, !dbg !11
  store ptr %t25, ptr %t27, !dbg !11
  %t28 = load ptr, ptr %v_cat, !dbg !12
  %t29 = getelementptr inbounds %TAnimal, ptr %t28, i32 0, i32 0, !dbg !12
  %t30 = load ptr, ptr %t29, !dbg !12
  %t31 = getelementptr inbounds [1 x ptr], ptr %t30, i32 0, i32 0, !dbg !12
  %t32 = load ptr, ptr %t31, !dbg !12
  call void (ptr) %t32(ptr %t28), !dbg !12
  %t33 = call ptr @malloc(i64 32), !dbg !13
  %t34 = getelementptr inbounds %TDog, ptr %t33, i32 0, i32 0, !dbg !13
  store ptr @TDog_vtable, ptr %t34, !dbg !13
  store ptr %t33, ptr %v_dog, !dbg !14
  %t35 = getelementptr inbounds [4 x i8], ptr @.str.6, i64 0, i64 0, !dbg !15
  %t36 = load ptr, ptr %v_dog, !dbg !16
  %t37 = getelementptr inbounds %TDog, ptr %t36, i32 0, i32 1, !dbg !16
  store ptr %t35, ptr %t37, !dbg !16
  %t38 = getelementptr inbounds [5 x i8], ptr @.str.7, i64 0, i64 0, !dbg !17
  %t39 = load ptr, ptr %v_dog, !dbg !18
  %t40 = getelementptr inbounds %TDog, ptr %t39, i32 0, i32 2, !dbg !18
  store ptr %t38, ptr %t40, !dbg !18
  %t41 = getelementptr inbounds [9 x i8], ptr @.str.8, i64 0, i64 0, !dbg !19
  %t42 = load ptr, ptr %v_dog, !dbg !20
  %t43 = getelementptr inbounds %TDog, ptr %t42, i32 0, i32 3, !dbg !20
  store ptr %t41, ptr %t43, !dbg !20
  %t44 = load ptr, ptr %v_dog, !dbg !21
  %t45 = getelementptr inbounds %TDog, ptr %t44, i32 0, i32 0, !dbg !21
  %t46 = load ptr, ptr %t45, !dbg !21
  %t47 = getelementptr inbounds [2 x ptr], ptr %t46, i32 0, i32 0, !dbg !21
  %t48 = load ptr, ptr %t47, !dbg !21
  call void (ptr) %t48(ptr %t44), !dbg !21
  %t49 = load ptr, ptr %v_dog, !dbg !22
  %t50 = getelementptr inbounds %TDog, ptr %t49, i32 0, i32 0, !dbg !22
  %t51 = load ptr, ptr %t50, !dbg !22
  %t52 = getelementptr inbounds [2 x ptr], ptr %t51, i32 0, i32 1, !dbg !22
  %t53 = load ptr, ptr %t52, !dbg !22
  call void (ptr) %t53(ptr %t49), !dbg !22
  ret i32 0
}

; ===== String constants =====
@.str.0 = private unnamed_addr constant [7 x i8] c" says \00", align 1
@.str.1 = private unnamed_addr constant [6 x i8] c"Dog: \00", align 1
@.str.2 = private unnamed_addr constant [3 x i8] c" (\00", align 1
@.str.3 = private unnamed_addr constant [2 x i8] c")\00", align 1
@.str.4 = private unnamed_addr constant [9 x i8] c"Whiskers\00", align 1
@.str.5 = private unnamed_addr constant [5 x i8] c"Meow\00", align 1
@.str.6 = private unnamed_addr constant [4 x i8] c"Rex\00", align 1
@.str.7 = private unnamed_addr constant [5 x i8] c"Woof\00", align 1
@.str.8 = private unnamed_addr constant [9 x i8] c"Labrador\00", align 1

; ===== DWARF debug info (kylix -g) =====
!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "kylix", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 7, !"Dwarf Version", i32 4}
!2 = !{i32 2, !"Debug Info Version", i32 3}
!3 = !DIFile(filename: "example40_declarative_oop.klx", directory: "/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/04_oop")
!4 = distinct !DISubprogram(name: "main", scope: !3, file: !3, line: 10, type: !24, scopeLine: 10, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: !27)
!5 = !DILocation(line: 10, column: 9, scope: !4)
!6 = !DILocation(line: 42, column: 17, scope: !4)
!7 = !DILocation(line: 42, column: 8, scope: !4)
!8 = !DILocation(line: 43, column: 15, scope: !4)
!9 = !DILocation(line: 43, column: 13, scope: !4)
!10 = !DILocation(line: 44, column: 16, scope: !4)
!11 = !DILocation(line: 44, column: 14, scope: !4)
!12 = !DILocation(line: 45, column: 12, scope: !4)
!13 = !DILocation(line: 47, column: 14, scope: !4)
!14 = !DILocation(line: 47, column: 8, scope: !4)
!15 = !DILocation(line: 48, column: 15, scope: !4)
!16 = !DILocation(line: 48, column: 13, scope: !4)
!17 = !DILocation(line: 49, column: 16, scope: !4)
!18 = !DILocation(line: 49, column: 14, scope: !4)
!19 = !DILocation(line: 50, column: 16, scope: !4)
!20 = !DILocation(line: 50, column: 14, scope: !4)
!21 = !DILocation(line: 51, column: 12, scope: !4)
!22 = !DILocation(line: 52, column: 15, scope: !4)
!23 = !{null}
!24 = !DISubroutineType(types: !23)
!25 = !{}
!26 = !DIBasicType(name: "int64", size: 64, encoding: DW_ATE_signed)
!27 = !{}
