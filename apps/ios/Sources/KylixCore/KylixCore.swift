import CKylixCore
import Foundation

/// Copy a Kylix-owned C string and release it with `kylix_free`.
private func take(_ p: UnsafePointer<CChar>?) -> String {
    guard let p else { return "" }
    let s = String(cString: p)
    kylix_free(UnsafeMutableRawPointer(mutating: p))
    return s
}

/// Shared business core. HTTP stays in the SwiftUI shell (URLSession).
public enum KylixCore {
    public static func validateLogin(username: String, password: String) -> String {
        username.withCString { u in
            password.withCString { p in
                take(mc_validate_login(u, p))
            }
        }
    }

    public static func loginRequest(username: String, password: String) -> String {
        username.withCString { u in
            password.withCString { p in
                take(mc_login_request(u, p))
            }
        }
    }

    public static func parseLogin(status: Int64, body: String) -> String {
        body.withCString { b in
            take(mc_parse_login(status, b))
        }
    }

    public static func refreshRequest(refreshToken: String) -> String {
        refreshToken.withCString { t in
            take(mc_refresh_request(t))
        }
    }

    public static func parseRefresh(status: Int64, body: String) -> String {
        body.withCString { b in
            take(mc_parse_refresh(status, b))
        }
    }

    public static func parseList(status: Int64, body: String) -> String {
        body.withCString { b in
            take(mc_parse_list(status, b))
        }
    }

    public static func authHeader(token: String) -> String {
        token.withCString { t in
            take(mc_auth_header(t))
        }
    }

    public static func loginPath() -> String { take(mc_login_path()) }
    public static func refreshPath() -> String { take(mc_refresh_path()) }
    public static func notesPath() -> String { take(mc_notes_path()) }
}
