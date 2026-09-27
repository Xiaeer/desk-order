const http = require('../../utils/request')
const { ensureLogin } = require('../../utils/auth')
const cartUtils = require('../../utils/cart')
const { requestOrderPayment } = require('../../utils/pay')
const { getUserProfile } = require('../../utils/user-center')

function normalizeCartItem(item) {
	return Object.assign({}, item, {
		price_text: (item.price / 100).toFixed(2),
		subtotal_text: (item.subtotal / 100).toFixed(2),
		selected_options: item.selected_options || [],
		option_summary: item.option_summary || '',
		cart_item_key: item.cart_item_key || ''
	})
}

Page({
	data: {
		shopId: 0,
		shopName: '',
		shopTableId: 0,
		selectedTableNo: '',
		sceneToken: '',
		lockedByScene: false,
		tables: [],
		tableLoading: false,
		items: [],
		totalAmount: 0,
		remark: '',
		submitting: false,
		balanceAmount: 0,
		balanceText: '0.00',
		balanceEnough: false,
		selectedPayMethod: 'wechat'
	},

	onShow() {
		const cart = cartUtils.getCartState()
		const totalAmount = cartUtils.getTotalAmount(cart)
		const shopId = Number(cart.shopId || 0)
		this.setData({
			shopId,
			shopName: cart.shopName,
			items: (cart.items || []).map(normalizeCartItem),
			totalAmount,
			totalAmountText: (totalAmount / 100).toFixed(2)
		})
		this.syncTableContext(shopId)
		this.loadShopTables(shopId)
		this.loadWallet(totalAmount)
	},

	syncTableContext(shopId) {
		const app = getApp()
		const context = app.getActiveTableContext ? app.getActiveTableContext() : null
		if (context && Number(context.shopId) === Number(shopId)) {
			this.setData({
				shopTableId: Number(context.shopTableId || 0),
				selectedTableNo: context.tableNo || '',
				sceneToken: context.sceneToken || '',
				lockedByScene: !!context.lockedByScene
			})
			return
		}
		this.setData({
			shopTableId: 0,
			selectedTableNo: '',
			sceneToken: '',
			lockedByScene: false
		})
	},

	loadShopTables(shopId) {
		if (!shopId) {
			this.setData({ tables: [] })
			return
		}
		this.setData({ tableLoading: true })
		http.get('/shop/' + shopId + '/tables', null, { noAuth: true })
			.then(list => {
				const tables = list || []
				const matched = tables.find(item => Number(item.id) === Number(this.data.shopTableId))
				if (!matched && this.data.shopTableId) {
					this.clearTableSelection(false)
				}
				this.setData({ tables })
			})
			.catch(() => {
				this.setData({ tables: [] })
			})
			.finally(() => {
				this.setData({ tableLoading: false })
			})
	},

	loadWallet(totalAmount) {
		getUserProfile()
			.then(profile => {
				const balanceAmount = Number(profile.balance_amount) || 0
				const balanceEnough = balanceAmount >= totalAmount
				this.setData({
					balanceAmount,
					balanceText: (balanceAmount / 100).toFixed(2),
					balanceEnough,
					selectedPayMethod: balanceEnough ? 'balance' : 'wechat'
				})
			})
			.catch(() => {})
	},

	onRemarkInput(e) {
		this.setData({ remark: e.detail.value })
	},

	selectTable(e) {
		if (this.data.lockedByScene) {
			wx.showToast({ title: '扫码桌号已绑定，如需切换请扫描对应桌码', icon: 'none' })
			return
		}
		const shopTableId = Number(e.currentTarget.dataset.id || 0)
		const tableNo = e.currentTarget.dataset.no || ''
		const sceneToken = e.currentTarget.dataset.scene || ''
		const app = getApp()
		const context = {
			shopId: this.data.shopId,
			shopName: this.data.shopName,
			shopTableId,
			tableNo,
			sceneToken,
			lockedByScene: false
		}
		if (app.setActiveTableContext) {
			app.setActiveTableContext(context)
		}
		this.setData({
			shopTableId,
			selectedTableNo: tableNo,
			sceneToken
		})
	},

	clearTableSelection(showFeedback = true) {
		if (this.data.lockedByScene) {
			wx.showToast({ title: '扫码桌号已绑定，如需取消请重新从门店进入', icon: 'none' })
			return
		}
		const app = getApp()
		if (app.clearActiveTableContext) {
			app.clearActiveTableContext()
		}
		this.setData({
			shopTableId: 0,
			selectedTableNo: '',
			sceneToken: '',
			lockedByScene: false
		})
		if (showFeedback) {
			wx.showToast({ title: '已取消绑定桌号', icon: 'none' })
		}
	},

	selectPayMethod(e) {
		const method = e.currentTarget.dataset.method
		if (method === 'balance' && !this.data.balanceEnough) {
			wx.showToast({ title: '余额不足，请先充值', icon: 'none' })
			return
		}
		this.setData({ selectedPayMethod: method })
	},

	goRecharge() {
		wx.navigateTo({ url: '/pages/recharge/recharge' })
	},

	submitOrder() {
		const cart = cartUtils.getCartState()
		if (!cart.shopId || !(cart.items || []).length) {
			wx.showToast({ title: '购物车为空', icon: 'none' })
			return
		}

		this.setData({ submitting: true })
		let createdOrder = null
		ensureLogin()
			.then(() => {
				const payload = {
					shop_id: cart.shopId,
					remark: this.data.remark,
					items: cart.items.map(item => ({
						product_id: item.product_id,
						quantity: item.quantity,
						selected_options: item.selected_options || []
					}))
				}
				if (this.data.shopTableId) {
					payload.shop_table_id = this.data.shopTableId
				}
				if (this.data.sceneToken) {
					payload.entry_scene = this.data.sceneToken
				}
				return http.post('/order', payload)
			})
			.then(order => {
				createdOrder = order
				cartUtils.clearCart()
				return requestOrderPayment(order.id, this.data.selectedPayMethod)
					.then(payResult => ({ success: true, payResult }))
					.catch(error => ({ success: false, error }))
			})
			.then(result => {
				if (!createdOrder) {
					return
				}

				if (result && result.success) {
					const payMode = result.payResult && result.payResult.mode
					const toastTitle = payMode === 'mock'
						? '模拟支付成功'
						: (payMode === 'balance' ? '余额支付成功' : '支付成功')
					wx.showToast({ title: toastTitle, icon: 'success' })
				} else if (result && result.error && result.error.type === 'cancel') {
					wx.showToast({ title: result.error.msg, icon: 'none' })
				} else if (result && result.error && result.error.type === 'pay') {
					wx.showToast({ title: result.error.msg, icon: 'none' })
				}

				wx.redirectTo({ url: '/pages/order-detail/order-detail?id=' + createdOrder.id })
			})
			.catch(() => {})
			.finally(() => {
				this.setData({ submitting: false })
			})
	}
})
