const { getUserApiBase } = require('./config')

function isUnauthorized(res, result) {
	return !!((res && res.statusCode === 401) || (result && result.code === 401))
}

function retryWithFreshToken(url, method, data, options) {
	const auth = require('./auth')
	auth.clearToken()
	return auth.ensureLogin({ force: true }).then(() => {
		return request(url, method, data, { ...(options || {}), _retriedAfterAuthFailure: true })
	})
}

function buildQuery(data) {
	if (!data) {
		return ''
	}
	const query = Object.keys(data)
		.filter(key => data[key] !== undefined && data[key] !== null && data[key] !== '')
		.map(key => encodeURIComponent(key) + '=' + encodeURIComponent(data[key]))
		.join('&')
	return query ? '?' + query : ''
}

function request(url, method, data, options) {
	const requestOptions = options || {}
	const token = wx.getStorageSync('token') || ''
	const isGet = method === 'GET'
	const requestUrl = getUserApiBase() + url + (isGet ? buildQuery(data) : '')

	return new Promise((resolve, reject) => {
		wx.request({
			url: requestUrl,
			method,
			data: isGet ? undefined : data,
			header: {
				'Content-Type': 'application/json',
				'Authorization': requestOptions.noAuth ? '' : (token ? 'Bearer ' + token : '')
			},
			success(res) {
				const result = res.data || {}
				if (result.code === 0) {
					resolve(result.data)
					return
				}
				if (!requestOptions.noAuth && !requestOptions._retriedAfterAuthFailure && isUnauthorized(res, result)) {
					retryWithFreshToken(url, method, data, requestOptions)
						.then(resolve)
						.catch(err => {
							wx.showToast({ title: '登录已失效，请重试', icon: 'none' })
							reject(err)
						})
					return
				}
				if (!requestOptions.silent) {
					wx.showToast({ title: result.msg || '请求失败', icon: 'none' })
				}
				reject(result)
			},
			fail(err) {
				console.error('[DeskOrder][request] failed', {
					url: requestUrl,
					method,
					errMsg: err && err.errMsg,
					error: err
				})
				if (!requestOptions.silent) {
					wx.showToast({ title: '网络错误', icon: 'none' })
				}
				reject(err)
			}
		})
	})
}

module.exports = {
	get(url, data, options) {
		return request(url, 'GET', data, options)
	},
	post(url, data, options) {
		return request(url, 'POST', data, options)
	},
	put(url, data, options) {
		return request(url, 'PUT', data, options)
	},
	delete(url, data, options) {
		return request(url, 'DELETE', data, options)
	}
}
