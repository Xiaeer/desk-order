import type { App, BrowserWindow, IpcMain } from 'electron';
import { getPrinters, printReceipt } from './printer';
import QRCode from 'qrcode';

interface RegisterIPCOptions {
	app: App;
	ipcMain: IpcMain;
	getMainWindow: () => BrowserWindow | null;
}

export function registerPOSIpcHandlers({ app, ipcMain, getMainWindow }: RegisterIPCOptions): void {
	ipcMain.handle('pos:getPrinters', async () => {
		const mainWindow = getMainWindow();
		if (!mainWindow) {
			return [];
		}
		return getPrinters(mainWindow);
	});

	ipcMain.handle('pos:printReceipt', async (_, payload) => printReceipt(payload));

	ipcMain.handle('pos:generateQRCode', async (_, text: unknown, size: unknown) => {
		const source = String(text || '').trim();
		if (!source) {
			return '';
		}
		const width = Number.isFinite(Number(size)) ? Math.max(128, Math.min(640, Number(size))) : 280;
		return QRCode.toDataURL(source, {
			type: 'image/png',
			margin: 1,
			width,
		});
	});

	ipcMain.handle('pos:getAppInfo', () => ({
		version: app.getVersion(),
		electron: process.versions.electron,
		node: process.versions.node,
	}));
}
