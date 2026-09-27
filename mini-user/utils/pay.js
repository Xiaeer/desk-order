const http = require('./request')
const { ensureLogin } = require('./auth')

function normalizePaymentError(error) {
	const errMsg = error && error.errMsg ? error.errMsg : ''
	if (/cancel/i.test(errMsg)) {
		return { type: 'cancel', msg: '已取消支付', detail: errMsg }
	}
	return { type: 'pay', msg: '拉起微信支付失败', detail: errMsg }
}

function simulateMockPayment(payParams) {
	const mockResult = String(payParams && payParams.mock_result ? payParams.mock_result : 'success').toLowerCase()
	if (mockResult === 'cancel') {
		return Promise.reject({ type: 'cancel', msg: '已取消模拟支付', detail: mockResult })
	}
	if (mockResult === 'fail') {
		return Promise.reject({ type: 'pay', msg: '模拟支付失败', detail: mockResult })
	}
	return Promise.resolve(payParams)
}

function handlePaymentParams(payParams) {
	if (payParams && payParams.mode === 'mock') {
		return simulateMockPayment(payParams)
	}
	if (payParams && payParams.mode === 'balance') {
		return Promise.resolve(payParams)
	}

	return new Promise((resolve, reject) => {
		wx.requestPayment(Object.assign({}, payParams, {
			success() {
				resolve(payParams)
			},
			fail(error) {
				reject(normalizePaymentError(error))
			}
		}))
	})
}

function requestOrderPayment(orderId, method) {
	const payload = method ? { method } : undefined
	return ensureLogin()
		.then(() => http.post('/order/' + orderId + '/pay', payload))
		.then(handlePaymentParams)
}

function requestRechargePayment(orderId) {
	return ensureLogin()
		.then(() => http.post('/recharge/order/' + orderId + '/pay'))
		.then(handlePaymentParams)
}

module.exports = {
	requestOrderPayment,
	requestRechargePayment
}