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
 */
class MainActivity : Activity() {
    private val io = Executors.newSingleThreadExecutor()
    private val http = ApiClient()
    private var token: String = ""

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

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
            token = ""
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

        logout.setOnClickListener { showLogin("") }
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
        val body = KylixBridge.loginRequest(user, pass)
        val resp = http.postJson(base + KylixBridge.loginPath(), body)
        val parsed = JSONObject(KylixBridge.parseLogin(resp.status.toLong(), resp.body))
        if (!parsed.optBoolean("ok")) {
            return parsed.optString("error", "sign in failed")
        }
        token = parsed.optString("token")
        return if (token.isEmpty()) "sign in failed" else null
    }

    /** First value is true when the list should be shown. */
    private fun loadNotes(base: String): Pair<Boolean, String> {
        val resp = http.getBearer(base + KylixBridge.notesPath(), KylixBridge.authHeader(token))
        val parsed = JSONObject(KylixBridge.parseList(resp.status.toLong(), resp.body))
        if (!parsed.optBoolean("ok")) {
            if (parsed.optBoolean("relogin")) {
                token = ""
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
