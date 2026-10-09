/* c_host_check.c — load libkylixlogic.so and call the exported core.
 * Built by apps/shared/host_check.sh. Not an Android/iOS run: it checks the
 * same C ABI those shells link. */
#include <dlfcn.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

typedef const char *(*fn2)(const char *, const char *);
typedef const char *(*fn_status)(int64_t, const char *);
typedef const char *(*fn1)(const char *);
typedef const char *(*fn0)(void);
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

static const char *must(void *sym, const char *name) {
    if (sym == NULL) {
        fprintf(stderr, "missing symbol %s\n", name);
        fails = 1;
    }
    return name;
}

int main(int argc, char **argv) {
    const char *path = argc > 1 ? argv[1] : "./libkylixlogic.so";
    void *h = dlopen(path, RTLD_NOW);
    if (h == NULL) {
        fprintf(stderr, "dlopen: %s\n", dlerror());
        return 1;
    }

    fn2 validate = (fn2)dlsym(h, "mc_validate_login");
    fn2 login_req = (fn2)dlsym(h, "mc_login_request");
    fn_status parse_login = (fn_status)dlsym(h, "mc_parse_login");
    fn1 refresh_req = (fn1)dlsym(h, "mc_refresh_request");
    fn_status parse_refresh = (fn_status)dlsym(h, "mc_parse_refresh");
    fn1 list_req = (fn1)dlsym(h, "mc_list_request");
    fn_status parse_list = (fn_status)dlsym(h, "mc_parse_list");
    fn1 auth = (fn1)dlsym(h, "mc_auth_header");
    fn0 login_path = (fn0)dlsym(h, "mc_login_path");
    fn0 refresh_path = (fn0)dlsym(h, "mc_refresh_path");
    fn0 notes_path = (fn0)dlsym(h, "mc_notes_path");
    fn_free kylix_free = (fn_free)dlsym(h, "kylix_free");
    must(validate, "mc_validate_login");
    must(login_req, "mc_login_request");
    must(parse_login, "mc_parse_login");
    must(refresh_req, "mc_refresh_request");
    must(parse_refresh, "mc_parse_refresh");
    must(list_req, "mc_list_request");
    must(parse_list, "mc_parse_list");
    must(auth, "mc_auth_header");
    must(login_path, "mc_login_path");
    must(refresh_path, "mc_refresh_path");
    must(notes_path, "mc_notes_path");
    must(kylix_free, "kylix_free");
    if (fails) {
        return 1;
    }

    const char *s = validate("", "x");
    expect("validate", s, "Username is required");
    kylix_free((void *)s);

    s = validate("admin", "Admin@123");
    expect("validate_ok", s, "");
    kylix_free((void *)s);

    s = login_path();
    expect("login_path", s, "/api/login");
    kylix_free((void *)s);

    s = refresh_path();
    expect("refresh_path", s, "/api/refresh");
    kylix_free((void *)s);

    s = notes_path();
    expect("notes_path", s, "/api/notes");
    kylix_free((void *)s);

    s = auth("tok.en");
    expect("auth", s, "Bearer tok.en");
    kylix_free((void *)s);

    s = login_req("admin", "Admin@123");
    expect("login_req", s, "{\"username\":\"admin\",\"password\":\"Admin@123\"}");
    kylix_free((void *)s);

    const char *ok = "{\"ok\":true,\"token\":\"tok.en\",\"refresh_token\":\"ref.tok\",\"username\":\"admin\",\"display_name\":\"Administrator\",\"expires_in\":86400,\"refresh_expires_in\":2592000}";
    s = parse_login(200, ok);
    expect("parse_login", s, ok);
    kylix_free((void *)s);

    s = refresh_req("ref.tok");
    expect("refresh_req", s, "{\"refresh_token\":\"ref.tok\"}");
    kylix_free((void *)s);

    s = parse_refresh(200, ok);
    expect("parse_refresh", s, ok);
    kylix_free((void *)s);

    s = parse_refresh(401, "{\"ok\":false,\"error\":\"unauthorized\"}");
    expect("refresh_bad", s, "{\"ok\":false,\"error\":\"unauthorized\",\"relogin\":true}");
    kylix_free((void *)s);

    s = parse_login(401, "{\"ok\":false,\"error\":\"Invalid username or password\"}");
    expect("parse_bad", s, "{\"ok\":false,\"error\":\"Invalid username or password\",\"relogin\":false}");
    kylix_free((void *)s);

    s = list_req("tok.en");
    expect("list_req", s, "{\"method\":\"GET\",\"path\":\"/api/notes\",\"authorization\":\"Bearer tok.en\"}");
    kylix_free((void *)s);

    s = parse_list(401, "Unauthorized");
    expect("relogin", s, "{\"ok\":false,\"error\":\"unauthorized\",\"relogin\":true}");
    kylix_free((void *)s);

    dlclose(h);
    if (fails) {
        return 1;
    }
    printf("c abi host check: PASS\n");
    return 0;
}
