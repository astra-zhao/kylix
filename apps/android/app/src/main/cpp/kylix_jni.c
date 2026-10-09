#include <jni.h>
#include <stdint.h>

/* Declarations of the Kylix C ABI (apps/shared/mobilecore_lib.klx). */
const char *mc_validate_login(const char *username, const char *password);
const char *mc_login_request(const char *username, const char *password);
const char *mc_parse_login(int64_t status, const char *body);
const char *mc_refresh_request(const char *refresh_token);
const char *mc_parse_refresh(int64_t status, const char *body);
const char *mc_parse_list(int64_t status, const char *body);
const char *mc_auth_header(const char *token);
const char *mc_login_path(void);
const char *mc_refresh_path(void);
const char *mc_notes_path(void);
const char *mc_logout_path(void);
void kylix_free(void *p);

static const char *chars(JNIEnv *env, jstring s) {
    if (s == NULL) {
        return "";
    }
    return (*env)->GetStringUTFChars(env, s, NULL);
}

static void release(JNIEnv *env, jstring s, const char *p) {
    if (s != NULL && p != NULL) {
        (*env)->ReleaseStringUTFChars(env, s, p);
    }
}

/* Copy a Kylix-owned C string into a Java string, then kylix_free it. */
static jstring take(JNIEnv *env, const char *p) {
    jstring s;
    if (p == NULL) {
        return (*env)->NewStringUTF(env, "");
    }
    s = (*env)->NewStringUTF(env, p);
    kylix_free((void *)p);
    return s;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_validateLogin(JNIEnv *env, jobject thiz, jstring user, jstring pass) {
    const char *u = chars(env, user);
    const char *p = chars(env, pass);
    jstring out = take(env, mc_validate_login(u, p));
    release(env, user, u);
    release(env, pass, p);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_loginRequest(JNIEnv *env, jobject thiz, jstring user, jstring pass) {
    const char *u = chars(env, user);
    const char *p = chars(env, pass);
    jstring out = take(env, mc_login_request(u, p));
    release(env, user, u);
    release(env, pass, p);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_parseLogin(JNIEnv *env, jobject thiz, jlong status, jstring body) {
    const char *b = chars(env, body);
    jstring out = take(env, mc_parse_login((int64_t)status, b));
    release(env, body, b);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_parseList(JNIEnv *env, jobject thiz, jlong status, jstring body) {
    const char *b = chars(env, body);
    jstring out = take(env, mc_parse_list((int64_t)status, b));
    release(env, body, b);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_authHeader(JNIEnv *env, jobject thiz, jstring token) {
    const char *t = chars(env, token);
    jstring out = take(env, mc_auth_header(t));
    release(env, token, t);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_loginPath(JNIEnv *env, jobject thiz) {
    (void)thiz;
    return take(env, mc_login_path());
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_refreshPath(JNIEnv *env, jobject thiz) {
    (void)thiz;
    return take(env, mc_refresh_path());
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_refreshRequest(JNIEnv *env, jobject thiz, jstring token) {
    const char *t = chars(env, token);
    jstring out = take(env, mc_refresh_request(t));
    release(env, token, t);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_parseRefresh(JNIEnv *env, jobject thiz, jlong status, jstring body) {
    const char *b = chars(env, body);
    jstring out = take(env, mc_parse_refresh((int64_t)status, b));
    release(env, body, b);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_notesPath(JNIEnv *env, jobject thiz) {
    (void)thiz;
    return take(env, mc_notes_path());
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_admin_KylixBridge_logoutPath(JNIEnv *env, jobject thiz) {
    (void)thiz;
    return take(env, mc_logout_path());
}
