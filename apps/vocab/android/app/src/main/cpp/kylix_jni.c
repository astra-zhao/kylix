#include <jni.h>
#include <stdint.h>

/* Declarations of the Kylix C ABI (apps/vocab/vocab_lib.klx). */
const char *vc_words_query(const char *known);
const char *vc_review_query(const char *known);
const char *vc_mark_path(void);
const char *vc_mark_request(const char *known, const char *id);
const char *vc_parse(int64_t status, const char *body);
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
Java_dev_kylix_vocab_KylixBridge_wordsQuery(JNIEnv *env, jobject thiz, jstring known) {
    const char *k = chars(env, known);
    jstring out = take(env, vc_words_query(k));
    release(env, known, k);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_vocab_KylixBridge_reviewQuery(JNIEnv *env, jobject thiz, jstring known) {
    const char *k = chars(env, known);
    jstring out = take(env, vc_review_query(k));
    release(env, known, k);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_vocab_KylixBridge_markPath(JNIEnv *env, jobject thiz) {
    (void)thiz;
    return take(env, vc_mark_path());
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_vocab_KylixBridge_markRequest(JNIEnv *env, jobject thiz, jstring known, jstring id) {
    const char *k = chars(env, known);
    const char *i = chars(env, id);
    jstring out = take(env, vc_mark_request(k, i));
    release(env, known, k);
    release(env, id, i);
    (void)thiz;
    return out;
}

JNIEXPORT jstring JNICALL
Java_dev_kylix_vocab_KylixBridge_parse(JNIEnv *env, jobject thiz, jlong status, jstring body) {
    const char *b = chars(env, body);
    jstring out = take(env, vc_parse((int64_t)status, b));
    release(env, body, b);
    (void)thiz;
    return out;
}
