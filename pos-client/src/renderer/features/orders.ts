import type { OrderRecord, POSWSMessage } from '../types/domain.js';
import {
	ORDER_FINALIZED_MAX_AGE_MS,
	ORDER_RETENTION_ACTIVE_LIMIT,
	ORDER_RETENTION_FINAL_LIMIT,
	ORDER_RETENTION_TOTAL_LIMIT,
} from '../app/constants.js';

interface OrdersFeatureOptions {
	state: {
		orders: OrderRecord[];
		lastEventAt: string;
		notice: string;
	};
	saveOrders: (orders: OrderRecord[]) => void;
	formatAmount: (value: unknown) => string;
	formatTime: (value: unknown) => string;
	render: () => void;
	onNewOrder: (order: OrderRecord) => void;
}

export function createOrdersFeature({
	state,
	saveOrders,
	formatAmount,
	formatTime,
	render,
	onNewOrder,
}: OrdersFeatureOptions) {
	let saveOrdersTimer: number | null = null;
	let renderOrdersTimer: number | null = null;

	function scheduleSaveOrders(): void {
		if (saveOrdersTimer) {
			window.clearTimeout(saveOrdersTimer);
		}
		saveOrdersTimer = window.setTimeout(() => {
			saveOrders(state.orders);
			saveOrdersTimer = null;
		}, 300);
	}

	function scheduleOrdersRender(): void {
		if (renderOrdersTimer) {
			return;
		}
		renderOrdersTimer = window.setTimeout(() => {
			renderOrdersTimer = null;
			render();
		}, 80);
	}

	function normalizeOrder(payload: any): OrderRecord | null {
		if (!payload) {
			return null;
		}
		return {
			id: Number(payload.id || payload.order_id || Date.now()),
			order_no: payload.order_no || '',
			total_amount: Number(payload.total_amount != null ? payload.total_amount : payload.total || 0),
			remark: payload.remark || '',
			status: Number(payload.status != null ? payload.status : 1),
			created_at: payload.created_at || payload.paid_at || new Date().toISOString(),
			table_no_snapshot: payload.table_no_snapshot || '',
			shop_name: payload.shop_name || (payload.shop && payload.shop.name) || '',
			items: Array.isArray(payload.items)
				? payload.items.map((item) => ({
					id: Number(item.id || 0),
					name: item.name || '',
					quantity: Number(item.quantity || 0),
					price: Number(item.price != null ? item.price : 0),
					subtotal: Number(item.subtotal != null ? item.subtotal : 0),
					option_summary: item.option_summary || '',
				}))
				: [],
		};
	}

	function isActiveOrderStatus(status: number): boolean {
		return status === 0 || status === 1 || status === 2;
	}

	function isFinalizedOrderExpired(order: OrderRecord, now = Date.now()): boolean {
		if (isActiveOrderStatus(Number(order.status))) {
			return false;
		}
		const createdAt = new Date(order.created_at || '').getTime();
		if (Number.isNaN(createdAt)) {
			return false;
		}
		return now - createdAt > ORDER_FINALIZED_MAX_AGE_MS;
	}

	function trimOrdersByPriority(orders: OrderRecord[]): OrderRecord[] {
		const now = Date.now();
		const activeOrders: OrderRecord[] = [];
		const finalizedOrders: OrderRecord[] = [];

		for (const order of orders) {
			if (isFinalizedOrderExpired(order, now)) {
				continue;
			}
			if (isActiveOrderStatus(Number(order.status))) {
				if (activeOrders.length < ORDER_RETENTION_ACTIVE_LIMIT) {
					activeOrders.push(order);
				}
				continue;
			}
			if (finalizedOrders.length < ORDER_RETENTION_FINAL_LIMIT) {
				finalizedOrders.push(order);
			}
		}

		return [...activeOrders, ...finalizedOrders].slice(0, ORDER_RETENTION_TOTAL_LIMIT);
	}

	function upsertOrder(order: OrderRecord): void {
		state.orders = trimOrdersByPriority([order, ...state.orders.filter((item) => item.order_no !== order.order_no)]);
		scheduleSaveOrders();
	}

	const trimmedInitialOrders = trimOrdersByPriority(state.orders);
	if (trimmedInitialOrders.length !== state.orders.length) {
		state.orders = trimmedInitialOrders;
		saveOrders(state.orders);
	}

	function playBeep() {
		try {
			const audioContext = new (window.AudioContext || window.webkitAudioContext)();
			const oscillator = audioContext.createOscillator();
			const gainNode = audioContext.createGain();
			oscillator.connect(gainNode);
			gainNode.connect(audioContext.destination);
			oscillator.frequency.value = 880;
			gainNode.gain.value = 0.06;
			oscillator.start();
			window.setTimeout(() => {
				oscillator.stop();
				audioContext.close();
			}, 180);
		} catch {
			// ignore audio failures on locked-down terminals
		}
	}

	function handleSocketMessage(message: POSWSMessage): void {
		if (!message || !message.type) {
			return;
		}
		if (message.type === 'new_order' || message.type === 'order_update') {
			const order = normalizeOrder(message.data);
			if (!order) {
				return;
			}
			upsertOrder(order);
			state.lastEventAt = formatTime(order.created_at || new Date().toISOString());
			if (message.type === 'new_order') {
				state.notice = `新订单 ${order.order_no}，金额 ${formatAmount(order.total_amount)}`;
				playBeep();
				onNewOrder(order);
			}
			scheduleOrdersRender();
		}
	}

	function findOrderByID(orderID: number): OrderRecord | undefined {
		return state.orders.find((item) => Number(item.id) === Number(orderID));
	}

	return {
		normalizeOrder,
		upsertOrder,
		handleSocketMessage,
		findOrderByID,
	};
}
