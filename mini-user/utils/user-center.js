const http = require('./request')
const { getUserApiBase } = require('./config')
const { ensureLogin, getToken } = require('./auth')

function getUserProfile() {
	return ensureLogin().then(() => http.get('/me'))
}

function updateUserProfile(data) {
	return ensureLogin().then(() => http.put('/me/profile', data || {}))
}

function uploadUserAvatar(filePath) {
	return ensureLogin().then(() => new Promise((resolve, reject) => {
		const token = getToken()
		wx.uploadFile({
			url: getUserApiBase() + '/upload/avatar',
			filePath,
			name: 'file',
			header: {
				'Authorization': token ? 'Bearer ' + token : ''
			},
			success(res) {
				let result = null
				try {
					result = JSON.parse(res.data || '{}')
				} catch (error) {
					wx.showToast({ title: '头像上传失败', icon: 'none' })
					reject(error)
					return
				}
				if (result && result.code === 0 && result.data && result.data.url) {
					resolve(result.data.url)
					return
				}
				wx.showToast({ title: (result && result.msg) || '头像上传失败', icon: 'none' })
				reject(result)
			},
			fail(err) {
				wx.showToast({ title: '头像上传失败', icon: 'none' })
				reject(err)
			}
		})
	}))
}

function bindUserPhone(code) {
	return ensureLogin().then(() => http.post('/me/phone', { code: code || '' }))
}

function getRechargeActivities() {
	return ensureLogin().then(() => http.get('/recharge/activities'))
}

function createRechargeOrder(activityId) {
	return ensureLogin().then(() => http.post('/recharge/order', { activity_id: activityId }))
}

function getRechargeOrders(page, size, status) {
	return ensureLogin().then(() => http.get('/recharge/orders', { page, size, status }))
}

function cancelRechargeOrder(orderId) {
	return ensureLogin().then(() => http.delete('/recharge/order/' + orderId))
}

function applyRechargeRefund(orderId, requestNote) {
	return ensureLogin().then(() => http.post('/recharge/order/' + orderId + '/refund', { request_note: requestNote || '' }))
}

function getBalanceTransactions(page, size) {
	return ensureLogin().then(() => http.get('/balance/transactions', { page, size }))
}

function getRefundRecords(page, size) {
	return ensureLogin().then(() => http.get('/refund-records', { page, size }))
}

function getRefundNotifications(page, size) {
	return ensureLogin().then(() => http.get('/refund-notifications', { page, size }))
}

function readRefundNotifications() {
	return ensureLogin().then(() => http.put('/refund-notifications/read', {}))
}

function grantRefundSubscribePermissions(templateIds) {
	return ensureLogin().then(() => http.post('/refund-subscriptions/grant', { template_ids: templateIds || [] }))
}

function cancelRefundRequest(requestId) {
	return ensureLogin().then(() => http.post('/refund-request/' + requestId + '/cancel', {}))
}

function cancelUserOrder(orderId) {
	return ensureLogin().then(() => http.post('/order/' + orderId + '/cancel'))
}

module.exports = {
	getUserProfile,
	updateUserProfile,
	uploadUserAvatar,
	bindUserPhone,
	getRechargeActivities,
	createRechargeOrder,
	getRechargeOrders,
	cancelRechargeOrder,
	applyRechargeRefund,
	getBalanceTransactions,
	getRefundRecords,
	getRefundNotifications,
	readRefundNotifications,
	grantRefundSubscribePermissions,
	cancelRefundRequest,
	cancelUserOrder
}