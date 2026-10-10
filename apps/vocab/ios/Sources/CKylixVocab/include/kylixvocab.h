#pragma once
#include <stdint.h>

/* C ABI of apps/vocab/vocab_lib.klx. Returned strings are malloc'd;
 * call kylix_free when the caller has copied them. */
const char *vc_words_path(void);
const char *vc_review_path(void);
const char *vc_mark_path(void);
const char *vc_words_query(const char *known);
const char *vc_review_query(const char *known);
const char *vc_browse(const char *known);
const char *vc_review(const char *known);
const char *vc_mark_request(const char *known, const char *id);
const char *vc_apply_mark(const char *known, const char *id);
const char *vc_parse(int64_t status, const char *body);
void kylix_free(void *p);
