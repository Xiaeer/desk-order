const http = require('../../utils/request')
const cartUtils = require('../../utils/cart')

function normalizeProduct(product) {
	return Object.assign({}, product, {
		price_text: (product.price / 100).toFixed(2),
		has_options: Array.isArray(product.options) && product.options.length > 0
	})
}

function formatAmount(value) {
	return (Number(value || 0) / 100).toFixed(2)
}

function formatCompactAmount(value) {
	return formatAmount(value).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
}

function buildDefaultOptionSelectionMap(product) {
	const selectionMap = {}
	;(product.options || []).forEach(group => {
		if (group.select_type === 'single' && (group.required || Number(group.min_select || 0) > 0) && Array.isArray(group.values) && group.values[0]) {
			selectionMap[group.id] = [group.values[0].id]
			return
		}
		selectionMap[group.id] = []
	})
	return selectionMap
}

function getSelectedValueIDs(selectionMap, groupId) {
	return Array.isArray(selectionMap[groupId]) ? selectionMap[groupId] : []
}

function buildOptionHelperText(group) {
	const groupName = group.name || '该选项组'
	if (group.select_type === 'single') {
		return group.required ? `${groupName}必选，且只能选 1 项` : `${groupName}单选，可不选`
	}
	const minSelect = Number(group.min_select || 0)
	const maxSelect = Number(group.max_select || 0)
	if (minSelect > 0 && maxSelect > 0) {
		return `至少选 ${minSelect} 项，最多选 ${maxSelect} 项`
	}
	if (minSelect > 0) {
		return `至少选 ${minSelect} 项`
	}
	if (maxSelect > 0) {
		return `最多选 ${maxSelect} 项`
	}
	return '可自由多选'
}

function decorateOptionGroups(product, selectionMap) {
	return (product.options || []).map(group => {
		const selectedValueIDs = getSelectedValueIDs(selectionMap, group.id)
		return Object.assign({}, group, {
			helper_text: buildOptionHelperText(group),
			values: (group.values || []).map(value => Object.assign({}, value, {
				selected: selectedValueIDs.includes(value.id),
				price_delta_text: Number(value.price_delta || 0) > 0 ? `+¥${formatCompactAmount(value.price_delta)}` : ''
			}))
		})
	})
}

function buildSelectedOptionsPayload(product, selectionMap) {
	return (product.options || [])
		.map(group => ({
			group_id: group.id,
			value_ids: getSelectedValueIDs(selectionMap, group.id)
		}))
		.filter(selection => selection.value_ids.length)
}

function calculateSelectedOptionDelta(product, selectionMap) {
	let totalDelta = 0
	;(product.options || []).forEach(group => {
		const valueMap = new Map((group.values || []).map(value => [value.id, value]))
		getSelectedValueIDs(selectionMap, group.id).forEach(valueId => {
			const value = valueMap.get(valueId)
			if (value) {
				totalDelta += Number(value.price_delta || 0)
			}
		})
	})
	return totalDelta
}

function buildOptionSummary(product, selectionMap) {
	return (product.options || [])
		.map(group => {
			const valueMap = new Map((group.values || []).map(value => [value.id, value]))
			const selectedNames = getSelectedValueIDs(selectionMap, group.id)
				.map(valueId => valueMap.get(valueId))
				.filter(Boolean)
				.map(value => value.name)
			if (!selectedNames.length) {
				return ''
			}
			return `${group.name}:${selectedNames.join('/')}`
		})
		.filter(Boolean)
		.join('；')
}

function validateOptionSelection(product, selectionMap) {
	for (const group of product.options || []) {
		const selectedCount = getSelectedValueIDs(selectionMap, group.id).length
		if (group.select_type === 'single') {
			if ((group.required || Number(group.min_select || 0) > 0) && selectedCount !== 1) {
				return `${group.name}需要且只能选择 1 项`
			}
			if (selectedCount > 1) {
				return `${group.name}只能选择 1 项`
			}
			continue
		}
		if (selectedCount < Number(group.min_select || 0)) {
			return `${group.name}至少选择 ${group.min_select} 项`
		}
		if (Number(group.max_select || 0) > 0 && selectedCount > Number(group.max_select || 0)) {
			return `${group.name}最多选择 ${group.max_select} 项`
		}
	}
	return ''
}

function buildCountMap(cart) {
	const map = {}
	;(cart.items || []).forEach(item => {
		map[item.product_id] = (map[item.product_id] || 0) + item.quantity
	})
	return map
}

Page({
	data: {
		loading: false,
		shopId: 0,
		shopTableId: 0,
		shopName: '',
		tableNo: '',
		sceneToken: '',
		categories: [],
		activeCategoryIndex: 0,
		currentProducts: [],
		productCountMap: {},
		showOptionPopup: false,
		optionProduct: null,
		optionGroups: [],
		optionSelectionMap: {},
		optionPopupPriceText: '0.00',
		optionSummaryPreview: '',
		cartCount: 0,
		totalAmount: 0
	},

	onLoad(options) {
		this.setData({
			shopId: Number(options.shopId || 0),
			shopTableId: Number(options.shopTableId || 0),
			sceneToken: decodeURIComponent(options.sceneToken || '')
		})
		this.syncTableContext()
		this.loadMenu()
	},

	onShow() {
		this.syncTableContext()
		this.syncCartState()
	},

	syncTableContext() {
		const app = getApp()
		const context = app.getActiveTableContext ? app.getActiveTableContext() : null
		if (context && Number(context.shopId) === Number(this.data.shopId)) {
			this.setData({
				shopTableId: Number(context.shopTableId || 0),
				tableNo: context.tableNo || '',
				sceneToken: context.sceneToken || this.data.sceneToken || ''
			})
			return
		}
		if (context && Number(context.shopId) !== Number(this.data.shopId) && app.clearActiveTableContext) {
			app.clearActiveTableContext()
		}
		this.setData({
			shopTableId: 0,
			tableNo: '',
			sceneToken: ''
		})
	},

	onPullDownRefresh() {
		this.loadMenu(true)
	},

	loadMenu(isRefresh) {
		if (!this.data.shopId) {
			return
		}
		this.setData({ loading: true })
		http.get('/shop/' + this.data.shopId + '/menu', null, { noAuth: true })
			.then(menu => {
				const categories = (menu.categories || []).map(category => Object.assign({}, category, {
					products: (category.products || []).map(normalizeProduct)
				}))
				const currentProducts = categories[0] ? categories[0].products || [] : []
				wx.setStorageSync('selected_shop', {
					id: menu.shop_id,
					name: menu.shop_name
				})
				cartUtils.ensureShopCart(menu.shop_id, menu.shop_name)
				wx.setNavigationBarTitle({ title: menu.shop_name || '店铺菜单' })
				this.setData({
					shopName: menu.shop_name || '',
					categories,
					activeCategoryIndex: 0,
					currentProducts
				})
				this.syncCartState()
			})
			.finally(() => {
				this.setData({ loading: false })
				if (isRefresh) {
					wx.stopPullDownRefresh()
				}
			})
	},

	syncCartState() {
		const cart = cartUtils.getCartState()
		this.setData({
			productCountMap: buildCountMap(cart),
			cartCount: cartUtils.getTotalCount(cart),
			totalAmount: cartUtils.getTotalAmount(cart),
			totalAmountText: (cartUtils.getTotalAmount(cart) / 100).toFixed(2)
		})
	},

	switchCategory(e) {
		const index = Number(e.currentTarget.dataset.index)
		const category = this.data.categories[index] || {}
		this.setData({
			activeCategoryIndex: index,
			currentProducts: category.products || []
		})
	},

	addItem(e) {
		const index = Number(e.currentTarget.dataset.index)
		const product = this.data.currentProducts[index]
		if (!product) {
			return
		}
		if (product.has_options) {
			this.openOptionPopup(product)
			return
		}
		cartUtils.addItem(this.data.shopId, this.data.shopName, product)
		this.syncCartState()
	},

	subItem(e) {
		const productId = Number(e.currentTarget.dataset.id)
		const currentCount = this.data.productCountMap[productId] || 0
		if (currentCount <= 0) {
			return
		}
		cartUtils.updateItemQuantity(productId, currentCount - 1)
		this.syncCartState()
	},

	openCart() {
		if (!this.data.cartCount) {
			return
		}
		wx.navigateTo({ url: '/pages/cart/cart' })
	},

	goConfirm() {
		if (!this.data.cartCount) {
			wx.showToast({ title: '请先选择商品', icon: 'none' })
			return
		}
		wx.navigateTo({ url: '/pages/order-confirm/order-confirm' })
	},

	noop() {},

	refreshOptionPopup(product, selectionMap) {
		const optionDelta = calculateSelectedOptionDelta(product, selectionMap)
		this.setData({
			optionProduct: product,
			optionSelectionMap: selectionMap,
			optionGroups: decorateOptionGroups(product, selectionMap),
			optionPopupPriceText: formatAmount(Number(product.price || 0) + optionDelta),
			optionSummaryPreview: buildOptionSummary(product, selectionMap)
		})
	},

	openOptionPopup(product) {
		const selectionMap = buildDefaultOptionSelectionMap(product)
		this.setData({ showOptionPopup: true })
		this.refreshOptionPopup(product, selectionMap)
	},

	closeOptionPopup() {
		this.setData({
			showOptionPopup: false,
			optionProduct: null,
			optionGroups: [],
			optionSelectionMap: {},
			optionPopupPriceText: '0.00',
			optionSummaryPreview: ''
		})
	},

	toggleOptionValue(e) {
		const optionProduct = this.data.optionProduct
		if (!optionProduct) {
			return
		}
		const groupId = e.currentTarget.dataset.groupId
		const valueId = e.currentTarget.dataset.valueId
		const group = (optionProduct.options || []).find(item => item.id === groupId)
		if (!group) {
			return
		}
		const nextSelectionMap = Object.assign({}, this.data.optionSelectionMap)
		const currentValues = getSelectedValueIDs(nextSelectionMap, groupId).slice()
		if (group.select_type === 'single') {
			const alreadySelected = currentValues.length === 1 && currentValues[0] === valueId
			nextSelectionMap[groupId] = alreadySelected && !group.required ? [] : [valueId]
			this.refreshOptionPopup(optionProduct, nextSelectionMap)
			return
		}
		const selectedIndex = currentValues.indexOf(valueId)
		if (selectedIndex >= 0) {
			currentValues.splice(selectedIndex, 1)
			nextSelectionMap[groupId] = currentValues
			this.refreshOptionPopup(optionProduct, nextSelectionMap)
			return
		}
		if (Number(group.max_select || 0) > 0 && currentValues.length >= Number(group.max_select || 0)) {
			wx.showToast({ title: `${group.name} 最多选择 ${group.max_select} 项`, icon: 'none' })
			return
		}
		currentValues.push(valueId)
		nextSelectionMap[groupId] = currentValues
		this.refreshOptionPopup(optionProduct, nextSelectionMap)
	},

	confirmOptionSelection() {
		const optionProduct = this.data.optionProduct
		if (!optionProduct) {
			return
		}
		const validationMessage = validateOptionSelection(optionProduct, this.data.optionSelectionMap)
		if (validationMessage) {
			wx.showToast({ title: validationMessage, icon: 'none' })
			return
		}
		const selectedOptions = buildSelectedOptionsPayload(optionProduct, this.data.optionSelectionMap)
		const optionSummary = buildOptionSummary(optionProduct, this.data.optionSelectionMap)
		const optionDelta = calculateSelectedOptionDelta(optionProduct, this.data.optionSelectionMap)
		cartUtils.addItem(this.data.shopId, this.data.shopName, optionProduct, {
			selected_options: selectedOptions,
			option_summary: optionSummary,
			price: Number(optionProduct.price || 0) + optionDelta
		})
		this.closeOptionPopup()
		this.syncCartState()
		wx.showToast({ title: '已加入购物车', icon: 'success' })
	}
})
