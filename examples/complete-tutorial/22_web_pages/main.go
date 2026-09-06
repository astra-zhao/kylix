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
}

func (self *TTemplateEngine) PushScope(prefix string, idx int64) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:59
if (self.PrefixCount < int64(len(self.Prefixes)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:60
self.Prefixes[self.PrefixCount] = prefix
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:62
self.Prefixes = append(self.Prefixes, prefix)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:63
if (self.PrefixCount < int64(len(self.Indexes)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:64
self.Indexes[self.PrefixCount] = idx
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:66
self.Indexes = append(self.Indexes, idx)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:67
self.PrefixCount = (self.PrefixCount + 1)
}

func (self *TTemplateEngine) PopScope() {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:72
if (self.PrefixCount > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:73
self.PrefixCount = (self.PrefixCount - 1)
}	
}

func (self *TTemplateEngine) Create() {
var tmpMap map[string]string = map[string]string{}
var tmpMapI map[string]int64 = map[string]int64{}
var tmpArr []string
var tmpArrI []int64
_ = tmpMap
_ = tmpMapI
_ = tmpArr
_ = tmpArrI
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:84
self.Scalars = tmpMap
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:85
self.ListLens = tmpMapI
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:86
self.Prefixes = tmpArr
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:87
self.Indexes = tmpArrI
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:88
self.PrefixCount = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:89
self.CurList = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:90
self.CurIdx = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:91
self.ItemStarted = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:92
self.LastError = ""
}

func (self *TTemplateEngine) AddVar(name string, value string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:100
self.Scalars[name] = value
}

func (self *TTemplateEngine) AddInt(name string, value int64) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:106
self.Scalars[name] = fmt.Sprintf("%d", value)
}

func (self *TTemplateEngine) AddVariant(name string, value interface{}) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:115
self.Scalars[name] = fmt.Sprintf("%v", value)
}

func (self *TTemplateEngine) SetContext(m map[string]string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:124
self.Scalars = m
}

func (self *TTemplateEngine) AddListLen(name string, count int64) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:133
self.ListLens[name] = count
}

func (self *TTemplateEngine) BeginList(name string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:139
self.CurList = name
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:140
self.CurIdx = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:141
self.ItemStarted = false
}

func (self *TTemplateEngine) AddItem(value string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:148
self.Scalars[((self.CurList + ".") + fmt.Sprintf("%d", self.CurIdx))] = value
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:149
self.CurIdx = (self.CurIdx + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:150
self.ItemStarted = false
}

func (self *TTemplateEngine) BeginItem() {
}

func (self *TTemplateEngine) ItemField(key string, value string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:169
self.Scalars[((((self.CurList + ".") + fmt.Sprintf("%d", self.CurIdx)) + ".") + key)] = value
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:170
self.ItemStarted = true
}

func (self *TTemplateEngine) NextItem() {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:176
if self.ItemStarted	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:177
self.CurIdx = (self.CurIdx + 1)
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:178
self.ItemStarted = false
}

func (self *TTemplateEngine) EndList() {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:185
if self.ItemStarted	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:186
self.ListLens[self.CurList] = (self.CurIdx + 1)
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:188
self.ListLens[self.CurList] = self.CurIdx
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:189
self.CurList = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:190
self.CurIdx = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:191
self.ItemStarted = false
}

func (self *TTemplateEngine) ListLen(name string) int64 {
var result int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:197
result = self.ListLens[name]
	return result
}

func (self *TTemplateEngine) RenderString(tpl string) string {
var result string
var out string
_ = out
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:212
self.LastError = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:213
self.PrefixCount = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:214
out = TplRenderInto(self, tpl)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:215
result = out
	return result
}

func (self *TTemplateEngine) ErrorMsg() string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:222
result = self.LastError
	return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:232
func TplFindFrom(hay string, needle string, from int64) int64 {
var result int64
var n int64
var m int64
var i int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:238
n = int64(len(hay))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:239
m = int64(len(needle))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:240
result = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:241
i = from
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:242
for ((i + m) <= n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:244
if (hay[i:(i + m)] == needle)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:246
result = i
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:247
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:249
i = (i + 1)
	}
_ = n
_ = m
_ = i
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:254
func TplSlice(hay string, from int64, n int64) string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:256
result = hay[from:(from + n)]
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:260
func TplStartsWith(hay string, needle string, start int64) bool {
var result bool
var part string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:264
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:265
if ((start + int64(len(needle))) > int64(len(hay)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:266
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:267
part = hay[start:(start + int64(len(needle)))]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:268
if (part == needle)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:269
result = true
}	
_ = part
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:273
func TplIsSpace(c string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:275
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:276
if (c == " ")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:277
result = true
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:278
if (c == "\t")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:279
result = true
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:280
if (c == "\n")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:281
result = true
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:282
if (c == "\r")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:283
result = true
}	
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:287
func TplTrim(s string) string {
var result string
var n int64
var start int64
var fin int64
var c string
var running bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:295
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:296
start = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:297
running = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:298
for running	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:300
if (start >= n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:301
running = false
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:304
c = s[start:(start + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:305
if TplIsSpace(c)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:306
start = (start + 1)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:308
running = false
			}
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:311
fin = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:312
running = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:313
for running	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:315
if (fin <= start)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:316
running = false
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:319
c = s[(fin - 1):fin]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:320
if TplIsSpace(c)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:321
fin = (fin - 1)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:323
running = false
			}
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:326
result = s[start:fin]
_ = n
_ = start
_ = fin
_ = c
_ = running
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:330
func TplEscape(s string) string {
var result string
var i int64
var n int64
var c string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:336
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:337
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:338
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:339
for (i < n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:341
c = s[i:(i + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:342
if (c == "&")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:343
result = (result + "&amp;")
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:344
if (c == "<")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:345
result = (result + "&lt;")
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:346
if (c == ">")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:347
result = (result + "&gt;")
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:348
if (c == "\"")					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:349
result = (result + "&quot;")
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:350
if (func() int64 { if len(c) == 0 { return 0 }; return int64(c[0]) }() == 39)						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:353
result = (result + "&#39;")
}						 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:355
result = (result + c)
						}
					}
				}
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:356
i = (i + 1)
	}
_ = i
_ = n
_ = c
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:363
func TplCharIsLetter(c string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:365
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:366
if (c >= "a")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:368
if (c <= "z")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:369
result = true
}		
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:371
if (result == false)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:373
if (c >= "A")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:375
if (c <= "Z")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:376
result = true
}			
}		
}	
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:382
func TplIsTruthy(s string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:384
result = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:385
if (s == "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:386
result = false
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:387
if (s == "0")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:388
result = false
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:389
if (s == "false")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:390
result = false
}	
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:400
func TplResolve(eng *TTemplateEngine, path string) string {
var result string
var k int64
var p string
var v string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:406
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:407
if (path == "@index")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:409
if (eng.PrefixCount > 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:411
result = fmt.Sprintf("%d", eng.Indexes[(eng.PrefixCount - 1)])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:412
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:414
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:416
if (path == ".")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:418
if (eng.PrefixCount > 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:421
p = eng.Prefixes[(eng.PrefixCount - 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:422
result = eng.Scalars[p[0:(int64(len(p)) - 1)]]
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:424
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:427
v = eng.Scalars[path]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:428
if (v != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:430
result = v
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:431
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:434
k = eng.PrefixCount
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:435
for (k > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:437
p = eng.Prefixes[(k - 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:438
v = eng.Scalars[(p + path)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:439
if (v != "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:441
result = v
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:442
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:444
k = (k - 1)
	}
_ = k
_ = p
_ = v
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:450
func TplResolveKey(eng *TTemplateEngine, path string) string {
var result string
var k int64
var p string
var v string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:456
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:457
v = eng.Scalars[path]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:458
if (v != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:460
result = path
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:461
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:463
k = eng.PrefixCount
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:464
for (k > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:466
p = eng.Prefixes[(k - 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:467
v = eng.Scalars[(p + path)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:468
if (v != "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:470
result = (p + path)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:471
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:473
k = (k - 1)
	}
_ = k
_ = p
_ = v
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:482
func TplResolveListLen(eng *TTemplateEngine, name string) int64 {
var result int64
var k int64
var p string
var v int64
var s string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:489
result = eng.ListLens[name]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:490
if (result > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:491
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:496
s = eng.Scalars[(name + "#len")]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:497
if (s != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:499
result = func() int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:500
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:502
k = eng.PrefixCount
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:503
for (k > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:505
p = eng.Prefixes[(k - 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:506
v = eng.ListLens[(p + name)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:507
if (v > 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:509
result = v
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:510
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:512
s = eng.Scalars[((p + name) + "#len")]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:513
if (s != "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:515
result = func() int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:516
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:518
k = (k - 1)
	}
_ = k
_ = p
_ = v
_ = s
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:528
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
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:543
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:544
colon = TplFindFrom(spec, ":", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:545
if (colon < 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:547
name = spec
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:548
arg = ""
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:552
name = TplTrim(spec[0:colon])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:553
arg = TplTrim(spec[(colon + 1):int64(len(spec))])
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:556
if (name == "upper")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:558
result = strings.ToUpper(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:559
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:561
if (name == "lower")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:563
result = strings.ToLower(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:564
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:566
if (name == "trim")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:568
result = TplTrim(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:569
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:571
if (name == "escape")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:573
result = TplEscape(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:574
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:576
if (name == "length")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:578
result = fmt.Sprintf("%d", int64(len(s)))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:579
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:581
if (name == "nl2br")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:585
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:586
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:587
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:588
for (i < n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:590
c = s[i:(i + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:591
if (c == "\\")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:593
if TplStartsWith(s, "\n", i)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:595
result = (result + "<br>")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:596
i = (i + 2)
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:600
result = (result + c)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:601
i = (i + 1)
				}
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:606
result = (result + c)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:607
i = (i + 1)
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:610
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:612
if (name == "capitalize")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:614
if (int64(len(s)) == 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:615
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:616
first = s[0:1]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:617
rest = s[1:int64(len(s))]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:618
result = (strings.ToUpper(first) + rest)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:619
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:621
if (name == "title")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:624
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:625
prevLetter = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:626
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:627
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:628
for (i < n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:630
c = s[i:(i + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:631
if TplCharIsLetter(c)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:633
if prevLetter				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:634
result = (result + c)
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:636
result = (result + strings.ToUpper(c))
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:637
prevLetter = true
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:641
result = (result + c)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:642
prevLetter = false
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:644
i = (i + 1)
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:646
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:648
if (name == "default")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:650
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:651
if (s == "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:652
result = arg
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:653
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:655
if (name == "truncate")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:657
n = func() int64 { v, _ := strconv.ParseInt(arg, 10, 64); return v }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:658
if (int64(len(s)) <= n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:660
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:661
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:663
if (n > 3)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:664
result = (s[0:(n - 3)] + "...")
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:666
result = s[0:n]
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:667
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:669
if (name == "replace")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:672
sep = TplFindFrom(arg, ":", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:673
if (sep < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:674
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:675
a = arg[0:sep]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:676
b = arg[(sep + 1):int64(len(arg))]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:677
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:678
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:679
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:680
for (i < n)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:682
if TplStartsWith(s, a, i)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:684
result = (result + b)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:685
i = (i + int64(len(a)))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:686
if (int64(len(a)) == 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:687
i = (i + 1)
}				
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:691
result = (result + s[i:(i + 1)])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:692
i = (i + 1)
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:695
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:697
if (name == "wordwrap")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:700
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:701
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:703
if (name == "raw")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:706
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:707
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:710
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

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:716
func TplHasRawFilter(spec string) bool {
var result bool
var rest string
var pipe int64
var part string
var fname string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:723
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:724
rest = TplTrim(spec)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:725
for (rest != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:727
pipe = TplFindFrom(rest, "|", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:728
if (pipe < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:730
part = TplTrim(rest)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:731
rest = ""
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:735
part = TplTrim(rest[0:pipe])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:736
rest = TplTrim(rest[(pipe + 1):int64(len(rest))])
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:738
if (part == "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:739
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:740
fname = part
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:741
if (TplFindFrom(part, ":", 0) >= 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:742
fname = TplTrim(part[0:TplFindFrom(part, ":", 0)])
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:743
if (fname == "raw")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:745
result = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:746
return result
}		
	}
_ = rest
_ = pipe
_ = part
_ = fname
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:753
func TplApplyFilters(s string, spec string) string {
var result string
var rest string
var pipe int64
var part string
var fname string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:760
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:761
rest = TplTrim(spec)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:762
for (rest != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:764
pipe = TplFindFrom(rest, "|", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:765
if (pipe < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:767
part = TplTrim(rest)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:768
rest = ""
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:772
part = TplTrim(rest[0:pipe])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:773
rest = TplTrim(rest[(pipe + 1):int64(len(rest))])
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:775
if (part == "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:776
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:777
result = TplApplyOneFilter(result, part)
	}
_ = rest
_ = pipe
_ = part
_ = fname
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:788
func TplFindBlockEnd(tpl string, from int64, kind string) int64 {
var result int64
var openTag string
var closeTag string
var depth int64
var pos int64
var o int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:796
result = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:797
openTag = ("{{#" + kind)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:798
closeTag = ("{{/" + kind)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:799
depth = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:800
pos = from
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:801
for (pos < int64(len(tpl)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:803
o = TplFindFrom(tpl, "{{", pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:804
if (o < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:805
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:806
if TplStartsWith(tpl, openTag, o)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:808
depth = (depth + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:809
pos = (o + 2)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:811
if TplStartsWith(tpl, closeTag, o)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:813
depth = (depth - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:814
if (depth == 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:816
result = o
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:817
return result
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:819
pos = (o + 2)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:822
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

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:828
func TplFindElse(tpl string, from int64, endPos int64) int64 {
var result int64
var depth int64
var pos int64
var o int64
var close int64
var inner string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:836
result = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:837
depth = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:838
pos = from
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:839
for (pos < endPos)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:841
o = TplFindFrom(tpl, "{{", pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:842
if (o < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:843
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:844
if (o >= endPos)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:845
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:846
if TplStartsWith(tpl, "{{#", o)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:848
depth = (depth + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:849
pos = (o + 2)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:851
if TplStartsWith(tpl, "{{/", o)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:853
depth = (depth - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:854
pos = (o + 2)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:858
if (depth == 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:860
close = TplFindFrom(tpl, "}}", o)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:861
if (close < 0)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:862
return result
}					
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:863
if (close <= endPos)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:865
inner = TplTrim(tpl[(o + 2):close])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:866
if (inner == "else")						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:868
result = o
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:869
return result
}						
}					
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:873
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

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:883
func TplRenderErr(eng *TTemplateEngine, msg string) {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:885
if (eng.LastError == "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:886
eng.LastError = msg
}	
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:891
func TplRenderVar(eng *TTemplateEngine, inner string) string {
var result string
var pipe int64
var path string
var spec string
var val string
var hasRaw bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:899
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:900
pipe = TplFindFrom(inner, "|", 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:901
if (pipe < 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:903
path = TplTrim(inner)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:904
spec = ""
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:908
path = TplTrim(inner[0:pipe])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:909
spec = inner[(pipe + 1):int64(len(inner))]
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:911
val = TplResolve(eng, path)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:912
hasRaw = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:913
if (spec != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:915
val = TplApplyFilters(val, spec)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:916
hasRaw = TplHasRawFilter(spec)
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:919
if (hasRaw == false)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:920
val = TplEscape(val)
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:921
result = val
_ = pipe
_ = path
_ = spec
_ = val
_ = hasRaw
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:925
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
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:951
out = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:952
pos = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:953
n = int64(len(tpl))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:954
for (pos < n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:956
open = TplFindFrom(tpl, "{{", pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:957
if (open < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:959
out = (out + tpl[pos:n])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:960
			break
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:962
out = (out + tpl[pos:open])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:964
if TplStartsWith(tpl, "{{{", open)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:966
close = TplFindFrom(tpl, "}}}", (open + 3))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:967
if (close < 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:969
TplRenderErr(eng, "unclosed {{{ tag")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:970
				break
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:972
inner = TplTrim(tpl[(open + 3):close])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:973
raw = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:974
after = (close + 3)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:978
close = TplFindFrom(tpl, "}}", (open + 2))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:979
if (close < 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:981
TplRenderErr(eng, "unclosed {{ tag")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:982
				break
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:984
inner = TplTrim(tpl[(open + 2):close])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:985
raw = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:986
after = (close + 2)
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:989
if (inner == "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:991
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:992
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:995
c = inner[0:1]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:996
if (c == "!")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:999
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1000
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1003
if (c == "/")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1005
TplRenderErr(eng, (("unexpected close tag {{" + inner) + "}}"))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1006
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1007
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1010
if (c == "#")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1012
sp = TplFindFrom(inner, " ", 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1013
if (sp < 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1015
TplRenderErr(eng, (("missing block name in {{" + inner) + "}}"))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1016
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1017
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1019
kw = TplTrim(inner[1:sp])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1020
arg = TplTrim(inner[(sp + 1):int64(len(inner))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1022
if (kw == "each")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1024
endPos = TplFindBlockEnd(tpl, after, "each")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1025
if (endPos < 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1027
TplRenderErr(eng, "unclosed {{#each}}")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1028
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1029
					continue
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1031
body = tpl[after:endPos]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1032
ln = TplResolveListLen(eng, arg)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1033
listKey = TplResolveKey(eng, arg)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1034
if (listKey == "")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1035
listKey = arg
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1036
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1037
for (i < ln)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1039
prefix = (((listKey + ".") + fmt.Sprintf("%d", i)) + ".")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1040
eng.PushScope(prefix, i)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1041
child = TplRenderInto(eng, body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1042
if (eng.LastError != "")					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1045
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1046
pos = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1047
						continue
}					
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1049
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1050
eng.PopScope()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1051
i = (i + 1)
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1054
close = TplFindFrom(tpl, "}}", endPos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1055
if (close < 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1056
close = n
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1058
close = (close + 2)
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1059
pos = close
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1060
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1063
if (kw == "if")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1065
endPos = TplFindBlockEnd(tpl, after, "if")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1066
if (endPos < 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1068
TplRenderErr(eng, "unclosed {{#if}}")
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1069
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1070
					continue
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1072
elsePos = TplFindElse(tpl, after, endPos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1073
cond = TplResolve(eng, arg)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1074
if TplIsTruthy(cond)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1076
if (elsePos >= 0)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1077
body = tpl[after:elsePos]
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1079
body = tpl[after:endPos]
					}
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1083
if (elsePos >= 0)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1086
elseEnd = TplFindFrom(tpl, "}}", elsePos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1087
if (elseEnd < 0)						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1088
elseEnd = endPos
}						 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1090
elseEnd = (elseEnd + 2)
						}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1091
body = tpl[elseEnd:endPos]
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1094
body = ""
					}
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1096
child = TplRenderInto(eng, body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1097
if (eng.LastError != "")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1099
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1100
pos = n
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1101
					continue
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1103
out = (out + child)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1104
close = TplFindFrom(tpl, "}}", endPos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1105
if (close < 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1106
close = n
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1108
close = (close + 2)
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1109
pos = close
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1110
				continue
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1113
TplRenderErr(eng, (("unknown block {{#" + kw) + "}}"))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1114
pos = after
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1115
			continue
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1119
out = (out + TplRenderVarRaw(eng, inner, raw))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1120
if (eng.LastError != "")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1121
pos = n
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1123
pos = after
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1125
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
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1131
func TplRenderVarRaw(eng *TTemplateEngine, inner string, raw bool) string {
var result string
var spec string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1135
if raw	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1137
if (TplFindFrom(inner, "|", 0) >= 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1138
result = TplRenderVar(eng, (inner + "|raw"))
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1140
result = TplRenderVar(eng, (inner + " | raw"))
		}
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1143
result = TplRenderVar(eng, inner)
	}
_ = spec
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1151
func RenderTemplate(eng *TTemplateEngine, tpl string) (string, error) {
var out string
var msg string
_ = out
_ = msg
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1156
out = eng.RenderString(tpl)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1157
msg = eng.ErrorMsg()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1158
if (msg != "")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1159
return "", errors.New(msg)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1160
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/template_engine.klx:1162
return out, nil
}

var eng *TTemplateEngine
var page string
var err error
//line example59_template.klx:15
func Show(title string, tpl string) {
var eng *TTemplateEngine
var out string
var err error
//line example59_template.klx:21
eng = func() *TTemplateEngine { v := &TTemplateEngine{}; v.Create(); return v }()
//line example59_template.klx:22
eng.AddVar("name", "<b>World</b>")
//line example59_template.klx:23
eng.AddVar("title", "The Kylix Report")
//line example59_template.klx:24
eng.AddVar("user.name", "Li Lei")
//line example59_template.klx:25
eng.AddVar("user.email", "li@kylix.dev")
//line example59_template.klx:26
out, err = RenderTemplate(eng, tpl)
//line example59_template.klx:27
if (err != nil)	 {
//line example59_template.klx:28
fmt.Println(title, " ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example59_template.klx:30
fmt.Println(title, " [", out, "]")
	}
_ = eng
_ = out
_ = err
}

func main() {
//line example59_template.klx:40
Show("escape", "Hello {{ name }}!")
//line example59_template.klx:41
Show("raw", "Hello {{{ name }}}!")
//line example59_template.klx:44
Show("dots", "{{ user.name }} <{{ user.email }}>")
//line example59_template.klx:47
Show("upper", "{{ title | upper }}")
//line example59_template.klx:48
Show("chain", "{{ title | lower | capitalize }}")
//line example59_template.klx:49
Show("trunc", "{{ title | truncate:11 }}")
//line example59_template.klx:50
Show("deflt", "{{ missing | default:N/A }}")
//line example59_template.klx:51
Show("len", "len={{ title | length }}")
//line example59_template.klx:54
eng = func() *TTemplateEngine { v := &TTemplateEngine{}; v.Create(); return v }()
//line example59_template.klx:55
eng.AddVar("title", "Team")
//line example59_template.klx:56
eng.BeginList("users")
//line example59_template.klx:57
eng.ItemField("name", "Li Lei")
//line example59_template.klx:58
eng.ItemField("role", "admin")
//line example59_template.klx:59
eng.NextItem()
//line example59_template.klx:60
eng.ItemField("name", "Han Mei")
//line example59_template.klx:61
eng.ItemField("role", "dev")
//line example59_template.klx:62
eng.NextItem()
//line example59_template.klx:63
eng.ItemField("name", "Tom")
//line example59_template.klx:64
eng.ItemField("role", "guest")
//line example59_template.klx:65
eng.EndList()
//line example59_template.klx:67
page, err = RenderTemplate(eng, "<h1>{{ title }}</h1>{{#each users}}<li>{{ @index }}: {{ name }} ({{ role }})</li>{{/each}}")
//line example59_template.klx:68
if (err != nil)	 {
//line example59_template.klx:69
fmt.Println("each ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example59_template.klx:71
fmt.Println("each [", page, "]")
	}
//line example59_template.klx:74
eng = func() *TTemplateEngine { v := &TTemplateEngine{}; v.Create(); return v }()
//line example59_template.klx:75
eng.BeginList("tags")
//line example59_template.klx:76
eng.AddItem("go")
//line example59_template.klx:77
eng.AddItem("pascal")
//line example59_template.klx:78
eng.AddItem("llvm")
//line example59_template.klx:79
eng.EndList()
//line example59_template.klx:80
page, err = RenderTemplate(eng, "tags:{{#each tags}} {{ . }}{{/each}}")
//line example59_template.klx:81
if (err != nil)	 {
//line example59_template.klx:82
fmt.Println("tags ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example59_template.klx:84
fmt.Println("tags [", page, "]")
	}
//line example59_template.klx:87
eng = func() *TTemplateEngine { v := &TTemplateEngine{}; v.Create(); return v }()
//line example59_template.klx:88
eng.AddVar("flag", "1")
//line example59_template.klx:89
eng.AddVar("zero", "0")
//line example59_template.klx:90
page, err = RenderTemplate(eng, "{{#if flag}}ON{{else}}OFF{{/if}} {{#if zero}}ON{{else}}OFF{{/if}} {{#if nope}}ON{{else}}OFF{{/if}}")
//line example59_template.klx:91
if (err != nil)	 {
//line example59_template.klx:92
fmt.Println("if ERROR: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example59_template.klx:94
fmt.Println("if [", page, "]")
	}
//line example59_template.klx:97
Show("comment", "a{{! hidden }}b")
//line example59_template.klx:100
Show("error", "broken {{#each users}}")
}
