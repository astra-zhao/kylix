package main

import (
	"kylix/stdlib"
	"fmt"
)

func main() {
//line example48_phase6_net_crypto_encoding.klx:5
fmt.Println(("SHA256: " + stdlib.Sha256("hello")))
//line example48_phase6_net_crypto_encoding.klx:6
fmt.Println(("Base64: " + stdlib.Base64Encode("hello")))
//line example48_phase6_net_crypto_encoding.klx:7
fmt.Println(("Hex: " + stdlib.HexEncode("hello")))
//line example48_phase6_net_crypto_encoding.klx:8
fmt.Println(("MD5: " + stdlib.Md5("abc")))
//line example48_phase6_net_crypto_encoding.klx:9
fmt.Println(("HMAC: " + stdlib.HmacSha256("key", "data")))
//line example48_phase6_net_crypto_encoding.klx:11
hash := func() string { _v, _ := stdlib.BCryptHash("test", 4); return _v }()
//line example48_phase6_net_crypto_encoding.klx:12
if stdlib.BCryptCompare("test", hash)	 {
//line example48_phase6_net_crypto_encoding.klx:13
fmt.Println("BCrypt: OK")
}	 else {
//line example48_phase6_net_crypto_encoding.klx:15
fmt.Println("BCrypt: FAIL")
	}
//line example48_phase6_net_crypto_encoding.klx:17
enc := func() string { _v, _ := stdlib.AesEncrypt("mykey", "hello world"); return _v }()
//line example48_phase6_net_crypto_encoding.klx:18
dec := func() string { _v, _ := stdlib.AesDecrypt("mykey", enc); return _v }()
//line example48_phase6_net_crypto_encoding.klx:19
if (dec == "hello world")	 {
//line example48_phase6_net_crypto_encoding.klx:20
fmt.Println("AES round-trip: OK")
}	 else {
//line example48_phase6_net_crypto_encoding.klx:22
fmt.Println("AES round-trip: FAIL")
	}
//line example48_phase6_net_crypto_encoding.klx:24
fmt.Println("Phase 6 stdlib OK")
}
