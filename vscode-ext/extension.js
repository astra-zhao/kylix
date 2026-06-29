const { LanguageClient, Executable } = require('vscode-languageclient/node');
const vscode = require('vscode');
const path = require('path');

let client;
let statusBarItem;

/**
 * Resolve the Kylix compiler path.
 *
 * Order of precedence:
 *  1. kylix.compiler.path setting (if set and non-empty)
 *  2. KYLIX_PATH environment variable
 *  3. "kylix" (assumed on PATH)
 */
function resolveCompilerPath(config) {
    const fromConfig = config.get('compiler.path', 'kylix');
    if (fromConfig && fromConfig.trim() !== '') {
        return fromConfig.trim();
    }
    if (process.env.KYLIX_PATH) {
        return process.env.KYLIX_PATH;
    }
    return 'kylix';
}

/**
 * Build the LSP server executable descriptor.
 */
function serverOptions(compilerPath) {
    const exec = {
        command: compilerPath,
        args: ['lsp'],
        options: { cwd: vscode.workspace.rootPath || process.cwd() }
    };
    return {
        run: exec,
        debug: exec
    };
}

function activate(context) {
    const config = vscode.workspace.getConfiguration('kylix');

    if (!config.get('lsp.enabled', true)) {
        console.log('Kylix LSP is disabled');
        return;
    }

    const compilerPath = resolveCompilerPath(config);

    // Status bar indicator showing LSP connection state.
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 50);
    statusBarItem.text = '$(symbol-keyword) Kylix';
    statusBarItem.tooltip = 'Kylix Language Server: starting…';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);

    // Create language client (the client manages the server process lifecycle —
    // do NOT spawn the process manually, or two servers will start).
    client = new LanguageClient(
        'kylix',
        'Kylix Language Server',
        serverOptions(compilerPath),
        {
            documentSelector: [{ scheme: 'file', language: 'kylix' }],
            synchronize: {
                fileEvents: vscode.workspace.createFileSystemWatcher('**/*.klx')
            }
        }
    );

    client.start().then(() => {
        statusBarItem.tooltip = `Kylix Language Server: ready (${compilerPath})`;
    }).catch((err) => {
        statusBarItem.text = '$(error) Kylix';
        statusBarItem.tooltip = `Kylix LSP failed to start: ${err.message}`;
        vscode.window.showErrorMessage(
            `Failed to start Kylix Language Server via "${compilerPath}". ` +
            `Set "kylix.compiler.path" or KYLIX_PATH. (${err.message})`
        );
    });

    // Command: compile the active .klx file with the Kylix compiler.
    const compileCmd = vscode.commands.registerCommand('kylix.compile', async () => {
        const doc = vscode.window.activeTextEditor && vscode.window.activeTextEditor.document;
        if (!doc || doc.languageId !== 'kylix') {
            vscode.window.showWarningMessage('Open a .klx file to compile.');
            return;
        }
        await doc.save();
        const terminal = vscode.window.createTerminal('Kylix Compile');
        terminal.show(true);
        terminal.sendText(`${compilerPath} build ${shellQuote(doc.fileName)}`);
    });
    context.subscriptions.push(compileCmd);

    // Command: run the active .klx file.
    const runCmd = vscode.commands.registerCommand('kylix.run', async () => {
        const doc = vscode.window.activeTextEditor && vscode.window.activeTextEditor.document;
        if (!doc || doc.languageId !== 'kylix') {
            vscode.window.showWarningMessage('Open a .klx file to run.');
            return;
        }
        await doc.save();
        const terminal = vscode.window.createTerminal('Kylix Run');
        terminal.show(true);
        terminal.sendText(`${compilerPath} run ${shellQuote(doc.fileName)}`);
    });
    context.subscriptions.push(runCmd);

    context.subscriptions.push({
        dispose: () => {
            if (client) {
                client.stop();
            }
        }
    });

    console.log('Kylix extension activated');
}

/**
 * shellQuote wraps a path in double quotes for the integrated terminal,
 * escaping any embedded quotes. Safe for paths with spaces.
 */
function shellQuote(p) {
    return '"' + String(p).replace(/"/g, '\\"') + '"';
}

function deactivate() {
    if (client) {
        return client.stop();
    }
}

module.exports = {
    activate,
    deactivate
};
