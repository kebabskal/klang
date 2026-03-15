import * as path from 'path';
import * as fs from 'fs';
import { workspace, ExtensionContext, window } from 'vscode';
import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
} from 'vscode-languageclient/node';

let client: LanguageClient | undefined;
let outputChannel = window.createOutputChannel('Klang');

function log(msg: string) {
	outputChannel.appendLine(`[${new Date().toLocaleTimeString()}] ${msg}`);
}

export function activate(context: ExtensionContext) {
	log('Klang extension activating...');
	log(`Extension path: ${context.extensionPath}`);

	const serverPath = findLspServer(context);
	if (!serverPath) {
		const msg = 'Klang LSP server (kl-lsp) not found. Build it with: go build -o bin/kl-lsp ./cmd/kl-lsp';
		log('ERROR: ' + msg);
		window.showWarningMessage(msg);
		return;
	}

	log(`LSP server found: ${serverPath}`);

	const serverOptions: ServerOptions = {
		run: { command: serverPath },
		debug: { command: serverPath },
	};

	const clientOptions: LanguageClientOptions = {
		documentSelector: [{ scheme: 'file', language: 'klang' }],
		outputChannel: outputChannel,
	};

	client = new LanguageClient(
		'klang',
		'Klang Language Server',
		serverOptions,
		clientOptions
	);

	client.start().then(() => {
		log('LSP client started successfully');
	}, (err) => {
		log(`ERROR starting LSP client: ${err}`);
	});
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}

function findLspServer(context: ExtensionContext): string | undefined {
	// 1. Check user setting
	const configured = workspace.getConfiguration('klang').get<string>('lspPath');
	if (configured && configured.length > 0 && fs.existsSync(configured)) {
		log(`Found LSP via user setting: ${configured}`);
		return configured;
	}

	// 2. Check next to the extension itself (symlinked/junctioned from klang repo)
	const extDir = context.extensionPath;
	const extDirs = [extDir];
	try {
		const realDir = fs.realpathSync(extDir);
		if (realDir !== extDir) {
			log(`Extension junction resolved: ${extDir} -> ${realDir}`);
			extDirs.push(realDir);
		}
	} catch {}
	for (const dir of extDirs) {
		const candidates = [
			path.join(dir, '..', '..', 'bin', 'kl-lsp.exe'),
			path.join(dir, '..', '..', 'bin', 'kl-lsp'),
		];
		for (const c of candidates) {
			const resolved = path.resolve(c);
			log(`Checking: ${resolved}`);
			if (fs.existsSync(resolved)) {
				log(`Found LSP via extension path: ${resolved}`);
				return resolved;
			}
		}
	}

	// 3. Check workspace root bin/ directory
	const workspaceFolders = workspace.workspaceFolders;
	if (workspaceFolders) {
		for (const folder of workspaceFolders) {
			const candidates = [
				path.join(folder.uri.fsPath, 'bin', 'kl-lsp.exe'),
				path.join(folder.uri.fsPath, 'bin', 'kl-lsp'),
			];
			for (const c of candidates) {
				log(`Checking: ${c}`);
				if (fs.existsSync(c)) {
					log(`Found LSP via workspace: ${c}`);
					return c;
				}
			}
		}
	}

	// 4. Check if kl-lsp is on PATH
	const ext = process.platform === 'win32' ? '.exe' : '';
	const pathDirs = (process.env.PATH || '').split(path.delimiter);
	for (const dir of pathDirs) {
		const candidate = path.join(dir, 'kl-lsp' + ext);
		if (fs.existsSync(candidate)) {
			log(`Found LSP on PATH: ${candidate}`);
			return candidate;
		}
	}

	log('LSP server not found in any location');
	return undefined;
}
