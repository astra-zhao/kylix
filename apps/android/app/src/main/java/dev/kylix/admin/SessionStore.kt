package dev.kylix.admin

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey

/**
 * Access and refresh tokens for a cold start. EncryptedSharedPreferences
 * (AES256-GCM values, AES256-SIV keys) when the Android Keystore can build a
 * master key. If that fails, this store is a no-op and the process keeps the
 * pair in memory only — the activity must not crash.
 */
class SessionStore(context: Context) {
    data class Session(
        val server: String,
        val token: String,
        val refreshToken: String,
        val accessExpiresAtMs: Long
    )

    private val prefs: SharedPreferences? = openPrefs(context.applicationContext)

    fun load(): Session? {
        val p = prefs ?: return null
        val refresh = p.getString(K_REFRESH, "") ?: ""
        if (refresh.isEmpty()) {
            return null
        }
        return Session(
            server = p.getString(K_SERVER, "") ?: "",
            token = p.getString(K_TOKEN, "") ?: "",
            refreshToken = refresh,
            accessExpiresAtMs = p.getLong(K_EXP, 0L)
        )
    }

    fun save(server: String, token: String, refreshToken: String, accessExpiresAtMs: Long) {
        val p = prefs ?: return
        p.edit()
            .putString(K_SERVER, server)
            .putString(K_TOKEN, token)
            .putString(K_REFRESH, refreshToken)
            .putLong(K_EXP, accessExpiresAtMs)
            .apply()
    }

    fun clear() {
        prefs?.edit()?.clear()?.apply()
    }

    private companion object {
        const val FILE = "kylix_session"
        const val K_SERVER = "server"
        const val K_TOKEN = "token"
        const val K_REFRESH = "refresh_token"
        const val K_EXP = "access_expires_at"

        fun openPrefs(context: Context): SharedPreferences? {
            return try {
                val master = MasterKey.Builder(context)
                    .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
                    .build()
                EncryptedSharedPreferences.create(
                    context,
                    FILE,
                    master,
                    EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
                    EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
                )
            } catch (e: Exception) {
                null
            }
        }
    }
}
