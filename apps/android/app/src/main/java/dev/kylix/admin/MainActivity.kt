package dev.kylix.admin

import android.app.Activity
import android.os.Bundle
import android.view.View
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import org.json.JSONObject
import java.util.concurrent.Executors

/**
 * Login + notes list. Business rules live in the Kylix core; this activity
 * only collects input, calls OkHttp, and renders the normalized JSON.
 * Tokens are written to [SessionStore] and restored on a cold start.
 */
class MainActivity : Activity() {
    private val io = Executors.newSingleThreadExecutor()
    private val http = ApiClient()
    private lateinit var session: SessionStore
    private var serverBase: String = ""
    private var token: String = ""
    private var refreshToken: String = ""
    private var accessExpiresAtMs: Long = 0

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        session = SessionStore(this)

        val server = findViewById<EditText>(R.id.server)
        val username = findViewById<EditText>(R.id.username)
        val password = findViewById<EditText>(R.id.password)
        val message = findViewById<TextView>(R.id.message)
        val notes = findViewById<TextView>(R.id.notes)
        val login = findViewById<Button>(R.id.login)
        val logout = findViewById<Button>(R.id.logout)
        val loginBox = findViewById<View>(R.id.loginBox)
        val listBox = findViewById<View>(R.id.listBox)

        fun showLogin(text: String) {
            dropSession()
            loginBox.visibility = View.VISIBLE
            listBox.visibility = View.GONE
            message.text = text
        }

        fun showList(text: String) {
            loginBox.visibility = View.GONE
            listBox.visibility = View.VISIBLE
            notes.text = text
        }

        login.setOnClickListener {
            val base = server.text.toString().trim().trimEnd('/')
            val user = username.text.toString()
            val pass = password.text.toString()
            message.text = "Signing in…"
            login.isEnabled = false
            io.execute {
                var err: String? = null
                var listText = ""
                var stayIn = false
                try {
                    err = signIn(base, user, pass)
                    if (err == null) {
                        val loaded = loadNotes(base)
                        listText = loaded.second
                        stayIn = loaded.first
                    }
                } catch (e: Exception) {
                    err = e.message ?: "network error"
                }
                runOnUiThread {
                    login.isEnabled = true
                    if (err != null || !stayIn) {
                        showLogin(err ?: listText)
                    } else {
                        showList(listText)
                    }
                }
            }
        }

        logout.setOnClickListener {
            val base = serverBase
            val raw = refreshToken
            logout.isEnabled = false
            io.execute {
                try {
                    if (base.isNotEmpty() && raw.isNotEmpty()) {
                        http.postJson(base + KylixBridge.logoutPath(), KylixBridge.refreshRequest(raw))
                    }
                } catch (_: Exception) {
                    // Offline logout still clears this device. The server row
                    // stays until that token is presented or evicted.
                }
                runOnUiThread {
                    logout.isEnabled = true
                    showLogin("")
                }
            }
        }

        val saved = session.load()
        if (saved != null && saved.refreshToken.isNotEmpty()) {
            server.setText(saved.server)
            serverBase = saved.server.trim().trimEnd('/')
            token = saved.token
            refreshToken = saved.refreshToken
            accessExpiresAtMs = saved.accessExpiresAtMs
            loginBox.visibility = View.GONE
            listBox.visibility = View.VISIBLE
            notes.text = "Restoring session…"
            io.execute {
                var listText = ""
                var stay = false
                var err: String? = null
                try {
                    val loaded = loadNotes(serverBase)
                    listText = loaded.second
                    stay = loaded.first
                } catch (e: Exception) {
                    err = e.message ?: "network error"
                }
                runOnUiThread {
                    if (err != null) {
                        notes.text = err
                    } else if (!stay) {
                        showLogin(listText)
                    } else {
                        showList(listText)
                    }
                }
            }
        }
    }

    override fun onDestroy() {
        io.shutdownNow()
        super.onDestroy()
    }

    /** Returns an error string, or null when [token] is set. */
    private fun signIn(base: String, user: String, pass: String): String? {
        if (base.isEmpty()) {
            return "Server URL is required"
        }
        val local = KylixBridge.validateLogin(user, pass)
        if (local.isNotEmpty()) {
            return local
        }
        serverBase = base
        val body = KylixBridge.loginRequest(user, pass)
        val resp = http.postJson(base + KylixBridge.loginPath(), body)
        val parsed = JSONObject(KylixBridge.parseLogin(resp.status.toLong(), resp.body))
        if (!parsed.optBoolean("ok")) {
            return parsed.optString("error", "sign in failed")
        }
        return storeSession(parsed)
    }

    /** Keeps the new pair and writes it to secure storage. */
    private fun storeSession(parsed: JSONObject): String? {
        token = parsed.optString("token")
        refreshToken = parsed.optString("refresh_token")
        val ttl = parsed.optLong("expires_in", 0)
        accessExpiresAtMs = if (ttl > 0) System.currentTimeMillis() + ttl * 1000 else 0
        if (token.isEmpty() || refreshToken.isEmpty()) {
            dropSession()
            return "sign in failed"
        }
        session.save(serverBase, token, refreshToken, accessExpiresAtMs)
        return null
    }

    private fun dropSession() {
        token = ""
        refreshToken = ""
        accessExpiresAtMs = 0
        if (::session.isInitialized) {
            session.clear()
        }
    }

    /**
     * One refresh. Success replaces both tokens (the server rotates this jti
     * only). Failure drops the local pair so the next step is the login screen.
     */
    private fun refresh(base: String): Boolean {
        if (refreshToken.isEmpty()) {
            return false
        }
        val resp = http.postJson(base + KylixBridge.refreshPath(), KylixBridge.refreshRequest(refreshToken))
        val parsed = JSONObject(KylixBridge.parseRefresh(resp.status.toLong(), resp.body))
        if (!parsed.optBoolean("ok") || parsed.optBoolean("relogin")) {
            dropSession()
            return false
        }
        return storeSession(parsed) == null
    }

    /** First value is true when the list should be shown. */
    private fun loadNotes(base: String): Pair<Boolean, String> {
        if (accessExpiresAtMs > 0 && System.currentTimeMillis() >= accessExpiresAtMs - 60_000L) {
            if (!refresh(base)) {
                return Pair(false, "session expired")
            }
        }
        var resp = http.getBearer(base + KylixBridge.notesPath(), KylixBridge.authHeader(token))
        var parsed = JSONObject(KylixBridge.parseList(resp.status.toLong(), resp.body))
        if (!parsed.optBoolean("ok") && parsed.optBoolean("relogin") && refresh(base)) {
            resp = http.getBearer(base + KylixBridge.notesPath(), KylixBridge.authHeader(token))
            parsed = JSONObject(KylixBridge.parseList(resp.status.toLong(), resp.body))
        }
        if (!parsed.optBoolean("ok")) {
            if (parsed.optBoolean("relogin")) {
                dropSession()
            }
            return Pair(false, parsed.optString("error", "could not load notes"))
        }
        return Pair(true, renderNotes(parsed))
    }

    private fun renderNotes(parsed: JSONObject): String {
        val items = parsed.optJSONArray("items") ?: return "No notes yet."
        if (items.length() == 0) {
            return "No notes yet.\nCreate one in the KylixAdmin web console (Notes), then sign in again."
        }
        val b = StringBuilder()
        for (i in 0 until items.length()) {
            val row = items.getJSONObject(i)
            if (i > 0) {
                b.append("\n\n")
            }
            val done = if (row.optBoolean("done")) "done" else "open"
            b.append(row.optString("title")).append("  (").append(done).append(")\n")
            b.append(row.optString("body"))
        }
        return b.toString()
    }
}
