// generator_stdlib.go — stdlib function dispatch for `uses` modules.
//
// When a Kylix program uses `uses strutil;` and calls `Reverse('hello')`,
// this file maps `Reverse` → `stdlib.Reverse(...)`.
//
// Functions that return (value, error) are wrapped in an inline func to discard
// the error and return only the value, matching Kylix's single-return semantics.
package generator

import (
	"kylix/ast"
)

// stdlibModuleFuncs maps module name → set of exported function names.
// Only includes functions actually exported by the Go stdlib package.
// Pure Kylix modules (strutil, mathutil, arrayutil, collections) are NOT included
// here — they require unit compilation and cannot be called from program files yet.
var stdlibModuleFuncs = map[string]map[string]bool{
	"sysutil": strToSet(
		"FileOpen", "ReadFile", "WriteFile", "AppendFile",
		"FileExists", "DirExists", "CreateDir", "DeleteFile",
		"CopyFile", "ListDir", "ListFiles", "GetFileSize",
		"ReadLines", "WriteLines",
		"PathJoin", "PathDir", "PathBase", "PathExt",
		"GetWorkingDir", "SetWorkingDir", "GetTempDir",
		"GetEnv", "SetEnv", "Sleep",
	),
	"jsonutil": strToSet(
		"JsonEncode", "JsonEncodePretty",
		"JsonDecode", "JsonDecodeMap", "JsonDecodeArray",
		"JsonGetString", "JsonGetInt", "JsonGetFloat", "JsonGetBool",
		"JsonGetMap", "JsonGetArray", "JsonHasKey", "JsonIsValid",
		"JsonReadFile", "JsonWriteFile",
	),
	"datetime": strToSet(
		"Now", "Today", "MakeDate", "MakeTime", "ParseDate", "ParseDateTime",
		"GetTimestamp", "GetTimestampMs",
	),
	"regex": strToSet(
		"IsAlpha", "IsAlphaNumeric",
		"IsNumeric", "IsEmail", "IsURL", "IsIP",
		"RegexCompile", "RegexMustCompile",
		"RegexMatch", "RegexFind", "RegexFind2", "RegexReplace", "RegexSplit",
	),
	"httpclient": strToSet(
		"NewHttpClient", "HttpGet", "HttpPost", "HttpGetJSON",
	),
	"web": strToSet(
		"NewServer",
	),
	"orm": strToSet(
		"NewDatabase", "NewORM", "NewQueryBuilder", "NewMigrationManager",
	),
	"container":  strToSet("NewContainer"),
	"config":     strToSet("NewConfig", "NewAppConfig"),
	"autoconfig": strToSet("NewAutoConfig"),
	"middleware": strToSet(
		"NewCORSMiddleware", "NewLoggingMiddleware", "NewRecoveryMiddleware",
		"NewAuthMiddleware", "NewRateLimitMiddleware", "LoggerMiddleware",
		"NewRequestIDMiddleware", "GetRequestID", "GetAuthToken",
	),
	"validation": strToSet("NewValidator", "NewRequestValidator"),
	"template":   strToSet("NewTemplateEngine", "NewView"),
	"boot": strToSet(
		"BootRun", "BootGET", "BootPOST", "BootPUT", "BootDELETE",
		"BootUseLogger", "BootUseRecover", "BootUseCORS", "BootUseRequestID",
		"BootText", "BootJSON", "BootHTML",
		"BootConfigSet", "BootConfigGetString", "BootConfigGetInt",
		"BootRegisterInstance", "BootResolve",
		"BootRegisterAuth", "BootRegisterRoles", "BootEnforceAuth", "BootEnforceRole",
	),
}

// stdlibErrorFuncReturnTypes maps error-returning stdlib functions to their
// concrete Go return type. Used to avoid `interface{}` wrapping issues when
// assigning the result to a typed variable.
var stdlibErrorFuncReturnTypes = map[string]string{
	"ReadFile":        "string",
	"ParseDate":       "*stdlib.TDateTime",
	"ParseDateTime":   "*stdlib.TDateTime",
	"RegexCompile":    "*stdlib.TRegex",
	"HttpGet":         "string",
	"HttpPost":        "string",
	"HttpGetJSON":     "map[string]interface{}",
	"ListDir":         "[]string",
	"ListFiles":       "[]string",
	"ReadLines":       "[]string",
	"GetFileSize":     "int64",
	"FileOpen":        "*stdlib.TTextFile",
	"JsonDecode":      "interface{}",
	"JsonDecodeMap":   "map[string]interface{}",
	"JsonDecodeArray": "[]interface{}",
	"JsonReadFile":    "interface{}",
}

// stdlibErrorFuncs are stdlib functions that return (T, error) in Go.
// The generator wraps them to discard the error.
var stdlibErrorFuncs = map[string]bool{
	"ReadFile": true, "WriteFile": true, "AppendFile": true,
	"FileOpen": true, "CreateDir": true, "DeleteFile": true,
	"CopyFile": true, "ListDir": true, "ListFiles": true,
	"ReadLines": true, "WriteLines": true, "GetFileSize": true,
	"ParseDate": true, "ParseDateTime": true,
	"RegexCompile": true,
	"HttpGet":      true, "HttpPost": true, "HttpGetJSON": true,
	"JsonDecode": true, "JsonDecodeMap": true, "JsonDecodeArray": true,
	"JsonReadFile": true,
}

// stdlibProcedureFuncs are stdlib functions that return no value (procedures).
var stdlibProcedureFuncs = map[string]bool{
	"WriteFile": true, "AppendFile": true, "CreateDir": true,
	"DeleteFile": true, "CopyFile": true, "WriteLines": true,
	"SetWorkingDir": true, "SetEnv": true, "Sleep": true,
	"Stdout": true, "Stderr": true, "WasiExit": true,
}

func strToSet(names ...string) map[string]bool {
	s := make(map[string]bool, len(names))
	for _, n := range names {
		s[n] = true
	}
	return s
}

// resolveStdlibFunc returns (goFuncName, moduleName) if funcName belongs to
// one of the `uses` modules, otherwise ("", "").
func (g *Generator) resolveStdlibFunc(funcName string) (string, string) {
	// Skip if the function is user-defined (takes precedence)
	if g.userFuncs[funcName] {
		return "", ""
	}
	for mod, funcs := range stdlibModuleFuncs {
		if g.usedModules[mod] && funcs[funcName] {
			return funcName, mod
		}
	}
	return "", ""
}

// generateStdlibCall emits a stdlib.FuncName(...) call.
// Functions returning (T, error) are wrapped to discard the error.
// Procedures (void) call directly.
func (g *Generator) generateStdlibCall(funcName string, args []ast.Expression) {
	goCall := "stdlib." + funcName

	if stdlibErrorFuncs[funcName] {
		if stdlibProcedureFuncs[funcName] {
			// procedure that returns error: discard silently
			g.write("func() { " + goCall + "(")
			g.writeArgList(args)
			g.write(") }()")
		} else {
			// function returning (T, error): wrap to return T only
			// Use concrete type when known to avoid type-assertion friction.
			retType := "interface{}"
			if t, ok := stdlibErrorFuncReturnTypes[funcName]; ok {
				retType = t
			}
			g.write("func() " + retType + " { _v, _ := " + goCall + "(")
			g.writeArgList(args)
			g.write("); return _v }()")
		}
	} else if stdlibProcedureFuncs[funcName] {
		g.write(goCall + "(")
		g.writeArgList(args)
		g.write(")")
	} else {
		g.write(goCall + "(")
		g.writeArgList(args)
		g.write(")")
	}
}

// writeArgList emits comma-separated generated arguments.
func (g *Generator) writeArgList(args []ast.Expression) {
	for i, arg := range args {
		if i > 0 {
			g.write(", ")
		}
		g.generateExpression(arg)
	}
}
