const http = require('../../utils/request')

Page({
  data: {
    status: null
  },

  onLoad(options) {
    if (options && options.status) {
      const s = parseInt(options.status, 10)
      this.setData({ status: s })
      if (s === 1) {
        wx.switchTab({ url: '/pages/home/home' })
      }
      return
    }

    const token = wx.getStorageSync('token')
    if (!token) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }

    // 获取店铺状态
    http.get('/shop').then(shop => {
      if (shop && shop.status === 1) {
        wx.switchTab({ url: '/pages/home/home' })
      } else {
        this.setData({ status: (shop && shop.status) || 0 })
      }
    }).catch(() => {
      wx.redirectTo({ url: '/pages/register/register' })
    })
  },

  reapply() {
    wx.redirectTo({ url: '/pages/register/register' })
  }
})
