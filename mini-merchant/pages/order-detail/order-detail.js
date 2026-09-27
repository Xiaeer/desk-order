const http = require('../../utils/request')

Page({
  data: {
    order: null,
    statusText: '',
    loading: false
  },

  onLoad(options) {
    if (options.id) {
      this.loadOrder(options.id)
    }
  },

  loadOrder(id) {
    http.get('/order/' + id).then(data => {
      const statusMap = { 0: '待支付', 1: '待接单', 2: '已接单', 3: '已完成', 4: '已取消' }
      this.setData({
        order: data,
        statusText: statusMap[data.status] || '未知'
      })
    })
  },

  // 接单
  acceptOrder() {
    this.setData({ loading: true })
    http.put('/order/' + this.data.order.id + '/accept').then(() => {
      wx.showToast({ title: '已接单', icon: 'success' })
      this.loadOrder(this.data.order.id)
    }).finally(() => {
      this.setData({ loading: false })
    })
  },

  // 完成订单
  completeOrder() {
    this.setData({ loading: true })
    http.put('/order/' + this.data.order.id + '/complete').then(() => {
      wx.showToast({ title: '已完成', icon: 'success' })
      this.loadOrder(this.data.order.id)
    }).finally(() => {
      this.setData({ loading: false })
    })
  }
})
