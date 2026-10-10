package dev.kylix.vocab

import android.app.Activity
import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import org.json.JSONObject
import java.util.concurrent.Executors

/** One card from the Kylix deck JSON. */
data class Card(val id: String, val en: String, val zh: String, val known: Boolean)

/**
 * Browse the six words, mark one known, or review the ones still unknown.
 * The known-id list is stored on the device and sent back on every request.
 * The server does not keep it.
 */
class MainActivity : Activity() {
    private val http = ApiClient()
    private val io = Executors.newSingleThreadExecutor()
    private val prefsName = "vocab"

    private lateinit var server: EditText
    private lateinit var progress: TextView
    private lateinit var cardEn: TextView
    private lateinit var cardZh: TextView
    private lateinit var message: TextView
    private lateinit var btnKnown: Button
    private lateinit var btnNext: Button
    private lateinit var btnAll: Button
    private lateinit var btnReview: Button

    private var known = ""
    private var cards: List<Card> = emptyList()
    private var index = 0
    private var left = 6
    private var total = 6
    private var reviewing = false
    private var busy = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        server = findViewById(R.id.server)
        progress = findViewById(R.id.progress)
        cardEn = findViewById(R.id.cardEn)
        cardZh = findViewById(R.id.cardZh)
        message = findViewById(R.id.message)
        btnKnown = findViewById(R.id.btnKnown)
        btnNext = findViewById(R.id.btnNext)
        btnAll = findViewById(R.id.btnAll)
        btnReview = findViewById(R.id.btnReview)

        val prefs = getSharedPreferences(prefsName, MODE_PRIVATE)
        known = prefs.getString("known", "") ?: ""
        val savedServer = prefs.getString("server", "") ?: ""
        if (savedServer.isNotEmpty()) {
            server.setText(savedServer)
        }

        btnAll.setOnClickListener { reviewing = false; load() }
        btnReview.setOnClickListener { reviewing = true; load() }
        btnKnown.setOnClickListener { markCurrent() }
        btnNext.setOnClickListener {
            if (cards.isNotEmpty()) {
                index = (index + 1) % cards.size
                showCard()
            }
        }
        load()
    }

    override fun onDestroy() {
        io.shutdownNow()
        super.onDestroy()
    }

    private fun baseUrl(): String {
        return server.text.toString().trim().trimEnd('/')
    }

    private fun setBusy(on: Boolean) {
        busy = on
        btnKnown.isEnabled = !on
        btnNext.isEnabled = !on
        btnAll.isEnabled = !on
        btnReview.isEnabled = !on
    }

    private fun load() {
        val base = baseUrl()
        if (base.isEmpty()) {
            message.text = "Server URL is required"
            return
        }
        setBusy(true)
        message.text = ""
        val snapshot = known
        io.execute {
            var err: String? = null
            var parsed: JSONObject? = null
            try {
                val path = if (reviewing) {
                    KylixBridge.reviewQuery(snapshot)
                } else {
                    KylixBridge.wordsQuery(snapshot)
                }
                val resp = http.get(base + path)
                parsed = JSONObject(KylixBridge.parse(resp.status.toLong(), resp.body))
                if (!parsed.optBoolean("ok")) {
                    err = parsed.optString("error", "request failed")
                    parsed = null
                }
            } catch (e: Exception) {
                err = e.message ?: "network error"
            }
            val body = parsed
            runOnUiThread {
                setBusy(false)
                if (err != null || body == null) {
                    message.text = err ?: "request failed"
                } else {
                    applyDeck(body)
                }
            }
        }
    }

    private fun markCurrent() {
        if (cards.isEmpty()) {
            message.text = "No card"
            return
        }
        val base = baseUrl()
        if (base.isEmpty()) {
            message.text = "Server URL is required"
            return
        }
        val id = cards[index].id
        val snapshot = known
        setBusy(true)
        message.text = ""
        io.execute {
            var err: String? = null
            var parsed: JSONObject? = null
            try {
                val resp = http.postJson(base + KylixBridge.markPath(), KylixBridge.markRequest(snapshot, id))
                parsed = JSONObject(KylixBridge.parse(resp.status.toLong(), resp.body))
                if (!parsed.optBoolean("ok")) {
                    err = parsed.optString("error", "request failed")
                    parsed = null
                }
            } catch (e: Exception) {
                err = e.message ?: "network error"
            }
            val body = parsed
            runOnUiThread {
                setBusy(false)
                if (err != null || body == null) {
                    message.text = err ?: "request failed"
                } else {
                    known = body.optString("known")
                    save()
                    if (reviewing) {
                        load()
                    } else {
                        applyDeck(body)
                    }
                }
            }
        }
    }

    private fun applyDeck(body: JSONObject) {
        known = body.optString("known")
        left = body.optInt("left", 0)
        total = body.optInt("total", 0)
        save()
        val arr = body.optJSONArray("cards")
        val next = ArrayList<Card>()
        if (arr != null) {
            for (i in 0 until arr.length()) {
                val row = arr.getJSONObject(i)
                next.add(Card(row.optInt("id").toString(), row.optString("en"), row.optString("zh"), row.optBoolean("known")))
            }
        }
        cards = next
        if (cards.isEmpty()) {
            index = 0
        } else if (index >= cards.size) {
            index = 0
        }
        showCard()
    }

    private fun showCard() {
        val mode = if (reviewing) "复习" else "全部"
        progress.text = "$mode · 还剩 $left / $total"
        if (cards.isEmpty()) {
            cardEn.text = "全部掌握"
            cardZh.text = ""
            return
        }
        val card = cards[index]
        cardEn.text = card.en
        cardZh.text = card.zh
        if (card.known) {
            message.text = "已认识"
        } else {
            message.text = ""
        }
    }

    private fun save() {
        getSharedPreferences(prefsName, MODE_PRIVATE).edit()
            .putString("known", known)
            .putString("server", baseUrl())
            .apply()
    }
}
