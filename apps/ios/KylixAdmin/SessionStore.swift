import Foundation
import Security

/// One Keychain item (generic password, service `dev.kylix.admin`) holding the
/// server URL and both tokens. A failed read or write leaves the in-memory
/// session alone; the shell must not crash.
struct StoredSession {
    var server: String
    var token: String
    var refreshToken: String
    var accessExpiresAt: TimeInterval
}

enum SessionStore {
    private static let service = "dev.kylix.admin"
    private static let account = "session"

    static func save(_ session: StoredSession) {
        guard let data = try? JSONSerialization.data(withJSONObject: [
            "server": session.server,
            "token": session.token,
            "refresh_token": session.refreshToken,
            "access_expires_at": session.accessExpiresAt
        ]) else {
            return
        }
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account
        ]
        var add = query
        add[kSecValueData as String] = data
        add[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
        let status = SecItemAdd(add as CFDictionary, nil)
        if status == errSecDuplicateItem {
            SecItemUpdate(query as CFDictionary, [kSecValueData as String: data] as CFDictionary)
        }
    }

    static func load() -> StoredSession? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        guard status == errSecSuccess, let data = item as? Data else {
            return nil
        }
        guard let obj = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else {
            return nil
        }
        let refresh = (obj["refresh_token"] as? String) ?? ""
        if refresh.isEmpty {
            return nil
        }
        let exp = (obj["access_expires_at"] as? NSNumber)?.doubleValue ?? 0
        return StoredSession(
            server: (obj["server"] as? String) ?? "",
            token: (obj["token"] as? String) ?? "",
            refreshToken: refresh,
            accessExpiresAt: exp
        )
    }

    static func clear() {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account
        ]
        SecItemDelete(query as CFDictionary)
    }
}
