package dev.kylix.vocab

/**
 * JNI entry points into libkylixvocab.so. Every string the core returns is
 * copied into a Java string and released with kylix_free inside the C bridge.
 */
object KylixBridge {
    init {
        System.loadLibrary("kylixvocab")
        System.loadLibrary("vocabjni")
    }

    external fun wordsQuery(known: String): String
    external fun reviewQuery(known: String): String
    external fun markPath(): String
    external fun markRequest(known: String, id: String): String
    external fun parse(status: Long, body: String): String
}
