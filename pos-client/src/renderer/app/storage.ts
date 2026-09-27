import { AUTO_PRINT_HISTORY_KEY, AUTO_PRINT_HISTORY_LIMIT, CONFIG_KEY, DEFAULT_CONFIG, ORDERS_KEY, ORDER_RETENTION_TOTAL_LIMIT, PAY_SESSION_KEY } from './constants.js';
import type { OrderRecord, POSConfig, PaySessionRecord } from '../types/domain.js';

export function normalizeApiBase(value: unknown): string {
	return String(value || '').trim().replace(/\/+$/, '');
}

function sanitizeConfig(config: Partial<POSConfig>): POSConfig {
	const normalized = { ...DEFAULT_CONFIG, ...config };
	delete (normalized as Record<string, unknown>).printerName;
	delete (normalized as Record<string, unknown>).autoPrint;
	return normalized;
}

export function loadConfig(): POSConfig {
	try {
		const stored = JSON.parse(localStorage.getItem(CONFIG_KEY) || '{}') as Record<string, unknown>;
		const legacyPrinterName = String(stored.printerName || '').trim();
		const hasLegacyAutoPrint = typeof stored.autoPrint === 'boolean';
		return sanitizeConfig({
			...DEFAULT_CONFIG,
			...stored,
			customerPrinterName: String(stored.customerPrinterName != null ? stored.customerPrinterName : legacyPrinterName).trim(),
			kitchenPrinterName: String(stored.kitchenPrinterName || '').trim(),
			autoPrintCustomer: typeof stored.autoPrintCustomer === 'boolean'
				? stored.autoPrintCustomer
				: hasLegacyAutoPrint
					? Boolean(stored.autoPrint)
					: DEFAULT_CONFIG.autoPrintCustomer,
			autoPrintKitchen: typeof stored.autoPrintKitchen === 'boolean'
				? stored.autoPrintKitchen
				: hasLegacyAutoPrint
					? Boolean(stored.autoPrint)
					: DEFAULT_CONFIG.autoPrintKitchen,
		});
	} catch {
		return { ...DEFAULT_CONFIG };
	}
}

export function saveConfig(config: POSConfig): void {
	localStorage.setItem(CONFIG_KEY, JSON.stringify(sanitizeConfig(config)));
}

export function loadOrders(): OrderRecord[] {
	try {
		const stored = JSON.parse(localStorage.getItem(ORDERS_KEY) || '[]') as unknown;
		return Array.isArray(stored) ? stored.slice(0, ORDER_RETENTION_TOTAL_LIMIT) : [];
	} catch {
		return [];
	}
}

export function saveOrders(orders: OrderRecord[]): void {
	localStorage.setItem(ORDERS_KEY, JSON.stringify((orders || []).slice(0, ORDER_RETENTION_TOTAL_LIMIT)));
}

export function loadPaySession(): PaySessionRecord {
	try {
		const stored = JSON.parse(localStorage.getItem(PAY_SESSION_KEY) || '{}') as Record<string, unknown>;
		return {
			payingOrder: stored && typeof stored.payingOrder === 'object' ? (stored.payingOrder as Record<string, unknown>) : null,
			payExpireAt: Number(stored.payExpireAt || 0),
			payCodeUrl: String(stored.payCodeUrl || ''),
			payQRCodeDataUrl: String(stored.payQRCodeDataUrl || ''),
			cashierMessage: String(stored.cashierMessage || ''),
			pendingCreateRequestId: String(stored.pendingCreateRequestId || ''),
		};
	} catch {
		return {
			payingOrder: null,
			payExpireAt: 0,
			payCodeUrl: '',
			payQRCodeDataUrl: '',
			cashierMessage: '',
			pendingCreateRequestId: '',
		};
	}
}

export function savePaySession(session: PaySessionRecord): void {
	localStorage.setItem(PAY_SESSION_KEY, JSON.stringify({
		payingOrder: session && session.payingOrder && typeof session.payingOrder === 'object' ? session.payingOrder : null,
		payExpireAt: Number(session && session.payExpireAt ? session.payExpireAt : 0),
		payCodeUrl: String(session && session.payCodeUrl ? session.payCodeUrl : ''),
		payQRCodeDataUrl: String(session && session.payQRCodeDataUrl ? session.payQRCodeDataUrl : ''),
		cashierMessage: String(session && session.cashierMessage ? session.cashierMessage : ''),
		pendingCreateRequestId: String(session && session.pendingCreateRequestId ? session.pendingCreateRequestId : ''),
	}));
}

export function clearPaySession(): void {
	localStorage.removeItem(PAY_SESSION_KEY);
}

export function loadAutoPrintHistory(): string[] {
	try {
		const stored = JSON.parse(localStorage.getItem(AUTO_PRINT_HISTORY_KEY) || '[]') as unknown;
		return Array.isArray(stored)
			? stored.filter((item) => typeof item === 'string').slice(0, AUTO_PRINT_HISTORY_LIMIT)
			: [];
	} catch {
		return [];
	}
}

export function saveAutoPrintHistory(history: string[]): void {
	localStorage.setItem(
		AUTO_PRINT_HISTORY_KEY,
		JSON.stringify((history || []).filter((item) => typeof item === 'string').slice(0, AUTO_PRINT_HISTORY_LIMIT)),
	);
}
