package com.kylix.plugin

import com.intellij.ide.IconProvider
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiFile
import javax.swing.Icon

/**
 * Gives `.klx` files the Kylix icon in the project tree, editor tabs, and other
 * file-listing UIs (v6.8.0). The TextMate bundle already associates `.klx` with
 * the generic TextMate file type; this icon provider overrides the icon for any
 * file whose extension is `klx`.
 */
class KylixIconProvider : IconProvider() {
    override fun getIcon(element: PsiElement, flags: Int): Icon? {
        if (element is PsiFile) {
            val vf = element.virtualFile ?: return null
            if (vf.extension.equals("klx", ignoreCase = true)) {
                return KylixIcons.KylixFile
            }
        }
        return null
    }
}
