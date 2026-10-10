import KylixVocab
import SwiftUI

struct VocabCard: Identifiable {
    let id: String
    let en: String
    let zh: String
    let known: Bool
}

struct HttpResult {
    let status: Int
    let body: String
}

@main
struct VocabApp: App {
    var body: some Scene {
        WindowGroup {
            VocabRoot()
        }
    }
}

struct VocabRoot: View {
    @State private var server = "http://127.0.0.1:8091"
    @State private var known = ""
    @State private var cards: [VocabCard] = []
    @State private var index = 0
    @State private var left = 6
    @State private var total = 6
    @State private var reviewing = false
    @State private var busy = false
    @State private var message = ""
    @State private var didLoad = false

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: 12) {
                TextField("Server URL", text: $server)
                    .textFieldStyle(.roundedBorder)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                Text(progressText)
                    .font(.headline)
                Text(current?.en ?? "全部掌握")
                    .font(.largeTitle)
                Text(current?.zh ?? "")
                    .font(.title2)
                Button(busy ? "…" : "认识") { Task { await markCurrent() } }
                    .buttonStyle(.borderedProminent)
                    .disabled(busy || current == nil)
                Button("下一张") { advance() }
                    .disabled(busy || cards.isEmpty)
                Button("全部单词") {
                    reviewing = false
                    Task { await load() }
                }
                .disabled(busy)
                Button("复习未掌握") {
                    reviewing = true
                    Task { await load() }
                }
                .disabled(busy)
                Text(message)
                    .foregroundStyle(.secondary)
                Spacer()
            }
            .padding()
            .navigationTitle("背单词")
        }
        .task {
            if !didLoad {
                didLoad = true
                if let saved = UserDefaults.standard.string(forKey: "vocab.server"), !saved.isEmpty {
                    server = saved
                }
                known = UserDefaults.standard.string(forKey: "vocab.known") ?? ""
                await load()
            }
        }
    }

    private var current: VocabCard? {
        guard !cards.isEmpty else { return nil }
        return cards[index % cards.count]
    }

    private var progressText: String {
        let mode = reviewing ? "复习" : "全部"
        return "\(mode) · 还剩 \(left) / \(total)"
    }

    private func advance() {
        guard !cards.isEmpty else { return }
        index = (index + 1) % cards.count
        message = current?.known == true ? "已认识" : ""
    }

    private func baseURL() -> String {
        server.trimmingCharacters(in: .whitespacesAndNewlines).trimmingCharacters(in: CharacterSet(charactersIn: "/"))
    }

    private func save() {
        UserDefaults.standard.set(known, forKey: "vocab.known")
        UserDefaults.standard.set(baseURL(), forKey: "vocab.server")
    }

    private func load() async {
        let base = baseURL()
        if base.isEmpty {
            message = "Server URL is required"
            return
        }
        busy = true
        message = ""
        defer { busy = false }
        let path = reviewing ? KylixVocab.reviewQuery(known: known) : KylixVocab.wordsQuery(known: known)
        do {
            let resp = try await httpGet(base + path)
            try apply(KylixVocab.parse(status: Int64(resp.status), body: resp.body))
        } catch {
            message = error.localizedDescription
        }
    }

    private func markCurrent() async {
        guard let card = current else {
            message = "No card"
            return
        }
        let base = baseURL()
        if base.isEmpty {
            message = "Server URL is required"
            return
        }
        busy = true
        message = ""
        defer { busy = false }
        do {
            let body = KylixVocab.markRequest(known: known, id: card.id)
            let resp = try await httpPost(base + KylixVocab.markPath(), body)
            let text = KylixVocab.parse(status: Int64(resp.status), body: resp.body)
            if reviewing {
                if let obj = try jsonObject(text), boolField(obj, "ok") {
                    known = obj["known"] as? String ?? known
                    save()
                }
                await load()
            } else {
                try apply(text)
            }
        } catch {
            message = error.localizedDescription
        }
    }

    private func apply(_ text: String) throws {
        guard let obj = try jsonObject(text) else {
            message = "request failed"
            return
        }
        if !boolField(obj, "ok") {
            message = obj["error"] as? String ?? "request failed"
            return
        }
        known = obj["known"] as? String ?? ""
        left = intField(obj, "left")
        total = intField(obj, "total")
        save()
        var next: [VocabCard] = []
        if let rows = obj["cards"] as? [Any] {
            for item in rows {
                guard let row = item as? [String: Any] else { continue }
                next.append(VocabCard(
                    id: String(intField(row, "id")),
                    en: row["en"] as? String ?? "",
                    zh: row["zh"] as? String ?? "",
                    known: boolField(row, "known")
                ))
            }
        }
        cards = next
        if cards.isEmpty {
            index = 0
        } else if index >= cards.count {
            index = 0
        }
        message = current?.known == true ? "已认识" : ""
    }

    private func intField(_ obj: [String: Any], _ key: String) -> Int {
        if let n = obj[key] as? Int { return n }
        if let n = obj[key] as? NSNumber { return n.intValue }
        return 0
    }

    private func boolField(_ obj: [String: Any], _ key: String) -> Bool {
        if let n = obj[key] as? Bool { return n }
        if let n = obj[key] as? NSNumber { return n.boolValue }
        return false
    }

    private func jsonObject(_ text: String) throws -> [String: Any]? {
        guard let data = text.data(using: .utf8),
              let obj = try JSONSerialization.jsonObject(with: data) as? [String: Any] else {
            return nil
        }
        return obj
    }

    private func httpGet(_ url: String) async throws -> HttpResult {
        guard let u = URL(string: url) else { throw URLError(.badURL) }
        let (data, resp) = try await URLSession.shared.data(from: u)
        let code = (resp as? HTTPURLResponse)?.statusCode ?? 0
        return HttpResult(status: code, body: String(data: data, encoding: .utf8) ?? "")
    }

    private func httpPost(_ url: String, _ body: String) async throws -> HttpResult {
        guard let u = URL(string: url) else { throw URLError(.badURL) }
        var req = URLRequest(url: u)
        req.httpMethod = "POST"
        req.setValue("application/json; charset=utf-8", forHTTPHeaderField: "Content-Type")
        req.httpBody = body.data(using: .utf8)
        let (data, resp) = try await URLSession.shared.data(for: req)
        let code = (resp as? HTTPURLResponse)?.statusCode ?? 0
        return HttpResult(status: code, body: String(data: data, encoding: .utf8) ?? "")
    }
}
