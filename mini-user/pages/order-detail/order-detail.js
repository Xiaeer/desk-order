const http = require('../../utils/request')
const { ensureLogin } = require('../../utils/auth')
const { requestOrderPayment } = require('../../utils/pay')
const { getUserProfile, cancelUserOrder } = require('../../utils/user-center')

function formatTime(value) {
	if (!value) {
		return '-'
	}
	return String(value).replace('T', ' ').replace('Z', '').slice(0, 19)
}

function formatStatus(status) {
	return ['待支付', '已支付', '已接单', '已完成', '已取消'][status] || '未知状态'
}

function normalizeOrder(order) {
	return Object.assign({}, order, {
		can_pay: Number(order.status) === 0,
		can_cancel: typeof order.can_user_cancel === 'boolean' ? order.can_user_cancel : Number(order.status) === 0,
		can_refund_balance: !!order.can_user_refund,
		refund_entry_visible: !!order.refund_entry_visible,
		refund_action_text: order.refund_action_text || '退款回余额',
		pay_channel_text: order.pay_channel === 'balance' ? '余额支付' : '微信支付',
		created_at_text: formatTime(order.created_at),
		paid_at_text: formatTime(order.paid_at),
		total_amount_text: (order.total_amount / 100).toFixed(2),
		status_text: formatStatus(order.status),
		items: (order.items || []).map(item => Object.assign({}, item, {
			price_text: (item.price / 100).toFixed(2),
			subtotal_text: (item.subtotal / 100).toFixed(2),
			option_summary: item.option_summary || ''
		}))
	})
}

Page({
	data: {
		id: 0,
		loading: false,
		order: null,
		paying: false,
		actioning: false,
		balanceText: '0.00',
		balanceEnough: false
	},

	onLoad(options) {
		this.setData({ id: Number(options.id || 0) })
	},

	onShow() {
		this.loadDetail()
		this.loadProfile()
	},

	loadDetail() {
		if (!this.data.id) {
			return
		}
		this.setData({ loading: true })
		ensureLogin()
			.then(() => http.get('/order/' + this.data.id))
			.then(order => {
				const normalizedOrder = normalizeOrder(order)
				const balanceAmount = Math.round(Number(this.data.balanceText || 0) * 100)
				this.setData({
					order: normalizedOrder,
					balanceEnough: balanceAmount >= Number(normalizedOrder.total_amount || 0)
				})
			})
			.finally(() => {
				this.setData({ loading: false })
			})
	},

	loadProfile() {
		getUserProfile()
			.then(profile => {
				const balanceAmount = Number(profile.balance_amount) || 0
				this.setData({
					balanceText: (balanceAmount / 100).toFixed(2),
					balanceEnough: this.data.order ? balanceAmount >= Number(this.data.order.total_amount || 0) : false
				})
			})
			.catch(() => {})
	},

	payOrder(e) {
		if (!this.data.order || !this.data.order.can_pay || this.data.paying) {
			return
		}
		const method = e.currentTarget.dataset.method || 'wechat'
		if (method === 'balance' && !this.data.balanceEnough) {
			wx.showToast({ title: '余额不足，请先充值', icon: 'none' })
			return
		}

		this.setData({ paying: true })
		requestOrderPayment(this.data.id, method)
			.then(payResult => {
				const mode = payResult && payResult.mode
				const title = mode === 'mock' ? '模拟支付成功' : (mode === 'balance' ? '余额支付成功' : '支付成功')
				wx.showToast({ title, icon: 'success' })
				this.loadDetail()
				this.loadProfile()
			})
			.catch(error => {
				if (error && error.type === 'cancel') {
					wx.showToast({ title: error.msg, icon: 'none' })
					return
				}
				if (error && error.type === 'pay') {
					wx.showToast({ title: error.msg, icon: 'none' })
				}
			})
			.finally(() => {
				this.setData({ paying: false })
			})
	},

	goRecharge() {
		wx.navigateTo({ url: '/pages/recharge/recharge' })
	},

	handleOrderAction() {
		const order = this.data.order
		if (!order || this.data.actioning || (!order.can_cancel && !order.can_refund_balance)) {
			return
		}
		const isRefund = order.can_refund_balance
		wx.showModal({
			title: isRefund ? (order.refund_action_text || '退款回余额') : '取消订单',
			content: isRefund ? '该订单将按原余额支付金额退回钱包。' : '待支付订单取消后将无法继续支付。',
			success: res => {
				if (!res.confirm) {
					return
				}
				this.setData({ actioning: true })
				cancelUserOrder(this.data.id)
					.then(() => {
						wx.showToast({ title: isRefund ? '已退回余额' : '订单已取消', icon: 'success' })
						this.loadDetail()
						this.loadProfile()
					})
					.finally(() => {
						this.setData({ actioning: false })
					})
			}
		})
	}
})
