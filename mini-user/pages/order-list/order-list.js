const http = require('../../utils/request')
const { ensureLogin } = require('../../utils/auth')

const FILTERS = [
	{ key: 'all', label: '全部' },
	{ key: 'pending', label: '待支付' },
	{ key: 'progress', label: '进行中' },
	{ key: 'done', label: '已完成' }
]

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
		created_at_text: formatTime(order.created_at),
		total_amount_text: (order.total_amount / 100).toFixed(2),
		status_text: formatStatus(order.status)
	})
}

Page({
	data: {
		loading: false,
		list: [],
		filteredList: [],
		filters: FILTERS,
		activeFilter: 'all'
	},

	onShow() {
		this.loadOrders()
	},

	onPullDownRefresh() {
		this.loadOrders(true)
	},

	loadOrders(isRefresh) {
		this.setData({ loading: true })
		ensureLogin()
			.then(() => http.get('/orders', { page: 1, size: 20 }))
			.then(data => {
				const list = (data.list || []).map(normalizeOrder)
				this.setData({
					list
				})
				this.applyFilter(this.data.activeFilter, list)
			})
			.finally(() => {
				this.setData({ loading: false })
				if (isRefresh) {
					wx.stopPullDownRefresh()
				}
			})
	},

	applyFilter(filterKey, sourceList) {
		const list = Array.isArray(sourceList) ? sourceList : this.data.list
		const filteredList = list.filter(item => {
			if (filterKey === 'pending') {
				return Number(item.status) === 0
			}
			if (filterKey === 'progress') {
				return Number(item.status) === 1 || Number(item.status) === 2
			}
			if (filterKey === 'done') {
				return Number(item.status) === 3 || Number(item.status) === 4
			}
			return true
		})
		this.setData({ filteredList, activeFilter: filterKey })
	},

	switchFilter(e) {
		this.applyFilter(e.currentTarget.dataset.key)
	},

	openDetail(e) {
		const id = e.currentTarget.dataset.id
		wx.navigateTo({ url: '/pages/order-detail/order-detail?id=' + id })
	}
})
