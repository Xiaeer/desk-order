const { ensureLogin } = require('./utils/auth')
const { getRuntimeEnvName, getApiHost } = require('./utils/config')
const http = require('./utils/request')

function decodeSceneValue(rawValue) {
	try {
		return decodeURIComponent(rawValue || '').trim()
	} catch (error) {
		return String(rawValue || '').trim()
	}
}

function looksLikeShopTableSceneToken(rawValue) {
	const value = String(rawValue || '').trim()
	return /^st[a-z0-9]{8,}$/i.test(value)
}

function extractSceneToken(options) {
	if (!options) {
		return ''
	}
	const queryScene = decodeSceneValue(options.query && options.query.scene)
	if (looksLikeShopTableSceneToken(queryScene)) {
		return queryScene
	}
	const directScene = decodeSceneValue(options.scene)
	if (looksLikeShopTableSceneToken(directScene)) {
		return directScene
	}
	return ''
}

function normalizeActiveTableContext(context) {
	if (!context || typeof context !== 'object') {
		return null
	}
	return {
		shopId: Number(context.shopId || 0),
		shopName: String(context.shopName || ''),
		shopTableId: Number(context.shopTableId || 0),
		tableNo: String(context.tableNo || ''),
		sceneToken: String(context.sceneToken || ''),
		lockedByScene: !!context.lockedByScene
	}
}

App({
	globalData: {
		location: null,
		nearbyShops: [],
		activeTableContext: null,
		handlingSceneToken: '',
		brandName: 'DeskOrder'
	},
	onLaunch(options) {
		console.log('[DeskOrder][mini-user] runtime_env=', getRuntimeEnvName(), 'api_host=', getApiHost())
		this.globalData.activeTableContext = normalizeActiveTableContext(wx.getStorageSync('selected_shop_table'))
		this.globalData.brandName = this.normalizeBrandName(wx.getStorageSync('mini_user_brand_name'))
		this.fetchMiniUserPublicConfig().finally(() => {
			this.applyBrandToCurrentPage()
		})
		ensureLogin().catch(() => {})
		this.handleSceneEntry(options)
	},
	onShow(options) {
		this.handleSceneEntry(options)
		this.applyBrandToCurrentPage()
	},
	normalizeBrandName(rawValue) {
		const value = String(rawValue || '').trim()
		return value || 'DeskOrder'
	},
	getBrandName() {
		return this.normalizeBrandName(this.globalData.brandName)
	},
	fetchMiniUserPublicConfig() {
		return http.get('/config/public', {}, { noAuth: true, silent: true })
			.then(resp => {
				const brandName = this.normalizeBrandName(resp && resp.mini_user_brand_name)
				this.globalData.brandName = brandName
				wx.setStorageSync('mini_user_brand_name', brandName)
				return brandName
			})
			.catch(() => this.getBrandName())
	},
	applyBrandToCurrentPage() {
		const pages = getCurrentPages()
		if (!pages || !pages.length) {
			return
		}
		wx.setNavigationBarTitle({ title: this.getBrandName() })
	},
	getActiveTableContext() {
		if (this.globalData.activeTableContext) {
			return this.globalData.activeTableContext
		}
		const cached = normalizeActiveTableContext(wx.getStorageSync('selected_shop_table'))
		this.globalData.activeTableContext = cached
		return cached
	},
	setActiveTableContext(context) {
		const normalized = normalizeActiveTableContext(context)
		this.globalData.activeTableContext = normalized
		if (normalized) {
			wx.setStorageSync('selected_shop_table', normalized)
			return
		}
		wx.removeStorageSync('selected_shop_table')
	},
	clearActiveTableContext() {
		this.globalData.activeTableContext = null
		wx.removeStorageSync('selected_shop_table')
	},
	isCurrentTableMenu(context) {
		const pages = getCurrentPages()
		const currentPage = pages && pages.length ? pages[pages.length - 1] : null
		return !!(
			currentPage &&
			currentPage.route === 'pages/menu/menu' &&
			Number(currentPage.data && currentPage.data.shopId) === Number(context.shopId) &&
			Number(currentPage.data && currentPage.data.shopTableId) === Number(context.shopTableId) &&
			String(currentPage.data && currentPage.data.sceneToken || '') === String(context.sceneToken || '')
		)
	},
	openTableMenu(context) {
		if (this.isCurrentTableMenu(context)) {
			return
		}
		wx.reLaunch({
			url: `/pages/menu/menu?shopId=${context.shopId}&shopTableId=${context.shopTableId}&sceneToken=${encodeURIComponent(context.sceneToken)}`
		})
	},
	handleSceneEntry(options) {
		const sceneToken = extractSceneToken(options)
		console.log('[DeskOrder][scene] launch options=', options, 'scene_token=', sceneToken)
		if (!sceneToken || this.globalData.handlingSceneToken === sceneToken) {
			return
		}
		const existing = this.getActiveTableContext()
		if (existing && existing.sceneToken === sceneToken) {
			this.openTableMenu(existing)
			return
		}
		this.globalData.handlingSceneToken = sceneToken
		http.get('/entry/scene', { scene: sceneToken }, { noAuth: true })
			.then(entry => {
				const context = {
					shopId: entry.shop_id,
					shopName: entry.shop_name,
					shopTableId: entry.shop_table_id,
					tableNo: entry.table_no,
					sceneToken: entry.scene_token,
					lockedByScene: true
				}
				this.setActiveTableContext(context)
				wx.setStorageSync('selected_shop', {
					id: context.shopId,
					name: context.shopName
				})
				this.openTableMenu(context)
			})
			.catch(error => {
				console.error('[DeskOrder][scene] resolve failed', {
					sceneToken,
					error
				})
			})
			.finally(() => {
				this.globalData.handlingSceneToken = ''
			})
	}
})
