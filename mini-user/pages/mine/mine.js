const { getUserProfile, updateUserProfile, uploadUserAvatar, bindUserPhone, getRechargeActivities } = require('../../utils/user-center')

function formatAmount(value) {
	return ((Number(value) || 0) / 100).toFixed(2)
}

function formatCompactAmount(value) {
	return formatAmount(value).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
}

function getGreeting() {
	const hour = new Date().getHours()
	if (hour < 6) {
		return '夜深了'
	}
	if (hour < 12) {
		return '早上好'
	}
	if (hour < 18) {
		return '下午好'
	}
	return '晚上好'
}

function formatPhone(phone) {
	const normalized = String(phone || '').trim()
	if (!normalized) {
		return '未绑定手机号'
	}
	if (normalized.length !== 11) {
		return normalized
	}
	return `${normalized.slice(0, 3)}****${normalized.slice(-4)}`
}

function buildDisplayName(profile) {
	const currentProfile = profile || {}
	const nickname = String(currentProfile.nickname || '').trim()
	if (nickname) {
		return nickname
	}
	if (currentProfile.id) {
		return `用户#${currentProfile.id}`
	}
	return '用户'
}

function buildProfileState(profile) {
	const currentProfile = profile || {}
	return {
		profile: currentProfile,
		displayName: buildDisplayName(currentProfile),
		showRefundCenter: !!(currentProfile && (currentProfile.refund_feature_enabled || currentProfile.has_refund_records)),
		profileActionDesc: currentProfile.nickname || currentProfile.avatar
			? '修改用户名和头像'
			: '设置用户名和头像，方便商家识别',
		phoneDisplay: formatPhone(currentProfile.phone),
		profileView: {
			balanceText: formatAmount(currentProfile.balance_amount),
			rechargeText: formatAmount(currentProfile.total_recharge_amount),
			giftText: formatAmount(currentProfile.total_gift_amount),
			consumeText: formatAmount(currentProfile.total_consume_amount)
		}
	}
}

Page({
	data: {
		loading: false,
		greeting: getGreeting(),
		profile: null,
		displayName: '用户',
		showRefundCenter: false,
		profileActionDesc: '设置用户名和头像，方便商家识别',
		phoneDisplay: '未绑定手机号',
		profileView: null,
		showProfileEditor: false,
		profileDraftNickname: '',
		profileDraftAvatar: '',
		featuredActivity: null,
		featuredActivityText: '充值活动进行中'
	},

	onShow() {
		this.loadPageData()
	},

	loadPageData() {
		this.setData({ loading: true, greeting: getGreeting() })
		Promise.all([
			getUserProfile(),
			getRechargeActivities().catch(() => [])
		])
			.then(([profile, activities]) => {
				const featuredActivity = Array.isArray(activities) && activities.length ? activities[0] : null
				this.setData({
					...buildProfileState(profile),
					featuredActivity,
					featuredActivityText: featuredActivity
						? `充${formatCompactAmount(featuredActivity.recharge_amount)}送${formatCompactAmount(featuredActivity.gift_amount)}`
						: '当前暂无额外赠送活动'
				})
			})
			.finally(() => {
				this.setData({ loading: false })
			})
	},

	bindPhoneNumber(event) {
		const detail = event && event.detail ? event.detail : {}
		const code = typeof detail.code === 'string' ? detail.code.trim() : ''
		if (!code) {
			if (detail.errMsg && detail.errMsg.indexOf('cancel') !== -1) {
				return
			}
			wx.showToast({ title: '手机号授权失败', icon: 'none' })
			return
		}

		this.setData({ loading: true })
		bindUserPhone(code)
			.then(profile => {
				this.setData(buildProfileState(profile))
				wx.showToast({ title: '手机号已绑定', icon: 'success' })
			})
			.finally(() => {
				this.setData({ loading: false })
			})
	},

	openProfileEditor() {
		const profile = this.data.profile || {}
		this.setData({
			showProfileEditor: true,
			profileDraftNickname: profile.nickname || '',
			profileDraftAvatar: profile.avatar || ''
		})
	},

	closeProfileEditor() {
		this.setData({ showProfileEditor: false })
	},

	noop() {},

	handleProfileNicknameInput(event) {
		this.setData({
			profileDraftNickname: event && event.detail ? event.detail.value : ''
		})
	},

	handleChooseAvatar(event) {
		const detail = event && event.detail ? event.detail : {}
		const avatarUrl = typeof detail.avatarUrl === 'string' ? detail.avatarUrl.trim() : ''
		if (!avatarUrl) {
			wx.showToast({ title: '请选择头像', icon: 'none' })
			return
		}
		this.setData({ profileDraftAvatar: avatarUrl })
	},

	saveProfile() {
		const nickname = String(this.data.profileDraftNickname || '').trim()
		const draftAvatar = String(this.data.profileDraftAvatar || '').trim()
		if (!nickname && !draftAvatar) {
			wx.showToast({ title: '请设置用户名或头像', icon: 'none' })
			return
		}

		const shouldUploadAvatar = draftAvatar && !/^https?:\/\//i.test(draftAvatar)
		const avatarTask = shouldUploadAvatar ? uploadUserAvatar(draftAvatar) : Promise.resolve(draftAvatar)

		this.setData({ loading: true })
		avatarTask
			.then(avatar => updateUserProfile({ nickname, avatar }))
			.then(profile => {
				this.setData({
					...buildProfileState(profile),
					showProfileEditor: false,
					profileDraftNickname: profile.nickname || '',
					profileDraftAvatar: profile.avatar || ''
				})
				wx.showToast({ title: '资料已保存', icon: 'success' })
			})
			.finally(() => {
				this.setData({ loading: false })
			})
	},

	goRecharge() {
		wx.navigateTo({ url: '/pages/recharge/recharge' })
	},

	goTransactions() {
		wx.navigateTo({ url: '/pages/balance-transactions/balance-transactions' })
	},

	goRefundCenter() {
		wx.navigateTo({ url: '/pages/refund-center/refund-center' })
	},

	showAbout() {
		const app = getApp()
		const brandName = app && typeof app.getBrandName === 'function' ? app.getBrandName() : 'DeskOrder'
		wx.showModal({
			title: `关于 ${brandName}`,
			content: '这是用户端点单小程序，可查看余额、参与充值活动并在店内点单。',
			showCancel: false
		})
	}
})