package com.kylix.plugin

import com.intellij.openapi.util.IconLoader
import javax.swing.Icon

/**
 * Kylix plugin icons (v6.8.0). The .klx file icon is used by KylixIconProvider
 * and the Kylix run configuration type; the compiler icon is reserved for a
 * toolbar/action if added later.
 */
object KylixIcons {
    @JvmField
    val KylixFile: Icon = IconLoader.getIcon("/icons/kylix.svg", KylixIcons::class.java)
}
