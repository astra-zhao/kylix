package com.kylix.plugin

import com.intellij.execution.Executor
import com.intellij.execution.actions.ConfigurationContext
import com.intellij.execution.actions.RunConfigurationProducer
import com.intellij.execution.configurations.ConfigurationFactory
import com.intellij.execution.configurations.ConfigurationType
import com.intellij.execution.configurations.ConfigurationTypeBase
import com.intellij.execution.configurations.ConfigurationTypeUtil
import com.intellij.execution.configurations.RunConfiguration
import com.intellij.execution.configurations.RunConfigurationBase
import com.intellij.execution.configurations.RunProfileState
import com.intellij.execution.runners.ExecutionEnvironment
import com.intellij.openapi.options.SettingsEditor
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.TextBrowseFolderListener
import com.intellij.openapi.ui.TextFieldWithBrowseButton
import com.intellij.openapi.util.Ref
import com.intellij.psi.PsiElement
import org.jdom.Element
import java.awt.BorderLayout
import javax.swing.JComponent
import javax.swing.JLabel
import javax.swing.JPanel

/**
 * v6.8.0: IntelliJ run configuration for Kylix — compiles and runs a `.klx`
 * script via `kylix run <file>` (the KylixRT backend). Right-clicking a `.klx`
 * file in the project tree offers "Run '<file>.klx'"; a manual run configuration
 * can also pick a script in the Kylix settings editor.
 */

class KylixConfigurationType : ConfigurationTypeBase(
    ID, "Kylix", "Run a Kylix (.klx) program", KylixIcons.KylixFile
) {
    companion object {
        const val ID = "KylixRunConfiguration"

        @JvmStatic
        fun getInstance(): KylixConfigurationType =
            ConfigurationTypeUtil.findConfigurationType(KylixConfigurationType::class.java)
    }

    init {
        addFactory(KylixConfigurationFactory(this))
    }

    private class KylixConfigurationFactory(type: ConfigurationType) : ConfigurationFactory(type) {
        override fun getName(): String = "Kylix"
        override fun getIcon() = KylixIcons.KylixFile
        override fun createTemplateConfiguration(project: Project): RunConfiguration =
            KylixRunConfiguration(project, this)
    }
}

class KylixRunConfiguration(
    project: Project,
    factory: ConfigurationFactory,
) : RunConfigurationBase<KylixCommandLineState>(project, factory, "Kylix") {

    /** Absolute path of the .klx script to run. */
    var scriptPath: String? = null

    override fun getState(executor: Executor, environment: ExecutionEnvironment): KylixCommandLineState? {
        if (scriptPath.isNullOrBlank()) return null
        return KylixCommandLineState(environment, scriptPath!!)
    }

    override fun getConfigurationEditor(): SettingsEditor<out RunConfiguration> =
        KylixSettingsEditor()

    override fun writeExternal(element: Element) {
        super.writeExternal(element)
        scriptPath?.let { element.setAttribute("script", it) }
    }

    override fun readExternal(element: Element) {
        super.readExternal(element)
        scriptPath = element.getAttributeValue("script")
    }
}

/** Settings editor — a single file picker for the .klx script. */
class KylixSettingsEditor : SettingsEditor<KylixRunConfiguration>() {
    private val scriptField = TextFieldWithBrowseButton()

    init {
        scriptField.addBrowseFolderListener(
            TextBrowseFolderListener(
                com.intellij.openapi.fileChooser.FileChooserDescriptorFactory.createSingleFileDescriptor("klx")
            )
        )
    }

    override fun resetEditorFrom(config: KylixRunConfiguration) {
        scriptField.text = config.scriptPath ?: ""
    }

    override fun applyEditorTo(config: KylixRunConfiguration) {
        config.scriptPath = scriptField.text.ifBlank { null }
    }

    override fun createEditor(): JComponent = JPanel(BorderLayout(8, 0)).apply {
        add(JLabel("Script:"), BorderLayout.WEST)
        add(scriptField, BorderLayout.CENTER)
    }
}

/** Process state: `kylix run <file>` in a console. */
class KylixCommandLineState(
    environment: ExecutionEnvironment,
    private val scriptPath: String,
) : com.intellij.execution.configurations.CommandLineState(environment) {

    override fun startProcess(): com.intellij.execution.process.ProcessHandler {
        val commandLine = com.intellij.execution.configurations.GeneralCommandLine(
            "kylix", "run", scriptPath
        )
        val file = com.intellij.openapi.vfs.LocalFileSystem.getInstance().findFileByPath(scriptPath)
        file?.parent?.path?.let { commandLine.setWorkDirectory(it) }
        val handler = com.intellij.execution.process.OSProcessHandler(commandLine)
        com.intellij.execution.process.ProcessTerminatedListener.attach(handler, environment.project)
        return handler
    }
}

/**
 * Right-click / gutter run for a .klx file: "Run 'example.klx'" becomes
 * available in the Run context menu.
 */
class KylixRunConfigurationProducer :
    RunConfigurationProducer<KylixRunConfiguration>(
        ConfigurationTypeUtil.findConfigurationType(KylixConfigurationType::class.java)
    ) {

    override fun setupConfigurationFromContext(
        configuration: KylixRunConfiguration,
        context: ConfigurationContext,
        sourceElement: Ref<PsiElement>,
    ): Boolean {
        val vf = context.psiLocation?.containingFile?.virtualFile ?: return false
        if (!vf.extension.equals("klx", ignoreCase = true)) return false
        configuration.scriptPath = vf.path
        configuration.setName(vf.name)
        return true
    }

    override fun isConfigurationFromContext(
        configuration: KylixRunConfiguration,
        context: ConfigurationContext,
    ): Boolean =
        context.psiLocation?.containingFile?.virtualFile?.path == configuration.scriptPath
}
