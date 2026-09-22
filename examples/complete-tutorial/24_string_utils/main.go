package main

import (
	"strings"
	"fmt"
)

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:42
func IndexOf(s, sub string) int64 {
var result int64
var i int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:46
result = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:47
if (int64(len(sub)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:49
result = 0
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:53
if (int64(len(sub)) <= int64(len(s)))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:55
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:56
for (i <= (int64(len(s)) - int64(len(sub))))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:58
if (s[i:(i + int64(len(sub)))] == sub)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:60
result = i
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:61
i = int64(len(s))
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:64
i = (i + 1)
				}
			}
}		
	}
_ = i
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:71
func LastIndexOf(s, sub string) int64 {
var result int64
var i int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:75
result = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:76
if (int64(len(sub)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:78
result = int64(len(s))
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:82
if (int64(len(sub)) <= int64(len(s)))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:84
i = (int64(len(s)) - int64(len(sub)))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:85
for (i >= 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:87
if (s[i:(i + int64(len(sub)))] == sub)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:89
result = i
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:90
i = (-1)
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:93
i = (i - 1)
				}
			}
}		
	}
_ = i
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:100
func Contains(s, sub string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:102
result = (IndexOf(s, sub) >= 0)
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:106
func StartsWith(s, prefix string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:108
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:109
if (int64(len(prefix)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:110
result = true
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:113
if (int64(len(prefix)) <= int64(len(s)))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:115
if (s[0:int64(len(prefix))] == prefix)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:116
result = true
}			
}		
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:122
func EndsWith(s, suffix string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:124
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:125
if (int64(len(suffix)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:126
result = true
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:129
if (int64(len(suffix)) <= int64(len(s)))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:131
if (s[(int64(len(s)) - int64(len(suffix))):int64(len(s))] == suffix)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:132
result = true
}			
}		
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:138
func Count(s, sub string) int64 {
var result int64
var i int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:142
result = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:143
if (int64(len(sub)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:145
result = 0
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:149
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:150
for ((i + int64(len(sub))) <= int64(len(s)))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:152
if (s[i:(i + int64(len(sub)))] == sub)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:154
result = (result + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:155
i = (i + int64(len(sub)))
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:158
i = (i + 1)
			}
		}
	}
_ = i
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:167
func Replace(s, fromS, toS string) string {
var result string
var i int64
var outp string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:172
if (int64(len(fromS)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:174
result = s
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:178
outp = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:179
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:180
for (i < int64(len(s)))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:182
if ((i + int64(len(fromS))) <= int64(len(s)))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:184
if (s[i:(i + int64(len(fromS)))] == fromS)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:186
outp = (outp + toS)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:187
i = (i + int64(len(fromS)))
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:191
outp = (outp + s[i:(i + 1)])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:192
i = (i + 1)
				}
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:197
outp = (outp + s[i:(i + 1)])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:198
i = (i + 1)
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:201
result = outp
	}
_ = i
_ = outp
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:206
func ReverseString(s string) string {
var result string
var i int64
var outp string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:211
outp = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:212
i = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:213
for (i > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:215
i = (i - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:216
outp = (outp + s[i:(i + 1)])
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:218
result = outp
_ = i
_ = outp
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:222
func DupeString(s string, n int64) string {
var result string
var i int64
var outp string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:227
outp = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:228
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:229
for (i < n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:231
outp = (outp + s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:232
i = (i + 1)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:234
result = outp
_ = i
_ = outp
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:239
func Substring(s string, start int64, clampLen int64) string {
var result string
var b int64
var e int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:243
b = start
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:244
if (b < 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:245
b = 0
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:246
if (b > int64(len(s)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:247
b = int64(len(s))
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:248
e = (b + clampLen)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:249
if (e < b)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:250
e = b
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:251
if (e > int64(len(s)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:252
e = int64(len(s))
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:253
result = s[b:e]
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:260
func SUAt(s string, i int64) string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:262
if ((i >= 0) && (i < int64(len(s))))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:263
result = s[i:(i + 1)]
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:265
result = " "
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:271
func SUIsSpace(c string) bool {
var result bool
var o int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:275
o = func() int64 { if len(c) == 0 { return 0 }; return int64(c[0]) }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:276
result = ((((o == 32) || (o == 9)) || (o == 13)) || (o == 10))
_ = o
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:280
func TrimLeft(s string) string {
var result string
var i int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:284
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:287
for ((i < int64(len(s))) && SUIsSpace(SUAt(s, i)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:288
i = (i + 1)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:289
result = s[i:int64(len(s))]
_ = i
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:293
func TrimRight(s string) string {
var result string
var e int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:297
e = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:298
for ((e > 0) && SUIsSpace(SUAt(s, (e - 1))))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:299
e = (e - 1)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:300
result = s[0:e]
_ = e
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:304
func Trim(s string) string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:306
result = TrimLeft(TrimRight(s))
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:310
func IsBlank(s string) bool {
var result bool
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:312
result = (int64(len(Trim(s))) == 0)
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:319
func PadStart(s string, width int64, pad string) string {
var result string
var outp string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:323
if (int64(len(pad)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:324
pad = " "
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:325
outp = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:326
for (int64(len(outp)) < width)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:327
outp = (pad + outp)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:328
result = outp
_ = outp
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:333
func PadEnd(s string, width int64, pad string) string {
var result string
var outp string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:337
if (int64(len(pad)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:338
pad = " "
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:339
outp = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:340
for (int64(len(outp)) < width)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:341
outp = (outp + pad)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:342
result = outp
_ = outp
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:351
func Split(s, sep string) []string {
var result []string
var arr []string
var i int64
var seg int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:356
result = arr
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:357
if (int64(len(s)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:359
result = arr
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:363
if (int64(len(sep)) == 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:365
arr = append(arr, s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:366
result = arr
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:370
seg = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:371
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:372
for (i <= (int64(len(s)) - int64(len(sep))))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:374
if (s[i:(i + int64(len(sep)))] == sep)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:376
arr = append(arr, s[seg:i])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:377
i = (i + int64(len(sep)))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:378
seg = i
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:381
i = (i + 1)
				}
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:383
arr = append(arr, s[seg:int64(len(s))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:384
result = arr
		}
	}
_ = arr
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:391
func SplitWhitespace(s string) []string {
var result []string
var arr []string
var i int64
var seg int64
var outp string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:397
result = arr
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:398
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:399
seg = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:400
for (i < int64(len(s)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:402
if SUIsSpace(SUAt(s, i))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:404
if (seg >= 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:406
arr = append(arr, s[seg:i])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:407
seg = (-1)
}			
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:412
if (seg < 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:413
seg = i
}			
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:415
i = (i + 1)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:417
if (seg >= 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:419
outp = s[seg:int64(len(s))]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:420
arr = append(arr, outp)
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:422
result = arr
_ = arr
_ = outp
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:426
func Join(parts []string, sep string) string {
var result string
var i int64
var outp string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:431
outp = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:432
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:433
for (i < int64(len(parts)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:435
if (i > 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:436
outp = (outp + sep)
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:437
outp = (outp + parts[i])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:438
i = (i + 1)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:440
result = outp
_ = i
_ = outp
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:446
func Capitalize(s string) string {
var result string
var first string
var rest string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:450
if (int64(len(s)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:452
result = ""
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:456
first = strings.ToUpper(s[0:1])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:457
rest = strings.ToLower(s[1:int64(len(s))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:458
result = (first + rest)
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:464
func TitleCase(s string) string {
var result string
var arr []string
var i int64
var outp string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:470
arr = SplitWhitespace(s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:471
outp = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:472
i = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:473
for (i < int64(len(arr)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:475
if (i > 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:476
outp = (outp + " ")
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:477
outp = (outp + Capitalize(arr[i]))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:478
i = (i + 1)
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/stringutil.klx:480
result = outp
_ = arr
_ = i
_ = outp
return result
}

//line example62_string_utils.klx:19
func Show(label_, v string) {
//line example62_string_utils.klx:21
fmt.Println((((label_ + " |") + v) + "|"))
}

func main() {
//line example62_string_utils.klx:26
Show("Trim", Trim("  hello world  "))
//line example62_string_utils.klx:27
Show("TrimLeft", TrimLeft("  hello  "))
//line example62_string_utils.klx:28
Show("TrimRight", TrimRight("  hello  "))
//line example62_string_utils.klx:29
if IsBlank("     ")	 {
//line example62_string_utils.klx:30
fmt.Println("IsBlank |yes|")
}	 else {
//line example62_string_utils.klx:32
fmt.Println("IsBlank |no|")
	}
//line example62_string_utils.klx:33
if IsBlank(Trim("  x "))	 {
//line example62_string_utils.klx:34
fmt.Println("IsBlank x |yes|")
}	 else {
//line example62_string_utils.klx:36
fmt.Println("IsBlank x |no|")
	}
//line example62_string_utils.klx:39
if StartsWith("Kylix.pas", "Kylix")	 {
//line example62_string_utils.klx:40
fmt.Println("StartsWith |yes|")
}	 else {
//line example62_string_utils.klx:42
fmt.Println("StartsWith |no|")
	}
//line example62_string_utils.klx:43
if EndsWith("report_2026.pdf", ".pdf")	 {
//line example62_string_utils.klx:44
fmt.Println("EndsWith |yes|")
}	 else {
//line example62_string_utils.klx:46
fmt.Println("EndsWith |no|")
	}
//line example62_string_utils.klx:47
if Contains("the quick fox", "quick")	 {
//line example62_string_utils.klx:48
fmt.Println("Contains |yes|")
}	 else {
//line example62_string_utils.klx:50
fmt.Println("Contains |no|")
	}
//line example62_string_utils.klx:51
fmt.Println((("IndexOf |" + fmt.Sprintf("%d", IndexOf("banana", "an"))) + "|"))
//line example62_string_utils.klx:52
fmt.Println((("LastIndexOf |" + fmt.Sprintf("%d", LastIndexOf("banana", "an"))) + "|"))
//line example62_string_utils.klx:53
fmt.Println((("IndexOf missing |" + fmt.Sprintf("%d", IndexOf("banana", "xy"))) + "|"))
//line example62_string_utils.klx:54
fmt.Println((("Count |" + fmt.Sprintf("%d", Count("a-b-a-b-a", "a-b"))) + "|"))
//line example62_string_utils.klx:57
Show("Replace", Replace("one fish, two fish", "fish", "cat"))
//line example62_string_utils.klx:58
Show("Reverse", ReverseString("kylix"))
//line example62_string_utils.klx:59
Show("Dupe", DupeString("ab", 3))
//line example62_string_utils.klx:60
Show("Substring", Substring("abcdefgh", 2, 3))
//line example62_string_utils.klx:61
Show("Substring clamp", Substring("abc", 2, 99))
//line example62_string_utils.klx:64
fmt.Println(((PadEnd("id", 6, " ") + PadEnd("name", 10, " ")) + "score"))
//line example62_string_utils.klx:65
fmt.Println(((PadEnd("7", 6, " ") + PadEnd("alice", 10, " ")) + PadStart("91", 5, " ")))
//line example62_string_utils.klx:66
fmt.Println(((PadEnd("42", 6, " ") + PadEnd("bob", 10, " ")) + PadStart("7", 5, " ")))
//line example62_string_utils.klx:67
Show("PadStart zero", PadStart("42", 5, "0"))
//line example62_string_utils.klx:70
parts := Split("2026-09-11", "-")
//line example62_string_utils.klx:71
fmt.Println((("Split len |" + fmt.Sprintf("%d", int64(len(parts)))) + "|"))
//line example62_string_utils.klx:72
fmt.Println((((((("y |" + parts[0]) + "| m |") + parts[1]) + "| d |") + parts[2]) + "|"))
//line example62_string_utils.klx:73
fmt.Println((("Join |" + Join(parts, "/")) + "|"))
//line example62_string_utils.klx:74
ws := SplitWhitespace("  keep   the  change ")
//line example62_string_utils.klx:75
fmt.Println((("Words |" + fmt.Sprintf("%d", int64(len(ws)))) + "|"))
//line example62_string_utils.klx:76
fmt.Println((("Joined words |" + Join(ws, "+")) + "|"))
//line example62_string_utils.klx:77
csv := Split("a,,b", ",")
//line example62_string_utils.klx:78
fmt.Println(((("Empty middle |" + fmt.Sprintf("%d", int64(len(csv)))) + "| ") + Join(csv, ";")))
//line example62_string_utils.klx:81
Show("Capitalize", Capitalize("kYLIX lang"))
//line example62_string_utils.klx:82
Show("TitleCase", TitleCase("the quick brown fox"))
//line example62_string_utils.klx:85
row := Split("a,b,c", ",")
//line example62_string_utils.klx:86
fmt.Println((("Rebuilt |" + Join(row, ",")) + "|"))
}
