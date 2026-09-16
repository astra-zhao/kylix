#!/usr/bin/env python3
"""extract_stdlib_ir.py — v0.6.9 P3: bake host stdlib IR into the bootstrap emitter.

Instead of hand-porting pkg/llvmgen/stdlib_*.go (15.5k lines of Go) to
src/llvmgen.klx, we let the HOST compiler emit each stdlib module's final IR,
extract the module-level `define @__kylix_<seg>_*` bodies (plus their string
constants), and bake them into src/stdlib_ir.klx as Kylix string data. The
bootstrap emitter only implements the call-site dispatch: evaluate args, emit
`call @__kylix_<seg>_<Fn>`, mark the segment used; used segments' IR lines are
appended verbatim at module end. Unreferenced segments stay out of the .o, so
their external deps (sqlite3/curl/…) never reach the linker — same
conditionality as the host's enqueueStdlib.

Sources: 51-tutorial stdlib examples + /tmp/stdir_cover/cover.ll (a coverage
program calling every stdlib function the host dispatch implements).

Usage: python3 scripts/extract_stdlib_ir.py > src/stdlib_ir.klx
"""
import re, os, sys

# Sources: scripts/cover.klx (coverage program calling every stdlib function the
# host dispatch implements) + tutorial IRs. All .ll files are generated into
# /tmp/stdir_cover/ by scripts/rebake_stdlib_ir.sh (host `build --backend=llvm
# -o <dir>/<name>.ll`); only compilation is required — the .ll text is what the
# extractor parses.
TUT = '/Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/'
OUT = '/tmp/stdir_cover/'
FILES = [OUT + 'cover.ll'] + [OUT + p for p in [
    'example36_sysutil.ll',
    'example37_jsonutil.ll',
    'example38_datetime.ll',
    'example39_regex.ll',
    'example48_phase6_net_crypto_encoding.ll',
    'example50_jwt_auth.ll',
    'example52_database.ll',
    'example53_cache.ll',
    'example54_http.ll',
    'example55_websocket.ll',
    # boot segment source: BootRun/read_headers/parse_request/route_lookup/
    # path_match/serve_static/BootText/BootHTML/error pages/session/csrf/
    # multipart defines (v0.9.0 P2). Program-specific handler_N wrappers and
    # route-table globals are excluded below — the bootstrap emitter emits
    # those itself (annotation auto-wiring).
    'example60_web_framework.ll',
    # boot route-registration coverage: [Put]/[Delete]/[Authenticated]/proc
    # handler verbs example60 does not exercise (scripts/cover_boot.klx).
    'cover_boot.ll',
]]

# Segment order: 'runtime' first — it is emitted unconditionally (tiny, and
# referenced by the others). 'boot' last (v0.9.0 P2): the KylixBoot server
# defines (BootRun/read_headers/parse_request/... from example60.ll).
SEGMENTS = ['runtime', 'sysutil', 'regex', 'datetime', 'encoding', 'net',
            'cache', 'crypto', 'db', 'json', 'jwt', 'httpclient', 'websocket',
            'boot']

# Symbols extracted into the 'runtime' segment. Includes the variant helpers
# and htab_get_variant/htab_keys that llvmgen.klx lacks — this upgrades the
# bootstrap Variant runtime to host parity for free. is_subtype is NOT here:
# it is wired to the host's exception-table class RTTI (@__kylix_exctab), while
# the bootstrap emits its own RTTI via EmitClassRuntime.
RUNTIME_SYMS = set([
    'now_ms',
    'htab_get_variant', 'htab_keys',
    'variant_add', 'variant_as_bool', 'variant_as_double', 'variant_as_int',
    'variant_as_str', 'variant_box_bool', 'variant_box_float', 'variant_box_int',
    'variant_box_map', 'variant_box_str', 'variant_compare', 'variant_div',
    'variant_idiv', 'variant_map_get', 'variant_mod', 'variant_mul',
    'variant_print', 'variant_println', 'variant_sub',
])
# Unprefixed helpers that belong to their caller's segment.
SEG_EXTRA = {
    'httpclient': ['http_write_cb'],
    'websocket': ['ws_b64', 'ws_buildframe', 'ws_rand', 'ws_readheaders',
                  'ws_recvn', 'ws_sendall', 'ws_sha1'],
}
SEG_EXTRA_SYMS = set('@__kylix_' + s for v in SEG_EXTRA.values() for s in v)


def sym_seg(sym):
    s = sym[len('@__kylix_'):]
    if s in RUNTIME_SYMS:
        return 'runtime'
    if sym in SEG_EXTRA_SYMS:
        for seg, names in SEG_EXTRA.items():
            if s in names:
                return seg
    # __kylix_boot_handler_N wrappers are program-specific — the bootstrap
    # emitter generates them itself (annotation auto-wiring).
    if re.match(r'^boot_handler_\d+$', s):
        return None
    # The boot arena helpers are shared by the baked boot defines AND the
    # bootstrap's own TResponse fluent methods; keep them out of 'runtime'
    # (the unconditional segment) so non-server programs don't get the 1MiB
    # arena bss.
    if s in ('arena_alloc', 'arena_reset'):
        return 'boot'
    for seg in SEGMENTS[1:]:
        if s.startswith(seg + '_'):
            return seg
    return None


def parse_ll(path):
    """Return {sym: [define lines]}, {strname: line}, [declare lines],
    {globname: line}, {typename: line}."""
    bodies, strs, decls, globs, types = {}, {}, [], {}, {}
    lines = open(path).read().split('\n')
    i = 0
    while i < len(lines):
        ln = lines[i]
        m = re.match(r'^define [^@]*@(__kylix_[A-Za-z0-9_]+)\(', ln)
        if m:
            sym = '@' + m.group(1)
            body = []
            while i < len(lines):
                body.append(lines[i])
                if lines[i].startswith('}'):
                    break
                i += 1
            bodies[sym] = body
            i += 1
            continue
        m = re.match(r'^(@\.str\.\d+) = .*$', ln)
        if m:
            strs[m.group(1)] = ln
        m = re.match(r'^@(__kylix_[A-Za-z0-9_]+) = .*$', ln)
        if m:
            globs[m.group(1)] = ln
        # Named IR types referenced by baked bodies. Only %__ret_* (method
        # multi-return aggregates, e.g. %__ret_BootRequest_File used by the
        # baked req_file/req_savefile defines) go into a segment; other named
        # types are bootstrap-owned (%__kylix_edge) or program-specific
        # (%Exception, %TWebController) and must NOT be baked.
        m = re.match(r'^(%__ret_[A-Za-z0-9_]+) = .*$', ln)
        if m:
            types[m.group(1)] = ln
        if ln.startswith('declare '):
            decls.append(ln)
        i += 1
    return bodies, strs, decls, globs, types


def main():
    # Bodies must stay bound to the file they came from: the same @.str.N name
    # carries DIFFERENT content across .ll files (each file numbers its own
    # constant pool), so name-keyed cross-file lookup would paste the wrong
    # string into a body (this silently corrupted HexEncode once).
    file_data = []
    all_decls, all_globs, all_types = [], {}, {}
    for f in FILES:
        if not os.path.exists(f):
            sys.exit('missing source: ' + f)
        b, s, d, gb, ty = parse_ll(f)
        file_data.append((f, b, s))
        for ln in d:
            if ln not in all_decls:
                all_decls.append(ln)
        all_globs.update(gb)
        all_types.update(ty)

    seg_bodies = {seg: [] for seg in SEGMENTS}
    seg_kstrs = {seg: [] for seg in SEGMENTS}  # kstr constant lines, kept OFF
    # the body stream (a mid-define kstr line is an opcode error for llc)
    kstr_of_content = {}   # constant-pool line body (name stripped) → kstr name
    kstr_counter = [0]
    sigs = []

    seen_syms = set()
    for fname, bodies, strs in file_data:
        for sym in sorted(bodies):
            seg = sym_seg(sym)
            if seg is None:
                print('; dropped (no segment): ' + sym, file=sys.stderr)
                continue
            if sym in seen_syms:
                continue  # same define emitted in every source file — keep first
            seen_syms.add(sym)
            body = bodies[sym]
            # rewrite @.str.N → @kstr.<seg>.M, deduped by CONTENT so the same
            # string shared across files maps to one kstr constant
            def sub_str(m):
                sn = m.group(0)
                line = strs.get(sn)
                if line is None:
                    sys.exit('str %s not found in %s (%s)' % (sn, seg, fname))
                content = seg + '\x00' + line.split(' = ', 1)[1]
                # dedup per (segment, content): a cross-segment shared kstr
                # would dangle when the other segment isn't emitted
                kn = kstr_of_content.get(content)
                if kn is None:
                    kn = '@kstr.%s.%d' % (seg, kstr_counter[0])
                    kstr_counter[0] += 1
                    kstr_of_content[content] = kn
                    seg_kstrs[seg].append(line.replace(sn, kn, 1))
                return kn
            for ln in body:
                seg_bodies[seg].append(re.sub(r'@\.str\.\d+', sub_str, ln))
            # signature for the dispatch table (define line); ret may be a
            # struct type like `{ ptr, i64 }`
            d = body[0]
            m = re.match(r'^define\s+(.+?)\s+@[\w.]+\((.*)\)\s*\{$', d)
            if not m:
                sys.exit('cannot parse define line: ' + d)
            ret, params = m.group(1), m.group(2)
            ptypes = []
            for p in params.split(','):
                p = p.strip()
                if not p:
                    continue
                ptypes.append(p.split()[0])  # drop %name
            sigs.append((sym, seg, ret, ptypes))

    # Per-segment text with @.str.N rewritten to segment-local @kstr.<seg>.N.
    # stdlib bodies call libc/OpenSSL/curl/sqlite helpers — carry every
    # declare from the source IRs into the runtime segment, EXCEPT the ones
    # llvmgen.klx's EmitRuntimeDecls already declares (LLVM rejects duplicate
    # declares, even with identical signatures). Keep in sync with that list.
    # Mirrors TLLVMGenerator.EmitRuntimeDecls (v0.6.9 P3 full host parity).
    BOOTSTRAP_DECLARED = {
        'printf', 'puts', 'malloc', 'free', 'strlen', 'strcpy', 'strcat',
        'strcmp', 'memcpy', 'memmove', 'atoll', 'snprintf', 'strtod', 'fabs',
        'strchr', 'strstr', 'strncmp', 'gettimeofday', 'setjmp', '_setjmp',
        '_longjmp', 'longjmp', 'exit', 'fopen', 'fclose', 'fread', 'fwrite',
        'fputs', 'fseek', 'ftell', 'access', 'mkdir', 'remove', 'getcwd',
        'chdir', 'getenv', 'opendir', 'closedir',
        'socket', 'connect', 'bind', 'listen', 'accept', 'send', 'recv',
        'close', 'arc4random_buf', 'getrandom', 'setsockopt', 'inet_pton',
        'SHA256', 'MD5', 'strncpy', 'EVP_CIPHER_CTX_new', 'EVP_CIPHER_CTX_free',
        'EVP_aes_256_cbc', 'EVP_EncryptInit_ex', 'EVP_EncryptUpdate',
        'EVP_EncryptFinal_ex', 'EVP_DecryptInit_ex', 'EVP_DecryptUpdate',
        'EVP_DecryptFinal_ex', 'EVP_CIPHER_CTX_block_size', 'RAND_bytes',
        'EVP_sha256', 'PKCS5_PBKDF2_HMAC', 'sscanf',
        'sqlite3_open', 'sqlite3_close', 'sqlite3_prepare_v2',
        'sqlite3_bind_text', 'sqlite3_bind_int64', 'sqlite3_step',
        'sqlite3_column_text', 'sqlite3_column_count', 'sqlite3_column_name',
        'sqlite3_column_type', 'sqlite3_column_int64', 'sqlite3_column_double',
        'sqlite3_finalize', 'sqlite3_changes',
        'curl_easy_init', 'curl_easy_setopt', 'curl_easy_perform',
        'curl_easy_cleanup', 'curl_slist_append', 'curl_slist_free_all',
        'regcomp', 'regexec', 'regfree',
        'time', 'localtime', 'localtime_r', 'localtime_s', 'mktime', 'strftime',
        'calloc', 'write', 'RAND_bytes',
        'llvm.memset.p0.i64', 'llvm.memcpy.p0.p0.i64',
    }
    # module-level @__kylix_* globals referenced by the extracted bodies.
    # Exc/jmpbuf/args/emptystr/datetime-arena are bootstrap-owned; the boot
    # routes table is emitted by the bootstrap's auto-wiring (see glob_seg).
    GLOBAL_SKIP = {'__kylix_args', '__kylix_emptystr', '__kylix_exc_active',
                   '__kylix_exc_obj', '__kylix_exc_type', '__kylix_exctab',
                   '__kylix_jmpbuf', '__kylix_datetime_arena',
                   '__kylix_datetime_arena_ptr'}
    def glob_seg(name):
        if name in GLOBAL_SKIP:
            return None
        if name.startswith('__kylix_boot_'):
            rest = name[len('__kylix_boot_'):]
            # routes/nroutes/ctrl_* are program-specific — the bootstrap's
            # annotation auto-wiring emits them itself. Everything else
            # (404_page/500_page/sessions/static_dir/csrf_enabled) backs the
            # baked boot defines.
            if rest in ('routes', 'nroutes') or rest.startswith('ctrl_'):
                return None
            return 'boot'
        if name in ('__kylix_b64_table', '__kylix_b64url_table'):
            return 'encoding'
        return 'runtime'
    seg_globals = {seg: [] for seg in SEGMENTS}
    for gname in sorted(all_globs):
        gseg = glob_seg(gname)
        if gseg is not None:
            seg_globals[gseg].append(all_globs[gname])

    seg_text, seg_deps = {}, {}
    extra_decls = [ln for ln in all_decls
                   if re.match(r'declare [^@]*@([\w.]+)\(', ln).group(1).lstrip('@')
                   not in BOOTSTRAP_DECLARED]
    for seg in SEGMENTS:
        lines_out = list(seg_bodies[seg]) + seg_kstrs[seg]
        if seg == 'runtime':
            lines_out = extra_decls + lines_out
        if seg == 'boot':
            # %__ret_* multi-return aggregate types first (LLVM accepts named
            # types anywhere at module level, but keep the segment readable).
            lines_out = [all_types[t] for t in sorted(all_types)] + lines_out
        # module-level globals referenced by this segment's bodies
        lines_out.extend(seg_globals[seg])
        deps = set()
        for ln in lines_out:
            for sym2 in re.findall(r'@__kylix_[A-Za-z0-9_]+', ln):
                s2 = sym_seg(sym2)
                if s2 is not None and s2 != seg:
                    deps.add(s2)
        seg_text[seg] = lines_out
        seg_deps[seg] = sorted(deps)

    # JsonEncode/JsonEncodePretty take a Variant box in the baked IR, but Kylix
    # programs pass a map variable (htab handle). Synthesize htab-arg wrappers
    # so the bootstrap dispatch needs no AST type inspection (its parser can't
    # lower `(expr as TIdentifier).Value` for function-argument exprs).
    WRAPPERS = {
        'json': [('JsonEncode', '__kylix_json_JsonEncode'),
                 ('JsonEncodePretty', '__kylix_json_JsonEncodePretty')],
    }
    base_syms = set(a[0].lstrip('@') for a in sigs)
    for seg, pairs in WRAPPERS.items():
        for kyname, sym in pairs:
            symn = sym[1:] if sym.startswith('@') else sym
            if symn not in base_syms:
                print('; wrapper skipped, base missing: ' + symn, file=sys.stderr)
                continue
            wrapper = [
                'define ptr @' + symn + '_htab(ptr %m) {',
                'entry:',
                '  %t0 = call ptr @__kylix_variant_box_map(ptr %m)',
                '  %t1 = call ptr @' + symn + '(ptr %t0)',
                '  ret ptr %t1',
                '}',
            ]
            seg_text[seg].extend(wrapper)
            sigs.append(('@' + symn + '_htab', seg, 'ptr', ['ptr']))

    # ---- Emit the Kylix data unit. ----------------------------------------
    def q(s):
        return "'" + s.replace("'", "''") + "'"

    out = []
    w = out.append
    w('// stdlib_ir.klx — AUTO-GENERATED by scripts/extract_stdlib_ir.py; do not edit.')
    w('// Host-emitted LLVM IR for the stdlib modules, baked as Kylix string data')
    w('// (v0.6.9 P3 bootstrap no-Go closed loop). Segments are appended verbatim')
    w('// when a program calls into them; unreferenced segments are never emitted.')
    w('')
    w('var')
    w('  StdIrLines: array of String;   // flattened IR lines, per-segment ranges below')
    w('  StdSegNames: array of String;  // segment order: runtime first (unconditional)')
    w('  StdSegStart: array of Integer; // first line of the segment in StdIrLines')
    w('  StdSegCount: array of Integer; // line count of the segment')
    w('  StdSegDeps: array of String;   // comma-separated dependent segment names')
    w('  StdFnNames: array of String;   // "@__kylix_<seg>_<Fn>" (PathJoin_N per arity)')
    w('  StdFnRet: array of String;     // LLVM return type ("void" allowed)')
    w('  StdFnParams: array of String;  // LLVM param types joined by "," ("" = none)')
    w('  StdIrInitDone: Boolean;')
    w('')
    w('procedure StdIrInit();')
    w('begin')
    w('  if StdIrInitDone then Exit;')
    w('  StdIrInitDone := true;')
    # first pass of Init: compute segment starts
    starts, pos = {}, 0
    for seg in SEGMENTS:
        starts[seg] = pos
        pos += len(seg_text[seg])
    for seg in SEGMENTS:
        w("  append(StdSegNames, %s);" % q(seg))
        w('  append(StdSegStart, %d);' % starts[seg])
        w('  append(StdSegCount, %d);' % len(seg_text[seg]))
        w("  append(StdSegDeps, %s);" % q(','.join(seg_deps[seg])))
    w('  // ---- IR lines (define bodies + segment-local @kstr.* constants) ----')
    for seg in SEGMENTS:
        w('  // ---- segment: %s ----' % seg)
        for ln in seg_text[seg]:
            w('  append(StdIrLines, %s);' % q(ln))
    w('  // ---- dispatch signatures ----')
    for sym, seg, ret, ptypes in sigs:
        w('  append(StdFnNames, %s);' % q(sym[1:]))  # drop leading @
        w('  append(StdFnRet, %s);' % q(ret))
        w('  append(StdFnParams, %s);' % q(','.join(ptypes)))
    w('end;')
    sys.stdout.write('\n'.join(out) + '\n')
    # stats to stderr
    total = sum(len(seg_text[s]) for s in SEGMENTS)
    for seg in SEGMENTS:
        print('; %-11s %4d lines  deps=%s' % (seg, len(seg_text[seg]), ','.join(seg_deps[seg]) or '-'),
              file=sys.stderr)
    print('; total IR lines: %d, signatures: %d' % (total, len(sigs)), file=sys.stderr)


if __name__ == '__main__':
    main()
