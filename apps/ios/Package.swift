// swift-tools-version: 5.9
import PackageDescription

// Swift Package 薄壳：C 头（module map）+ Swift 包装。
// libkylixcore.a 由 apps/ios/build_core.sh 放到 Sources/CKylixCore/lib/。
// 链接搜索路径是包目录下的相对路径；Xcode 工程另在 project.yml 里再指定一次，
// 避免 SPM 把 -L 解析到构建目录。详见 docs/MOBILE_APPS.md。
let package = Package(
    name: "KylixCore",
    platforms: [
        .iOS(.v16),
        .macOS(.v13),
    ],
    products: [
        .library(name: "KylixCore", targets: ["KylixCore"]),
    ],
    targets: [
        .target(
            name: "CKylixCore",
            path: "Sources/CKylixCore",
            publicHeadersPath: "include",
            linkerSettings: [
                .linkedLibrary("kylixcore"),
                .unsafeFlags(["-L", "Sources/CKylixCore/lib"]),
            ]
        ),
        .target(
            name: "KylixCore",
            dependencies: ["CKylixCore"],
            path: "Sources/KylixCore"
        ),
    ]
)
