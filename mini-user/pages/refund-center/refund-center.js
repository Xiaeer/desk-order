const { getRefundRecords, getRefundNotifications, readRefundNotifications, getUserProfile, cancelRefundRequest } = require('../../utils/user-center')
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

function normalizeRecord(item) {
	return Object.assign({}, item, {
		request_note: item.request_note || '',
		can_cancel: !!item.can_cancel,
		cancel_action_text: item.cancel_action_text || '撤回申请',
		requested_total_amount_text: formatAmount(item.requested_total_amount),
		approved_total_amount_text: formatAmount(item.approved_total_amount),
		created_at_text: formatTime(item.created_at),
		updated_at_text: formatTime(item.updated_at)
	})
}

function normalizeNotification(item) {
	return Object.assign({}, item, {
		created_at_text: formatTime(item.created_at)
	})
}

Page({
	data: {
		loading: false,
		subscribing: false,
		cancelingRefundId: 0,
		activeTab: 'records',
		records: [],
		notifications: [],
		subscribeEnabled: false,
		subscribeTemplateIds: []
	},

	onShow() {
		this.loadPageData()
	},

	onPullDownRefresh() {
		this.loadPageData(true)
	},

	switchTab(e) {
		const tab = e.currentTarget.dataset.tab || 'records'
		if (tab === this.data.activeTab) {
			return
		}
		this.setData({ activeTab: tab })
	},

	loadPageData(isRefresh) {
		this.setData({ loading: true })
		Promise.all([
			getUserProfile().catch(() => null),
			getRefundRecords(1, 20).catch(() => ({ list: [] })),
			getRefundNotifications(1, 20).catch(() => ({ list: [] }))
		])
			.then(([profile, recordPage, notificationPage]) => {
				const records = (recordPage.list || []).map(normalizeRecord)
				const notifications = (notificationPage.list || []).map(normalizeNotification)
				this.setData({
					records,
					notifications,
					subscribeEnabled: !!(profile && profile.refund_subscribe_enabled),
					subscribeTemplateIds: profile && profile.refund_subscribe_template_ids ? profile.refund_subscribe_template_ids : []
				})
				if (notifications.some(item => !item.is_read)) {
					readRefundNotifications()
						.then(() => {
							this.setData({
								notifications: notifications.map(item => Object.assign({}, item, { is_read: true }))
							})
						})
						.catch(() => {})
				}
			})
			.finally(() => {
				this.setData({ loading: false })
				if (isRefresh) {
					wx.stopPullDownRefresh()
				}
			})
	},

	handleSubscribeRefund() {
		if (this.data.subscribing) {
			return
		}
		if (!this.data.subscribeTemplateIds.length) {
			wx.showToast({ title: '当前未配置微信订阅消息', icon: 'none' })
			return
		}
		this.setData({ subscribing: true })
		requestRefundSubscribe(this.data.subscribeTemplateIds)
			.then(result => {
				if (result && result.acceptedTemplateIds && result.acceptedTemplateIds.length) {
					wx.showToast({ title: '已开启微信通知', icon: 'success' })
					return
				}
				wx.showToast({ title: '你还未授权订阅消息', icon: 'none' })
			})
			.catch(() => {
				wx.showToast({ title: '开启失败，请稍后重试', icon: 'none' })
			})
			.finally(() => {
				this.setData({ subscribing: false })
			})
	},

	handleCancelRefund(e) {
		const requestId = Number(e.currentTarget.dataset.id || 0)
		if (!requestId || this.data.cancelingRefundId) {
			return
		}
		wx.showModal({
			title: '撤回退款申请',
			content: '仅待审核的充值退款申请可撤回，撤回后可以重新提交。',
			success: res => {
				if (!res.confirm) {
					return
				}
				this.setData({ cancelingRefundId: requestId })
				cancelRefundRequest(requestId)
					.then(() => {
						wx.showToast({ title: '退款申请已撤回', icon: 'success' })
						this.loadPageData()
					})
					.finally(() => {
						this.setData({ cancelingRefundId: 0 })
					})
			}
		})
	}
})