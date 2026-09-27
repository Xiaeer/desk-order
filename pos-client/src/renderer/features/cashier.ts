export function createCashierFeature({
	state,
	render,
	updatePayCountdownDisplay,
	updateCashierCartDisplay,
	generatePayQRCode,
	hasConfig,
	buildPOSAuthPayload,
	requestPOSPost,
	requestPOSGet,
	normalizeOrder,
	upsertOrder,
	savePaySession,
	clearPaySession,
	onPaidOrder,
	PAY_TIMEOUT_SECONDS,
}) {
	function hasActivePayingOrder() {
		if (!state.payingOrder || !state.payingOrder.id) {
			return false;
		}
		const status = Number(state.payingOrder.status || 0);
		return status !== 1 && status !== 2 && status !== 3 && status !== 4;
	}

	function persistPaySession() {
		if ((!state.payingOrder || !state.payingOrder.id) && !String(state.pendingCreateRequestId || '').trim()) {
			clearPaySession();
			return;
		}
		savePaySession({
			payingOrder: state.payingOrder,
			payExpireAt: Number(state.payExpireAt || 0),
			payCodeUrl: String(state.payCodeUrl || ''),
			payQRCodeDataUrl: String(state.payQRCodeDataUrl || ''),
			cashierMessage: String(state.cashierMessage || ''),
			pendingCreateRequestId: String(state.pendingCreateRequestId || ''),
		});
	}

	function createClientRequestID() {
		return `pos-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
	}

	function resetPendingCreateRequestID() {
		state.pendingCreateRequestId = '';
		persistPaySession();
	}

	function getAllProductsFromMenu() {
		if (!state.menu || !Array.isArray(state.menu.categories)) {
			return [];
		}
		return state.menu.categories.flatMap((category) => Array.isArray(category.products) ? category.products : []);
	}

	function findMenuProduct(productId) {
		return getAllProductsFromMenu().find((product) => Number(product.id) === Number(productId));
	}

	function buildCartItemKey(productId, selectedOptions) {
		return `${Number(productId)}:${JSON.stringify(selectedOptions || [])}`;
	}

	function openOptionPicker(product) {
		const groups = Array.isArray(product && product.options) ? product.options : [];
		const selections = {};
		for (const group of groups) {
			const values = Array.isArray(group.values) ? group.values : [];
			if (!values.length) {
				continue;
			}
			const selectType = String(group.select_type || '').toLowerCase();
			const required = Boolean(group.required) || Number(group.min_select || 0) > 0;
			if (selectType === 'single' && required) {
				selections[String(group.id)] = [String(values[0].id)];
			}
		}
		state.optionPicker = {
			product,
			selections,
			error: '',
		};
		render();
	}

	function toggleOptionPickerValue(groupID, valueID) {
		if (!state.optionPicker || !state.optionPicker.product) {
			return;
		}
		const product = state.optionPicker.product;
		const groups = Array.isArray(product.options) ? product.options : [];
		const group = groups.find((item) => String(item.id) === String(groupID));
		if (!group) {
			return;
		}
		const key = String(group.id);
		const current = Array.isArray(state.optionPicker.selections[key]) ? [...state.optionPicker.selections[key]] : [];
		const valueKey = String(valueID);
		const selectType = String(group.select_type || '').toLowerCase();

		if (selectType === 'single') {
			const required = Boolean(group.required) || Number(group.min_select || 0) > 0;
			if (current.length === 1 && current[0] === valueKey && !required) {
				state.optionPicker.selections[key] = [];
			} else {
				state.optionPicker.selections[key] = [valueKey];
			}
		} else {
			const idx = current.indexOf(valueKey);
			if (idx >= 0) {
				current.splice(idx, 1);
			} else {
				current.push(valueKey);
			}
			state.optionPicker.selections[key] = current;
		}

		state.optionPicker.error = '';
		render();
	}

	function getOptionPickerSelectionResult() {
		if (!state.optionPicker || !state.optionPicker.product) {
			return null;
		}
		const groups = Array.isArray(state.optionPicker.product.options) ? state.optionPicker.product.options : [];
		const selections = [];
		let optionDelta = 0;
		const summaryParts = [];

		for (const group of groups) {
			const values = Array.isArray(group.values) ? group.values : [];
			if (!values.length) {
				continue;
			}
			const selectType = String(group.select_type || '').toLowerCase();
			const minSelect = Number(group.min_select || 0);
			const maxSelect = Number(group.max_select || 0);
			const required = Boolean(group.required) || minSelect > 0;
			const picked = Array.isArray(state.optionPicker.selections[String(group.id)]) ? state.optionPicker.selections[String(group.id)] : [];

			if (!picked.length) {
				if (required) {
					return { error: `${group.name} 为必选项` };
				}
				continue;
			}

			if (selectType === 'single' && picked.length !== 1) {
				return { error: `${group.name} 只能选择 1 项` };
			}
			if (selectType !== 'single') {
				if (minSelect > 0 && picked.length < minSelect) {
					return { error: `${group.name} 至少选择 ${minSelect} 项` };
				}
				if (maxSelect > 0 && picked.length > maxSelect) {
					return { error: `${group.name} 最多选择 ${maxSelect} 项` };
				}
			}

			const pickedValues = picked
				.map((id) => values.find((value) => String(value.id) === String(id)))
				.filter(Boolean);
			if (!pickedValues.length) {
				if (required) {
					return { error: `${group.name} 选择项无效，请重选` };
				}
				continue;
			}

			selections.push({
				group_id: String(group.id),
				value_ids: pickedValues.map((value) => String(value.id)),
			});
			summaryParts.push(`${group.name}:${pickedValues.map((value) => value.name).join('/')}`);
			for (const value of pickedValues) {
				optionDelta += Number(value.price_delta || 0);
			}
		}

		return {
			selectedOptions: selections,
			optionSummary: summaryParts.join('；'),
			optionDelta,
		};
	}

	function confirmOptionPicker() {
		if (!state.optionPicker || !state.optionPicker.product) {
			return;
		}
		const result = getOptionPickerSelectionResult();
		if (!result || result.error) {
			state.optionPicker.error = result && result.error ? result.error : '规格选择失败，请重试';
			render();
			return;
		}
		addCartItemResolved(state.optionPicker.product, result);
		state.optionPicker = null;
		state.cashierMessage = '';
		render();
	}

	function addCartItemResolved(product, picked) {
		const finalPrice = Number(product.price || 0) + Number(picked.optionDelta || 0);
		const cartKey = buildCartItemKey(product.id, picked.selectedOptions);
		const existing = state.cart.find((item) => item.cart_key === cartKey);
		if (existing) {
			existing.quantity += 1;
			return;
		}
		state.cart.push({
			cart_key: cartKey,
			product_id: Number(product.id),
			name: product.name,
			price: finalPrice,
			quantity: 1,
			selected_options: picked.selectedOptions,
			option_summary: picked.optionSummary,
		});
	}

	function addCartItem(productId) {
		const product = findMenuProduct(productId);
		if (!product) {
			return;
		}
		if (Array.isArray(product.options) && product.options.length > 0) {
			openOptionPicker(product);
			return;
		}
		addCartItemResolved(product, {
			selectedOptions: [],
			optionSummary: '',
			optionDelta: 0,
		});
		state.cashierMessage = '';
		resetPendingCreateRequestID();
		updateCashierCartDisplay();
	}

	function removeCartItem(cartIndex) {
		if (!Number.isInteger(cartIndex) || cartIndex < 0 || cartIndex >= state.cart.length) {
			return;
		}
		state.cart.splice(cartIndex, 1);
		resetPendingCreateRequestID();
		updateCashierCartDisplay();
	}

	function changeCartItemQuantity(cartIndex, delta) {
		if (!Number.isInteger(cartIndex) || cartIndex < 0 || cartIndex >= state.cart.length) {
			return;
		}
		const item = state.cart[cartIndex];
		if (!item) {
			return;
		}
		item.quantity += delta;
		if (item.quantity <= 0) {
			removeCartItem(cartIndex);
			return;
		}
		resetPendingCreateRequestID();
		updateCashierCartDisplay();
	}

	function getCartTotalAmount() {
		return state.cart.reduce((sum, item) => sum + Number(item.price || 0) * Number(item.quantity || 0), 0);
	}

	function syncExpandedCategories(menuData) {
		const categories = menuData && Array.isArray(menuData.categories) ? menuData.categories : [];
		if (!categories.length) {
			state.cashierExpandedCategoryIDs = [];
			return;
		}
		const allIDs = categories.map((category, index) => String(category && category.id != null ? category.id : `idx-${index}`));
		const current = Array.isArray(state.cashierExpandedCategoryIDs) ? state.cashierExpandedCategoryIDs : [];
		const next = current.filter((id) => allIDs.includes(String(id)));
		if (!next.length) {
			next.push(allIDs[0]);
		}
		state.cashierExpandedCategoryIDs = next;
	}

	function resetCashierPayingState() {
		if (state.payPollTimer) {
			window.clearInterval(state.payPollTimer);
			state.payPollTimer = null;
		}
		if (state.payCountdownTimer) {
			window.clearInterval(state.payCountdownTimer);
			state.payCountdownTimer = null;
		}
		state.payingOrder = null;
		state.payExpireAt = 0;
		state.payCodeUrl = '';
		state.payQRCodeDataUrl = '';
		state.pendingCreateRequestId = '';
		clearPaySession();
	}

	function getPayCountdownSeconds() {
		if (!state.payExpireAt) {
			return 0;
		}
		return Math.max(0, Math.ceil((state.payExpireAt - Date.now()) / 1000));
	}

	function startPayCountdown() {
		if (!state.payExpireAt) {
			return;
		}
		updatePayCountdownDisplay(getPayCountdownSeconds());
		if (state.payCountdownTimer) {
			window.clearInterval(state.payCountdownTimer);
			state.payCountdownTimer = null;
		}
		state.payCountdownTimer = window.setInterval(() => {
			const seconds = getPayCountdownSeconds();
			if (seconds <= 0) {
				cancelPayingOrder('支付超时，订单已自动取消');
				return;
			}
			updatePayCountdownDisplay(seconds);
		}, 1000);
	}

	async function cancelPayingOrder(successMessage) {
		if (!state.payingOrder || !state.payingOrder.id) {
			return;
		}
		try {
			const orderData = await requestPOSPost(state.config, `/order/${state.payingOrder.id}/cancel`, buildPOSAuthPayload());
			upsertOrder(normalizeOrder(orderData));
			state.cashierMessage = successMessage || `订单 ${orderData.order_no} 已取消`;
		} catch (error) {
			state.cashierMessage = `取消订单失败：${error.message}`;
			persistPaySession();
			render();
			return;
		}
		resetCashierPayingState();
		render();
	}

	function clearPayingSession() {
		if (hasActivePayingOrder()) {
			state.cashierMessage = `订单 ${state.payingOrder.order_no} 仍在待支付，请先取消或继续支付`;
			render();
			return;
		}
		resetCashierPayingState();
		render();
	}

	async function ensureCashierMenu(force = false) {
		if (!hasConfig()) {
			state.menuMessage = '请先完成 POS 配置';
			render();
			return;
		}
		if (!force && state.menu && Array.isArray(state.menu.categories)) {
			return;
		}
		state.menuLoading = true;
		state.menuMessage = '';
		render();
		try {
			const data = await requestPOSPost(state.config, '/menu', buildPOSAuthPayload());
			state.menu = data;
			syncExpandedCategories(data);
			if (!data || !Array.isArray(data.categories) || !data.categories.length) {
				state.menuMessage = '当前门店暂无在售商品';
			}
		} catch (error) {
			state.menuMessage = error && error.message ? String(error.message) : '菜单获取失败';
		} finally {
			state.menuLoading = false;
			render();
		}
	}

	async function createCashierOrderAndPay() {
		if (state.cashierBusy) {
			return;
		}
		if (hasActivePayingOrder()) {
			state.cashierMessage = `订单 ${state.payingOrder.order_no} 尚未完成，请先继续支付、刷新状态或取消本单`;
			render();
			return;
		}
		if (!state.cart.length) {
			state.cashierMessage = '请先加入商品到购物车';
			render();
			return;
		}
		state.cashierBusy = true;
		state.cashierMessage = '正在创建订单...';
		render();
		let createdOrder = null;
		try {
			const clientRequestID = String(state.pendingCreateRequestId || '').trim() || createClientRequestID();
			state.pendingCreateRequestId = clientRequestID;
			persistPaySession();
			const orderData = await requestPOSPost(state.config, '/order', {
				...buildPOSAuthPayload(),
				client_request_id: clientRequestID,
				table_no: String(state.cashierTableNo || '').trim(),
				remark: String(state.cashierRemark || '').trim(),
				terminal_ref: state.config.terminalName || 'POS-01',
				items: state.cart.map((item) => ({
					product_id: item.product_id,
					quantity: item.quantity,
					selected_options: Array.isArray(item.selected_options) ? item.selected_options : [],
				})),
			});
			createdOrder = orderData;
			state.payingOrder = orderData;
			state.pendingCreateRequestId = '';
			upsertOrder(normalizeOrder(orderData));
			state.cashierMessage = '订单已创建，正在发起微信支付...';
			persistPaySession();
			render();

			const payData = await requestPOSPost(state.config, `/order/${orderData.id}/pay`, buildPOSAuthPayload());
			state.payCodeUrl = payData.code_url || '';
			state.payQRCodeDataUrl = state.payCodeUrl ? await generatePayQRCode(state.payCodeUrl) : '';
			state.payingOrder = {
				...orderData,
				status: Number(payData.status != null ? payData.status : orderData.status),
			};
			upsertOrder(normalizeOrder(state.payingOrder));

			if (String(payData.mode || '') === 'mock') {
				const mockResult = String(payData.mock_result || '').trim();
				if (mockResult) {
					state.cashierMessage = `当前为 Mock 支付：${mockResult}`;
				}
			}

			if (state.payingOrder.status === 1 || state.payingOrder.status === 2 || state.payingOrder.status === 3) {
				state.cashierMessage = `订单 ${state.payingOrder.order_no} 已支付成功`;
				state.cart = [];
				onPaidOrder(normalizeOrder(state.payingOrder));
				resetCashierPayingState();
				render();
				return;
			}

			if (!state.payCodeUrl) {
				state.cashierMessage = '未拿到二维码链接，请稍后刷新订单状态';
			} else {
				state.cashierMessage = `请顾客扫码支付：${state.payingOrder.order_no}`;
				state.payExpireAt = Date.now() + PAY_TIMEOUT_SECONDS * 1000;
				startPayCountdown();
			}
			persistPaySession();
			startPayStatusPolling();
		} catch (error) {
			const errorMessage = error && error.message ? String(error.message) : '下单失败，请重试';
			if (createdOrder && createdOrder.order_no) {
				state.cashierMessage = `订单 ${createdOrder.order_no} 已创建，但支付发起失败：${errorMessage}`;
				state.notice = `订单 ${createdOrder.order_no} 已创建，但支付发起失败：${errorMessage}`;
				state.payExpireAt = 0;
				state.payCodeUrl = '';
				state.payQRCodeDataUrl = '';
				persistPaySession();
			} else {
				state.cashierMessage = errorMessage;
				state.notice = `下单失败：${errorMessage}`;
			}
		} finally {
			state.cashierBusy = false;
			render();
		}
	}

	function startPayStatusPolling() {
		if (!state.payingOrder) {
			return;
		}
		if (state.payPollTimer) {
			window.clearInterval(state.payPollTimer);
			state.payPollTimer = null;
		}
		state.payPollTimer = window.setInterval(() => {
			refreshPayingOrderStatus();
		}, 3000);
	}

	async function refreshPayingOrderStatus() {
		if (!state.payingOrder || !state.payingOrder.id) {
			return;
		}
		try {
			const orderData = await requestPOSGet(state.config, `/order/${state.payingOrder.id}`, buildPOSAuthPayload());
			state.payingOrder = orderData;
			upsertOrder(normalizeOrder(orderData));
			const status = Number(orderData.status || 0);
			if (status === 1 || status === 2 || status === 3) {
				state.cashierMessage = `订单 ${orderData.order_no} 已支付成功`;
				state.cart = [];
				onPaidOrder(normalizeOrder(orderData));
				resetCashierPayingState();
			}
			if (status === 4) {
				state.cashierMessage = `订单 ${orderData.order_no} 已取消`;
				resetCashierPayingState();
			}
			persistPaySession();
			render();
		} catch (error) {
			state.cashierMessage = `刷新支付状态失败：${error.message}`;
			persistPaySession();
			render();
		}
	}

	function copyPayCodeUrl() {
		const value = String(state.payCodeUrl || '').trim();
		if (!value) {
			return;
		}
		if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
			navigator.clipboard.writeText(value)
				.then(() => {
					state.cashierMessage = '二维码链接已复制';
					persistPaySession();
					render();
				})
				.catch(() => {
					state.cashierMessage = '复制失败，请手动复制';
					persistPaySession();
					render();
				});
		}
	}

	function resumePayingWatchers() {
		if (!state.payingOrder || !state.payingOrder.id) {
			return;
		}
		const status = Number(state.payingOrder.status || 0);
		if (status === 1 || status === 2 || status === 3 || status === 4) {
			return;
		}
		if (!state.payPollTimer) {
			startPayStatusPolling();
		}
		if (!state.payCountdownTimer && state.payExpireAt > Date.now()) {
			startPayCountdown();
		}
	}

	return {
		addCartItem,
		changeCartItemQuantity,
		removeCartItem,
		createCashierOrderAndPay,
		ensureCashierMenu,
		refreshPayingOrderStatus,
		cancelPayingOrder,
		copyPayCodeUrl,
		toggleOptionPickerValue,
		confirmOptionPicker,
		resetCashierPayingState,
		clearPayingSession,
		resumePayingWatchers,
		getPayCountdownSeconds,
		getCartTotalAmount,
	};
}
