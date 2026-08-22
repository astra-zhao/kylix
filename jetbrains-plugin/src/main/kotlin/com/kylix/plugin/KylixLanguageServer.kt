package com.kylix.plugin

import com.redhat.devtools.lsp4ij.server.ProcessStreamConnectionProvider
import java.nio.file.Files
import java.nio.file.Paths

/**
 * Launches the Kylix LSP server (`kylix lsp`, stdio JSON-RPC) as an IntelliJ
 * language server via LSP4IJ. The compiler path is resolved from:
 *   1. the `KYLIX_PATH` environment variable,
 *   2. a `kylix` executable on `PATH`,
 *   3. the bare `kylix` command.
 * The LSP server reads `stdlib/klx` for completion via `$KYLIX_HOME/stdlib/klx`
 * (or by probing 5 directories above the executable — see pkg/lsp/document.go).
 */
class KylixLanguageServer : ProcessStreamConnectionProvider() {
    init {
        commands = listOf(resolveKylix(), "lsp")
    }

    private fun resolveKylix(): String {
        System.getenv("KYLIX_PATH")?.let { return it }
        val path = System.getenv("PATH") ?: return "kylix"
        for (dir in path.split(java.io.File.pathSeparatorChar)) {
            if (dir.isEmpty()) continue
            val exe = Paths.get(dir, "kylix")
            if (Files.isExecutable(exe)) return exe.toString()
        }
        return "kylix"
    }
}
