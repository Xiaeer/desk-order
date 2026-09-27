const { ensureLogin } = require('../../utils/auth')
const http = require('../../utils/request')
const { getLocation } = require('../../utils/location')
const { getRechargeActivities } = require('../../utils/user-center')

function formatAmount(value) {
	return ((Number(value) || 0) / 100).toFixed(2)
}

function formatCompactAmount(value) {
	return formatAmount(value).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
}

function cacheNearbyState(location, shops) {
	const app = getApp()
	if (location) {
		app.globalData.location = location
	}
	if (Array.isArray(shops)) {
		app.globalData.nearbyShops = shops
	}
}

function clearManualTableContext() {
	const app = getApp()
	if (app && typeof app.clearActiveTableContext === 'function') {
		app.clearActiveTableContext()
	}
}

function navigateToMenu(shop) {
	if (!shop || !shop.id) {
		wx.navigateTo({ url: '/pages/shop-select/shop-select' })
		return
	}

	clearManualTableContext()
	wx.setStorageSync('selected_shop', {
		id: shop.id,
		name: shop.name,
		address: shop.address
	})
	wx.navigateTo({ url: '/pages/menu/menu?shopId=' + shop.id })
}

function normalizeShopPreview(shop) {
	if (!shop) {
		return null
	}
	return Object.assign({}, shop, {
		distance_text: typeof shop.distance === 'number' ? shop.distance.toFixed(2) : '--',
		address_text: shop.address || '请先选择附近门店',
		state_text: shop.is_open ? '营业中' : '暂停营业'
	})
}

function getCurrentShopPreview() {
	const app = getApp()
	if (!app.globalData.location) {
		return null
	}
	const selectedShop = wx.getStorageSync('selected_shop') || null
	const cachedShops = Array.isArray(app.globalData.nearbyShops) ? app.globalData.nearbyShops : []
	if (!cachedShops.length) {
		return null
	}
	if (selectedShop && selectedShop.id) {
		const matched = cachedShops.find(item => Number(item.id) === Number(selectedShop.id))
		if (matched) {
			return normalizeShopPreview(matched)
		}
	}
	return normalizeShopPreview(cachedShops[0] || null)
}

Page({
	data: {
		brandName: 'DeskOrder',
		locationText: '开始查找附近门店',
		hasLocation: false,
		startingOrder: false,
		currentShop: null,
		featuredActivityTitle: '储值有礼',
		featuredActivityDesc: '进入我的页面，查看当前可用充值优惠'
	},

	onLoad(options) {
		const app = getApp()
		if (app && typeof app.handleSceneEntry === 'function') {
			app.handleSceneEntry(options)
		}
		ensureLogin()
			.then(() => {
				this.loadActivityPreview()
			})
			.catch(() => {})
	},

	onShow() {
		const app = getApp()
		this.setData({ brandName: app.getBrandName ? app.getBrandName() : 'DeskOrder' })
		this.syncHomePreview()
	},

	syncHomePreview() {
		const app = getApp()
		const currentShop = getCurrentShopPreview()
		if (app.globalData.location) {
			this.setData({
				currentShop,
				hasLocation: true,
				locationText: currentShop ? '已获取当前位置，可开始点餐' : '已获取当前位置，附近暂无可用门店'
			})
			return
		}
		this.setData({
			currentShop: null,
			hasLocation: false,
			locationText: '开始查找附近门店'
		})
	},

	loadActivityPreview() {
		getRechargeActivities()
			.then(list => {
				const activity = Array.isArray(list) && list.length ? list[0] : null
				if (!activity) {
					return
				}
				this.setData({
					featuredActivityTitle: activity.name,
					featuredActivityDesc: `充${formatCompactAmount(activity.recharge_amount)}送${formatCompactAmount(activity.gift_amount)}，到账¥${formatAmount(activity.arrival_amount)}`
				})
			})
			.catch(() => {})
	},

	goCurrentShop() {
		if (this.data.hasLocation && this.data.currentShop && this.data.currentShop.id) {
			navigateToMenu(this.data.currentShop)
			return
		}
		this.goStartOrder()
	},

	goStartOrder() {
		if (this.data.startingOrder) {
			return
		}

		this.setData({ startingOrder: true })
		getLocation()
			.then(location => {
				cacheNearbyState(location)
				this.setData({
					hasLocation: true,
					locationText: '已获取当前位置，正在匹配最近门店'
				})
				return http.get('/shops/nearby', {
					latitude: location.latitude,
					longitude: location.longitude
				}, { noAuth: true })
					.then(shops => {
						cacheNearbyState(location, shops || [])
						return shops || []
					})
			})
			.then(shops => {
				if (shops.length) {
					const currentShop = normalizeShopPreview(shops[0])
					this.setData({ currentShop })
					navigateToMenu(currentShop)
					return
				}

				wx.navigateTo({ url: '/pages/shop-select/shop-select' })
			})
			.catch(() => {
				wx.navigateTo({ url: '/pages/shop-select/shop-select' })
			})
			.finally(() => {
				this.setData({ startingOrder: false })
			})
	},

	goShopSelect() {
		wx.navigateTo({ url: '/pages/shop-select/shop-select' })
	},

	goOrders() {
		wx.reLaunch({ url: '/pages/order-list/order-list' })
	},

	goRecharge() {
		wx.navigateTo({ url: '/pages/recharge/recharge' })
	}
})
