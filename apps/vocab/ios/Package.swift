// swift-tools-version: 5.9
import PackageDescription

// Swift Package 薄壳：C 头（module map）+ Swift 包装。
// libkylixvocab.a 由 apps/vocab/ios/build_core.sh 放到 Sources/CKylixVocab/lib/。
let package = Package(
    name: "KylixVocab",
    platforms: [
        .iOS(.v16),
        .macOS(.v13),
    ],
    products: [
        .library(name: "KylixVocab", targets: ["KylixVocab"]),
    ],
    targets: [
        .target(
            name: "CKylixVocab",
            path: "Sources/CKylixVocab",
            publicHeadersPath: "include",
            linkerSettings: [
                .linkedLibrary("kylixvocab"),
                .unsafeFlags(["-L", "Sources/CKylixVocab/lib"]),
            ]
        ),
        .target(
            name: "KylixVocab",
            dependencies: ["CKylixVocab"],
            path: "Sources/KylixVocab"
        ),
    ]
)
