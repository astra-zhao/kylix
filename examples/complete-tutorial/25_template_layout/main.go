package main

import (
	"fmt"
	"strconv"
	"strings"
	"errors"
)

type TTemplateEngine struct {
Scalars map[string]string
ListLens map[string]int64
CurList string
CurIdx int64
ItemStarted bool
Prefixes []string
PrefixCount int64
Indexes []int64
LastError string
Templates map[string]string
Depth int64
}

func (self *TTemplateEngine) PushScope(prefix string, idx int64) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:71
if (self.PrefixCount < int64(len(self.Prefixes)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:72
self.Prefixes[self.PrefixCount] = prefix
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:74
self.Prefixes = append(self.Prefixes, prefix)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:75
if (self.PrefixCount < int64(len(self.Indexes)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:76
self.Indexes[self.PrefixCount] = idx
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:78
self.Indexes = append(self.Indexes, idx)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:79
self.PrefixCount = (self.PrefixCount + 1)
}

func (self *TTemplateEngine) PopScope() {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:84
if (self.PrefixCount > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:85
self.PrefixCount = (self.PrefixCount - 1)
}	
}

func (self *TTemplateEngine) Create() {
var tmpMap map[string]string = map[string]string{}
var tmpMapI map[string]int64 = map[string]int64{}
var tmpArr []string
var tmpArrI []int64
var tmpTpl map[string]string = map[string]string{}
_ = tmpMap
_ = tmpMapI
_ = tmpArr
_ = tmpArrI
_ = tmpTpl
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:97
self.Scalars = tmpMap
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:98
self.ListLens = tmpMapI
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:99
self.Prefixes = tmpArr
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:100
self.Indexes = tmpArrI
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:101
self.PrefixCount = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:102
self.CurList = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:103
self.CurIdx = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:104
self.ItemStarted = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:105
self.LastError = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:106
self.Templates = tmpTpl
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:107
self.Depth = 0
}

func (self *TTemplateEngine) AddVar(name string, value string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:115
self.Scalars[name] = value
}

func (self *TTemplateEngine) AddInt(name string, value int64) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:121
self.Scalars[name] = fmt.Sprintf("%d", value)
}

func (self *TTemplateEngine) AddVariant(name string, value interface{}) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:130
self.Scalars[name] = func() string { __v := value; if __v == nil { return "" }; return fmt.Sprintf("%v", __v) }()
}

func (self *TTemplateEngine) SetContext(m map[string]string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:139
self.Scalars = m
}

func (self *TTemplateEngine) AddListLen(name string, count int64) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:148
self.ListLens[name] = count
}

func (self *TTemplateEngine) BeginList(name string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:154
self.CurList = name
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:155
self.CurIdx = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:156
self.ItemStarted = false
}

func (self *TTemplateEngine) AddTemplate(name string, src string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:167
self.Templates[name] = src
}

func (self *TTemplateEngine) HasTemplate(name string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:173
result = (self.Templates[name] != "")
	return result
}

func (self *TTemplateEngine) AddItem(value string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:180
self.Scalars[((self.CurList + ".") + fmt.Sprintf("%d", self.CurIdx))] = value
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:181
self.CurIdx = (self.CurIdx + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:182
self.ItemStarted = false
}

func (self *TTemplateEngine) BeginItem() {
}

func (self *TTemplateEngine) ItemField(key string, value string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:201
self.Scalars[((((self.CurList + ".") + fmt.Sprintf("%d", self.CurIdx)) + ".") + key)] = value
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:202
self.ItemStarted = true
}

func (self *TTemplateEngine) NextItem() {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:208
if self.ItemStarted	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:209
self.CurIdx = (self.CurIdx + 1)
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:210
self.ItemStarted = false
}

func (self *TTemplateEngine) EndList() {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:217
if self.ItemStarted	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:218
self.ListLens[self.CurList] = (self.CurIdx + 1)
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:220
self.ListLens[self.CurList] = self.CurIdx
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:221
self.CurList = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:222
self.CurIdx = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:223
self.ItemStarted = false
}

func (self *TTemplateEngine) ListLen(name string) int64 {
var result int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:229
result = self.ListLens[name]
	return result
}

func (self *TTemplateEngine) RenderString(tpl string) string {
var result string
var out string
_ = out
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:244
self.LastError = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:245
self.PrefixCount = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:246
self.Depth = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:247
out = TplRenderInto(self, tpl)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:248
result = out
	return result
}

func (self *TTemplateEngine) ErrorMsg() string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:255
result = self.LastError
	return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:265
func TplFindFrom(hay string, needle string, from int64) int64 {
var result int64
var n int64
var m int64
var i int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:271
n = int64(len(hay))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:272
m = int64(len(needle))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:273
result = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:274
i = from
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:275
for ((i + m) <= n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:277
if (hay[i:(i + m)] == needle)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:279
result = i
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:280
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:282
i = (i + 1)
	}
_ = n
_ = m
_ = i
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:287
func TplSlice(hay string, from int64, n int64) string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:289
result = hay[from:(from + n)]
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:293
func TplStartsWith(hay string, needle string, start int64) bool {
var result bool
var part string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:297
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:298
if ((start + int64(len(needle))) > int64(len(hay)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:299
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:300
part = hay[start:(start + int64(len(needle)))]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:301
if (part == needle)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:302
result = true
}	
_ = part
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:306
func TplIsSpace(c string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:308
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:309
if (c == " ")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:310
result = true
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:311
if (c == "\t")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:312
result = true
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:313
if (c == "\n")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:314
result = true
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:315
if (c == "\r")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:316
result = true
}	
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:320
func TplTrim(s string) string {
var result string
var n int64
var start int64
var fin int64
var c string
var running bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:328
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:329
start = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:330
running = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:331
for running	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:333
if (start >= n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:334
running = false
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:337
c = s[start:(start + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:338
if TplIsSpace(c)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:339
start = (start + 1)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:341
running = false
			}
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:344
fin = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:345
running = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:346
for running	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:348
if (fin <= start)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:349
running = false
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:352
c = s[(fin - 1):fin]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:353
if TplIsSpace(c)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:354
fin = (fin - 1)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:356
running = false
			}
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:359
result = s[start:fin]
_ = n
_ = start
_ = fin
_ = c
_ = running
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:363
func TplEscape(s string) string {
var result string
var i int64
var n int64
var c string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:369
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:370
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:371
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:372
for (i < n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:374
c = s[i:(i + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:375
if (c == "&")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:376
result = (result + "&amp;")
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:377
if (c == "<")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:378
result = (result + "&lt;")
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:379
if (c == ">")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:380
result = (result + "&gt;")
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:381
if (c == "\"")					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:382
result = (result + "&quot;")
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:383
if (func() int64 { if len(c) == 0 { return 0 }; return int64(c[0]) }() == 39)						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:386
result = (result + "&#39;")
}						 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:388
result = (result + c)
						}
					}
				}
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:389
i = (i + 1)
	}
_ = i
_ = n
_ = c
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:396
func TplCharIsLetter(c string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:398
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:399
if (c >= "a")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:401
if (c <= "z")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:402
result = true
}		
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:404
if (result == false)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:406
if (c >= "A")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:408
if (c <= "Z")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:409
result = true
}			
}		
}	
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:415
func TplIsTruthy(s string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:417
result = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:418
if (s == "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:419
result = false
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:420
if (s == "0")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:421
result = false
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:422
if (s == "false")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:423
result = false
}	
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:428
func TplCleanName(s string) string {
var result string
var n int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:432
result = TplTrim(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:433
if TplStartsWith(result, "layout ", 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:434
result = TplTrim(result[7:int64(len(result))])
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:435
n = int64(len(result))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:436
if (n >= 2)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:438
if (result[0:1] == "\"")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:440
if (result[(n - 1):n] == "\"")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:441
result = result[1:(n - 1)]
}			
}		
}	
_ = n
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:453
func TplResolve(eng *TTemplateEngine, path string) string {
var result string
var k int64
var p string
var v string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:459
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:460
if (path == "@index")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:462
if (eng.PrefixCount > 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:464
result = fmt.Sprintf("%d", eng.Indexes[(eng.PrefixCount - 1)])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:465
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:467
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:469
if (path == ".")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:471
if (eng.PrefixCount > 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:474
p = eng.Prefixes[(eng.PrefixCount - 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:475
result = eng.Scalars[p[0:(int64(len(p)) - 1)]]
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:477
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:480
v = eng.Scalars[path]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:481
if (v != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:483
result = v
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:484
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:487
k = eng.PrefixCount
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:488
for (k > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:490
p = eng.Prefixes[(k - 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:491
v = eng.Scalars[(p + path)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:492
if (v != "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:494
result = v
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:495
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:497
k = (k - 1)
	}
_ = k
_ = p
_ = v
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:503
func TplResolveKey(eng *TTemplateEngine, path string) string {
var result string
var k int64
var p string
var v string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:509
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:510
v = eng.Scalars[path]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:511
if (v != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:513
result = path
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:514
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:516
k = eng.PrefixCount
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:517
for (k > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:519
p = eng.Prefixes[(k - 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:520
v = eng.Scalars[(p + path)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:521
if (v != "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:523
result = (p + path)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:524
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:526
k = (k - 1)
	}
_ = k
_ = p
_ = v
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:535
func TplResolveListLen(eng *TTemplateEngine, name string) int64 {
var result int64
var k int64
var p string
var v int64
var s string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:542
result = eng.ListLens[name]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:543
if (result > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:544
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:549
s = eng.Scalars[(name + "#len")]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:550
if (s != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:552
result = func() int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:553
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:555
k = eng.PrefixCount
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:556
for (k > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:558
p = eng.Prefixes[(k - 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:559
v = eng.ListLens[(p + name)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:560
if (v > 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:562
result = v
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:563
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:565
s = eng.Scalars[((p + name) + "#len")]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:566
if (s != "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:568
result = func() int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:569
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:571
k = (k - 1)
	}
_ = k
_ = p
_ = v
_ = s
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:581
func TplApplyOneFilter(s string, spec string) string {
var result string
var name string
var arg string
var colon int64
var sep int64
var a string
var b string
var n int64
var i int64
var c string
var prevLetter bool
var first string
var rest string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:596
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:597
colon = TplFindFrom(spec, ":", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:598
if (colon < 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:600
name = spec
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:601
arg = ""
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:605
name = TplTrim(spec[0:colon])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:606
arg = TplTrim(spec[(colon + 1):int64(len(spec))])
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:609
if (name == "upper")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:611
result = strings.ToUpper(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:612
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:614
if (name == "lower")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:616
result = strings.ToLower(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:617
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:619
if (name == "trim")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:621
result = TplTrim(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:622
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:624
if (name == "escape")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:626
result = TplEscape(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:627
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:629
if (name == "length")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:631
result = fmt.Sprintf("%d", int64(len(s)))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:632
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:634
if (name == "nl2br")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:638
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:639
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:640
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:641
for (i < n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:643
c = s[i:(i + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:644
if (c == "\\")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:646
if TplStartsWith(s, "\n", i)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:648
result = (result + "<br>")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:649
i = (i + 2)
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:653
result = (result + c)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:654
i = (i + 1)
				}
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:659
result = (result + c)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:660
i = (i + 1)
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:663
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:665
if (name == "capitalize")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:667
if (int64(len(s)) == 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:668
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:669
first = s[0:1]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:670
rest = s[1:int64(len(s))]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:671
result = (strings.ToUpper(first) + rest)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:672
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:674
if (name == "title")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:677
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:678
prevLetter = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:679
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:680
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:681
for (i < n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:683
c = s[i:(i + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:684
if TplCharIsLetter(c)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:686
if prevLetter				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:687
result = (result + c)
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:689
result = (result + strings.ToUpper(c))
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:690
prevLetter = true
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:694
result = (result + c)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:695
prevLetter = false
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:697
i = (i + 1)
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:699
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:701
if (name == "default")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:703
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:704
if (s == "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:705
result = arg
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:706
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:708
if (name == "truncate")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:710
n = func() int64 { v, _ := strconv.ParseInt(arg, 10, 64); return v }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:711
if (int64(len(s)) <= n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:713
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:714
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:716
if (n > 3)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:717
result = (s[0:(n - 3)] + "...")
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:719
result = s[0:n]
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:720
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:722
if (name == "replace")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:725
sep = TplFindFrom(arg, ":", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:726
if (sep < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:727
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:728
a = arg[0:sep]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:729
b = arg[(sep + 1):int64(len(arg))]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:730
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:731
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:732
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:733
for (i < n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:735
if TplStartsWith(s, a, i)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:737
result = (result + b)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:738
i = (i + int64(len(a)))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:739
if (int64(len(a)) == 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:740
i = (i + 1)
}				
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:744
result = (result + s[i:(i + 1)])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:745
i = (i + 1)
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:748
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:750
if (name == "wordwrap")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:753
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:754
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:756
if (name == "raw")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:759
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:760
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:763
result = s
_ = name
_ = arg
_ = colon
_ = sep
_ = a
_ = b
_ = n
_ = i
_ = c
_ = prevLetter
_ = first
_ = rest
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:769
func TplHasRawFilter(spec string) bool {
var result bool
var rest string
var pipe int64
var part string
var fname string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:776
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:777
rest = TplTrim(spec)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:778
for (rest != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:780
pipe = TplFindFrom(rest, "|", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:781
if (pipe < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:783
part = TplTrim(rest)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:784
rest = ""
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:788
part = TplTrim(rest[0:pipe])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:789
rest = TplTrim(rest[(pipe + 1):int64(len(rest))])
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:791
if (part == "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:792
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:793
fname = part
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:794
if (TplFindFrom(part, ":", 0) >= 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:795
fname = TplTrim(part[0:TplFindFrom(part, ":", 0)])
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:796
if (fname == "raw")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:798
result = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:799
return result
}		
	}
_ = rest
_ = pipe
_ = part
_ = fname
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:806
func TplApplyFilters(s string, spec string) string {
var result string
var rest string
var pipe int64
var part string
var fname string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:813
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:814
rest = TplTrim(spec)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:815
for (rest != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:817
pipe = TplFindFrom(rest, "|", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:818
if (pipe < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:820
part = TplTrim(rest)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:821
rest = ""
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:825
part = TplTrim(rest[0:pipe])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:826
rest = TplTrim(rest[(pipe + 1):int64(len(rest))])
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:828
if (part == "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:829
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:830
result = TplApplyOneFilter(result, part)
	}
_ = rest
_ = pipe
_ = part
_ = fname
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:841
func TplFindBlockEnd(tpl string, from int64, kind string) int64 {
var result int64
var openTag string
var closeTag string
var depth int64
var pos int64
var o int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:849
result = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:850
openTag = ("{{#" + kind)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:851
closeTag = ("{{/" + kind)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:852
depth = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:853
pos = from
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:854
for (pos < int64(len(tpl)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:856
o = TplFindFrom(tpl, "{{", pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:857
if (o < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:858
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:859
if TplStartsWith(tpl, openTag, o)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:861
depth = (depth + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:862
pos = (o + 2)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:864
if TplStartsWith(tpl, closeTag, o)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:866
depth = (depth - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:867
if (depth == 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:869
result = o
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:870
return result
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:872
pos = (o + 2)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:875
pos = (o + 2)
			}
		}
	}
_ = openTag
_ = closeTag
_ = depth
_ = pos
_ = o
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:881
func TplFindElse(tpl string, from int64, endPos int64) int64 {
var result int64
var depth int64
var pos int64
var o int64
var close int64
var inner string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:889
result = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:890
depth = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:891
pos = from
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:892
for (pos < endPos)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:894
o = TplFindFrom(tpl, "{{", pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:895
if (o < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:896
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:897
if (o >= endPos)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:898
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:899
if TplStartsWith(tpl, "{{#", o)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:901
depth = (depth + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:902
pos = (o + 2)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:904
if TplStartsWith(tpl, "{{/", o)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:906
depth = (depth - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:907
pos = (o + 2)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:911
if (depth == 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:913
close = TplFindFrom(tpl, "}}", o)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:914
if (close < 0)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:915
return result
}					
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:916
if (close <= endPos)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:918
inner = TplTrim(tpl[(o + 2):close])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:919
if (inner == "else")						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:921
result = o
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:922
return result
}						
}					
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:926
pos = (o + 2)
			}
		}
	}
_ = depth
_ = pos
_ = o
_ = close
_ = inner
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:936
func TplRenderErr(eng *TTemplateEngine, msg string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:938
if (eng.LastError == "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:939
eng.LastError = msg
}	
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:944
func TplRenderVar(eng *TTemplateEngine, inner string) string {
var result string
var pipe int64
var path string
var spec string
var val string
var hasRaw bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:952
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:953
pipe = TplFindFrom(inner, "|", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:954
if (pipe < 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:956
path = TplTrim(inner)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:957
spec = ""
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:961
path = TplTrim(inner[0:pipe])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:962
spec = inner[(pipe + 1):int64(len(inner))]
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:964
val = TplResolve(eng, path)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:965
hasRaw = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:966
if (spec != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:968
val = TplApplyFilters(val, spec)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:969
hasRaw = TplHasRawFilter(spec)
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:972
if (hasRaw == false)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:973
val = TplEscape(val)
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:974
result = val
_ = pipe
_ = path
_ = spec
_ = val
_ = hasRaw
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:978
func TplRenderInto(eng *TTemplateEngine, tpl string) string {
var result string
var out string
var pos int64
var n int64
var open int64
var close int64
var inner string
var raw bool
var after int64
var c string
var sp int64
var kw string
var arg string
var endPos int64
var body string
var bodyEnd int64
var elsePos int64
var cond string
var ln int64
var i int64
var listKey string
var prefix string
var child string
var elseEnd int64
var pname string
var psrc string
var lname string
var lsrc string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1008
out = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1009
pos = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1010
n = int64(len(tpl))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1011
for (pos < n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1013
open = TplFindFrom(tpl, "{{", pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1014
if (open < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1016
out = (out + tpl[pos:n])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1017
			break
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1019
out = (out + tpl[pos:open])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1021
if TplStartsWith(tpl, "{{{", open)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1023
close = TplFindFrom(tpl, "}}}", (open + 3))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1024
if (close < 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1026
TplRenderErr(eng, "unclosed {{{ tag")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1027
				break
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1029
inner = TplTrim(tpl[(open + 3):close])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1030
raw = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1031
after = (close + 3)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1035
close = TplFindFrom(tpl, "}}", (open + 2))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1036
if (close < 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1038
TplRenderErr(eng, "unclosed {{ tag")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1039
				break
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1041
inner = TplTrim(tpl[(open + 2):close])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1042
raw = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1043
after = (close + 2)
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1046
if (inner == "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1048
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1049
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1052
c = inner[0:1]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1053
if (c == "!")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1056
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1057
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1060
if (c == "/")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1062
TplRenderErr(eng, (("unexpected close tag {{" + inner) + "}}"))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1063
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1064
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1067
if (c == ">")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1071
pname = TplCleanName(inner[1:int64(len(inner))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1072
psrc = eng.Templates[pname]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1073
if (psrc == "")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1075
TplRenderErr(eng, ("unknown partial: " + pname))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1076
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1077
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1079
eng.Depth = (eng.Depth + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1080
if (eng.Depth > 32)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1083
eng.Depth = (eng.Depth - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1084
TplRenderErr(eng, ("partial recursion too deep: " + pname))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1085
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1086
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1088
child = TplRenderInto(eng, psrc)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1089
eng.Depth = (eng.Depth - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1090
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1091
if (eng.LastError != "")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1092
pos = n
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1094
pos = after
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1095
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1098
if (c == "<")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1103
lname = TplCleanName(inner[1:int64(len(inner))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1104
lsrc = eng.Templates[lname]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1105
if (lsrc == "")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1107
TplRenderErr(eng, ("unknown layout: " + lname))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1108
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1109
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1111
body = TplRenderInto(eng, tpl[after:n])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1112
if (eng.LastError != "")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1114
out = (out + body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1115
pos = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1116
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1118
eng.Scalars["content"] = body
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1119
eng.Depth = (eng.Depth + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1120
if (eng.Depth > 32)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1122
eng.Depth = (eng.Depth - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1123
eng.Scalars["content"] = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1124
TplRenderErr(eng, ("layout recursion too deep: " + lname))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1125
pos = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1126
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1128
child = TplRenderInto(eng, lsrc)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1129
eng.Depth = (eng.Depth - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1130
eng.Scalars["content"] = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1131
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1132
pos = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1133
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1136
if (c == "#")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1138
sp = TplFindFrom(inner, " ", 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1139
if (sp < 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1141
TplRenderErr(eng, (("missing block name in {{" + inner) + "}}"))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1142
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1143
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1145
kw = TplTrim(inner[1:sp])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1146
arg = TplTrim(inner[(sp + 1):int64(len(inner))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1148
if (kw == "each")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1150
endPos = TplFindBlockEnd(tpl, after, "each")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1151
if (endPos < 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1153
TplRenderErr(eng, "unclosed {{#each}}")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1154
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1155
					continue
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1157
body = tpl[after:endPos]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1158
ln = TplResolveListLen(eng, arg)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1159
listKey = TplResolveKey(eng, arg)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1160
if (listKey == "")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1161
listKey = arg
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1162
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1163
for (i < ln)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1165
prefix = (((listKey + ".") + fmt.Sprintf("%d", i)) + ".")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1166
eng.PushScope(prefix, i)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1167
child = TplRenderInto(eng, body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1168
if (eng.LastError != "")					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1171
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1172
pos = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1173
						continue
}					
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1175
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1176
eng.PopScope()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1177
i = (i + 1)
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1180
close = TplFindFrom(tpl, "}}", endPos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1181
if (close < 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1182
close = n
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1184
close = (close + 2)
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1185
pos = close
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1186
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1189
if (kw == "if")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1191
endPos = TplFindBlockEnd(tpl, after, "if")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1192
if (endPos < 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1194
TplRenderErr(eng, "unclosed {{#if}}")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1195
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1196
					continue
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1198
elsePos = TplFindElse(tpl, after, endPos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1199
cond = TplResolve(eng, arg)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1200
if TplIsTruthy(cond)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1202
if (elsePos >= 0)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1203
body = tpl[after:elsePos]
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1205
body = tpl[after:endPos]
					}
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1209
if (elsePos >= 0)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1212
elseEnd = TplFindFrom(tpl, "}}", elsePos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1213
if (elseEnd < 0)						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1214
elseEnd = endPos
}						 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1216
elseEnd = (elseEnd + 2)
						}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1217
body = tpl[elseEnd:endPos]
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1220
body = ""
					}
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1222
child = TplRenderInto(eng, body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1223
if (eng.LastError != "")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1225
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1226
pos = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1227
					continue
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1229
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1230
close = TplFindFrom(tpl, "}}", endPos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1231
if (close < 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1232
close = n
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1234
close = (close + 2)
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1235
pos = close
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1236
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1239
TplRenderErr(eng, (("unknown block {{#" + kw) + "}}"))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1240
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1241
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1245
out = (out + TplRenderVarRaw(eng, inner, raw))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1246
if (eng.LastError != "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1247
pos = n
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1249
pos = after
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1251
result = out
_ = out
_ = pos
_ = n
_ = open
_ = close
_ = inner
_ = raw
_ = after
_ = c
_ = sp
_ = kw
_ = arg
_ = endPos
_ = body
_ = bodyEnd
_ = elsePos
_ = cond
_ = ln
_ = i
_ = listKey
_ = prefix
_ = child
_ = elseEnd
_ = pname
_ = psrc
_ = lname
_ = lsrc
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1257
func TplRenderVarRaw(eng *TTemplateEngine, inner string, raw bool) string {
var result string
var spec string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1261
if raw	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1263
if (TplFindFrom(inner, "|", 0) >= 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1264
result = TplRenderVar(eng, (inner + "|raw"))
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1266
result = TplRenderVar(eng, (inner + " | raw"))
		}
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1269
result = TplRenderVar(eng, inner)
	}
_ = spec
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1277
func RenderTemplate(eng *TTemplateEngine, tpl string) (string, error) {
var out string
var msg string
_ = out
_ = msg
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1282
out = eng.RenderString(tpl)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1283
msg = eng.ErrorMsg()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1284
if (msg != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1285
return "", errors.New(msg)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1286
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1288
return out, nil
}

var eng *TTemplateEngine
var out string
var err error
func main() {
//line example63_template_layout.klx:20
eng = func() *TTemplateEngine { v := &TTemplateEngine{}; v.Create(); return v }()
//line example63_template_layout.klx:23
eng.AddVar("title", "Admin")
//line example63_template_layout.klx:24
eng.AddVar("nav", "Home")
//line example63_template_layout.klx:25
eng.AddVar("page", "Dashboard")
//line example63_template_layout.klx:26
eng.AddVar("v", "kylix")
//line example63_template_layout.klx:29
eng.AddTemplate("base", "<html><title>{{ title }}</title><body>{{ nav }} {{{ content }}} <footer>v0.9</footer></body></html>")
//line example63_template_layout.klx:30
out, err = RenderTemplate(eng, "{{< base}}<h1>{{ page }}</h1>")
//line example63_template_layout.klx:31
if (err != nil)	 {
//line example63_template_layout.klx:32
fmt.Println("1 layout: ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:34
fmt.Println("1 layout: ", out)
	}
//line example63_template_layout.klx:37
out, err = RenderTemplate(eng, "{{< layout base}}X")
//line example63_template_layout.klx:38
if (err != nil)	 {
//line example63_template_layout.klx:39
fmt.Println("2 alias: ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:41
fmt.Println("2 alias: ", out)
	}
//line example63_template_layout.klx:44
out, err = RenderTemplate(eng, "c=[{{{ content }}}]")
//line example63_template_layout.klx:45
if (err != nil)	 {
//line example63_template_layout.klx:46
fmt.Println("3 cleared: ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:48
fmt.Println("3 cleared: ", out)
	}
//line example63_template_layout.klx:51
eng.AddTemplate("row", "<li>{{ @index }}:{{ . }}</li>")
//line example63_template_layout.klx:52
eng.BeginList("tags")
//line example63_template_layout.klx:53
eng.AddItem("go")
//line example63_template_layout.klx:54
eng.AddItem("llvm")
//line example63_template_layout.klx:55
eng.AddItem("pascal")
//line example63_template_layout.klx:56
eng.EndList()
//line example63_template_layout.klx:57
out, err = RenderTemplate(eng, "<ul>{{#each tags}}{{> row}}{{/each}}</ul>")
//line example63_template_layout.klx:58
if (err != nil)	 {
//line example63_template_layout.klx:59
fmt.Println("4 partial: ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:61
fmt.Println("4 partial: ", out)
	}
//line example63_template_layout.klx:64
eng.AddTemplate("usercard", "<div class=\"card\">{{ name }}</div>")
//line example63_template_layout.klx:65
eng.BeginList("users")
//line example63_template_layout.klx:66
eng.ItemField("name", "Li")
//line example63_template_layout.klx:67
eng.NextItem()
//line example63_template_layout.klx:68
eng.ItemField("name", "Yu")
//line example63_template_layout.klx:69
eng.EndList()
//line example63_template_layout.klx:70
out, err = RenderTemplate(eng, "{{#each users}}{{> usercard}}{{/each}}")
//line example63_template_layout.klx:71
if (err != nil)	 {
//line example63_template_layout.klx:72
fmt.Println("5 scoped: ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:74
fmt.Println("5 scoped: ", out)
	}
//line example63_template_layout.klx:77
out, err = RenderTemplate(eng, "a {{> nothere}} b")
//line example63_template_layout.klx:78
if (err != nil)	 {
//line example63_template_layout.klx:79
fmt.Println("6 unknown: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:81
fmt.Println("6 unknown: ok [", out, "]")
	}
//line example63_template_layout.klx:84
out, err = RenderTemplate(eng, "{{< ghost}}body")
//line example63_template_layout.klx:85
if (err != nil)	 {
//line example63_template_layout.klx:86
fmt.Println("7 unknown: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:88
fmt.Println("7 unknown: ok [", out, "]")
	}
//line example63_template_layout.klx:91
eng.AddTemplate("loop", "x{{> loop}}")
//line example63_template_layout.klx:92
out, err = RenderTemplate(eng, "{{> loop}}")
//line example63_template_layout.klx:93
if (err != nil)	 {
//line example63_template_layout.klx:94
fmt.Println("8 deep: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:96
fmt.Println("8 deep: ok (guard failed) [", out, "]")
	}
//line example63_template_layout.klx:99
eng.AddTemplate("cell", "<td>{{ v | upper }}</td>")
//line example63_template_layout.klx:100
out, err = RenderTemplate(eng, "{{< base}}<table>{{> cell}}</table>")
//line example63_template_layout.klx:101
if (err != nil)	 {
//line example63_template_layout.klx:102
fmt.Println("9 filters: ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example63_template_layout.klx:104
fmt.Println("9 filters: ", out)
	}
}
