const { LanguageClient } = require('vscode-languageclient/node');
const vscode = require('vscode');
const { spawn } = require('child_process');
const path = require('path');

let client;

function activate(context) {
    const config = vscode.workspace.getConfiguration('kylix');

    if (!config.get('lsp.enabled', true)) {
        console.log('Kylix LSP is disabled');
        return;
    }

    const compilerPath = config.get('compiler.path', 'kylix');

    // Start the LSP server
    const serverProcess = spawn(compilerPath, ['lsp'], {
        cwd: vscode.workspace.rootPath || process.cwd()
    });

    serverProcess.stderr.on('data', (data) => {
        console.error(`Kylix LSP stderr: ${data}`);
    });

    serverProcess.on('error', (err) => {
        vscode.window.showErrorMessage(`Failed to start Kylix LSP: ${err.message}`);
    });

    // Create language client
    client = new LanguageClient(
        'kylix',
        'Kylix Language Server',
        {
            run: {
                command: compilerPath,
                args: ['lsp'],
                options: { cwd: vscode.workspace.rootPath || process.cwd() }
            },
            debug: {
                command: compilerPath,
                args: ['lsp'],
                options: { cwd: vscode.workspace.rootPath || process.cwd() }
            }
        },
        {
            documentSelector: [{ scheme: 'file', language: 'kylix' }],
            synchronize: {
                fileEvents: vscode.workspace.createFileSystemWatcher('**/*.klx')
            }
        }
    );

    // Start the client
    client.start();

    context.subscriptions.push({
        dispose: () => {
            if (client) {
                client.stop();
            }
            if (serverProcess) {
                serverProcess.kill();
            }
        }
    });

    console.log('Kylix extension activated');
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
