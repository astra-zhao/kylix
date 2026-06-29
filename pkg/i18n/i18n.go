// i18n — Internationalization for Kylix compiler error messages.
//
// Language is selected via the KYLIX_LANG environment variable.
// Supported values: "en" (default), "zh" (Chinese).
//
// Each error code has a translation table; missing translations fall back to English.
package i18n

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// Lang represents a supported language.
type Lang string

const (
	LangEn Lang = "en"
	LangZh Lang = "zh"
)

var (
	currentLang Lang = LangEn
	once        sync.Once
)

// Init reads KYLIX_LANG and sets the current language.
// Called automatically on first use.
func Init() {
	once.Do(func() {
		v := strings.ToLower(os.Getenv("KYLIX_LANG"))
		switch v {
		case "zh", "zh-cn", "zh_cn", "chinese":
			currentLang = LangZh
		default:
			currentLang = LangEn
		}
	})
}

// SetLang explicitly sets the active language (for tests).
func SetLang(l Lang) {
	once.Do(func() {}) // mark initialized so Init() doesn't override
	currentLang = l
}

// CurrentLang returns the active language.
func CurrentLang() Lang {
	Init()
	return currentLang
}

// translations maps error codes to per-language message templates.
// Templates use %s/%d Go-style verbs; callers pass args in the same order.
var translations = map[string]map[Lang]string{
	"KLX001": {
		LangEn: "unexpected token: %s",
		LangZh: "意外的标记: %s",
	},
	"KLX002": {
		LangEn: "missing token: expected %s",
		LangZh: "缺少标记: 期望 %s",
	},
	"KLX003": {
		LangEn: "unterminated string literal",
		LangZh: "字符串字面量未闭合",
	},
	"KLX004": {
		LangEn: "%s",
		LangZh: "%s",
	},
	"KLX005": {
		LangEn: "circular dependency detected",
		LangZh: "检测到循环依赖",
	},
	"KLX101": {
		LangEn: "cannot assign %s literal to variable of type '%s'",
		LangZh: "无法将 %s 字面量赋给类型为 '%s' 的变量",
	},
	"KLX102": {
		LangEn: "cannot infer type for variable",
		LangZh: "无法推导变量类型",
	},
	"KLX103": {
		LangEn: "invalid type cast",
		LangZh: "无效的类型转换",
	},
	"KLX104": {
		LangEn: "type '%s' does not satisfy constraint '%s' for parameter '%s' of generic type '%s'",
		LangZh: "类型 '%s' 不满足约束 '%s'(用于泛型 '%[4]s' 的参数 '%[3]s')",
	},
	"KLX105": {
		LangEn: "type alias '%s' is recursive (cycle detected)",
		LangZh: "类型别名 '%s' 是递归的（检测到循环）",
	},
	"KLX201": {
		LangEn: "undeclared variable or function '%s'",
		LangZh: "未声明的变量或函数 '%s'",
	},
	"KLX202": {
		LangEn: "wrong number of arguments to '%s': expected %d, got %d",
		LangZh: "调用 '%s' 的参数数量错误: 期望 %d 个，实际 %d 个",
	},
	"KLX203": {
		LangEn: "duplicate declaration: '%s'",
		LangZh: "重复声明: '%s'",
	},
	"KLX204": {
		LangEn: "uninitialized variable: '%s'",
		LangZh: "变量未初始化: '%s'",
	},
	"KLX205": {
		LangEn: "'break' outside of loop",
		LangZh: "'break' 不在循环中",
	},
	"KLX206": {
		LangEn: "function '%s' missing return type",
		LangZh: "函数 '%s' 缺少返回类型",
	},
	"KLX214": {
		LangEn: "[Body] annotation error: %s",
		LangZh: "[Body] 注解错误: %s",
	},
	"KLX301": {
		LangEn: "class %q implements %q but is missing method %q",
		LangZh: "类 %q 声明实现了 %q 但缺少方法 %q",
	},
	"KLX302": {
		LangEn: "method '%s' signature mismatch",
		LangZh: "方法 '%s' 签名不匹配",
	},
	"KLX303": {
		LangEn: "unknown interface: '%s'",
		LangZh: "未知接口: '%s'",
	},
	"KLX401": {
		LangEn: "internal compiler error: %s",
		LangZh: "编译器内部错误: %s",
	},
	"KLX402": {
		LangEn: "cannot read file: %v",
		LangZh: "无法读取文件: %v",
	},
	"KLX403": {
		LangEn: "cannot write file: %v",
		LangZh: "无法写入文件: %v",
	},
}

// hints holds localized fix suggestions for common errors.
var hints = map[string]map[Lang]string{
	"KLX101_str_to_int": {
		LangEn: "use StrToInt(s) or StrToInt64(s) to convert a String to Integer",
		LangZh: "使用 StrToInt(s) 或 StrToInt64(s) 把 String 转为 Integer",
	},
	"KLX101_str_to_float": {
		LangEn: "use StrToFloat(s) to convert a String to Real",
		LangZh: "使用 StrToFloat(s) 把 String 转为 Real",
	},
	"KLX101_str_to_bool": {
		LangEn: "use (s = 'true') to convert a String to Boolean",
		LangZh: "使用 (s = 'true') 把 String 转为 Boolean",
	},
	"KLX101_int_to_str": {
		LangEn: "use IntToStr(n) to convert an Integer to String",
		LangZh: "使用 IntToStr(n) 把 Integer 转为 String",
	},
	"KLX201_did_you_mean": {
		LangEn: "did you mean '%s'?",
		LangZh: "你是否想用 '%s'?",
	},
	"KLX301_add_method": {
		LangEn: "add 'procedure/function %s' to class %s",
		LangZh: "在类 %s 中添加 'procedure/function %s'",
	},
	"KLX105_no_recursive": {
		LangEn: "type aliases cannot reference themselves directly or indirectly",
		LangZh: "类型别名不能直接或间接引用自身",
	},
}

// HasCode reports whether a translation table exists for the given error code.
func HasCode(code string) bool {
	_, ok := translations[code]
	return ok
}

// T translates an error code template with arguments. Falls back to the English
// version if the current language has no translation for the code.
func T(code string, args ...interface{}) string {
	Init()
	templates, ok := translations[code]
	if !ok {
		// Unknown code — return raw args formatted with %v.
		if len(args) == 1 {
			return fmt.Sprintf("%v", args[0])
		}
		return fmt.Sprint(args...)
	}
	tmpl, ok := templates[currentLang]
	if !ok {
		tmpl = templates[LangEn]
	}
	return fmt.Sprintf(tmpl, args...)
}

// Hint returns a localized fix suggestion. hintKey is a logical name like
// "KLX101_str_to_int". Falls back to English when the language is missing.
func Hint(hintKey string, args ...interface{}) string {
	Init()
	templates, ok := hints[hintKey]
	if !ok {
		return ""
	}
	tmpl, ok := templates[currentLang]
	if !ok {
		tmpl = templates[LangEn]
	}
	return fmt.Sprintf(tmpl, args...)
}
