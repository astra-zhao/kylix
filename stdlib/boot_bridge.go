// boot_bridge.go — Re-export pkg/boot symbols under stdlib for Kylix users.
//
// Kylix programs using `uses boot;` will see these symbols mapped to
// stdlib.BootXxx (because the codegen prefixes stdlib functions). To keep
// names natural in Kylix, we wrap the boot package APIs here.
package stdlib

import (
	"kylix/pkg/boot"
)

// ===== Type aliases for use in Kylix =====

// BootApp is the application object (App).
type BootApp = boot.App

// BootRequest / BootResponse are HTTP request/response.
type BootRequest = boot.Request
type BootResponse = boot.Response

// BootHandler is the canonical Kylix handler signature.
type BootHandler = boot.Handler

// BootMiddleware is the middleware function type.
type BootMiddleware = boot.Middleware

// ===== Top-level functions =====

// BootRun starts the default KylixBoot app on the given port.
func BootRun(port int64) error {
	return boot.Run(int(port))
}

// BootGET registers a GET handler.
func BootGET(path string, h boot.Handler) {
	boot.GET(path, h)
}

// BootPOST registers a POST handler.
func BootPOST(path string, h boot.Handler) {
	boot.POST(path, h)
}

// BootPUT registers a PUT handler.
func BootPUT(path string, h boot.Handler) {
	boot.PUT(path, h)
}

// BootDELETE registers a DELETE handler.
func BootDELETE(path string, h boot.Handler) {
	boot.DELETE(path, h)
}

// BootUseLogger installs the Logger middleware.
func BootUseLogger() {
	boot.Use(boot.Logger())
}

// BootUseRecover installs the Recover (panic-catching) middleware.
func BootUseRecover() {
	boot.Use(boot.Recover())
}

// BootUseCORS installs the CORS middleware (any origin).
func BootUseCORS() {
	boot.Use(boot.CORS())
}

// BootUseRequestID installs the RequestID middleware.
func BootUseRequestID() {
	boot.Use(boot.RequestID())
}

// BootText creates a plain text response.
func BootText(status int64, body string) *boot.Response {
	return boot.Text(int(status), body)
}

// BootJSON creates a JSON response from a value.
func BootJSON(status int64, value interface{}) *boot.Response {
	return boot.JSON(int(status), value)
}

// BootHTML creates an HTML response.
func BootHTML(status int64, body string) *boot.Response {
	return boot.HTML(int(status), body)
}

// BootConfigSet stores a key/value pair in the default config.
func BootConfigSet(key string, value interface{}) {
	boot.SetConfig(key, value)
}

// BootConfigGetString retrieves a string config value with fallback.
func BootConfigGetString(key, fallback string) string {
	return boot.GetConfigString(key, fallback)
}

// BootConfigGetInt retrieves an int config value with fallback.
func BootConfigGetInt(key string, fallback int64) int64 {
	return int64(boot.GetConfigInt(key, int(fallback)))
}

// BootRegisterInstance binds an instance to the default DI container.
func BootRegisterInstance(name string, instance interface{}) {
	boot.RegisterInstance(name, instance)
}

// BootResolve resolves a registered instance.
func BootResolve(name string) interface{} {
	return boot.Resolve(name)
}

// BootRegisterAuth registers the global token → user validator used by
// [Authenticated] and [Role] annotation guards.
func BootRegisterAuth(v func(string) (string, bool)) {
	boot.RegisterAuthValidator(v)
}

// BootRegisterRoles registers the global user → roles provider used by
// [Role] annotation guards.
func BootRegisterRoles(p func(string) []string) {
	boot.RegisterRolesProvider(p)
}

// BootEnforceAuth verifies the Bearer token and populates req.User on success.
// Returns a 401 *Response on any failure, nil on success. Called by generated
// route closures for methods carrying [Authenticated] or [Role(...)].
func BootEnforceAuth(req *boot.Request) *boot.Response {
	return boot.EnforceAuth(req)
}

// BootEnforceRole returns a 403 *Response unless req.User has the given role.
// Implicitly calls EnforceAuth if req.User is empty.
func BootEnforceRole(req *boot.Request, role string) *boot.Response {
	return boot.EnforceRole(req, role)
}

// BootReadJSON unmarshals the request body as JSON into the given pointer.
// Used by generated route closures annotated with [Body(TEntity)].
func BootReadJSON(req *boot.Request, out interface{}) error {
	return req.JSON(out)
}
