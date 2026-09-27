const auth = require('./auth')
const http = require('./request')

const LOGIN_PAGE = '/pages/login/login'
const REGISTER_PAGE = '/pages/register/register'
const AUDIT_STATUS_PAGE = '/pages/audit-status/audit-status'
const HOME_PAGE = '/pages/home/home'

function goLogin() {
  wx.reLaunch({ url: LOGIN_PAGE })
}

function goRegister() {
  wx.reLaunch({ url: REGISTER_PAGE })
}

function goAuditStatus(status) {
  wx.reLaunch({ url: AUDIT_STATUS_PAGE + '?status=' + status })
}

function goHome() {
  wx.switchTab({ url: HOME_PAGE })
}

function routeMerchantEntry(options = {}) {
  if (!auth.isLoggedIn()) {
    goLogin()
    return Promise.resolve('login')
  }

  const hasShop = typeof options.hasShop === 'boolean'
    ? options.hasShop
    : !!wx.getStorageSync('has_shop')

  if (!hasShop) {
    goRegister()
    return Promise.resolve('register')
  }

  return http.get('/shop').then(shop => {
    const status = shop && typeof shop.status === 'number' ? shop.status : 0
    if (status === 1) {
      goHome()
      return 'home'
    }

    goAuditStatus(status)
    return 'audit'
  }).catch(error => {
    if (error && error.code === 401) {
      auth.logout()
      goLogin()
      return 'login'
    }

    return Promise.reject(error)
  })
}

module.exports = {
  LOGIN_PAGE,
  REGISTER_PAGE,
  AUDIT_STATUS_PAGE,
  HOME_PAGE,
  goLogin,
  goRegister,
  goAuditStatus,
  goHome,
  routeMerchantEntry
}