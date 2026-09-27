const { grantRefundSubscribePermissions } = require('./user-center')

function normalizeTemplateIds(templateIds) {
	return (templateIds || []).filter(item => !!item)
}

function requestRefundSubscribe(templateIds) {
	const tmplIds = normalizeTemplateIds(templateIds)
	if (!tmplIds.length) {
		return Promise.resolve({ enabled: false, acceptedTemplateIds: [] })
	}
	if (typeof wx.requestSubscribeMessage !== 'function') {
		return Promise.resolve({ unsupported: true, acceptedTemplateIds: [] })
	}
	return new Promise((resolve, reject) => {
		wx.requestSubscribeMessage({
			tmplIds,
			success(res) {
				const acceptedTemplateIds = tmplIds.filter(id => res[id] === 'accept')
				if (!acceptedTemplateIds.length) {
					resolve({ acceptedTemplateIds, result: res })
					return
				}
				grantRefundSubscribePermissions(acceptedTemplateIds)
					.then(() => resolve({ acceptedTemplateIds, result: res }))
					.catch(reject)
			},
			fail: reject
		})
	})
}

module.exports = {
	requestRefundSubscribe
}