const { getMerchantApiBase } = require('./config')
const REGISTER_PAGE = '/pages/register/register'

function getCurrentRoute() {
  const pages = getCurrentPages()
  if (!pages.length) {
    return ''
  }
  return '/' + pages[pages.length - 1].route
}

function handleMissingShop() {
  const currentRoute = getCurrentRoute()
  wx.setStorageSync('has_shop', false)

  if (currentRoute === REGISTER_PAGE) {
    return
  }

  wx.reLaunch({ url: REGISTER_PAGE })
}

function isMissingShopResponse(response) {
  return response && response.msg === '店铺不存在'
}

function request(url, method, data) {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('token') || ''
    wx.request({
      url: getMerchantApiBase() + url,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        'Authorization': token ? 'Bearer ' + token : ''
      },
      success(res) {
        if (res.data.code === 0) {
          resolve(res.data.data)
        } else {
          if (isMissingShopResponse(res.data)) {
            handleMissingShop()
            reject(res.data)
            return
          }
          wx.showToast({ title: res.data.msg || '请求失败', icon: 'none' })
          reject(res.data)
        }
      },
      fail(err) {
        wx.showToast({ title: '网络错误', icon: 'none' })
        reject(err)
      }
    })
  })
}

function upload(url, filePath, name = 'file', formData = {}) {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('token') || ''
    wx.uploadFile({
      url: getMerchantApiBase() + url,
      filePath,
      name,
      formData,
      header: {
        'Authorization': token ? 'Bearer ' + token : ''
      },
      success(res) {
        let result = {}
        try {
          result = typeof res.data === 'string' ? JSON.parse(res.data) : (res.data || {})
        } catch (err) {
          wx.showToast({ title: '上传响应解析失败', icon: 'none' })
          reject(err)
          return
        }

        if (result.code === 0) {
          resolve(result.data)
        } else {
          if (isMissingShopResponse(result)) {
            handleMissingShop()
            reject(result)
            return
          }
          wx.showToast({ title: result.msg || '上传失败', icon: 'none' })
          reject(result)
        }
      },
      fail(err) {
        wx.showToast({ title: '网络错误', icon: 'none' })
        reject(err)
      }
    })
  })
}

module.exports = {
  get: (url, data) => request(url, 'GET', data),
  post: (url, data) => request(url, 'POST', data),
  put: (url, data) => request(url, 'PUT', data),
  del: (url, data) => request(url, 'DELETE', data),
  upload
}
