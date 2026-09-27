import { app, BrowserWindow, ipcMain, type BrowserWindow as ElectronBrowserWindow } from 'electron';
import { createMainWindow } from './window';
import { registerPOSIpcHandlers } from './ipc';

let mainWindow: ElectronBrowserWindow | null = null;

app.disableHardwareAcceleration();

function openMainWindow() {
	mainWindow = createMainWindow();
	mainWindow.on('closed', () => {
		mainWindow = null;
	});
	return mainWindow;
}

app.whenReady().then(() => {
	openMainWindow();

	registerPOSIpcHandlers({
		app,
		ipcMain,
		getMainWindow: () => mainWindow,
	});

	app.on('activate', () => {
		if (BrowserWindow.getAllWindows().length === 0) {
			openMainWindow();
		}
	});
});

app.on('window-all-closed', () => {
	if (process.platform !== 'darwin') {
		app.quit();
	}
});
