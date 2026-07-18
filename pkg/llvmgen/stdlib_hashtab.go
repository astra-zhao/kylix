package llvmgen

import (
	"fmt"
)

// stdlib_hashtab.go — internal string→string hash table IR runtime.
//
// Used by stdlib modules that need key→value storage (cache, and eventually
// map[K]V / jsonutil). NOT exposed to user Kylix code directly — there is no
// `hashtab` module. Instead, cache's TCache wraps these helpers.
//
// Design: separate chaining with a fixed bucket count (256). Each node is a
// heap-allocated {ptr key, ptr value, ptr next} struct (24 bytes). The table
// itself is a heap-allocated {ptr buckets, i64 size} header (16 bytes) —
// `buckets` points to a [256 x ptr] array of bucket head pointers.
//
//   htab_new()               -> ptr (table header, or null on OOM)
//   htab_put(t, k, v)        -> void (insert or update; copies k)
//   htab_get(t, k)           -> ptr (value, or null if absent)
//   htab_has(t, k)           -> i1
//   htab_del(t, k)           -> void (no-op if absent)
//   htab_size(t)             -> i64
//   htab_clear(t)            -> void (free all nodes; keep buckets)
//   htab_free(t)             -> void (clear + free buckets + header)
//
// Key strings are strdup'd on insert so the caller can free/reuse the input
// buffer. Value strings are stored as-is (caller-managed; cache semantics:
// the latest Put wins, the old value ptr is simply overwritten — leaked until
// htab_clear/free, matching the Go backend's no-GC behavior).

const htabBucketCount = 256

// htabNodeFields: {ptr key, ptr value, ptr next} = 24 bytes.
const htabNodeSize = 24

// emitHashtabBodies emits all hash-table runtime functions, once per module.
// Idempotent via hashtabEmitted.
func (g *Generator) emitHashtabBodies() {
	if g.hashtabEmitted {
		return
	}
	g.hashtabEmitted = true
	g.emitHashtabNew()
	g.emitHashtabHash()
	g.emitHashtabFind() // helper: returns ptr-to-node-or-null
	g.emitHashtabPut()
	g.emitHashtabGet()
	// v5.1.0: htab_get_variant (Variant-box-valued map reads) references the
	// Variant nilbox global, so emit it only when the Variant runtime is in
	// use — keeps cache/string-map modules free of Variant bloat.
	if g.needVariantRuntime {
		g.emitHashtabGetVariant()
	}
	g.emitHashtabHas()
	g.emitHashtabDel()
	g.emitHashtabSize()
	g.emitHashtabClear()
	g.emitHashtabStrdup()
}

// htab_new: ptr @__kylix_htab_new()
//
//	header = malloc(16)
//	buckets = malloc(256 * 8)  ; [256 x ptr]
//	memset(buckets, 0, 256*8)
//	header->buckets = buckets; header->size = 0
//	ret header
func (g *Generator) emitHashtabNew() {
	g.line("define ptr @__kylix_htab_new() {")
	g.line("entry:")
	hdr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", hdr))
	buckets := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", buckets, htabBucketCount*8))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 %d, i1 false)", buckets, htabBucketCount*8))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", buckets, hdr))
	sizePtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", sizePtr, hdr))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", sizePtr))
	g.line(fmt.Sprintf("  ret ptr %s", hdr))
	g.line("}")
	g.line("")
}

// htab_hash: i64 @__kylix_htab_hash(ptr %key) — djb2, masked to bucket range.
//
//	hash = 5381
//	for each byte c: hash = hash*33 + c
//	return hash & 255
func (g *Generator) emitHashtabHash() {
	g.line("define i64 @__kylix_htab_hash(ptr %key) {")
	g.line("entry:")
	hSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", hSlot))
	g.line(fmt.Sprintf("  store i64 5381, ptr %s", hSlot))
	pSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", pSlot))
	g.line(fmt.Sprintf("  store ptr %%key, ptr %s", pSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	curP := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", curP, pSlot))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, curP))
	isZero := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 0", isZero, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isZero, exitLbl, bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	cv := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", cv, c))
	curH := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curH, hSlot))
	h33 := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 33", h33, curH))
	newH := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", newH, h33, cv))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newH, hSlot))
	nextP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", nextP, curP))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nextP, pSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	finalH := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", finalH, hSlot))
	masked := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 255", masked, finalH))
	g.line(fmt.Sprintf("  ret i64 %s", masked))
	g.line("}")
	g.line("")
}

// htab_find: ptr @__kylix_htab_find(ptr %t, ptr %key)
//
//	Returns the node pointer whose key == %key, or null. (Used internally;
//	exposed so htab_get/has share one lookup walk.)
func (g *Generator) emitHashtabFind() {
	g.line("define ptr @__kylix_htab_find(ptr %t, ptr %key) {")
	g.line("entry:")
	buckets := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %%t", buckets))
	idx := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_htab_hash(ptr %%key)", idx))
	bucketPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", bucketPtr, buckets, idx))
	// node lives in an alloca slot (mutated in the loop — SSA can't reassign).
	nodeSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", nodeSlot))
	firstNode := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", firstNode, bucketPtr))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", firstNode, nodeSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	exitFoundLbl := g.label()
	exitMissLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	node := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", node, nodeSlot))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, node))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNull, exitMissLbl, bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	// node->key
	keyPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 0", keyPtr, node))
	nodeKey := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", nodeKey, keyPtr))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %%key)", cmp, nodeKey))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%next", eq, exitFoundLbl))
	g.line("next:")
	nextPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", nextPtr, node))
	nxt := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", nxt, nextPtr))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nxt, nodeSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitFoundLbl))
	g.line(fmt.Sprintf("  ret ptr %s", node))
	g.line(fmt.Sprintf("%s:", exitMissLbl))
	g.line("  ret ptr null")
	g.line("}")
	g.line("")
}

// htab_put: void @__kylix_htab_put(ptr %t, ptr %key, ptr %val)
//
//	If node exists: update node->value = val.
//	Else: new = malloc(24); new->key = strdup(key); new->value = val;
//	      new->next = bucket_head; bucket_head = new; t->size++.
func (g *Generator) emitHashtabPut() {
	g.line("define void @__kylix_htab_put(ptr %t, ptr %key, ptr %val) {")
	g.line("entry:")
	existing := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_find(ptr %%t, ptr %%key)", existing))
	hasExisting := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", hasExisting, existing))
	updateLbl := g.label()
	insertLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasExisting, updateLbl, insertLbl))
	// update path: existing->value (offset 8) = val
	g.line(fmt.Sprintf("%s:", updateLbl))
	valPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", valPtr, existing))
	g.line(fmt.Sprintf("  store ptr %%val, ptr %s", valPtr))
	g.line("  ret void")
	// insert path
	g.line(fmt.Sprintf("%s:", insertLbl))
	newNode := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", newNode, htabNodeSize))
	// strdup(key) into node->key
	keyField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 0", keyField, newNode))
	dup := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %%key)", dup))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", dup, keyField))
	// node->value = val
	valField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", valField, newNode))
	g.line(fmt.Sprintf("  store ptr %%val, ptr %s", valField))
	// buckets = t->buckets; idx = hash(key); bucketPtr = &buckets[idx]
	buckets := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %%t", buckets))
	idx := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_htab_hash(ptr %%key)", idx))
	bucketPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", bucketPtr, buckets, idx))
	oldHead := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", oldHead, bucketPtr))
	// node->next = oldHead
	nextField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", nextField, newNode))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", oldHead, nextField))
	// bucket_head = node
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", newNode, bucketPtr))
	// t->size++
	sizePtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%t, i64 8", sizePtr))
	curSize := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curSize, sizePtr))
	newSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", newSize, curSize))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newSize, sizePtr))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// htab_get: ptr @__kylix_htab_get(ptr %t, ptr %key) — value or null.
func (g *Generator) emitHashtabGet() {
	emptyStr := g.addString("") // @.str.N for "" — returned on miss so callers
	g.line("define ptr @__kylix_htab_get(ptr %t, ptr %key) {")
	g.line("entry:")
	// Compute the empty-string pointer inside the function body (GEP must be
	// in a function, not at module top level).
	emptyPtr := g.ptrTo(emptyStr, 1)
	node := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_find(ptr %%t, ptr %%key)", node))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, node))
	retNullLbl := g.label()
	retValLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNull, retNullLbl, retValLbl))
	g.line(fmt.Sprintf("%s:", retNullLbl))
	g.line(fmt.Sprintf("  ret ptr %s", emptyPtr))
	g.line(fmt.Sprintf("%s:", retValLbl))
	valPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", valPtr, node))
	v := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", v, valPtr))
	g.line(fmt.Sprintf("  ret ptr %s", v))
	g.line("}")
	g.line("")
}

// htab_get_variant: ptr @__kylix_htab_get_variant(ptr %t, ptr %key)
// v5.1.0: like htab_get but the value slot holds a Variant box pointer. On
// miss returns the global nil-box (@__kylix_variant_nilbox, tag=0) so callers
// (Variant map reads, JsonGet*) always receive a valid box — as_* dispatch on
// tag 0 → the typed zero (str→"", int→0, float→0.0, bool→false). Variant
// runtime is required (the nilbox global lives there); htab_find is shared.
func (g *Generator) emitHashtabGetVariant() {
	g.needVariantRuntime = true
	g.line("define ptr @__kylix_htab_get_variant(ptr %t, ptr %key) {")
	g.line("entry:")
	nilPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr @__kylix_variant_nilbox, i32 0, i32 0", nilPtr))
	node := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_find(ptr %%t, ptr %%key)", node))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, node))
	retNilLbl := g.label()
	retValLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNull, retNilLbl, retValLbl))
	g.line(fmt.Sprintf("%s:", retNilLbl))
	g.line(fmt.Sprintf("  ret ptr %s", nilPtr))
	g.line(fmt.Sprintf("%s:", retValLbl))
	valPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", valPtr, node))
	v := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", v, valPtr))
	g.line(fmt.Sprintf("  ret ptr %s", v))
	g.line("}")
	g.line("")
}

// htab_has: i1 @__kylix_htab_has(ptr %t, ptr %key)
func (g *Generator) emitHashtabHas() {
	g.line("define i1 @__kylix_htab_has(ptr %t, ptr %key) {")
	g.line("entry:")
	node := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_find(ptr %%t, ptr %%key)", node))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", r, node))
	g.line(fmt.Sprintf("  ret i1 %s", r))
	g.line("}")
	g.line("")
}

// htab_del: void @__kylix_htab_del(ptr %t, ptr %key)
//
//	Walk the bucket's linked list with a prev pointer; unlink matching node
//	and free it; t->size--. No-op if absent.
func (g *Generator) emitHashtabDel() {
	g.line("define void @__kylix_htab_del(ptr %t, ptr %key) {")
	g.line("entry:")
	buckets := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %%t", buckets))
	idx := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_htab_hash(ptr %%key)", idx))
	bucketPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", bucketPtr, buckets, idx))
	// prev = bucketPtr (pointer-to-pointer), cur = *bucketPtr
	prevSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", prevSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", bucketPtr, prevSlot))
	// cur lives in an alloca slot (mutated across loop iterations — SSA can't
	// reassign a value name).
	curSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", curSlot))
	firstCur := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", firstCur, bucketPtr))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", firstCur, curSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	notFoundLbl := g.label()
	foundLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	cur := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", cur, curSlot))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, cur))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNull, notFoundLbl, bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	keyPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 0", keyPtr, cur))
	curKey := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", curKey, keyPtr))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %%key)", cmp, curKey))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%advance", eq, foundLbl))
	// advance: prev = &cur->next; cur = cur->next
	g.line("advance:")
	nextField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", nextField, cur))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nextField, prevSlot))
	newCur := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", newCur, nextField))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", newCur, curSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	// found: unlink cur via *prev = cur->next; free(cur); size--
	g.line(fmt.Sprintf("%s:", foundLbl))
	curNext := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", curNext, cur))
	nextVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", nextVal, curNext))
	prevPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", prevPtr, prevSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nextVal, prevPtr))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", cur))
	sizePtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%t, i64 8", sizePtr))
	curSize := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curSize, sizePtr))
	newSize := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 1", newSize, curSize))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newSize, sizePtr))
	g.line("  ret void")
	g.line(fmt.Sprintf("%s:", notFoundLbl))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// htab_size: i64 @__kylix_htab_size(ptr %t)
func (g *Generator) emitHashtabSize() {
	g.line("define i64 @__kylix_htab_size(ptr %t) {")
	g.line("entry:")
	sizePtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%t, i64 8", sizePtr))
	s := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", s, sizePtr))
	g.line(fmt.Sprintf("  ret i64 %s", s))
	g.line("}")
	g.line("")
}

// htab_clear: void @__kylix_htab_clear(ptr %t) — free all nodes, reset size.
func (g *Generator) emitHashtabClear() {
	g.line("define void @__kylix_htab_clear(ptr %t) {")
	g.line("entry:")
	buckets := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %%t", buckets))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	done := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %d", done, curI, htabBucketCount))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, exitLbl, bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	bucketPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", bucketPtr, buckets, curI))
	// node lives in an alloca slot (mutated in the inner free loop).
	nodeSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", nodeSlot))
	firstNode := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", firstNode, bucketPtr))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", firstNode, nodeSlot))
	// inner loop: free chain
	innerCond := g.label()
	innerBody := g.label()
	g.line(fmt.Sprintf("  br label %%%s", innerCond))
	g.line(fmt.Sprintf("%s:", innerCond))
	node := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", node, nodeSlot))
	isEmpty := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isEmpty, node))
	g.line(fmt.Sprintf("  br i1 %s, label %%bucket_done, label %%%s", isEmpty, innerBody))
	g.line(fmt.Sprintf("%s:", innerBody))
	nextField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", nextField, node))
	nxt := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", nxt, nextField))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", node))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nxt, nodeSlot))
	g.line(fmt.Sprintf("  br label %%%s", innerCond))
	g.line("bucket_done:")
	g.line(fmt.Sprintf("  store ptr null, ptr %s", bucketPtr))
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	sizePtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%t, i64 8", sizePtr))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", sizePtr))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitHashtabStrdup: ptr @__kylix_htab_strdup(ptr %s) — malloc+strcpy helper.
// (libc has strdup but it's not in our declare list; this avoids adding it.)
func (g *Generator) emitHashtabStrdup() {
	g.line("define ptr @__kylix_htab_strdup(ptr %s) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%s)", ln))
	plus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", plus1, ln))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, plus1))
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %%s)", buf))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}
