Component({
	properties: {
		active: {
			type: String,
			value: 'home'
		}
	},
	data: {
		items: [
			{ key: 'home', label: '首页', desc: '点餐', path: '/pages/index/index' },
			{ key: 'orders', label: '订单', desc: '进度', path: '/pages/order-list/order-list' },
			{ key: 'mine', label: '我的', desc: '钱包', path: '/pages/mine/mine' }
		]
	},
	methods: {
		handleTap(event) {
			const { key, path } = event.currentTarget.dataset
			if (!path || key === this.properties.active) {
				return
			}
			wx.reLaunch({ url: path })
		}
	}
})