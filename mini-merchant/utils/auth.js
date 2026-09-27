const http = require('./request')

// 微信登录并获取 token
function login() {
  return new Promise((resolve, reject) => {
    wx.login({
      success(res) {
        if (res.code) {
          http.post('/login', { code: res.code }).then(data => {
            wx.setStorageSync('token', data.token)
            wx.setStorageSync('is_new', data.is_new)
            wx.setStorageSync('has_shop', data.has_shop)
            resolve(data)
          }).catch(reject)
        } else {
          reject(new Error('wx.login failed'))
        }
      },
      fail: reject
    })
  })
}

function getToken() {
  return wx.getStorageSync('token') || ''
}

function isLoggedIn() {
  return !!getToken()
}

function logout() {
  wx.removeStorageSync('token')
  wx.removeStorageSync('is_new')
  wx.removeStorageSync('has_shop')
}

module.exports = { login, getToken, isLoggedIn, logout }
