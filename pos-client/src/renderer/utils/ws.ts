import type { POSWSMessage } from '../types/domain.js';

export function buildWsUrl(apiBase: string, shopId: string, posToken: string): string {
	const normalizedBase = String(apiBase || '').trim().replace(/\/+$/, '');
	const wsBase = normalizedBase.replace(/^http:/i, 'ws:').replace(/^https:/i, 'wss:');
	return `${wsBase}/api/v1/pos/ws?shop_id=${encodeURIComponent(shopId)}&pos_token=${encodeURIComponent(posToken || '')}`;
}

interface SocketParams {
	url: string;
	onOpen?: () => void;
	onClose?: (details: { opened: boolean; code: number; reason: string }) => void;
	onError?: () => void;
	onMessage?: (message: POSWSMessage) => void;
}

interface SocketClient {
	close: () => void;
	send: (data: unknown) => void;
}

export function createPOSSocket({ url, onOpen, onClose, onError, onMessage }: SocketParams): SocketClient {
	const socket = new WebSocket(url);
	let manualClose = false;
	let opened = false;

	socket.addEventListener('open', () => {
		opened = true;
		if (typeof onOpen === 'function') {
			onOpen();
		}
	});

	socket.addEventListener('message', (event) => {
		try {
			const data = JSON.parse(event.data) as POSWSMessage;
			if (typeof onMessage === 'function') {
				onMessage(data);
			}
		} catch {
			// ignore malformed payloads
		}
	});

	socket.addEventListener('error', () => {
		if (typeof onError === 'function') {
			onError();
		}
	});

	socket.addEventListener('close', (event) => {
		if (!manualClose && typeof onClose === 'function') {
			onClose({
				opened,
				code: event.code,
				reason: event.reason || '',
			});
		}
	});

	return {
		close() {
			manualClose = true;
			socket.close();
		},
		send(data: unknown) {
			if (socket.readyState === WebSocket.OPEN) {
				socket.send(JSON.stringify(data));
			}
		},
	};
}
