package com.kylix.plugin

import org.jetbrains.plugins.textmate.api.TextMateBundleProvider
import org.jetbrains.plugins.textmate.api.TextMateBundleProvider.PluginBundle
import java.nio.file.Paths

/**
 * Registers the Kylix TextMate grammar (`textmate/kylix/syntaxes/kylix.tmLanguage.json`)
 * so `.klx` files get syntax highlighting (scope `source.kylix`, `fileTypes: ["klx"]`).
 * Registered via the `com.intellij.textmate.bundleProvider` extension point.
 */
class KylixTextMateBundleProvider : TextMateBundleProvider {
    override fun getBundles(): List<PluginBundle> {
        val url = javaClass.getResource("/textmate/kylix")!!
        return listOf(PluginBundle("Kylix", Paths.get(url.toURI())))
    }
}
