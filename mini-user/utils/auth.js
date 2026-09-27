const http = require('./request')

let pendingLoginPromise = null

function getToken() {
	return wx.getStorageSync('token') || ''
}

function setToken(token) {
	wx.setStorageSync('token', token)
}

function clearToken() {
	wx.removeStorageSync('token')
}

function login() {
	if (pendingLoginPromise) {
		return pendingLoginPromise
	}

	pendingLoginPromise = new Promise((resolve, reject) => {
		wx.login({
			success(loginRes) {
				if (!loginRes.code) {
					reject(new Error('获取微信登录凭证失败'))
					return
				}
				http.post('/login', { code: loginRes.code }, { noAuth: true })
					.then(data => {
						setToken(data.token)
						resolve(data.token)
					})
					.catch(reject)
			},
			fail(err) {
				reject(err)
			}
		})
	})

	return pendingLoginPromise.finally(() => {
		pendingLoginPromise = null
	})
}

function ensureLogin(options) {
	const force = !!(options && options.force)
	const token = getToken()
	if (token && !force) {
		return Promise.resolve(token)
	}
	if (force) {
		clearToken()
	}
	return login()
}

module.exports = {
	getToken,
	setToken,
	clearToken,
	login,
	ensureLogin
}
