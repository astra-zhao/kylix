/* c_host_check.c — load libkylixvocab.so and walk browse / mark / review.
 * Built by apps/vocab/host_check.sh. Not an Android/iOS run: it checks the
 * same C ABI those shells link. */
#include <dlfcn.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

typedef const char *(*fn0)(void);
typedef const char *(*fn1)(const char *);
typedef const char *(*fn2)(const char *, const char *);
typedef const char *(*fn_status)(int64_t, const char *);
typedef void (*fn_free)(void *);

static int fails;

static void expect(const char *name, const char *got, const char *want) {
    if (got == NULL || strcmp(got, want) != 0) {
        fprintf(stderr, "FAIL %s\n  got  [%s]\n  want [%s]\n",
                name, got ? got : "(null)", want);
        fails = 1;
    } else {
        printf("ok %s\n", name);
    }
}

static void expect_has(const char *name, const char *got, const char *part) {
    if (got == NULL || strstr(got, part) == NULL) {
        fprintf(stderr, "FAIL %s missing [%s]\n  got [%s]\n",
                name, part, got ? got : "(null)");
        fails = 1;
    } else {
        printf("ok %s\n", name);
    }
}

static void expect_not(const char *name, const char *got, const char *part) {
    if (got == NULL || strstr(got, part) != NULL) {
        fprintf(stderr, "FAIL %s should not contain [%s]\n  got [%s]\n",
                name, part, got ? got : "(null)");
        fails = 1;
    } else {
        printf("ok %s\n", name);
    }
}

static void must(void *sym, const char *name) {
    if (sym == NULL) {
        fprintf(stderr, "missing symbol %s\n", name);
        fails = 1;
    }
}

int main(int argc, char **argv) {
    const char *path = argc > 1 ? argv[1] : "./libkylixvocab.so";
    void *h = dlopen(path, RTLD_NOW);
    if (h == NULL) {
        fprintf(stderr, "dlopen: %s\n", dlerror());
        return 1;
    }

    fn0 words_path = (fn0)dlsym(h, "vc_words_path");
    fn0 review_path = (fn0)dlsym(h, "vc_review_path");
    fn0 mark_path = (fn0)dlsym(h, "vc_mark_path");
    fn1 words_query = (fn1)dlsym(h, "vc_words_query");
    fn1 review_query = (fn1)dlsym(h, "vc_review_query");
    fn1 browse = (fn1)dlsym(h, "vc_browse");
    fn1 review = (fn1)dlsym(h, "vc_review");
    fn2 mark_request = (fn2)dlsym(h, "vc_mark_request");
    fn2 apply_mark = (fn2)dlsym(h, "vc_apply_mark");
    fn_status parse = (fn_status)dlsym(h, "vc_parse");
    fn_free kylix_free = (fn_free)dlsym(h, "kylix_free");
    must(words_path, "vc_words_path");
    must(review_path, "vc_review_path");
    must(mark_path, "vc_mark_path");
    must(words_query, "vc_words_query");
    must(review_query, "vc_review_query");
    must(browse, "vc_browse");
    must(review, "vc_review");
    must(mark_request, "vc_mark_request");
    must(apply_mark, "vc_apply_mark");
    must(parse, "vc_parse");
    must(kylix_free, "kylix_free");
    if (fails) {
        return 1;
    }

    const char *s = words_path();
    expect("words_path", s, "/api/words");
    kylix_free((void *)s);

    s = review_path();
    expect("review_path", s, "/api/review");
    kylix_free((void *)s);

    s = mark_path();
    expect("mark_path", s, "/api/mark");
    kylix_free((void *)s);

    s = words_query("3,1,1");
    expect("words_query", s, "/api/words?known=1,3");
    kylix_free((void *)s);

    s = review_query("3,1");
    expect("review_query", s, "/api/review?known=1,3");
    kylix_free((void *)s);

    s = browse("");
    expect_has("browse_apple", s, "apple");
    expect_has("browse_zh", s, "\xe8\x8b\xb9\xe6\x9e\x9c");
    expect_has("browse_left", s, "\"left\":6");
    expect_has("browse_known", s, "\"known\":\"\"");
    kylix_free((void *)s);

    s = apply_mark("", "1");
    expect_has("mark_known", s, "\"known\":\"1\"");
    expect_has("mark_left", s, "\"left\":5");
    expect_has("mark_apple_known", s, "\"en\":\"apple\",\"zh\":\"\xe8\x8b\xb9\xe6\x9e\x9c\",\"known\":true");
    kylix_free((void *)s);

    s = review("1");
    expect_not("review_hides_apple", s, "apple");
    expect_has("review_book", s, "book");
    expect_has("review_left", s, "\"left\":5");
    kylix_free((void *)s);

    s = apply_mark("1", "1");
    expect_has("mark_again", s, "\"known\":\"1\"");
    kylix_free((void *)s);

    s = apply_mark("1", "9");
    expect("mark_bad", s, "{\"ok\":false,\"error\":\"Unknown word\"}");
    kylix_free((void *)s);

    s = mark_request("1", "2");
    expect("mark_req", s, "{\"known\":\"1\",\"id\":\"2\"}");
    kylix_free((void *)s);

    s = parse(400, "{\"ok\":false,\"error\":\"Unknown word\"}");
    expect("parse_keep", s, "{\"ok\":false,\"error\":\"Unknown word\"}");
    kylix_free((void *)s);

    s = parse(500, "nope");
    expect("parse_fail", s, "{\"ok\":false,\"error\":\"request failed\"}");
    kylix_free((void *)s);

    if (fails) {
        return 1;
    }
    printf("c abi PASS\n");
    return 0;
}
