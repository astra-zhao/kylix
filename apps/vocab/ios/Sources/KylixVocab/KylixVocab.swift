import CKylixVocab
import Foundation

/// Copy a Kylix-owned C string and release it with `kylix_free`.
private func take(_ p: UnsafePointer<CChar>?) -> String {
    guard let p else { return "" }
    let s = String(cString: p)
    kylix_free(UnsafeMutableRawPointer(mutating: p))
    return s
}

/// Shared flashcard core. HTTP stays in the SwiftUI shell (URLSession).
public enum KylixVocab {
    public static func wordsQuery(known: String) -> String {
        known.withCString { k in take(vc_words_query(k)) }
    }

    public static func reviewQuery(known: String) -> String {
        known.withCString { k in take(vc_review_query(k)) }
    }

    public static func markPath() -> String { take(vc_mark_path()) }

    public static func markRequest(known: String, id: String) -> String {
        known.withCString { k in
            id.withCString { i in
                take(vc_mark_request(k, i))
            }
        }
    }

    public static func parse(status: Int64, body: String) -> String {
        body.withCString { b in
            take(vc_parse(status, b))
        }
    }
}
