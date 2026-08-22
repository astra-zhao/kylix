package com.kylix.plugin

import com.intellij.openapi.project.Project
import com.redhat.devtools.lsp4ij.LanguageServerFactory
import com.redhat.devtools.lsp4ij.client.LanguageClientImpl
import com.redhat.devtools.lsp4ij.server.StreamConnectionProvider
import org.jetbrains.annotations.NotNull

/**
 * Wires the Kylix language server (a stdio process) into LSP4IJ. Declared in
 * plugin.xml via the `com.redhat.devtools.lsp4ij.server` extension point.
 */
class KylixLanguageServerFactory : LanguageServerFactory {
    override fun createConnectionProvider(@NotNull project: Project): StreamConnectionProvider =
        KylixLanguageServer()

    override fun createLanguageClient(@NotNull project: Project): LanguageClientImpl =
        LanguageClientImpl(project)
}
