function normalizeLocationError(error) {
	const errMsg = error && error.errMsg ? error.errMsg : ''
	if (/auth deny|auth denied|authorize|scope.userLocation/i.test(errMsg)) {
		return { type: 'permission', msg: '请允许定位权限', detail: errMsg }
	}
	if (/system permission denied|system denied/i.test(errMsg)) {
		return { type: 'system', msg: '请开启系统定位', detail: errMsg }
	}
	if (/timeout/i.test(errMsg)) {
		return { type: 'timeout', msg: '定位超时，请重试', detail: errMsg }
	}
	return { type: 'location', msg: '定位失败，请重试', detail: errMsg }
}

function getLocation() {
	return new Promise((resolve, reject) => {
		wx.getLocation({
			type: 'gcj02',
			success(res) {
				resolve({
					latitude: res.latitude,
					longitude: res.longitude,
					accuracy: typeof res.accuracy === 'number' ? res.accuracy : null,
					horizontalAccuracy: typeof res.horizontalAccuracy === 'number' ? res.horizontalAccuracy : null,
					verticalAccuracy: typeof res.verticalAccuracy === 'number' ? res.verticalAccuracy : null,
					coordinateType: 'gcj02'
				})
			},
			fail(err) {
				reject(normalizeLocationError(err))
			}
		})
	})
}

module.exports = {
		getLocation,
		normalizeLocationError
}
