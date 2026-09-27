import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('posBridge', {
	getPrinters: () => ipcRenderer.invoke('pos:getPrinters'),
	printReceipt: (payload) => ipcRenderer.invoke('pos:printReceipt', payload),
	generateQRCode: (text, size) => ipcRenderer.invoke('pos:generateQRCode', text, size),
	getAppInfo: () => ipcRenderer.invoke('pos:getAppInfo'),
});

export {};