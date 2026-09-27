import type { DocumentType, OrderRecord, POSConfig } from '../types/domain.js';

interface PrintFeatureOptions {
	state: {
		config: POSConfig;
		printMessage: string;
		autoPrintedOrderKeys: string[];
	};
	render: () => void;
	documentTypeLabelMap: Record<DocumentType, string>;
	saveAutoPrintHistory: (history: string[]) => void;
}

export function createPrintFeature({
	state,
	render,
	documentTypeLabelMap,
	saveAutoPrintHistory,
}: PrintFeatureOptions) {
	let printQueue: Promise<void> = Promise.resolve();

	function buildAutoPrintKey(orderNo: string, documentType: DocumentType): string {
		return `${orderNo}:${documentType}`;
	}

	function hasAutoPrinted(orderNo: string, documentType: DocumentType): boolean {
		return state.autoPrintedOrderKeys.includes(buildAutoPrintKey(orderNo, documentType));
	}

	function markAutoPrinted(orderNo: string, documentType: DocumentType): void {
		const nextKey = buildAutoPrintKey(orderNo, documentType);
		state.autoPrintedOrderKeys = [nextKey, ...state.autoPrintedOrderKeys.filter((item) => item !== nextKey)].slice(0, 200);
		saveAutoPrintHistory(state.autoPrintedOrderKeys);
	}

	function enqueuePrintTask(task: () => Promise<void>): Promise<void> {
		const queued = printQueue.then(task);
		printQueue = queued
			.then(() => undefined)
			.catch(() => undefined);
		return queued;
	}

	function buildSampleOrder(): OrderRecord {
		return {
			id: Date.now(),
			order_no: `TEST-${Date.now()}`,
			total_amount: 3600,
			remark: '少辣，米饭先上',
			status: 1,
			created_at: new Date().toISOString(),
			shop_name: 'DeskOrder 测试门店',
			table_no_snapshot: 'A08',
			items: [
				{ name: '辣子鸡', quantity: 1, subtotal: 2200, option_summary: '中辣，加年糕' },
				{ name: '米饭', quantity: 2, subtotal: 1400, option_summary: '' },
			],
		};
	}

	function getDocumentLabel(documentType: DocumentType): string {
		return documentTypeLabelMap[documentType] || '单据';
	}

	function resolvePrinterName(documentType: DocumentType): string {
		if (documentType === 'kitchen_ticket') {
			return state.config.kitchenPrinterName || state.config.customerPrinterName || '';
		}
		return state.config.customerPrinterName || '';
	}

	function resolvePrinterLabel(documentType: DocumentType): string {
		const printerName = resolvePrinterName(documentType);
		return printerName || '系统默认';
	}

	function isAutoPrintEnabled(documentType: DocumentType): boolean {
		return documentType === 'kitchen_ticket' ? state.config.autoPrintKitchen : state.config.autoPrintCustomer;
	}

	function buildPrintJobs(scope: string, manual: boolean) {
		const requestedTypes: DocumentType[] = scope === 'customer'
			? ['customer_receipt']
			: scope === 'kitchen'
				? ['kitchen_ticket']
				: ['customer_receipt', 'kitchen_ticket'];

		return requestedTypes
			.filter((documentType) => manual || isAutoPrintEnabled(documentType))
			.map((documentType) => ({
				documentType,
				label: getDocumentLabel(documentType),
				printerName: resolvePrinterName(documentType),
				printerLabel: resolvePrinterLabel(documentType),
			}));
	}

	function buildPrintSummary(documentType: DocumentType): string {
		const autoLabel = isAutoPrintEnabled(documentType) ? '自动开' : '自动关';
		return `${resolvePrinterLabel(documentType)} · ${autoLabel}`;
	}

	async function runPrintOrder(order: OrderRecord, options: { manual?: boolean; scope?: string } = {}): Promise<void> {
		const manual = Boolean(options.manual);
		const scope = options.scope || 'all';
		if (!window.posBridge || typeof window.posBridge.printReceipt !== 'function') {
			state.printMessage = '当前环境不支持 Electron 打印';
			render();
			return;
		}
		const jobs = buildPrintJobs(scope, manual);
		const orderNo = String(order.order_no || '').trim();
		const filteredJobs = manual || !orderNo
			? jobs
			: jobs.filter((job) => !hasAutoPrinted(orderNo, job.documentType));
		if (!filteredJobs.length) {
			state.printMessage = manual ? '当前没有可打印的单据' : `自动打印已关闭：${order.order_no}`;
			render();
			return;
		}
		const results = [];
		try {
			for (const job of filteredJobs) {
				try {
					const result = await window.posBridge.printReceipt({
						order,
						documentType: job.documentType,
						printerName: job.printerName,
					});
					const bridgeResult = (result && typeof result === 'object'
						? (result as { success?: boolean; error?: string })
						: null);
					if (bridgeResult && bridgeResult.success === false) {
						results.push(`${job.label}失败：${bridgeResult.error || '打印失败'}`);
						continue;
					}
					if (!manual && orderNo) {
						markAutoPrinted(orderNo, job.documentType);
					}
					results.push(`${job.label} -> ${job.printerLabel}`);
				} catch (error) {
					results.push(`${job.label}失败：${error.message}`);
				}
			}
			state.printMessage = `${manual ? '手动打印' : '自动打印'} ${order.order_no}：${results.join('；')}`;
		} catch (error) {
			state.printMessage = `打印失败：${error.message}`;
		}
		render();
	}

	function printOrder(order: OrderRecord, options: { manual?: boolean; scope?: string } = {}): Promise<void> {
		return enqueuePrintTask(() => runPrintOrder(order, options));
	}

	return {
		buildSampleOrder,
		buildPrintSummary,
		printOrder,
	};
}
