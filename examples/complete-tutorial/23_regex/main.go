package main

import (
	"fmt"
)

type TRegexParser struct {
Pat string
I int64
Frag string
Ranges string
Failed bool
}

func (self *TRegexParser) Fail() {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:210
self.Failed = true
}

func (self *TRegexParser) ParseClassAtom() {
var neg bool
var off int64
var cnt int64
var code int64
var ch string
var c2 string
var c3 string
_ = neg
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:221
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:222
neg = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:223
if ((self.I < int64(len(self.Pat))) && (self.Pat[self.I:(self.I + 1)] == "^"))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:225
neg = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:226
self.I = (self.I + 1)
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:228
if ((self.I >= int64(len(self.Pat))) || (self.Pat[self.I:(self.I + 1)] == "]"))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:230
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:231
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:233
off = (int64(len(self.Ranges)) / 8)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:234
cnt = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:235
for (self.I < int64(len(self.Pat)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:237
ch = self.Pat[self.I:(self.I + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:238
if (ch == "]")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:240
self.Frag = (self.Frag + REEncOp(1, off, cnt))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:241
if neg			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:242
self.Frag = REReplaceInstr(self.Frag, ((int64(len(self.Frag)) / 9) - 1), 2, off, cnt)
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:243
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:244
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:246
if (ch == "\\")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:248
if ((self.I + 1) >= int64(len(self.Pat)))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:250
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:251
return
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:253
c2 = self.Pat[(self.I + 1):(self.I + 2)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:254
if (((c2 == "d") || (c2 == "w")) || (c2 == "s"))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:256
self.Ranges = (self.Ranges + REEscRanges(c2))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:257
cnt = (cnt + 1)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:261
if (c2 == "n")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:261
code = 10
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:262
if (c2 == "t")					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:262
code = 9
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:263
if (c2 == "r")						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:263
code = 13
}						 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:264
if (c2 == "f")							 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:264
code = 12
}							 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:265
if (c2 == "v")								 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:265
code = 11
}								 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:266
code = func() int64 { if len(c2) == 0 { return 0 }; return int64(c2[0]) }()
								}
							}
						}
					}
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:267
self.Ranges = REAddRange(self.Ranges, code, code)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:268
cnt = (cnt + 1)
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:270
self.I = (self.I + 2)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:274
code = func() int64 { if len(ch) == 0 { return 0 }; return int64(ch[0]) }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:276
if ((((self.I + 2) < int64(len(self.Pat))) && (self.Pat[(self.I + 1):(self.I + 2)] == "-")) && (self.Pat[(self.I + 2):(self.I + 3)] != "]"))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:278
c3 = self.Pat[(self.I + 2):(self.I + 3)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:279
if (c3 == "\\")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:281
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:282
return
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:284
if (func() int64 { if len(c3) == 0 { return 0 }; return int64(c3[0]) }() < code)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:286
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:287
return
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:289
self.Ranges = REAddRange(self.Ranges, code, func() int64 { if len(c3) == 0 { return 0 }; return int64(c3[0]) }())
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:290
self.I = (self.I + 3)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:294
self.Ranges = REAddRange(self.Ranges, code, code)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:295
self.I = (self.I + 1)
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:297
cnt = (cnt + 1)
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:300
self.Fail()
}

func (self *TRegexParser) ParseAtom() {
var ch string
var nxt string
var code int64
var off int64
var n int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:309
if self.Failed	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:310
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:311
if (self.I >= int64(len(self.Pat)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:313
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:314
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:316
ch = self.Pat[self.I:(self.I + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:318
if (ch == "(")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:321
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:322
self.ParseAlt()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:323
if self.Failed		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:324
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:325
if ((self.I >= int64(len(self.Pat))) || (self.Pat[self.I:(self.I + 1)] != ")"))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:327
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:328
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:330
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:331
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:334
if (ch == "[")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:336
self.ParseClassAtom()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:337
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:340
if (ch == ".")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:343
off = (int64(len(self.Ranges)) / 8)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:344
self.Ranges = REAddRange(self.Ranges, 10, 10)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:345
self.Frag = (self.Frag + REEncOp(2, off, 1))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:346
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:347
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:350
if (ch == "^")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:352
self.Frag = (self.Frag + REEncOp(7, 0, 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:353
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:354
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:356
if (ch == "$")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:358
self.Frag = (self.Frag + REEncOp(8, 0, 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:359
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:360
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:363
if (ch == "\\")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:365
if ((self.I + 1) >= int64(len(self.Pat)))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:367
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:368
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:370
nxt = self.Pat[(self.I + 1):(self.I + 2)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:371
if ((((((nxt == "d") || (nxt == "w")) || (nxt == "s")) || (nxt == "D")) || (nxt == "W")) || (nxt == "S"))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:373
off = (int64(len(self.Ranges)) / 8)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:374
if (nxt == "d")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:374
self.Ranges = (self.Ranges + REEscRanges("d"))
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:375
if (nxt == "w")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:375
self.Ranges = (self.Ranges + REEscRanges("w"))
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:376
if (nxt == "s")					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:376
self.Ranges = (self.Ranges + REEscRanges("s"))
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:377
if (nxt == "D")						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:377
self.Ranges = (self.Ranges + REEscRanges("d"))
}						 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:378
if (nxt == "W")							 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:378
self.Ranges = (self.Ranges + REEscRanges("w"))
}							 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:379
self.Ranges = (self.Ranges + REEscRanges("s"))
							}
						}
					}
				}
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:380
n = ((int64(len(self.Ranges)) / 8) - off)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:381
if (((nxt == "D") || (nxt == "W")) || (nxt == "S"))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:382
self.Frag = (self.Frag + REEncOp(2, off, n))
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:384
self.Frag = (self.Frag + REEncOp(1, off, n))
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:385
self.I = (self.I + 2)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:386
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:388
if (nxt == "n")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:388
code = 10
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:389
if (nxt == "t")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:389
code = 9
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:390
if (nxt == "r")				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:390
code = 13
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:391
if (nxt == "f")					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:391
code = 12
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:392
if (nxt == "v")						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:392
code = 11
}						 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:393
if (nxt == "0")							 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:393
code = 0
}							 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:394
code = func() int64 { if len(nxt) == 0 { return 0 }; return int64(nxt[0]) }()
							}
						}
					}
				}
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:395
self.Frag = (self.Frag + REEncOp(0, code, 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:396
self.I = (self.I + 2)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:397
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:401
if (((((ch == "|") || (ch == ")")) || (ch == "*")) || (ch == "+")) || (ch == "?"))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:403
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:404
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:408
self.Frag = (self.Frag + REEncOp(0, func() int64 { if len(ch) == 0 { return 0 }; return int64(ch[0]) }(), 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:409
self.I = (self.I + 1)
}

func (self *TRegexParser) ParseRep() {
var base int64
var bodyCnt int64
var cnt int64
var m2 int64
var k int64
var cur int64
var closeI int64
var q string
var hasJ bool
var hasM bool
var body string
_ = q
_ = body
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:421
if self.Failed	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:422
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:423
base = (int64(len(self.Frag)) / 9)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:424
self.ParseAtom()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:425
if self.Failed	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:426
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:427
bodyCnt = ((int64(len(self.Frag)) / 9) - base)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:428
if (self.I >= int64(len(self.Pat)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:429
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:430
q = self.Pat[self.I:(self.I + 1)]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:432
if (((q == "*") || (q == "+")) || (q == "?"))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:434
hasM = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:435
if (((self.I + 1) < int64(len(self.Pat))) && (self.Pat[(self.I + 1):(self.I + 2)] == "?"))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:437
hasM = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:438
self.I = (self.I + 1)
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:440
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:442
if (q == "*")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:445
self.Frag = ((self.Frag[0:(base * 9)] + REEncOp(6, (base + 1), ((base + bodyCnt) + 2))) + self.Frag[(base * 9):int64(len(self.Frag))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:446
self.Frag = RERebaseJumps(self.Frag, (base + 1), ((base + bodyCnt) + 1), 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:447
self.Frag = (self.Frag + REEncOp(5, base, 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:448
if hasM			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:449
self.Frag = REReplaceInstr(self.Frag, base, 6, ((base + bodyCnt) + 2), (base + 1))
}			
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:451
if (q == "+")			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:454
self.Frag = (self.Frag + REEncOp(6, base, ((base + bodyCnt) + 1)))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:455
if hasM				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:456
self.Frag = REReplaceInstr(self.Frag, (base + bodyCnt), 6, ((base + bodyCnt) + 1), base)
}				
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:462
self.Frag = ((self.Frag[0:(base * 9)] + REEncOp(6, (base + 1), ((base + 1) + bodyCnt))) + self.Frag[(base * 9):int64(len(self.Frag))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:463
self.Frag = RERebaseJumps(self.Frag, (base + 1), ((base + bodyCnt) + 1), 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:464
if hasM				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:465
self.Frag = REReplaceInstr(self.Frag, base, 6, ((base + 1) + bodyCnt), (base + 1))
}				
			}
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:467
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:470
if (q == "{")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:473
k = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:474
cnt = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:475
for (((k < int64(len(self.Pat))) && (func() int64 { if len(self.Pat[k:(k + 1)]) == 0 { return 0 }; return int64(self.Pat[k:(k + 1)][0]) }() >= func() int64 { if len("0") == 0 { return 0 }; return int64("0"[0]) }())) && (func() int64 { if len(self.Pat[k:(k + 1)]) == 0 { return 0 }; return int64(self.Pat[k:(k + 1)][0]) }() <= func() int64 { if len("9") == 0 { return 0 }; return int64("9"[0]) }()))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:477
cnt = ((cnt * 10) + (func() int64 { if len(self.Pat[k:(k + 1)]) == 0 { return 0 }; return int64(self.Pat[k:(k + 1)][0]) }() - func() int64 { if len("0") == 0 { return 0 }; return int64("0"[0]) }()))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:478
k = (k + 1)
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:480
if (k == (self.I + 1))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:483
self.Frag = (self.Frag + REEncOp(0, func() int64 { if len("{") == 0 { return 0 }; return int64("{"[0]) }(), 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:484
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:485
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:487
body = self.Frag[(base * 9):int64(len(self.Frag))]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:488
hasJ = REHasJumps(body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:489
if (hasJ || (cnt > 200))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:491
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:492
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:494
if ((k < int64(len(self.Pat))) && (self.Pat[k:(k + 1)] == "}"))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:497
closeI = k
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:498
if (cnt == 0)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:499
self.Frag = self.Frag[0:(base * 9)]
}			
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:500
k = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:501
for (k < cnt)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:503
self.Frag = (self.Frag + body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:504
k = (k + 1)
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:506
self.I = (closeI + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:507
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:509
if ((k < int64(len(self.Pat))) && (self.Pat[k:(k + 1)] == ","))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:511
k = (k + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:512
m2 = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:513
hasM = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:514
for (((k < int64(len(self.Pat))) && (func() int64 { if len(self.Pat[k:(k + 1)]) == 0 { return 0 }; return int64(self.Pat[k:(k + 1)][0]) }() >= func() int64 { if len("0") == 0 { return 0 }; return int64("0"[0]) }())) && (func() int64 { if len(self.Pat[k:(k + 1)]) == 0 { return 0 }; return int64(self.Pat[k:(k + 1)][0]) }() <= func() int64 { if len("9") == 0 { return 0 }; return int64("9"[0]) }()))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:516
m2 = ((m2 * 10) + (func() int64 { if len(self.Pat[k:(k + 1)]) == 0 { return 0 }; return int64(self.Pat[k:(k + 1)][0]) }() - func() int64 { if len("0") == 0 { return 0 }; return int64("0"[0]) }()))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:517
k = (k + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:518
hasM = true
			}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:520
if ((k < int64(len(self.Pat))) && (self.Pat[k:(k + 1)] == "}"))			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:522
if ((m2 < cnt) || ((cnt + m2) > 400))				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:524
self.Fail()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:525
return
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:527
closeI = k
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:528
if (cnt == 0)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:529
self.Frag = self.Frag[0:(base * 9)]
}				
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:531
k = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:532
for (k < cnt)				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:534
self.Frag = (self.Frag + body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:535
k = (k + 1)
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:537
cur = (base + (cnt * bodyCnt))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:538
if hasM				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:541
k = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:542
for (k < (m2 - cnt))					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:544
self.Frag = ((self.Frag + REEncOp(6, (cur + 1), ((cur + 1) + bodyCnt))) + body)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:545
cur = ((cur + 1) + bodyCnt)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:546
k = (k + 1)
					}
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:552
self.Frag = (((self.Frag + REEncOp(6, (cur + 1), ((cur + bodyCnt) + 2))) + body) + REEncOp(5, cur, 0))
				}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:554
self.I = (closeI + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:555
return
}			
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:559
self.Frag = (self.Frag + REEncOp(0, func() int64 { if len("{") == 0 { return 0 }; return int64("{"[0]) }(), 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:560
self.I = (self.I + 1)
}	
}

func (self *TRegexParser) ParseCat() {
var c0 int64
_ = c0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:570
if self.Failed	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:571
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:572
c0 = (int64(len(self.Frag)) / 9)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:573
for (!self.Failed)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:575
if (self.I >= int64(len(self.Pat)))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:576
			break
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:577
if ((self.Pat[self.I:(self.I + 1)] == "|") || (self.Pat[self.I:(self.I + 1)] == ")"))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:578
			break
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:579
self.ParseRep()
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:581
if self.Failed	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:582
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:583
if ((int64(len(self.Frag)) / 9) == c0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:584
self.Frag = (self.Frag + REEncOp(4, 0, 0))
}	
}

func (self *TRegexParser) ParseAlt() {
var a0 int64
var c1 int64
var jIdx int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:594
if self.Failed	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:595
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:596
a0 = (int64(len(self.Frag)) / 9)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:597
self.ParseCat()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:598
if self.Failed	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:599
return
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:600
if ((self.I < int64(len(self.Pat))) && (self.Pat[self.I:(self.I + 1)] == "|"))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:602
self.I = (self.I + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:603
c1 = ((int64(len(self.Frag)) / 9) - a0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:605
self.Frag = ((self.Frag[0:(a0 * 9)] + REEncOp(6, (a0 + 1), ((a0 + c1) + 2))) + self.Frag[(a0 * 9):int64(len(self.Frag))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:606
self.Frag = RERebaseJumps(self.Frag, (a0 + 1), ((a0 + c1) + 1), 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:608
jIdx = ((a0 + c1) + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:609
self.Frag = (self.Frag + REEncOp(5, 0, 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:610
self.ParseAlt()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:611
if self.Failed		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:612
return
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:613
self.Frag = REReplaceInstr(self.Frag, jIdx, 5, (int64(len(self.Frag)) / 9), 0)
}	
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:45
func REHexNib(v int64) string {
var result string
var tab string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:49
tab = "0123456789abcdef"
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:50
result = tab[v:(v + 1)]
_ = tab
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:54
func REValNib(c string) int64 {
var result int64
var o int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:58
o = func() int64 { if len(c) == 0 { return 0 }; return int64(c[0]) }()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:59
if (o <= func() int64 { if len("9") == 0 { return 0 }; return int64("9"[0]) }())	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:60
result = (o - func() int64 { if len("0") == 0 { return 0 }; return int64("0"[0]) }())
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:62
result = ((o - func() int64 { if len("a") == 0 { return 0 }; return int64("a"[0]) }()) + 10)
	}
_ = o
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:66
func REEnc(v int64) string {
var result string
var r int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:70
r = v
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:71
result = (((REHexNib(((r / 4096) % 16)) + REHexNib(((r / 256) % 16))) + REHexNib(((r / 16) % 16))) + REHexNib((r % 16)))
_ = r
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:75
func REDec(s4 string) int64 {
var result int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:77
result = ((((REValNib(s4[0:1]) * 4096) + (REValNib(s4[1:2]) * 256)) + (REValNib(s4[2:3]) * 16)) + REValNib(s4[3:4]))
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:85
func REEncOp(op, a, b int64) string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:87
result = ((REHexNib(op) + REEnc(a)) + REEnc(b))
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:90
func REGetOp(prog string, i int64) int64 {
var result int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:92
result = REValNib(prog[(i * 9):((i * 9) + 1)])
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:95
func REGetA(prog string, i int64) int64 {
var result int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:97
result = REDec(prog[((i * 9) + 1):((i * 9) + 5)])
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:100
func REGetB(prog string, i int64) int64 {
var result int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:102
result = REDec(prog[((i * 9) + 5):((i * 9) + 9)])
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:106
func REReplaceInstr(f string, i, op, a, b int64) string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:108
if (((i * 9) + 9) >= int64(len(f)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:109
result = (f[0:(i * 9)] + REEncOp(op, a, b))
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:111
result = ((f[0:(i * 9)] + REEncOp(op, a, b)) + f[((i * 9) + 9):int64(len(f))])
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:117
func RERebaseJumps(f string, fromI, toI, delta int64) string {
var result string
var k int64
var op int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:121
result = f
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:122
k = fromI
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:123
for (k < toI)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:125
op = REGetOp(result, k)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:126
if (op == 5)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:127
result = REReplaceInstr(result, k, 5, (REGetA(result, k) + delta), 0)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:128
if (op == 6)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:129
result = REReplaceInstr(result, k, 6, (REGetA(result, k) + delta), (REGetB(result, k) + delta))
}			
		}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:130
k = (k + 1)
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:135
func REHasJumps(frag string) bool {
var result bool
var k int64
var cnt int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:139
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:140
cnt = (int64(len(frag)) / 9)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:141
k = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:142
for (k < cnt)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:144
if ((REGetOp(frag, k) == 5) || (REGetOp(frag, k) == 6))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:146
result = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:147
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:149
k = (k + 1)
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:156
func REAddRange(ranges string, lo, hi int64) string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:158
result = ((ranges + REEnc(lo)) + REEnc(hi))
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:163
func REInRanges(ranges string, off, n, c int64) bool {
var result bool
var k int64
var lo int64
var hi int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:167
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:168
k = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:169
for (k < n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:171
lo = REDec(ranges[((off + k) * 8):(((off + k) * 8) + 4)])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:172
hi = REDec(ranges[(((off + k) * 8) + 4):(((off + k) * 8) + 8)])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:173
if ((c >= lo) && (c <= hi))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:175
result = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:176
return result
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:178
k = (k + 1)
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:184
func REEscRanges(kind string) string {
var result string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:186
if (kind == "d")	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:187
result = REAddRange("", func() int64 { if len("0") == 0 { return 0 }; return int64("0"[0]) }(), func() int64 { if len("9") == 0 { return 0 }; return int64("9"[0]) }())
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:188
if (kind == "w")		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:189
result = REAddRange(REAddRange(REAddRange(REAddRange("", func() int64 { if len("0") == 0 { return 0 }; return int64("0"[0]) }(), func() int64 { if len("9") == 0 { return 0 }; return int64("9"[0]) }()), func() int64 { if len("A") == 0 { return 0 }; return int64("A"[0]) }(), func() int64 { if len("Z") == 0 { return 0 }; return int64("Z"[0]) }()), func() int64 { if len("_") == 0 { return 0 }; return int64("_"[0]) }(), func() int64 { if len("_") == 0 { return 0 }; return int64("_"[0]) }()), func() int64 { if len("a") == 0 { return 0 }; return int64("a"[0]) }(), func() int64 { if len("z") == 0 { return 0 }; return int64("z"[0]) }())
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:192
result = (REAddRange(REAddRange(REAddRange(REAddRange(REAddRange("", 9, 9), 10, 10), 11, 11), 12, 12), 13, 13) + REAddRange("", func() int64 { if len(" ") == 0 { return 0 }; return int64(" "[0]) }(), func() int64 { if len(" ") == 0 { return 0 }; return int64(" "[0]) }()))
		}
	}
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:623
func RECompile(pattern string) (string, string) {
var p *TRegexParser
var prog string
var rr string
_ = p
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:628
prog = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:629
rr = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:630
if (int64(len(pattern)) > 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:632
p = &TRegexParser{}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:633
p.Pat = pattern
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:634
p.I = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:635
p.Frag = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:636
p.Ranges = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:637
p.Failed = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:638
p.ParseAlt()
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:639
if p.Failed		 {
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:645
if (p.I < int64(len(pattern)))			 {
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:651
prog = (p.Frag + REEncOp(4, 0, 0))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:652
rr = p.Ranges
			}
		}
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:656
return prog, rr
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:666
func REFindAt(prog, ranges, s string, start int64) (int64, int64) {
var n int64
var pcCnt int64
var sp int64
var pc int64
var pos int64
var op int64
var a int64
var b int64
var midx int64
var me int64
var alive bool
var hit bool
var stkPC []int64
var stkPos []int64
var memo []int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:672
me = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:673
n = int64(len(s))
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:674
if (start <= n)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:676
pcCnt = (int64(len(prog)) / 9)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:677
stkPC = append(stkPC, 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:678
stkPos = append(stkPos, start)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:679
sp = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:680
for ((sp >= 0) && (me < 0))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:682
pc = stkPC[sp]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:683
pos = stkPos[sp]
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:684
sp = (sp - 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:685
alive = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:686
for alive			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:688
if ((pc >= pcCnt) || (pos > n))				 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:690
alive = false
}				 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:694
midx = ((pc * (n + 1)) + pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:695
for (int64(len(memo)) <= midx)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:696
memo = append(memo, 0)
					}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:697
if (memo[midx] == 1)					 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:699
alive = false
}					 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:709
memo[midx] = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:710
op = REGetOp(prog, pc)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:711
if (op == 0)						 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:714
if ((pos < n) && (func() int64 { if len(s[pos:(pos + 1)]) == 0 { return 0 }; return int64(s[pos:(pos + 1)][0]) }() == REGetA(prog, pc)))							 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:716
pos = (pos + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:717
pc = (pc + 1)
}							 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:721
memo[midx] = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:722
alive = false
							}
}						 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:725
if (op == 1)							 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:728
a = REGetA(prog, pc)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:729
b = REGetB(prog, pc)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:730
hit = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:731
if (pos < n)								 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:732
hit = REInRanges(ranges, a, b, func() int64 { if len(s[pos:(pos + 1)]) == 0 { return 0 }; return int64(s[pos:(pos + 1)][0]) }())
}								
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:733
if hit								 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:735
pos = (pos + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:736
pc = (pc + 1)
}								 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:740
memo[midx] = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:741
alive = false
								}
}							 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:744
if (op == 2)								 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:747
a = REGetA(prog, pc)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:748
b = REGetB(prog, pc)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:749
hit = true
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:750
if (pos < n)									 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:751
hit = (!REInRanges(ranges, a, b, func() int64 { if len(s[pos:(pos + 1)]) == 0 { return 0 }; return int64(s[pos:(pos + 1)][0]) }()))
}									
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:752
if hit									 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:754
pos = (pos + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:755
pc = (pc + 1)
}									 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:759
memo[midx] = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:760
alive = false
									}
}								 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:763
if (op == 3)									 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:766
if (pos < n)										 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:768
pos = (pos + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:769
pc = (pc + 1)
}										 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:773
memo[midx] = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:774
alive = false
										}
}									 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:777
if (op == 4)										 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:780
me = pos
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:781
alive = false
}										 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:783
if (op == 5)											 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:786
pc = REGetA(prog, pc)
}											 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:788
if (op == 6)												 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:791
b = REGetB(prog, pc)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:792
sp = (sp + 1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:793
for (int64(len(stkPC)) <= sp)													 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:795
stkPC = append(stkPC, 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:796
stkPos = append(stkPos, 0)
													}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:798
stkPC[sp] = b
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:799
stkPos[sp] = pos
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:800
pc = REGetA(prog, pc)
}												 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:802
if (op == 7)													 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:805
if (pos == 0)														 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:806
pc = (pc + 1)
}														 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:809
memo[midx] = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:810
alive = false
														}
}													 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:813
if (op == 8)														 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:816
if (pos == n)															 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:817
pc = (pc + 1)
}															 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:820
memo[midx] = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:821
alive = false
															}
}														 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:826
memo[midx] = 1
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:827
alive = false
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:834
if (me >= 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:835
return start, me
}	 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:837
return (-1), (-1)
	}
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:842
func RESearch(prog, ranges, s string, from int64) (int64, int64) {
var st int64
var ms int64
var me int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:846
ms = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:847
me = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:848
st = from
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:849
for ((st <= int64(len(s))) && (ms < 0))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:851
ms, me = REFindAt(prog, ranges, s, st)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:852
if (ms < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:853
st = (st + 1)
}		
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:855
return ms, me
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:861
func RegexMatch(pattern, s string) bool {
var result bool
var prog string
var ranges string
var ms int64
var me int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:866
result = false
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:867
prog, ranges = RECompile(pattern)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:868
if (int64(len(prog)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:869
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:870
ms, me = RESearch(prog, ranges, s, 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:871
result = ((ms >= 0) && (me >= 0))
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:875
func RegexFind(pattern, s string) string {
var result string
var prog string
var ranges string
var ms int64
var me int64
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:880
result = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:881
prog, ranges = RECompile(pattern)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:882
if (int64(len(prog)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:883
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:884
ms, me = RESearch(prog, ranges, s, 0)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:885
if (ms >= 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:886
result = s[ms:me]
}	
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:892
func RegexFindAll(pattern, s string) []string {
var result []string
var prog string
var ranges string
var ms int64
var me int64
var pos int64
var prevEnd int64
var arr []string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:898
result = arr
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:899
prog, ranges = RECompile(pattern)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:900
if (int64(len(prog)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:901
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:902
pos = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:903
prevEnd = (-1)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:904
for (pos <= int64(len(s)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:906
ms, me = RESearch(prog, ranges, s, pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:907
if (ms < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:908
			break
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:909
if ((me == ms) && (ms == prevEnd))		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:912
prevEnd = me
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:913
pos = (ms + 1)
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:917
arr = append(arr, s[ms:me])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:918
prevEnd = me
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:919
if (me == ms)			 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:920
pos = (ms + 1)
}			 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:922
pos = me
			}
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:925
result = arr
_ = arr
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:930
func RegexReplace(pattern, s, repl string) string {
var result string
var prog string
var ranges string
var ms int64
var me int64
var pos int64
var out string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:936
result = s
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:937
prog, ranges = RECompile(pattern)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:938
if (int64(len(prog)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:939
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:940
out = ""
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:941
pos = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:942
for (pos <= int64(len(s)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:944
ms, me = RESearch(prog, ranges, s, pos)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:945
if (ms < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:946
			break
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:947
if (me > ms)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:949
out = ((out + s[pos:ms]) + repl)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:950
pos = me
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:955
pos = (ms + 1)
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:958
result = (out + s[pos:int64(len(s))])
_ = out
return result
}

//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:963
func RegexSplit(pattern, s string) []string {
var result []string
var prog string
var ranges string
var ms int64
var me int64
var pos int64
var searchFrom int64
var arr []string
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:969
result = arr
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:970
prog, ranges = RECompile(pattern)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:971
if (int64(len(prog)) == 0)	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:973
result = append(result, s)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:974
return result
}	
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:976
pos = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:977
searchFrom = 0
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:978
for (searchFrom <= int64(len(s)))	 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:980
ms, me = RESearch(prog, ranges, s, searchFrom)
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:981
if (ms < 0)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:982
			break
}		
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:983
if (me > ms)		 {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:985
arr = append(arr, s[pos:ms])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:986
pos = me
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:987
searchFrom = me
}		 else {
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:991
searchFrom = (ms + 1)
		}
	}
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:994
arr = append(arr, s[pos:int64(len(s))])
//line /Users/astra/Documents/ai/learn/kylix/examples/complete-tutorial/../../stdlib/regex_engine.klx:995
result = arr
_ = arr
return result
}

//line example61_regex_engine.klx:18
func ShowMatch(pat, s string) {
var f string
//line example61_regex_engine.klx:22
if RegexMatch(pat, s)	 {
//line example61_regex_engine.klx:24
f = RegexFind(pat, s)
//line example61_regex_engine.klx:25
fmt.Println(((((("T " + pat) + " | ") + s) + " | ") + f))
}	 else {
//line example61_regex_engine.klx:28
fmt.Println(((("F " + pat) + " | ") + s))
	}
_ = f
}

//line example61_regex_engine.klx:31
func ShowFindAll(pat, s string) {
var a []string
var i int64
var out string
//line example61_regex_engine.klx:37
a = RegexFindAll(pat, s)
//line example61_regex_engine.klx:38
out = "["
//line example61_regex_engine.klx:39
i = 0
//line example61_regex_engine.klx:40
for (i < int64(len(a)))	 {
//line example61_regex_engine.klx:42
if (i > 0)		 {
//line example61_regex_engine.klx:43
out = (out + ",")
}		
//line example61_regex_engine.klx:44
out = (out + a[i])
//line example61_regex_engine.klx:45
i = (i + 1)
	}
//line example61_regex_engine.klx:47
fmt.Println((((((("A " + pat) + " | ") + s) + " | ") + out) + "]"))
_ = a
_ = i
_ = out
}

//line example61_regex_engine.klx:50
func ShowReplace(pat, s, repl string) {
//line example61_regex_engine.klx:52
fmt.Println(((((("R " + pat) + " | ") + s) + " | ") + RegexReplace(pat, s, repl)))
}

//line example61_regex_engine.klx:55
func ShowSplit(pat, s string) {
var a []string
var i int64
var out string
//line example61_regex_engine.klx:61
a = RegexSplit(pat, s)
//line example61_regex_engine.klx:62
out = "["
//line example61_regex_engine.klx:63
i = 0
//line example61_regex_engine.klx:64
for (i < int64(len(a)))	 {
//line example61_regex_engine.klx:66
if (i > 0)		 {
//line example61_regex_engine.klx:67
out = (out + ",")
}		
//line example61_regex_engine.klx:68
out = (out + a[i])
//line example61_regex_engine.klx:69
i = (i + 1)
	}
//line example61_regex_engine.klx:71
fmt.Println((((((("S " + pat) + " | ") + s) + " | ") + out) + "]"))
_ = a
_ = i
_ = out
}

func main() {
//line example61_regex_engine.klx:76
ShowMatch("abc", "xxabcyy")
//line example61_regex_engine.klx:77
ShowMatch("a.c", "abc")
//line example61_regex_engine.klx:78
ShowMatch("a[0-9]c", "a7c")
//line example61_regex_engine.klx:80
ShowMatch("^ab", "xab")
//line example61_regex_engine.klx:81
ShowMatch("^ab", "abx")
//line example61_regex_engine.klx:82
ShowMatch("ab$", "abx")
//line example61_regex_engine.klx:83
ShowMatch("ab$", "xab")
//line example61_regex_engine.klx:85
ShowMatch("\\d\\d\\d", "ab123")
//line example61_regex_engine.klx:86
ShowMatch("\\w\\w\\w", "ab1")
//line example61_regex_engine.klx:87
ShowMatch("\\D\\D", "12ab")
//line example61_regex_engine.klx:88
ShowMatch("\\s", "ab c")
//line example61_regex_engine.klx:90
ShowMatch("cat|dog", "hotdog")
//line example61_regex_engine.klx:91
ShowMatch("cat|dog", "hotbird")
//line example61_regex_engine.klx:92
ShowMatch("(ab)+", "xababab")
//line example61_regex_engine.klx:94
ShowMatch("(a|b)*abb", "babb")
//line example61_regex_engine.klx:95
ShowMatch("(a|b)*abb", "abxa")
//line example61_regex_engine.klx:96
ShowMatch("(a*)*b", "aaab")
//line example61_regex_engine.klx:98
ShowMatch("a{2,3}", "aaaa")
//line example61_regex_engine.klx:99
ShowMatch("a{2}", "a")
//line example61_regex_engine.klx:100
ShowMatch("colou?r", "color")
//line example61_regex_engine.klx:101
ShowMatch("colou?r", "colour")
//line example61_regex_engine.klx:103
ShowMatch("a+?", "aaa")
//line example61_regex_engine.klx:105
ShowMatch("[a-c]+", "zabcd")
//line example61_regex_engine.klx:106
ShowMatch("[^a-c]+", "abcz12")
//line example61_regex_engine.klx:108
ShowFindAll("\\d+", "a1b22c333")
//line example61_regex_engine.klx:109
ShowFindAll("\\w+", "hello world foo")
//line example61_regex_engine.klx:110
ShowFindAll("[a-c]+", "abcdc")
//line example61_regex_engine.klx:111
ShowFindAll("a*", "bab")
//line example61_regex_engine.klx:113
ShowReplace("\\d+", "a1b22c333", "#")
//line example61_regex_engine.klx:114
ShowReplace("cat", "hotcat concat", "dog")
//line example61_regex_engine.klx:115
ShowSplit(",", "a,b,c")
//line example61_regex_engine.klx:116
ShowSplit("\\s+", "a b  c")
//line example61_regex_engine.klx:117
fmt.Println("done")
}
