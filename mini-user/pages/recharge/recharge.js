const { getUserProfile, getRechargeActivities, createRechargeOrder, getRechargeOrders, cancelRechargeOrder, applyRechargeRefund } = require('../../utils/user-center')
const { requestRechargePayment } = require('../../utils/pay')
const { requestRefundSubscribe } = require('../../utils/refund-subscribe')

function formatAmount(value) {
	return ((Number(value) || 0) / 100).toFixed(2)
}

function formatTime(value) {
	if (!value) {
		return '-'
	}
	return String(value).replace('T', ' ').replace('Z', '').slice(0, 19)
}

function formatRechargeStatus(status) {
	return ['待支付', '已支付', '已取消'][Number(status)] || '未知状态'
}

function normalizeActivity(item) {
	return Object.assign({}, item, {
		recharge_amount_text: formatAmount(item.recharge_amount),
		gift_amount_text: formatAmount(item.gift_amount),
		arrival_amount_text: formatAmount(item.arrival_amount)
	})
}

function normalizeRechargeOrder(item) {
	if (!item) {
		return null
	}
	return Object.assign({}, item, {
		pay_amount_text: formatAmount(item.pay_amount),
		arrival_amount_text: formatAmount(item.total_arrival_amount),
		recharge_amount_text: formatAmount(item.recharge_amount),
		gift_amount_text: formatAmount(item.gift_amount),
		status_text: formatRechargeStatus(item.status),
		created_at_text: formatTime(item.created_at),
		paid_at_text: formatTime(item.paid_at),
		can_apply_refund: !!item.can_apply_refund,
		refund_entry_visible: !!item.refund_entry_visible,
		refund_action_text: item.refund_action_text || '申请退款',
		refund_status_text: item.refund_status_text || '',
		refund_result_text: item.refund_result_text || '',
		refund_updated_at_text: formatTime(item.refund_updated_at),
		has_refund_status: !!item.refund_status
	})
}

function getVisibleRecentRechargeOrders(list, expanded) {
	const orders = Array.isArray(list) ? list : []
	if (expanded || orders.length <= 3) {
		return orders
	}
	return orders.slice(0, 3)
}

Page({
	data: {
		loading: false,
		paying: false,
		canceling: false,
		applyingRefundId: 0,
		refundDialogVisible: false,
		refundDialogOrderId: 0,
		refundRequestNote: '',
		refundSubscribeTemplateIds: [],
		profile: null,
		balanceText: '0.00',
		activities: [],
		selectedActivityId: 0,
		pendingRechargeOrderId: 0,
		pendingRechargeOrder: null,
		recentRechargeOrders: [],
		recentRechargeOrdersVisible: [],
		recentRechargeOrdersExpanded: false
	},

	onShow() {
		this.loadPageData()
	},

	loadPageData() {
		this.setData({ loading: true })
		Promise.all([
			getUserProfile(),
			getRechargeActivities(),
			getRechargeOrders(1, 10).catch(() => ({ list: [] }))
		])
			.then(([profile, activities, rechargeOrderPage]) => {
				const normalized = (activities || []).map(normalizeActivity)
				const rechargeOrders = (rechargeOrderPage.list || []).map(normalizeRechargeOrder)
				const pendingRechargeOrder = rechargeOrders.find(item => Number(item.status) === 0) || null
				const recentRechargeOrders = rechargeOrders.filter(item => !pendingRechargeOrder || item.id !== pendingRechargeOrder.id).slice(0, 5)
				const recentRechargeOrdersVisible = getVisibleRecentRechargeOrders(recentRechargeOrders, false)
				const selectedActivityId = pendingRechargeOrder
					? pendingRechargeOrder.activity_id
					: (this.data.selectedActivityId || (normalized[0] ? normalized[0].id : 0))
				this.setData({
					profile,
					balanceText: formatAmount(profile.balance_amount),
					refundSubscribeTemplateIds: profile.refund_subscribe_template_ids || [],
					activities: normalized,
					selectedActivityId,
					pendingRechargeOrder,
					pendingRechargeOrderId: pendingRechargeOrder ? pendingRechargeOrder.id : 0,
					recentRechargeOrders,
					recentRechargeOrdersVisible,
					recentRechargeOrdersExpanded: false
				})
			})
			.finally(() => {
				this.setData({ loading: false })
			})
	},

	toggleRecentRechargeOrders() {
		if (!this.data.recentRechargeOrders.length || this.data.recentRechargeOrders.length <= 3) {
			return
		}
		const recentRechargeOrdersExpanded = !this.data.recentRechargeOrdersExpanded
		this.setData({
			recentRechargeOrdersExpanded,
			recentRechargeOrdersVisible: getVisibleRecentRechargeOrders(this.data.recentRechargeOrders, recentRechargeOrdersExpanded)
		})
	},

	selectActivity(e) {
		if (this.data.pendingRechargeOrderId) {
			wx.showToast({ title: '请先继续支付或取消当前待支付充值单', icon: 'none' })
			return
		}
		this.setData({
			selectedActivityId: Number(e.currentTarget.dataset.id),
			pendingRechargeOrderId: 0
		})
	},

	submitRecharge() {
		if (this.data.paying) {
			return
		}
		if (!this.data.selectedActivityId && !this.data.pendingRechargeOrderId) {
			wx.showToast({ title: '请先选择充值档位', icon: 'none' })
			return
		}

		this.setData({ paying: true })
		const createOrReuse = this.data.pendingRechargeOrderId
			? Promise.resolve(this.data.pendingRechargeOrder || { id: this.data.pendingRechargeOrderId })
			: createRechargeOrder(this.data.selectedActivityId)

		createOrReuse
			.then(order => {
				const pendingRechargeOrder = normalizeRechargeOrder(order)
				this.setData({
					pendingRechargeOrder,
					pendingRechargeOrderId: pendingRechargeOrder ? pendingRechargeOrder.id : 0
				})
				return requestRechargePayment(order.id)
					.then(payResult => ({ success: true, order, payResult }))
					.catch(error => ({ success: false, order, error }))
			})
			.then(result => {
				if (!result.success) {
					this.setData({
						pendingRechargeOrderId: result.order.id,
						pendingRechargeOrder: normalizeRechargeOrder(result.order)
					})
					if (result.error && result.error.type) {
						wx.showToast({ title: result.error.msg, icon: 'none' })
					}
					return
				}

				const isMock = result.payResult && result.payResult.mode === 'mock'
				wx.showToast({ title: isMock ? '模拟充值成功' : '充值成功', icon: 'success' })
				this.setData({ pendingRechargeOrderId: 0, pendingRechargeOrder: null })
				this.loadPageData()
				setTimeout(() => {
					wx.navigateBack({ delta: 1 })
				}, 400)
			})
			.finally(() => {
				this.setData({ paying: false })
			})
	},

	cancelPendingRecharge() {
		if (!this.data.pendingRechargeOrderId || this.data.canceling) {
			return
		}
		wx.showModal({
			title: '取消充值单',
			content: '仅待支付充值单可取消，取消后可重新选择活动档位。',
			success: res => {
				if (!res.confirm) {
					return
				}
				this.setData({ canceling: true })
				cancelRechargeOrder(this.data.pendingRechargeOrderId)
					.then(() => {
						wx.showToast({ title: '充值单已取消', icon: 'success' })
						this.setData({ pendingRechargeOrderId: 0, pendingRechargeOrder: null })
						this.loadPageData()
					})
					.finally(() => {
						this.setData({ canceling: false })
					})
			}
		})
	},

	noop() {},

	handleRefundNoteInput(e) {
		this.setData({ refundRequestNote: e.detail.value || '' })
	},

	closeRefundDialog() {
		if (this.data.applyingRefundId) {
			return
		}
		this.setData({
			refundDialogVisible: false,
			refundDialogOrderId: 0,
			refundRequestNote: ''
		})
	},

	handleApplyRefund(e) {
		const orderId = Number(e.currentTarget.dataset.id || 0)
		if (!orderId || this.data.applyingRefundId) {
			return
		}
		this.setData({
			refundDialogVisible: true,
			refundDialogOrderId: orderId,
			refundRequestNote: ''
		})
	},

	submitApplyRefund() {
		const orderId = this.data.refundDialogOrderId
		if (!orderId || this.data.applyingRefundId) {
			return
		}
		this.setData({ applyingRefundId: orderId })
		applyRechargeRefund(orderId, this.data.refundRequestNote.trim())
			.then(() => {
				wx.showToast({ title: '退款申请已提交', icon: 'success' })
				return requestRefundSubscribe(this.data.refundSubscribeTemplateIds)
					.then(result => {
						if (result && result.acceptedTemplateIds && result.acceptedTemplateIds.length) {
							wx.showToast({ title: '已开启微信通知', icon: 'success' })
						}
					})
					.catch(() => {})
			})
			.then(() => {
				this.setData({
					refundDialogVisible: false,
					refundDialogOrderId: 0,
					refundRequestNote: ''
				})
				this.loadPageData()
			})
			.finally(() => {
				this.setData({ applyingRefundId: 0 })
			})
	}
})