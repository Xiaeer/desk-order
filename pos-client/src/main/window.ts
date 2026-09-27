import * as path from 'path';
import { BrowserWindow } from 'electron';

export function createMainWindow(): BrowserWindow {
	const window = new BrowserWindow({
		width: 1280,
		height: 820,
		minWidth: 1080,
		minHeight: 720,
		autoHideMenuBar: true,
		backgroundColor: '#f1ede1',
		webPreferences: {
			preload: path.join(__dirname, 'preload.js'),
			contextIsolation: true,
			nodeIntegration: false,
			sandbox: false,
		},
	});

	window.loadFile(path.join(__dirname, '..', '..', 'src', 'renderer', 'index.html'));
	return window;
}
