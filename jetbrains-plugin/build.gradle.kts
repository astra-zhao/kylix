// Kylix JetBrains plugin — Modern Pascal language support for IntelliJ/GoLand.
// Syntax highlighting (TextMate), LSP integration (LSP4IJ bridging `kylix lsp`),
// and 25 live templates. Built with the IntelliJ Platform Gradle Plugin 2.x.
plugins {
    kotlin("jvm") version "2.1.20"
    id("org.jetbrains.intellij.platform") version "2.3.0"
}

group = "com.kylix"
version = "0.1.0"

repositories {
    mavenCentral()
    intellijPlatform {
        defaultRepositories()
    }
}

dependencies {
    intellijPlatform {
        // Community Edition SDK (TextMate + LSP4IJ both work on IC).
        intellijIdeaCommunity("2024.3")
        // TextMate syntax highlighting (bundled with the IDE).
        bundledPlugin("org.jetbrains.plugins.textmate")
        // LSP client bridge — `kylix lsp` becomes a language server in the IDE.
        plugin("com.redhat.devtools.lsp4ij", "0.17.0")
        instrumentationTools()
    }
}

kotlin {
    jvmToolchain(21)
}

intellijPlatform {
    instrumentCode = true
}
