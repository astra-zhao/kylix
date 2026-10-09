#pragma once
#include <stdint.h>

/* C ABI of apps/shared/mobilecore_lib.klx. Returned strings are malloc'd;
 * call kylix_free when the caller has copied them. */
const char *mc_validate_login(const char *username, const char *password);
const char *mc_login_request(const char *username, const char *password);
const char *mc_parse_login(int64_t status, const char *body);
const char *mc_refresh_request(const char *refresh_token);
const char *mc_parse_refresh(int64_t status, const char *body);
const char *mc_list_request(const char *token);
const char *mc_parse_list(int64_t status, const char *body);
const char *mc_auth_header(const char *token);
const char *mc_login_path(void);
const char *mc_refresh_path(void);
const char *mc_notes_path(void);
const char *mc_logout_path(void);
void kylix_free(void *p);
