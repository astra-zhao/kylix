package dev.kylix.admin

import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.concurrent.TimeUnit

/** One HTTP round-trip. The body is whatever the server sent, including errors. */
data class HttpResult(val status: Int, val body: String)

/**
 * Transport only. Request JSON, response interpretation and the paths all
 * come from [KylixBridge]. OkHttp is the TLS stack Android already trusts;
 * the Kylix core does not link libcurl on this target.
 */
class ApiClient {
    private val http = OkHttpClient.Builder()
        .callTimeout(20, TimeUnit.SECONDS)
        .build()
    private val json = "application/json; charset=utf-8".toMediaType()

    fun postJson(url: String, body: String): HttpResult {
        val req = Request.Builder()
            .url(url)
            .post(body.toRequestBody(json))
            .build()
        return execute(req)
    }

    fun getBearer(url: String, authorization: String): HttpResult {
        val req = Request.Builder()
            .url(url)
            .header("Authorization", authorization)
            .get()
            .build()
        return execute(req)
    }

    private fun execute(req: Request): HttpResult {
        http.newCall(req).execute().use { resp ->
            return HttpResult(resp.code, resp.body?.string() ?: "")
        }
    }
}
