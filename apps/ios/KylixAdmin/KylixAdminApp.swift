import KylixCore
import SwiftUI

@main
struct KylixAdminApp: App {
    var body: some Scene {
        WindowGroup {
            RootView()
        }
    }
}

struct NoteRow: Identifiable {
    let id: Int
    let title: String
    let body: String
    let done: Bool
}

struct RootView: View {
    @State private var server = "http://127.0.0.1:8090"
    @State private var username = "admin"
    @State private var password = ""
    @State private var message = ""
    @State private var busy = false
    @State private var token = ""
    @State private var refreshToken = ""
    @State private var accessExpiresAt: TimeInterval = 0
    @State private var notes: [NoteRow] = []
    @State private var signedIn = false

    var body: some View {
        NavigationStack {
            Group {
                if signedIn {
                    listScreen
                } else {
                    loginScreen
                }
            }
            .navigationTitle("KylixAdmin")
            .padding()
        }
    }

    private var loginScreen: some View {
        VStack(alignment: .leading, spacing: 12) {
            TextField("Server URL", text: $server)
                .textFieldStyle(.roundedBorder)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
            TextField("Username", text: $username)
                .textFieldStyle(.roundedBorder)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
            SecureField("Password", text: $password)
                .textFieldStyle(.roundedBorder)
            Button(busy ? "Signing in…" : "Sign in") {
                Task { await signIn() }
            }
            .buttonStyle(.borderedProminent)
            .disabled(busy)
            Text(message)
                .foregroundStyle(.secondary)
            Spacer()
        }
    }

    private var listScreen: some View {
        VStack(alignment: .leading, spacing: 12) {
            if notes.isEmpty {
                Text(message.isEmpty
                     ? "No notes yet.\nCreate one in the KylixAdmin web console (Notes), then sign in again."
                     : message)
                    .foregroundStyle(.secondary)
            } else {
                List(notes) { note in
                    VStack(alignment: .leading, spacing: 4) {
                        Text(note.title).font(.headline)
                        Text(note.body).font(.body)
                        Text(note.done ? "done" : "open")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
                .listStyle(.plain)
            }
            Button("Sign out") {
                clearSession()
                notes = []
                message = ""
                signedIn = false
            }
        }
    }

    private func signIn() async {
        busy = true
        message = ""
        defer { busy = false }
        let normalized = normalizeBase(server)
        if normalized.isEmpty {
            message = "Server URL is required"
            return
        }
        let local = KylixCore.validateLogin(username: username, password: password)
        if !local.isEmpty {
            message = local
            return
        }
        do {
            let body = KylixCore.loginRequest(username: username, password: password)
            let resp = try await postJSON(url: normalized + KylixCore.loginPath(), body: body)
            let parsed = try jsonObject(KylixCore.parseLogin(status: Int64(resp.status), body: resp.body))
            let ok = parsed["ok"] as? Bool ?? false
            if !ok {
                message = (parsed["error"] as? String) ?? "sign in failed"
                return
            }
            if storeSession(parsed) != nil {
                message = "sign in failed"
                return
            }
            try await loadNotes(base: normalized)
            signedIn = !token.isEmpty
        } catch {
            message = error.localizedDescription
        }
    }

    /// Keeps the new pair. Returns a message when either token is missing.
    private func storeSession(_ parsed: [String: Any]) -> String? {
        token = (parsed["token"] as? String) ?? ""
        refreshToken = (parsed["refresh_token"] as? String) ?? ""
        let ttl = (parsed["expires_in"] as? NSNumber)?.doubleValue ?? 0
        accessExpiresAt = ttl > 0 ? Date().timeIntervalSince1970 + ttl : 0
        return (token.isEmpty || refreshToken.isEmpty) ? "sign in failed" : nil
    }

    private func clearSession() {
        token = ""
        refreshToken = ""
        accessExpiresAt = 0
    }

    /// One refresh. Success replaces both tokens. Failure drops the local pair.
    private func refresh(base: String) async throws -> Bool {
        if refreshToken.isEmpty {
            return false
        }
        let resp = try await postJSON(url: base + KylixCore.refreshPath(), body: KylixCore.refreshRequest(refreshToken: refreshToken))
        let parsed = try jsonObject(KylixCore.parseRefresh(status: Int64(resp.status), body: resp.body))
        let ok = (parsed["ok"] as? Bool) ?? false
        let relogin = (parsed["relogin"] as? Bool) ?? false
        if !ok || relogin {
            clearSession()
            return false
        }
        return storeSession(parsed) == nil
    }

    private func loadNotes(base: String) async throws {
        if accessExpiresAt > 0 && Date().timeIntervalSince1970 >= accessExpiresAt - 60 {
            if try await refresh(base: base) == false {
                message = "session expired"
                signedIn = false
                return
            }
        }
        var resp = try await getBearer(url: base + KylixCore.notesPath(), authorization: KylixCore.authHeader(token: token))
        var parsed = try jsonObject(KylixCore.parseList(status: Int64(resp.status), body: resp.body))
        var ok = parsed["ok"] as? Bool ?? false
        if !ok && (parsed["relogin"] as? Bool) == true && (try await refresh(base: base)) {
            resp = try await getBearer(url: base + KylixCore.notesPath(), authorization: KylixCore.authHeader(token: token))
            parsed = try jsonObject(KylixCore.parseList(status: Int64(resp.status), body: resp.body))
            ok = parsed["ok"] as? Bool ?? false
        }
        if !ok {
            message = (parsed["error"] as? String) ?? "could not load notes"
            if (parsed["relogin"] as? Bool) == true {
                clearSession()
                signedIn = false
            }
            return
        }
        notes = decodeNotes(parsed)
        message = ""
    }

    private func normalizeBase(_ raw: String) -> String {
        var s = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        while s.hasSuffix("/") {
            s.removeLast()
        }
        return s
    }
}

private struct HttpResult {
    let status: Int
    let body: String
}

private func postJSON(url: String, body: String) async throws -> HttpResult {
    guard let u = URL(string: url) else { throw URLError(.badURL) }
    var req = URLRequest(url: u)
    req.httpMethod = "POST"
    req.setValue("application/json; charset=utf-8", forHTTPHeaderField: "Content-Type")
    req.httpBody = Data(body.utf8)
    let (data, resp) = try await URLSession.shared.data(for: req)
    let code = (resp as? HTTPURLResponse)?.statusCode ?? 0
    return HttpResult(status: code, body: String(data: data, encoding: .utf8) ?? "")
}

private func getBearer(url: String, authorization: String) async throws -> HttpResult {
    guard let u = URL(string: url) else { throw URLError(.badURL) }
    var req = URLRequest(url: u)
    req.httpMethod = "GET"
    req.setValue(authorization, forHTTPHeaderField: "Authorization")
    let (data, resp) = try await URLSession.shared.data(for: req)
    let code = (resp as? HTTPURLResponse)?.statusCode ?? 0
    return HttpResult(status: code, body: String(data: data, encoding: .utf8) ?? "")
}

private func jsonObject(_ text: String) throws -> [String: Any] {
    let data = Data(text.utf8)
    let obj = try JSONSerialization.jsonObject(with: data)
    return obj as? [String: Any] ?? [:]
}

private func decodeNotes(_ parsed: [String: Any]) -> [NoteRow] {
    let items = parsed["items"] as? [[String: Any]] ?? []
    return items.map { row in
        NoteRow(
            id: (row["id"] as? Int) ?? (row["id"] as? NSNumber)?.intValue ?? 0,
            title: (row["title"] as? String) ?? "",
            body: (row["body"] as? String) ?? "",
            done: (row["done"] as? Bool) ?? false
        )
    }
}
