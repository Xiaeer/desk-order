let cachedCashierMenuRef: unknown = null;
let cachedCashierMenuLoading = false;
let cachedCashierMenuMessage = '';
let cachedCashierExpandedKey = '';
let cachedCashierProductsHtml = '';
let cachedCashierCategoryCardsRef: unknown = null;
let cachedCashierCategoryCardsKeyword = '';
let cachedCashierCategoryCards = new Map<string, string>();

export function renderApp({
	state,
	statusLabelMap,
	orderStatusMap,
	helpers,
}) {
	const {
		escapeHtml,
		formatAmount,
		formatTime,
		formatCountdown,
		formatOptionPriceDelta,
		getQRImageUrl,
		getPayCountdownSeconds,
		getCartTotalAmount,
		buildPrintSummary,
	} = helpers;

	function renderOptionPicker() {
		if (!state.optionPicker || !state.optionPicker.product) {
			return '';
		}
		const product = state.optionPicker.product;
		const groups = Array.isArray(product.options) ? product.options : [];
		const groupsHtml = groups.map((group) => {
			const values = Array.isArray(group.values) ? group.values : [];
			const selected = Array.isArray(state.optionPicker.selections[String(group.id)])
				? state.optionPicker.selections[String(group.id)]
				: [];
			const tips = [];
			if (group.required || Number(group.min_select || 0) > 0) {
				tips.push('必选');
			}
			if (String(group.select_type || '').toLowerCase() === 'multi') {
				if (Number(group.min_select || 0) > 0) {
					tips.push(`至少 ${Number(group.min_select)} 项`);
				}
				if (Number(group.max_select || 0) > 0) {
					tips.push(`最多 ${Number(group.max_select)} 项`);
				}
			} else {
				tips.push('单选');
			}
			const valuesHtml = values.map((value) => {
				const checked = selected.includes(String(value.id));
				return `<button class="option-value-btn ${checked ? 'active' : ''}" data-action="option-picker-select" data-group-id="${escapeHtml(group.id)}" data-value-id="${escapeHtml(value.id)}">${escapeHtml(value.name)} ${escapeHtml(formatOptionPriceDelta(value.price_delta))}</button>`;
			}).join('');

			return `
				<div class="option-group">
				  <div class="option-group-title">${escapeHtml(group.name)} <span class="option-group-tip">${escapeHtml(tips.join(' · '))}</span></div>
				  <div class="option-value-list">${valuesHtml || '<span class="option-empty">暂无可选项</span>'}</div>
				</div>
			`;
		}).join('');

		return `
		  <div class="option-picker-mask">
		    <div class="option-picker-dialog">
		      <h3>${escapeHtml(product.name)} 规格选择</h3>
		      <div class="option-picker-subtitle">基础价格：${formatAmount(product.price || 0)}</div>
		      <div class="option-picker-body">${groupsHtml}</div>
		      ${state.optionPicker.error ? `<div class="option-picker-error">${escapeHtml(state.optionPicker.error)}</div>` : ''}
		      <div class="option-picker-actions">
		        <button class="ghost-btn small" data-action="option-picker-cancel">取消</button>
		        <button class="primary-btn small" data-action="option-picker-confirm">确认加入购物车</button>
		      </div>
		    </div>
		  </div>
		`;
	}

	function renderCashier() {
		const categories = state.menu && Array.isArray(state.menu.categories) ? state.menu.categories : [];
		const expandedCategoryIDs = Array.isArray(state.cashierExpandedCategoryIDs) ? state.cashierExpandedCategoryIDs : [];
		const menuKeyword = String(state.cashierMenuKeywordDebounced || '').trim().toLowerCase();
		const hasMenuKeyword = Boolean(menuKeyword);
		const expandedKey = `${expandedCategoryIDs.join(',')}|${menuKeyword}`;
		if (state.menu !== cachedCashierCategoryCardsRef || menuKeyword !== cachedCashierCategoryCardsKeyword) {
			cachedCashierCategoryCardsRef = state.menu;
			cachedCashierCategoryCardsKeyword = menuKeyword;
			cachedCashierCategoryCards = new Map<string, string>();
		}

		let productsHtml = cachedCashierProductsHtml;
		if (
			state.menu !== cachedCashierMenuRef
			|| state.menuLoading !== cachedCashierMenuLoading
			|| state.menuMessage !== cachedCashierMenuMessage
			|| expandedKey !== cachedCashierExpandedKey
		) {
			productsHtml = categories.length
				? categories.map((category, index) => {
					const categoryID = String(category && category.id != null ? category.id : `idx-${index}`);
					const products = Array.isArray(category.products)
						? category.products.filter((product) => {
							if (!menuKeyword) {
								return true;
							}
							const name = String(product && product.name ? product.name : '').toLowerCase();
							const desc = String(product && product.description ? product.description : '').toLowerCase();
							return name.includes(menuKeyword) || desc.includes(menuKeyword);
						})
						: [];
					if (hasMenuKeyword && !products.length) {
						return '';
					}
					const expanded = hasMenuKeyword ? products.length > 0 : expandedCategoryIDs.includes(categoryID);
					let productCards = '';
					if (expanded) {
						const cardsKey = `${categoryID}|${menuKeyword}`;
						if (cachedCashierCategoryCards.has(cardsKey)) {
							productCards = cachedCashierCategoryCards.get(cardsKey) || '';
						} else {
							productCards = products.length
								? products.map((product) => `
									<div class="cashier-product">
									  <div class="cashier-product-head">
									    <div class="cashier-product-name">${escapeHtml(product.name || '')}</div>
									    <div class="cashier-product-price">${formatAmount(product.price || 0)}</div>
									  </div>
									  <div class="cashier-product-desc">${escapeHtml(product.description || '暂无描述')}</div>
									  <button class="ghost-btn small" data-action="cart-add-item" data-product-id="${Number(product.id)}">加入购物车</button>
									</div>
								`).join('')
								: '<div class="empty-state">该分类暂无商品</div>';
							cachedCashierCategoryCards.set(cardsKey, productCards);
						}
					}
					return `
						<section class="cashier-category">
						  <h3>${hasMenuKeyword
								? `${escapeHtml(category.name || '未命名分类')}（${products.length}）`
								: `<button class="ghost-btn small" data-action="cashier-toggle-category" data-category-id="${escapeHtml(categoryID)}">${expanded ? '收起' : '展开'} ${escapeHtml(category.name || '未命名分类')}</button>`}
						  </h3>
						  ${expanded ? `<div class="cashier-product-grid">${productCards}</div>` : ''}
						</section>
					`;
				}).join('')
				: `<div class="empty-state">${escapeHtml(state.menuLoading ? '菜单加载中...' : state.menuMessage || '暂无菜单数据')}</div>`;
			if (!productsHtml && menuKeyword) {
				productsHtml = `<div class="empty-state">未找到匹配商品：${escapeHtml(menuKeyword)}</div>`;
			}

			cachedCashierMenuRef = state.menu;
			cachedCashierMenuLoading = state.menuLoading;
			cachedCashierMenuMessage = state.menuMessage;
			cachedCashierExpandedKey = expandedKey;
			cachedCashierProductsHtml = productsHtml;
		}

		const cartItemsHtml = state.cart.length
			? state.cart.map((item, index) => `
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
			`).join('')
			: '<li>购物车为空</li>';

		const qrImageUrl = getQRImageUrl(state.payQRCodeDataUrl);
		const payingStatus = state.payingOrder ? (orderStatusMap[state.payingOrder.status] || '未知状态') : '--';
		const countdown = getPayCountdownSeconds();

		return `
		  <section class="panel cashier-panel">
		    <div class="cashier-layout">
		      <div class="cashier-left">
		        <div class="toolbar">
		          <div class="toolbar-meta">门店：${escapeHtml(state.menu && state.menu.shop_name ? state.menu.shop_name : '--')}</div>
		          <input class="ghost-btn small" name="cashierMenuKeyword" value="${escapeHtml(state.cashierMenuKeyword || '')}" placeholder="搜索商品名称/描述" />
		          <button class="ghost-btn small" data-action="cashier-refresh-menu">刷新菜单</button>
		        </div>
		        <div class="cashier-menu-list">${productsHtml}</div>
		      </div>
		      <div class="cashier-right">
		        <div class="cashier-block">
		          <h3>购物车</h3>
		          <label>
		            <span>桌号（可选）</span>
		            <input name="cashierTableNo" value="${escapeHtml(state.cashierTableNo)}" placeholder="例如 A08" />
		          </label>
		          <label>
		            <span>备注（可选）</span>
		            <input name="cashierRemark" value="${escapeHtml(state.cashierRemark)}" placeholder="例如 少辣、先上饮料" />
		          </label>
		          <ul class="cashier-cart-list" data-role="cashier-cart-list">${cartItemsHtml}</ul>
		          <div class="cashier-total" data-role="cashier-cart-total">合计：${formatAmount(getCartTotalAmount())}</div>
		          <button class="primary-btn" data-action="cashier-create-order" ${state.cashierBusy ? 'disabled' : ''}>${state.cashierBusy ? '处理中...' : '下单并生成收款码'}</button>
		        </div>
		        <div class="cashier-block">
		          <h3>支付二维码</h3>
		          <div class="cashier-pay-meta">订单：${escapeHtml(state.payingOrder ? state.payingOrder.order_no : '--')}</div>
		          <div class="cashier-pay-meta">状态：${escapeHtml(payingStatus)}</div>
		          <div class="cashier-pay-meta">支付剩余：<span data-role="pay-countdown">${escapeHtml(formatCountdown(countdown))}</span></div>
		          ${qrImageUrl ? `<img class="cashier-qr" src="${qrImageUrl}" alt="pay qrcode" />` : '<div class="empty-state">创建订单后显示二维码</div>'}
		          ${state.payCodeUrl ? `<div class="cashier-code-url">${escapeHtml(state.payCodeUrl)}</div>` : ''}
		          <div class="cashier-pay-actions">
		            <button class="ghost-btn small" data-action="cashier-refresh-status">刷新状态</button>
		            <button class="ghost-btn small" data-action="cashier-cancel-order">取消本单</button>
		            <button class="ghost-btn small" data-action="cashier-copy-code-url">复制二维码链接</button>
		            <button class="ghost-btn small" data-action="cashier-reset-order">清空本单</button>
		          </div>
		          <div class="toolbar-meta">${escapeHtml(state.cashierMessage || '')}</div>
		        </div>
		      </div>
		    </div>
		  </section>
		`;
	}

	function renderSetup() {
		const customerPrinterOptions = [`<option value="" ${state.config.customerPrinterName ? '' : 'selected'}>系统默认打印机</option>`]
			.concat(
				state.printers.map((printer) => {
					const name = printer.name || printer.displayName || '';
					const selected = name === state.config.customerPrinterName ? 'selected' : '';
					const label = printer.isDefault ? `${name}（默认）` : name;
					return `<option value="${escapeHtml(name)}" ${selected}>${escapeHtml(label)}</option>`;
				}),
			)
			.join('');
		const kitchenPrinterOptions = [`<option value="" ${state.config.kitchenPrinterName ? '' : 'selected'}>跟随顾客单 / 系统默认</option>`]
			.concat(
				state.printers.map((printer) => {
					const name = printer.name || printer.displayName || '';
					const selected = name === state.config.kitchenPrinterName ? 'selected' : '';
					const label = printer.isDefault ? `${name}（默认）` : name;
					return `<option value="${escapeHtml(name)}" ${selected}>${escapeHtml(label)}</option>`;
				}),
			)
			.join('');

		return `
		  <section class="panel setup-panel">
		    <form data-form="config" class="form-grid">
		      <label>
		        <span>API 地址</span>
		        <input name="apiBase" value="${escapeHtml(state.config.apiBase)}" placeholder="http://127.0.0.1:8080" />
		      </label>
		      <label>
		        <span>shop_id</span>
		        <input name="shopId" value="${escapeHtml(state.config.shopId)}" placeholder="例如 12" />
		      </label>
		      <label>
		        <span>POS 密钥</span>
		        <input name="posToken" value="${escapeHtml(state.config.posToken)}" placeholder="从商户 H5 店铺页复制" />
		      </label>
		      <label>
		        <span>终端名称</span>
		        <input name="terminalName" value="${escapeHtml(state.config.terminalName)}" placeholder="POS-01" />
		      </label>
		      <label>
		        <span>顾客单打印机</span>
		        <select name="customerPrinterName">${customerPrinterOptions}</select>
		      </label>
		      <label>
		        <span>后厨单打印机</span>
		        <select name="kitchenPrinterName">${kitchenPrinterOptions}</select>
		      </label>
		      <label class="toggle-row">
		        <span>自动打印顾客单</span>
		        <input type="checkbox" name="autoPrintCustomer" ${state.config.autoPrintCustomer ? 'checked' : ''} />
		      </label>
		      <label class="toggle-row">
		        <span>自动打印后厨单</span>
		        <input type="checkbox" name="autoPrintKitchen" ${state.config.autoPrintKitchen ? 'checked' : ''} />
		      </label>
		      <div class="form-actions">
		        <button class="primary-btn" type="submit">保存并连接</button>
		      </div>
		    </form>
		  </section>
		`;
	}

	function renderOrderCard(order) {
		const items = Array.isArray(order.items) && order.items.length
			? order.items.map((item) => `
				<li>
				  <div class="item-main">
				    <span class="item-name">${escapeHtml(item.name)}</span>
				    ${item.option_summary ? `<div class="item-option">${escapeHtml(item.option_summary)}</div>` : ''}
				  </div>
				  <span class="item-qty">x${escapeHtml(item.quantity)}</span>
				  <span class="item-amount">${formatAmount(item.subtotal)}</span>
				</li>
			`).join('')
			: '<li><span>暂无明细</span><span>-</span><span>-</span></li>';
		const tableMeta = order.table_no_snapshot ? ` · 桌号 ${escapeHtml(order.table_no_snapshot)}` : '';

		return `
		  <article class="order-card">
		    <div class="order-head">
		      <div>
		        <div class="order-no">${escapeHtml(order.order_no)}</div>
		        <div class="order-meta">${escapeHtml(order.shop_name || '未命名门店')} · ${escapeHtml(formatTime(order.created_at))}${tableMeta}</div>
		      </div>
		      <div class="order-side">
		        <div class="order-total">${formatAmount(order.total_amount)}</div>
		        <div class="badge">${escapeHtml(orderStatusMap[order.status] || '未知状态')}</div>
		      </div>
		    </div>
		    <ul class="item-list">${items}</ul>
		    <div class="order-footer">
		      <div class="remark-block">
		        <div class="remark-text">备注：${escapeHtml(order.remark || '无')}</div>
		      </div>
		      <div class="order-actions">
		        <button class="ghost-btn small" data-action="print-order" data-print-scope="customer" data-order-id="${order.id}">顾客单</button>
		        <button class="ghost-btn small" data-action="print-order" data-print-scope="kitchen" data-order-id="${order.id}">后厨单</button>
		        <button class="primary-btn small" data-action="print-order" data-print-scope="all" data-order-id="${order.id}">全部重打</button>
		      </div>
		    </div>
		  </article>
		`;
	}

	function renderOrders() {
		const statusFilter = String(state.ordersStatusFilter || 'all');
		const windowMinutes = Math.max(0, Number(state.ordersTimeWindowMinutes || 0));
		const now = Date.now();
		const filteredOrders = state.orders.filter((order) => {
			if (statusFilter !== 'all' && String(order.status) !== statusFilter) {
				return false;
			}
			if (windowMinutes <= 0) {
				return true;
			}
			const createdAt = new Date(order.created_at || '').getTime();
			if (Number.isNaN(createdAt)) {
				return false;
			}
			return now - createdAt <= windowMinutes * 60 * 1000;
		});
		const safeLimit = Math.max(10, Number(state.ordersRenderLimit || 20));
		const visibleOrders = filteredOrders.slice(0, safeLimit);
		const hiddenCount = Math.max(0, filteredOrders.length - visibleOrders.length);
		const cards = filteredOrders.length
			? visibleOrders.map((order) => renderOrderCard(order)).join('')
			: `
			<div class="empty-state">
			  <div>还没有收到订单</div>
			  <div>可调整状态筛选或时间窗口，或确认小程序支付回调链路已打通。</div>
			</div>
		  `;
		const loadMore = hiddenCount > 0
			? `<div class="toolbar"><button class="ghost-btn small" data-action="orders-load-more">加载更多（剩余 ${hiddenCount} 条）</button></div>`
			: '';
 		const statusFilterOptions = [
			{ value: 'all', label: '全部状态' },
			{ value: '0', label: '待支付' },
			{ value: '1', label: '已支付' },
			{ value: '2', label: '已接单' },
			{ value: '3', label: '已完成' },
			{ value: '4', label: '已取消' },
		].map((option) => `<option value="${option.value}" ${statusFilter === option.value ? 'selected' : ''}>${option.label}</option>`).join('');
		const timeWindowOptions = [
			{ value: 0, label: '全部时间' },
			{ value: 15, label: '近 15 分钟' },
			{ value: 30, label: '近 30 分钟' },
			{ value: 60, label: '近 1 小时' },
			{ value: 180, label: '近 3 小时' },
		].map((option) => `<option value="${option.value}" ${windowMinutes === option.value ? 'selected' : ''}>${option.label}</option>`).join('');

		return `
		  <section class="panel order-panel">
		    <div class="toolbar">
		      <div class="toolbar-toggle-group">
		        <label class="toggle-row compact">
		          <span>顾客单自动打印</span>
		          <input type="checkbox" name="autoPrintCustomerToggle" ${state.config.autoPrintCustomer ? 'checked' : ''} />
		        </label>
		        <label class="toggle-row compact">
		          <span>后厨单自动打印</span>
		          <input type="checkbox" name="autoPrintKitchenToggle" ${state.config.autoPrintKitchen ? 'checked' : ''} />
		        </label>
		      </div>
		      <div class="toolbar-meta">共 ${filteredOrders.length}/${state.orders.length} 条</div>
		    </div>
		    <div class="toolbar">
		      <label class="toggle-row compact">
		        <span>状态筛选</span>
		        <select name="ordersStatusFilter">${statusFilterOptions}</select>
		      </label>
		      <label class="toggle-row compact">
		        <span>时间窗口</span>
		        <select name="ordersTimeWindowMinutes">${timeWindowOptions}</select>
		      </label>
		    </div>
		    <div class="order-list">${cards}</div>
		    ${loadMore}
		  </section>
		`;
	}

	return `
	  <div class="shell">
	    <aside class="sidebar">
	      <div>
	        <div class="brand">DeskOrder POS</div>
	        <div class="brand-sub">Windows 7 首版收单打印端</div>
	      </div>
	      <div class="status-card">
	        <div class="status-pill status-${state.connectionStatus}">${statusLabelMap[state.connectionStatus]}</div>
	        <div class="status-text">${escapeHtml(state.connectionMessage)}</div>
	        <div class="status-text">shop_id：${escapeHtml(state.config.shopId || '--')}</div>
	        <div class="status-text">顾客单：${escapeHtml(buildPrintSummary('customer_receipt'))}</div>
	        <div class="status-text">后厨单：${escapeHtml(buildPrintSummary('kitchen_ticket'))}</div>
	      </div>
	      <div class="sidebar-actions">
	        <button class="ghost-btn" data-action="switch-screen" data-screen="orders">订单台</button>
	        <button class="ghost-btn" data-action="switch-screen" data-screen="cashier">收银台</button>
	        <button class="ghost-btn" data-action="switch-screen" data-screen="setup">设置</button>
	        <button class="ghost-btn" data-action="refresh-printers">刷新打印机</button>
	        <button class="ghost-btn" data-action="test-print">测试双联打印</button>
	        <button class="ghost-btn" data-action="reconnect">重新连接</button>
	        <button class="ghost-btn" data-action="disconnect">断开连接</button>
	      </div>
	      <div class="meta-card">
	        <div>API：${escapeHtml(state.config.apiBase || '--')}</div>
	        <div>终端：${escapeHtml(state.config.terminalName || '--')}</div>
	        <div>POS 密钥：${state.config.posToken ? '已配置' : '未配置'}</div>
	        <div>最后事件：${escapeHtml(state.lastEventAt || '--')}</div>
	        <div>客户端：${escapeHtml(state.appInfo ? `Electron ${state.appInfo.electron}` : '--')}</div>
	      </div>
	    </aside>
	    <main class="content">
	      <header class="content-header">
	        <div>
	          <h1>${state.screen === 'orders' ? '实时订单' : state.screen === 'cashier' ? 'POS 收银台' : 'POS 设置'}</h1>
	          <p>${state.screen === 'orders' ? '顾客在小程序支付后，这里会收到订单，并按配置自动打印顾客单和后厨单。' : state.screen === 'cashier' ? '选择商品加入购物车，创建订单后展示微信二维码，支付成功后自动进入推单打印链路。' : '先配置 API 地址、shop_id、POS 密钥、顾客单打印机和后厨单打印机，再建立推单连接。'}</p>
	        </div>
	        <div class="header-note">${escapeHtml(state.printMessage || '')}</div>
	      </header>
	      ${state.screen === 'orders' ? renderOrders() : state.screen === 'cashier' ? renderCashier() : renderSetup()}
	    </main>
	    ${state.notice ? `<div class="notice" data-action="clear-notice">${escapeHtml(state.notice)}</div>` : ''}
	    ${renderOptionPicker()}
	  </div>
	`;
}
