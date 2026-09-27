import type { DocumentType, POSConfig } from '../types/domain.js';

export const CONFIG_KEY = 'deskorder-pos-config-v1';
export const ORDERS_KEY = 'deskorder-pos-orders-v1';
export const PAY_SESSION_KEY = 'deskorder-pos-pay-session-v1';
export const AUTO_PRINT_HISTORY_KEY = 'deskorder-pos-auto-print-history-v1';
export const PAY_TIMEOUT_SECONDS = 180;
export const ORDER_RETENTION_ACTIVE_LIMIT = 80;
export const ORDER_RETENTION_FINAL_LIMIT = 40;
export const ORDER_RETENTION_TOTAL_LIMIT = ORDER_RETENTION_ACTIVE_LIMIT + ORDER_RETENTION_FINAL_LIMIT;
export const ORDER_FINALIZED_MAX_AGE_HOURS = 24;
export const ORDER_FINALIZED_MAX_AGE_MS = ORDER_FINALIZED_MAX_AGE_HOURS * 60 * 60 * 1000;
export const AUTO_PRINT_HISTORY_LIMIT = 200;

export const DEFAULT_CONFIG: POSConfig = {
	apiBase: 'http://127.0.0.1:8080',
	shopId: '',
	posToken: '',
	terminalName: 'POS-01',
	customerPrinterName: '',
	kitchenPrinterName: '',
	autoPrintCustomer: true,
	autoPrintKitchen: true,
};

export const statusLabelMap: Record<'idle' | 'connecting' | 'connected' | 'error', string> = {
	idle: '未连接',
	connecting: '连接中',
	connected: '已连接',
	error: '连接异常',
};

export const orderStatusMap: Record<number, string> = {
	0: '待支付',
	1: '已支付',
	2: '已接单',
	3: '已完成',
	4: '已取消',
};

export const documentTypeLabelMap: Record<DocumentType, string> = {
	customer_receipt: '顾客单',
	kitchen_ticket: '后厨单',
};
