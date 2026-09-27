const http = require('../../utils/request')
const { getLocation } = require('../../utils/location')
const { isDevelopVersion } = require('../../utils/config')

function normalizeShop(shop) {
	return Object.assign({}, shop, {
		distance_text: Number(shop.distance || 0).toFixed(2)
	})
}

function formatCoordinate(value) {
	return typeof value === 'number' ? value.toFixed(6) : '--'
}

function formatAccuracy(location) {
	const accuracy = [location.horizontalAccuracy, location.accuracy].find(value => typeof value === 'number' && value >= 0)
	return typeof accuracy === 'number' ? `${accuracy.toFixed(1)} m` : '未知'
}

function showLocationError(error) {
	const message = (error && error.msg) || '定位失败，请重试'
	if (error && error.type === 'permission') {
		wx.showModal({
			title: '需要定位权限',
			content: '请在设置中允许小程序获取位置信息后再重试。',
			confirmText: '去设置',
			success(res) {
				if (res.confirm) {
					wx.openSetting()
				}
			}
		})
		return
	}
	wx.showToast({ title: message, icon: 'none' })
}

function showNearbyError(error) {
	const message = error && error.msg ? error.msg : '加载店铺失败'
	wx.showToast({ title: message, icon: 'none' })
}

function getCachedNearbyState() {
	const app = getApp()
	return {
		location: app.globalData.location || null,
		shops: Array.isArray(app.globalData.nearbyShops) ? app.globalData.nearbyShops : []
	}
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

function applyLocation(page, location) {
	page.setData({
		latitude: location.latitude,
		longitude: location.longitude,
		locationReady: true,
		latitudeText: formatCoordinate(location.latitude),
		longitudeText: formatCoordinate(location.longitude),
		accuracyText: formatAccuracy(location),
		coordinateTypeText: String(location.coordinateType || 'gcj02').toUpperCase()
	})
}

function clearLocation(page) {
	page.setData({
		locationReady: false,
		showLocationDebug: isDevelopVersion(),
		latitudeText: '--',
		longitudeText: '--',
		accuracyText: '--',
		shops: []
	})
}

function applyShops(page, shops) {
	page.setData({ shops: (shops || []).map(normalizeShop) })
}

Page({
	data: {
		loading: false,
		latitude: 0,
		longitude: 0,
		locationReady: false,
		showLocationDebug: isDevelopVersion(),
		latitudeText: '--',
		longitudeText: '--',
		accuracyText: '--',
		coordinateTypeText: 'GCJ-02',
		shops: []
	},

	onLoad() {
		this.initPage()
	},

	onPullDownRefresh() {
		this.initPage(true)
	},

	refreshLocation() {
		this.initPage(false)
	},

	initPage(isRefresh) {
		const cachedState = getCachedNearbyState()
		const cachedLocation = cachedState.location
		const cachedShops = cachedState.shops

		this.setData({ loading: true })

		if (!isRefresh && cachedLocation) {
			applyLocation(this, cachedLocation)
			if (cachedShops.length) {
				applyShops(this, cachedShops)
			}
		}

		getLocation()
			.catch(error => {
				if (cachedLocation) {
					return cachedLocation
				}
				throw error
			})
			.then(location => {
				applyLocation(this, location)
				return http.get('/shops/nearby', {
					latitude: location.latitude,
					longitude: location.longitude
				}, { noAuth: true })
					.then(shops => {
						cacheNearbyState(location, shops || [])
						return shops || []
					})
					.catch(error => {
						if (cachedShops.length) {
							wx.showToast({ title: '刷新失败，已显示上次结果', icon: 'none' })
							return cachedShops
						}
						showNearbyError(error)
						return []
					})
			})
			.then(shops => {
				applyShops(this, shops)
			})
			.catch(error => {
				clearLocation(this)
				showLocationError(error)
			})
			.finally(() => {
				this.setData({ loading: false })
				if (isRefresh) {
					wx.stopPullDownRefresh()
				}
			})
	},

	chooseShop(e) {
		const shop = e.currentTarget.dataset.shop
		clearManualTableContext()
		wx.setStorageSync('selected_shop', {
			id: shop.id,
			name: shop.name,
			address: shop.address
		})
		wx.navigateTo({ url: '/pages/menu/menu?shopId=' + shop.id })
	}
})
