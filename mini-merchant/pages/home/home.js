const http = require('../../utils/request')
const auth = require('../../utils/auth')

const AUDIT_STATUS_PAGE = '/pages/audit-status/audit-status'

Page({
  data: {
    shopName: '',
    isOpen: false,
    orders: [],
    todayCount: 0,
    todayAmount: 0
  },

  onShow() {
    if (!auth.isLoggedIn()) {
      wx.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.loadShopInfo().then(isApproved => {
      if (isApproved) {
        return this.loadPendingOrders()
      }
    }).catch(() => {})
  },

  loadShopInfo() {
    return http.get('/shop').then(data => {
      const status = data && typeof data.status === 'number' ? data.status : 0
      if (status !== 1) {
        wx.reLaunch({ url: AUDIT_STATUS_PAGE + '?status=' + status })
        return false
      }

      this.setData({
        shopName: data.name,
        isOpen: data.is_open
      })
      return true
    }).catch(() => false)
  },

  loadPendingOrders() {
    // 加载已支付的订单（待接单）
    return http.get('/orders', { page: 1, size: 50, status: 1 }).then(data => {
      const orders = data.list || []
      this.setData({
        orders,
        todayCount: orders.length,
        todayAmount: orders.reduce((sum, o) => sum + o.total_amount, 0)
      })
    }).catch(() => {})
  },

  // 切换营业状态
  toggleOpen() {
    const isOpen = !this.data.isOpen
    http.put('/shop/status', { is_open: isOpen }).then(() => {
      this.setData({ isOpen })
      wx.showToast({ title: isOpen ? '已开始营业' : '已休息', icon: 'success' })
    })
  },

  // 接单
  acceptOrder(e) {
    const id = e.currentTarget.dataset.id
    http.put('/order/' + id + '/accept').then(() => {
      wx.showToast({ title: '已接单', icon: 'success' })
      this.loadPendingOrders()
    })
  },

  // 查看订单详情
  viewOrder(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/order-detail/order-detail?id=' + id })
  },

  onPullDownRefresh() {
    this.loadShopInfo().then(isApproved => {
      if (isApproved) {
        return this.loadPendingOrders()
      }
    }).finally(() => {
      wx.stopPullDownRefresh()
    })
  }
})
