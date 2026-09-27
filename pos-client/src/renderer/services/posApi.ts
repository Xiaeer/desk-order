import type { POSConfig } from '../types/domain.js';

interface APIEnvelope<T> {
	code: number;
	msg?: string;
	data: T;
}

interface RequestJSONOptions extends RequestInit {
	headers?: HeadersInit;
	timeoutMs?: number;
	retry?: number;
	retryDelayMs?: number;
}

function normalizePOSApiBase(apiBase: string): string {
	return String(apiBase || '')
		.trim()
		.replace(/\/+$/, '')
		.replace(/\/api\/v1(?:\/pos)?$/i, '');
}

function wait(delayMs: number): Promise<void> {
	return new Promise((resolve) => {
		window.setTimeout(resolve, delayMs);
	});
}

async function fetchWithTimeout(url: string, options: RequestJSONOptions): Promise<Response> {
	const controller = new AbortController();
	const timeoutMs = Math.max(1000, Number(options.timeoutMs || 10000));
	const timeoutTimer = window.setTimeout(() => {
		controller.abort();
	}, timeoutMs);
	try {
		return await fetch(url, {
			headers: {
				'Content-Type': 'application/json',
				...(options.headers || {}),
			},
			...options,
			signal: controller.signal,
		});
	} catch (error) {
		if (error instanceof DOMException && error.name === 'AbortError') {
			throw new Error(`请求超时(${timeoutMs}ms)`);
		}
		throw error;
	} finally {
		window.clearTimeout(timeoutTimer);
	}
}

function shouldRetry(error: unknown): boolean {
	const message = error instanceof Error ? error.message : String(error || '');
	return /超时|timeout|network|failed to fetch/i.test(message);
}

export async function requestJSON<T = unknown>(url: string, options: RequestJSONOptions = {}): Promise<T> {
	const maxRetry = Math.max(0, Number(options.retry || 0));
	const retryDelayMs = Math.max(0, Number(options.retryDelayMs || 250));

	for (let attempt = 0; attempt <= maxRetry; attempt += 1) {
		try {
			const response = await fetchWithTimeout(url, options);
			const raw = await response.text();
			let payload: APIEnvelope<T> | null = null;
			try {
				payload = raw ? (JSON.parse(raw) as APIEnvelope<T>) : null;
			} catch {
				payload = null;
			}
			if (!response.ok) {
				throw new Error((payload && payload.msg) || `请求失败(${response.status})`);
			}
			if (!payload || payload.code !== 0) {
				throw new Error((payload && payload.msg) || '接口返回异常');
			}
			return payload.data;
		} catch (error) {
			if (attempt >= maxRetry || !shouldRetry(error)) {
				throw error;
			}
			await wait(retryDelayMs * (attempt + 1));
		}
	}

	throw new Error('请求失败');
}

export function requestPOSPost<TReq = unknown, TResp = unknown>(config: POSConfig, path: string, body: TReq): Promise<TResp> {
	const base = normalizePOSApiBase(config.apiBase);
	return requestJSON<TResp>(`${base}/api/v1/pos${path}`, {
		method: 'POST',
		body: JSON.stringify(body),
		timeoutMs: 10000,
		retry: 0,
	});
}

export function requestPOSGet<TResp = unknown>(config: POSConfig, path: string, params?: Record<string, string | number>): Promise<TResp> {
	const search = new URLSearchParams(
		Object.entries(params || {}).reduce<Record<string, string>>((acc, [key, value]) => {
			acc[key] = String(value);
			return acc;
		}, {}),
	).toString();
	const query = search ? `?${search}` : '';
	const base = normalizePOSApiBase(config.apiBase);
	return requestJSON<TResp>(`${base}/api/v1/pos${path}${query}`, {
		method: 'GET',
		timeoutMs: 8000,
		retry: 1,
	});
}

export async function pingPOSLogin(config: POSConfig): Promise<APIEnvelope<unknown>> {
	const base = normalizePOSApiBase(config.apiBase);
	const data = await requestJSON<unknown>(`${base}/api/v1/pos/login`, {
		method: 'POST',
		body: JSON.stringify({
			shop_id: Number(config.shopId),
			pos_token: config.posToken,
			terminal_name: config.terminalName,
		}),
		timeoutMs: 6000,
		retry: 1,
	});
	return {
		code: 0,
		msg: '',
		data,
	};
}
