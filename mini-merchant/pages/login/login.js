const auth = require('../../utils/auth')
const { goRegister, routeMerchantEntry } = require('../../utils/navigation')

Page({
  data: {
    loading: false
  },

  onLoad() {
    if (auth.isLoggedIn()) {
      routeMerchantEntry().catch(() => {})
    }
  },

  handleLogin() {
    if (this.data.loading) return
    this.setData({ loading: true })
    auth.login().then(data => {
      if (data.is_new || !data.has_shop) {
        goRegister()
        return
      }

      return routeMerchantEntry({ hasShop: true })
    }).catch(() => {
      wx.showToast({ title: '登录失败', icon: 'none' })
    }).finally(() => {
      this.setData({ loading: false })
    })
  }
})
