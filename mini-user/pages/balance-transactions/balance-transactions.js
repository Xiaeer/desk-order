const { getBalanceTransactions } = require('../../utils/user-center')

function formatAmount(value) {
	const amount = Number(value) || 0
	const prefix = amount > 0 ? '+' : ''
	return prefix + (amount / 100).toFixed(2)
}

function formatTime(value) {
	if (!value) {
		return '-'
	}
	return String(value).replace('T', ' ').replace('Z', '').slice(0, 19)
}

Page({
	data: {
		loading: false,
		list: []
	},

	onShow() {
		this.loadList()
	},

	onPullDownRefresh() {
		this.loadList(true)
	},

	loadList(isRefresh) {
		this.setData({ loading: true })
		getBalanceTransactions(1, 50)
			.then(data => {
				this.setData({
					list: (data.list || []).map(item => Object.assign({}, item, {
						change_text: formatAmount(item.change_amount),
						balance_text: ((Number(item.balance_after) || 0) / 100).toFixed(2),
						created_at_text: formatTime(item.created_at)
					}))
				})
			})
			.finally(() => {
				this.setData({ loading: false })
				if (isRefresh) {
					wx.stopPullDownRefresh()
				}
			})
	}
})