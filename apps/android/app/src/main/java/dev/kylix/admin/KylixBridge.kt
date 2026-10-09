package dev.kylix.admin

/**
 * JNI entry points into libkylixlogic.so. Every string the core returns is
 * copied into a Java string and released with kylix_free inside the C bridge.
 */
object KylixBridge {
    init {
        System.loadLibrary("kylixlogic")
        System.loadLibrary("kylixjni")
    }

    external fun validateLogin(username: String, password: String): String
    external fun loginRequest(username: String, password: String): String
    external fun parseLogin(status: Long, body: String): String
    external fun parseList(status: Long, body: String): String
    external fun authHeader(token: String): String
    external fun loginPath(): String
    external fun notesPath(): String
}
