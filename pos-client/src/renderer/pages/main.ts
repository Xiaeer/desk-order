import { buildWsUrl, createPOSSocket } from '../utils/ws.js';
import {
	PAY_TIMEOUT_SECONDS,
	documentTypeLabelMap,
	orderStatusMap,
	statusLabelMap,
} from '../app/constants.js';
import {
	loadAutoPrintHistory,
	loadConfig,
	loadOrders,
	loadPaySession,
	normalizeApiBase,
	clearPaySession,
	saveAutoPrintHistory,
	saveConfig,
	savePaySession,
	saveOrders,
} from '../app/storage.js';
import {
	escapeHtml,
	formatAmount,
	formatCountdown,
	formatOptionPriceDelta,
	formatTime,
	getQRImageUrl,
} from '../app/format.js';
import { pingPOSLogin, requestPOSGet, requestPOSPost } from '../services/posApi.js';
import { createPOSConnectionService } from '../services/posConnection.js';
import { createCashierFeature } from '../features/cashier.js';
import { createOrdersFeature } from '../features/orders.js';
import { createPrintFeature } from '../features/print.js';
import { renderApp } from '../components/render.js';

const initialPaySession = loadPaySession();

const state = {
	config: loadConfig(),
	orders: loadOrders(),
	ordersRenderLimit: 20,
	ordersStatusFilter: 'all',
	ordersTimeWindowMinutes: 0,
	menu: null,
	menuLoading: false,
	menuMessage: '',
	cashierMenuKeyword: '',
	cashierMenuKeywordDebounced: '',
	cashierMenuKeywordTimer: null,
	cashierExpandedCategoryIDs: [],
	cart: [],
	cashierTableNo: '',
	cashierRemark: '',
	cashierBusy: false,
	cashierMessage: initialPaySession.cashierMessage || '',
	payingOrder: initialPaySession.payingOrder,
	pendingCreateRequestId: initialPaySession.pendingCreateRequestId || '',
	payExpireAt: Number(initialPaySession.payExpireAt || 0),
	payCodeUrl: initialPaySession.payCodeUrl || '',
	payQRCodeDataUrl: initialPaySession.payQRCodeDataUrl || '',
	payPollTimer: null,
	payCountdownTimer: null,
	optionPicker: null,
	printers: [],
	screen: 'orders',
	connectionStatus: 'idle',
	connectionMessage: '请先绑定店铺',
	notice: '',
	lastEventAt: '',
	printMessage: '',
	autoPrintedOrderKeys: loadAutoPrintHistory(),
	reconnectTimer: null,
	socketClient: null,
	appInfo: null,
};

const root = document.getElementById('app');
let isRenderScheduled = false;
let noticeAutoHideTimer: number | null = null;
let lastNoticeText = '';

const printFeature = createPrintFeature({
	state,
	render,
	documentTypeLabelMap,
	saveAutoPrintHistory,
});

const ordersFeature = createOrdersFeature({
	state,
	saveOrders,
	formatAmount,
	formatTime,
	render,
	onNewOrder: (order) => {
		if (state.config.autoPrintCustomer || state.config.autoPrintKitchen) {
			printFeature.printOrder(order, { manual: false, scope: 'all' });
		}
	},
});

const cashierFeature = createCashierFeature({
	state,
	render,
	updatePayCountdownDisplay,
	updateCashierCartDisplay,
	generatePayQRCode,
	hasConfig,
	buildPOSAuthPayload,
	requestPOSPost,
	requestPOSGet,
	normalizeOrder: ordersFeature.normalizeOrder,
	upsertOrder: ordersFeature.upsertOrder,
	savePaySession,
	clearPaySession,
	onPaidOrder: (order) => {
		if (
			order
			&& (state.config.autoPrintCustomer || state.config.autoPrintKitchen)
			&& state.connectionStatus !== 'connected'
		) {
			printFeature.printOrder(order, { manual: false, scope: 'all' });
		}
	},
	PAY_TIMEOUT_SECONDS,
});

if (state.payingOrder && Number(state.payingOrder.id) > 0) {
	upsertOrder(normalizeOrder(state.payingOrder));
}

const connectionService = createPOSConnectionService({
	state,
	render,
	hasConfig,
	buildWsUrl,
	createPOSSocket,
	pingPOSLogin,
	formatTime,
	onMessage: handleSocketMessage,
});

bootstrap();

async function bootstrap() {
	state.screen = hasConfig() ? 'orders' : 'setup';
	render();
	bindEvents();
	void initializeStartupTasks();
}

async function initializeStartupTasks() {
	void refreshAppInfo();
	void refreshPrinters().then(render);
	if (state.payingOrder && Number(state.payingOrder.id) > 0) {
		resumePayingWatchers();
		void refreshPayingOrderStatus();
	}
	if (hasConfig()) {
		void connectPOS();
		window.setTimeout(() => {
			void ensureCashierMenu(false);
		}, 1200);
	}
}

async function refreshAppInfo() {
	if (!window.posBridge || typeof window.posBridge.getAppInfo !== 'function') {
		return;
	}
	try {
		state.appInfo = await window.posBridge.getAppInfo();
		render();
	} catch {
		// ignore app info fetch failure in startup path
	}
}

function bindEvents() {
	root.addEventListener('click', handleClick);
	root.addEventListener('input', handleInput);
	root.addEventListener('submit', handleSubmit);
	root.addEventListener('change', handleChange);
	document.addEventListener('visibilitychange', handleVisibilityChange);
}

function handleInput(event) {
	if (!event.target.matches('[name="cashierMenuKeyword"]')) {
		return;
	}
	state.cashierMenuKeyword = String(event.target.value || '');
	if (state.cashierMenuKeywordTimer) {
		window.clearTimeout(state.cashierMenuKeywordTimer);
		state.cashierMenuKeywordTimer = null;
	}
	state.cashierMenuKeywordTimer = window.setTimeout(() => {
		state.cashierMenuKeywordTimer = null;
		const nextKeyword = String(state.cashierMenuKeyword || '').trim().toLowerCase();
		if (nextKeyword === String(state.cashierMenuKeywordDebounced || '')) {
			return;
		}
		state.cashierMenuKeywordDebounced = nextKeyword;
		render();
	}, 180);
}

function handleVisibilityChange() {
	if (document.hidden) {
		if (state.payPollTimer) {
			window.clearInterval(state.payPollTimer);
			state.payPollTimer = null;
		}
		if (state.payCountdownTimer) {
			window.clearInterval(state.payCountdownTimer);
			state.payCountdownTimer = null;
		}
		return;
	}
	if (state.payingOrder && Number(state.payingOrder.id) > 0) {
		resumePayingWatchers();
		refreshPayingOrderStatus();
	}
}

function handleClick(event) {
	const actionTarget = event.target.closest('[data-action]');
	if (!actionTarget) {
		return;
	}
	const action = actionTarget.dataset.action;
	if (action === 'switch-screen') {
		state.screen = actionTarget.dataset.screen;
		if (state.screen === 'cashier') {
			ensureCashierMenu();
		}
		render();
		return;
	}
	if (action === 'reconnect') {
		connectPOS(true);
		return;
	}
	if (action === 'disconnect') {
		disconnectPOS('已手动断开');
		render();
		return;
	}
	if (action === 'refresh-printers') {
		refreshPrinters().then(render);
		return;
	}
	if (action === 'test-print') {
		printOrder(buildSampleOrder(), { manual: true, scope: 'all' });
		return;
	}
	if (action === 'print-order') {
		const orderId = Number(actionTarget.dataset.orderId);
		const order = state.orders.find((item) => Number(item.id) === orderId);
		if (order) {
			printOrder(order, {
				manual: true,
				scope: actionTarget.dataset.printScope || 'all',
			});
		}
		return;
	}
	if (action === 'clear-notice') {
		state.notice = '';
		render();
		return;
	}
	if (action === 'orders-load-more') {
		state.ordersRenderLimit = Math.min(200, Number(state.ordersRenderLimit || 20) + 20);
		render();
		return;
	}
	if (action === 'cashier-refresh-menu') {
		ensureCashierMenu(true);
		return;
	}
	if (action === 'cashier-toggle-category') {
		const categoryID = String(actionTarget.dataset.categoryId || '').trim();
		if (!categoryID) {
			return;
		}
		const current = Array.isArray(state.cashierExpandedCategoryIDs) ? [...state.cashierExpandedCategoryIDs] : [];
		const index = current.indexOf(categoryID);
		if (index >= 0) {
			current.splice(index, 1);
		} else {
			current.push(categoryID);
		}
		state.cashierExpandedCategoryIDs = current;
		render();
		return;
	}
	if (action === 'cart-add-item') {
		addCartItem(Number(actionTarget.dataset.productId));
		return;
	}
	if (action === 'cart-increase-item') {
		changeCartItemQuantity(Number(actionTarget.dataset.cartIndex), 1);
		return;
	}
	if (action === 'cart-decrease-item') {
		changeCartItemQuantity(Number(actionTarget.dataset.cartIndex), -1);
		return;
	}
	if (action === 'cart-remove-item') {
		removeCartItem(Number(actionTarget.dataset.cartIndex));
		return;
	}
	if (action === 'cashier-create-order') {
		createCashierOrderAndPay();
		return;
	}
	if (action === 'cashier-refresh-status') {
		refreshPayingOrderStatus();
		return;
	}
	if (action === 'cashier-cancel-order') {
		cancelPayingOrder('已手动取消订单');
		return;
	}
	if (action === 'cashier-reset-order') {
		cashierFeature.clearPayingSession();
		return;
	}
	if (action === 'cashier-copy-code-url') {
		copyPayCodeUrl();
		return;
	}
	if (action === 'option-picker-cancel') {
		state.optionPicker = null;
		render();
		return;
	}
	if (action === 'option-picker-select') {
		toggleOptionPickerValue(actionTarget.dataset.groupId, actionTarget.dataset.valueId);
		return;
	}
	if (action === 'option-picker-confirm') {
		confirmOptionPicker();
	}
}

async function handleSubmit(event) {
	if (!event.target.matches('[data-form="config"]')) {
		return;
	}
	event.preventDefault();
	const form = new FormData(event.target);
	const nextConfig = {
		apiBase: normalizeApiBase(form.get('apiBase')),
		shopId: String(form.get('shopId') || '').trim(),
		posToken: String(form.get('posToken') || '').trim(),
		terminalName: String(form.get('terminalName') || '').trim() || 'POS-01',
		customerPrinterName: String(form.get('customerPrinterName') || '').trim(),
		kitchenPrinterName: String(form.get('kitchenPrinterName') || '').trim(),
		autoPrintCustomer: form.get('autoPrintCustomer') === 'on',
		autoPrintKitchen: form.get('autoPrintKitchen') === 'on',
	};

	if (!nextConfig.apiBase || !nextConfig.shopId || !nextConfig.posToken) {
		state.connectionStatus = 'error';
		state.connectionMessage = '请填写 API 地址、shop_id 和 POS 密钥';
		render();
		return;
	}
	if (state.payingOrder && Number(state.payingOrder.id) > 0 && Number(state.payingOrder.status || 0) === 0) {
		state.connectionStatus = 'error';
		state.connectionMessage = `订单 ${state.payingOrder.order_no} 正在支付中，请先完成或取消本单后再修改配置`;
		render();
		return;
	}

	state.config = nextConfig;
	saveConfig(nextConfig);
	state.screen = 'orders';
	state.printMessage = '店铺绑定已保存';
	resetCashierPayingState();
	state.cart = [];
	state.menu = null;
	render();
	await connectPOS(true);
	await ensureCashierMenu(true);
}

function handleChange(event) {
	if (event.target.matches('[name="customerPrinterName"]')) {
		state.config.customerPrinterName = event.target.value;
		saveConfig(state.config);
		return;
	}
	if (event.target.matches('[name="kitchenPrinterName"]')) {
		state.config.kitchenPrinterName = event.target.value;
		saveConfig(state.config);
		return;
	}
	if (event.target.matches('[name="autoPrintCustomerToggle"]')) {
		state.config.autoPrintCustomer = event.target.checked;
		saveConfig(state.config);
		render();
		return;
	}
	if (event.target.matches('[name="autoPrintKitchenToggle"]')) {
		state.config.autoPrintKitchen = event.target.checked;
		saveConfig(state.config);
		render();
		return;
	}
	if (event.target.matches('[name="ordersStatusFilter"]')) {
		state.ordersStatusFilter = String(event.target.value || 'all');
		state.ordersRenderLimit = 20;
		render();
		return;
	}
	if (event.target.matches('[name="ordersTimeWindowMinutes"]')) {
		const minutes = Number(event.target.value || 0);
		state.ordersTimeWindowMinutes = Number.isFinite(minutes) ? Math.max(0, minutes) : 0;
		state.ordersRenderLimit = 20;
		render();
		return;
	}
	if (event.target.matches('[name="cashierTableNo"]')) {
		state.cashierTableNo = event.target.value;
		state.pendingCreateRequestId = '';
		savePaySession({
			payingOrder: state.payingOrder,
			payExpireAt: Number(state.payExpireAt || 0),
			payCodeUrl: String(state.payCodeUrl || ''),
			payQRCodeDataUrl: String(state.payQRCodeDataUrl || ''),
			cashierMessage: String(state.cashierMessage || ''),
			pendingCreateRequestId: '',
		});
		return;
	}
	if (event.target.matches('[name="cashierRemark"]')) {
		state.cashierRemark = event.target.value;
		state.pendingCreateRequestId = '';
		savePaySession({
			payingOrder: state.payingOrder,
			payExpireAt: Number(state.payExpireAt || 0),
			payCodeUrl: String(state.payCodeUrl || ''),
			payQRCodeDataUrl: String(state.payQRCodeDataUrl || ''),
			cashierMessage: String(state.cashierMessage || ''),
			pendingCreateRequestId: '',
		});
	}
}

function hasConfig() {
	return Boolean(state.config.apiBase && state.config.shopId && state.config.posToken);
}

function buildPOSAuthPayload() {
	return {
		shop_id: Number(state.config.shopId),
		pos_token: state.config.posToken,
	};
}

function buildSampleOrder() {
	return printFeature.buildSampleOrder();
}

function addCartItem(productID) {
	return cashierFeature.addCartItem(productID);
}

function changeCartItemQuantity(cartIndex, delta) {
	return cashierFeature.changeCartItemQuantity(cartIndex, delta);
}

function removeCartItem(cartIndex) {
	return cashierFeature.removeCartItem(cartIndex);
}

function createCashierOrderAndPay() {
	return cashierFeature.createCashierOrderAndPay();
}

function ensureCashierMenu(force = false) {
	return cashierFeature.ensureCashierMenu(force);
}

function refreshPayingOrderStatus() {
	return cashierFeature.refreshPayingOrderStatus();
}

function cancelPayingOrder(successMessage) {
	return cashierFeature.cancelPayingOrder(successMessage);
}

function copyPayCodeUrl() {
	return cashierFeature.copyPayCodeUrl();
}

function toggleOptionPickerValue(groupID, valueID) {
	return cashierFeature.toggleOptionPickerValue(groupID, valueID);
}

function confirmOptionPicker() {
	return cashierFeature.confirmOptionPicker();
}

function resetCashierPayingState() {
	return cashierFeature.resetCashierPayingState();
}

function resumePayingWatchers() {
	return cashierFeature.resumePayingWatchers();
}

function getPayCountdownSeconds() {
	return cashierFeature.getPayCountdownSeconds();
}

function getCartTotalAmount() {
	return cashierFeature.getCartTotalAmount();
}

async function refreshPrinters() {
	if (!window.posBridge || typeof window.posBridge.getPrinters !== 'function') {
		state.printers = [];
		return;
	}
	try {
		state.printers = await window.posBridge.getPrinters();
	} catch (error) {
		state.printMessage = `获取打印机失败：${error.message}`;
	}
}

async function connectPOS(forceReconnect = false) {
	return connectionService.connectPOS(forceReconnect);
}

function disconnectPOS(message) {
	return connectionService.disconnectPOS(message);
}

function handleSocketMessage(message) {
	return ordersFeature.handleSocketMessage(message);
}

function normalizeOrder(payload) {
	return ordersFeature.normalizeOrder(payload);
}

function upsertOrder(order) {
	return ordersFeature.upsertOrder(order);
}

function buildPrintSummary(documentType) {
	return printFeature.buildPrintSummary(documentType);
}

async function printOrder(order, options = {}) {
	return printFeature.printOrder(order, options);
}

function render() {
	if (isRenderScheduled) {
		return;
	}
	isRenderScheduled = true;
	window.requestAnimationFrame(() => {
		isRenderScheduled = false;
		renderNow();
	});
}

function renderNow() {
	root.innerHTML = renderApp({
		state,
		statusLabelMap,
		orderStatusMap,
		helpers: {
			escapeHtml,
			formatAmount,
			formatTime,
			formatCountdown,
			formatOptionPriceDelta,
			getQRImageUrl,
			getPayCountdownSeconds,
			getCartTotalAmount,
			buildPrintSummary,
		},
	});
	updateNoticeAutoHide();
}

function updateNoticeAutoHide() {
	const currentNotice = String(state.notice || '');
	if (!currentNotice) {
		lastNoticeText = '';
		if (noticeAutoHideTimer) {
			window.clearTimeout(noticeAutoHideTimer);
			noticeAutoHideTimer = null;
		}
		return;
	}
	if (currentNotice === lastNoticeText) {
		return;
	}
	lastNoticeText = currentNotice;
	if (noticeAutoHideTimer) {
		window.clearTimeout(noticeAutoHideTimer);
		noticeAutoHideTimer = null;
	}
	noticeAutoHideTimer = window.setTimeout(() => {
		noticeAutoHideTimer = null;
		if (!state.notice || state.notice !== currentNotice) {
			return;
		}
		state.notice = '';
		render();
	}, 5000);
}

function updatePayCountdownDisplay(seconds = getPayCountdownSeconds()) {
	const countdownNode = root.querySelector('[data-role="pay-countdown"]');
	if (!countdownNode) {
		return;
	}
	countdownNode.textContent = formatCountdown(seconds);
}

function buildCashierCartItemsHTML() {
	if (!state.cart.length) {
		return '<li>购物车为空</li>';
	}
	return state.cart.map((item, index) => `
		<li>
		  <div class="cashier-cart-main">
		    <div class="cashier-cart-name">${escapeHtml(item.name)}</div>
		    <div class="cashier-cart-price">${formatAmount(item.price)}</div>
		  </div>
		  ${item.option_summary ? `<div class="item-option">${escapeHtml(item.option_summary)}</div>` : ''}
		  <div class="cashier-cart-actions">
		    <button class="ghost-btn small" data-action="cart-decrease-item" data-cart-index="${index}">-</button>
		    <span>${item.quantity}</span>
		    <button class="ghost-btn small" data-action="cart-increase-item" data-cart-index="${index}">+</button>
		    <button class="ghost-btn small" data-action="cart-remove-item" data-cart-index="${index}">移除</button>
		  </div>
		</li>
	`).join('');
}

function updateCashierCartDisplay() {
	const listNode = root.querySelector('[data-role="cashier-cart-list"]');
	const totalNode = root.querySelector('[data-role="cashier-cart-total"]');
	if (!listNode || !totalNode) {
		render();
		return;
	}
	listNode.innerHTML = buildCashierCartItemsHTML();
	totalNode.textContent = `合计：${formatAmount(getCartTotalAmount())}`;
}

async function generatePayQRCode(rawCodeUrl) {
	const source = String(rawCodeUrl || '').trim();
	if (!source) {
		return '';
	}
	if (!window.posBridge || typeof window.posBridge.generateQRCode !== 'function') {
		return '';
	}
	try {
		return await window.posBridge.generateQRCode(source, 280);
	} catch {
		return '';
	}
}
