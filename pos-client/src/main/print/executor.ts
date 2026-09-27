import { BrowserWindow, type BrowserWindow as ElectronBrowserWindow, type PrinterInfo, type WebContents, type WebContentsPrintOptions } from 'electron';
import { buildReceiptHtml } from './templates';

const THERMAL_DPI = 203;
let sharedPrintWindow: BrowserWindow | null = null;
let printQueue: Promise<void> = Promise.resolve();

type PrintPayload = {
	order?: OrderLike;
	documentType?: string;
	printerName?: string;
};

type OrderLike = {
	order_no?: string;
	[key: string]: unknown;
};

type PrintResult = {
	success: boolean;
	error?: string;
};

function buildPrintOptions(payload?: PrintPayload): WebContentsPrintOptions {
	const options: WebContentsPrintOptions & { usePrinterDefaultPageSize?: boolean } = {
		silent: true,
		printBackground: false,
		color: false,
		margins: {
			marginType: 'none',
		},
		scaleFactor: 100,
		dpi: {
			horizontal: THERMAL_DPI,
			vertical: THERMAL_DPI,
		},
		usePrinterDefaultPageSize: true,
	};

	if (payload && payload.printerName) {
		options.deviceName = payload.printerName;
	}

	return options;
}

async function waitForPrintReady(webContents: WebContents): Promise<void> {
	try {
		await webContents.executeJavaScript(
			`document.fonts ? document.fonts.ready.then(() => true) : true`,
			true,
		);
	} catch {
		// Ignore font readiness failures and fall back to immediate printing.
	}
}

export async function getPrinters(mainWindow: ElectronBrowserWindow | null): Promise<PrinterInfo[]> {
	if (!mainWindow || mainWindow.isDestroyed()) {
		return [];
	}
	if (typeof mainWindow.webContents.getPrintersAsync === 'function') {
		return mainWindow.webContents.getPrintersAsync();
	}
	return mainWindow.webContents.getPrinters();
}

function getSharedPrintWindow(): BrowserWindow {
	if (sharedPrintWindow && !sharedPrintWindow.isDestroyed()) {
		return sharedPrintWindow;
	}
	sharedPrintWindow = new BrowserWindow({
		show: false,
		width: 320,
		height: 700,
		backgroundColor: '#ffffff',
		webPreferences: {
			sandbox: false,
		},
	});
	sharedPrintWindow.once('closed', () => {
		sharedPrintWindow = null;
	});
	return sharedPrintWindow;
}

async function runSinglePrint(payload?: PrintPayload): Promise<{ success: true }> {
	const order: OrderLike | undefined = payload && payload.order ? payload.order : (payload as OrderLike | undefined);
	if (!order || !order.order_no) {
		throw new Error('缺少可打印的订单数据');
	}

	const documentType = payload && payload.documentType ? payload.documentType : 'customer_receipt';
	const html = buildReceiptHtml(order as any, documentType);
	const printWindow = getSharedPrintWindow();

	return new Promise<{ success: true }>((resolve, reject) => {
		let settled = false;

		const cleanup = () => {
			printWindow.webContents.removeListener('did-finish-load', onDidFinishLoad);
			printWindow.webContents.removeListener('did-fail-load', onDidFailLoad);
		};

		const finish = (handler: (value: any) => void, value: any) => {
			if (settled) {
				return;
			}
			settled = true;
			cleanup();
			handler(value);
		};

		const onDidFinishLoad = async () => {
			await waitForPrintReady(printWindow.webContents);
			printWindow.webContents.print(
				buildPrintOptions(payload),
				(success, errorType) => {
					if (!success) {
						finish(reject, new Error(errorType || '打印失败'));
						return;
					}
					finish(resolve, { success: true });
				},
			);
		};

		const onDidFailLoad = (_: unknown, code: number, desc: string) => {
			finish(reject, new Error(`打印页面加载失败: ${code} ${desc}`));
		};

		printWindow.webContents.once('did-finish-load', onDidFinishLoad);
		printWindow.webContents.once('did-fail-load', onDidFailLoad);

		printWindow
			.loadURL(`data:text/html;charset=UTF-8,${encodeURIComponent(html)}`)
			.catch((error: Error) => {
				finish(reject, error);
			});
	});
}

export async function printReceipt(payload?: PrintPayload): Promise<PrintResult> {
	const task = printQueue.then(async () => {
		try {
			await runSinglePrint(payload);
			return { success: true };
		} catch (error) {
			const message = error instanceof Error ? error.message : String(error || '打印失败');
			return { success: false, error: message };
		}
	});
	printQueue = task
		.then(() => undefined)
		.catch(() => undefined);
	return task;
}
