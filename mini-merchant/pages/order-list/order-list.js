const http = require('../../utils/request')

const STATUS_MAP = {
  '': '全部',
  '1': '待接单',
  '2': '已接单',
  '3': '已完成',
  '4': '已取消'
}

Page({
  data: {
    tabs: [
      { key: '', label: '全部' },
      { key: '1', label: '待接单' },
      { key: '2', label: '已接单' },
      { key: '3', label: '已完成' }
    ],
    activeTab: '',
    orders: [],
    page: 1,
    size: 20,
    total: 0,
    loading: false,
    noMore: false
  },

  onShow() {
    this.setData({ page: 1, orders: [], noMore: false })
    this.loadOrders()
  },

  switchTab(e) {
    const key = e.currentTarget.dataset.key
    this.setData({ activeTab: key, page: 1, orders: [], noMore: false })
    this.loadOrders()
  },

  loadOrders() {
    if (this.data.loading || this.data.noMore) return
    this.setData({ loading: true })

    const params = { page: this.data.page, size: this.data.size }
    if (this.data.activeTab) params.status = this.data.activeTab

    http.get('/orders', params).then(data => {
      const list = data.list || []
      const orders = this.data.page === 1 ? list : this.data.orders.concat(list)
      this.setData({
        orders,
        total: data.total,
        noMore: orders.length >= data.total
      })
    }).catch(() => {}).finally(() => {
      this.setData({ loading: false })
    })
  },

  viewOrder(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/order-detail/order-detail?id=' + id })
  },

  onReachBottom() {
    if (!this.data.noMore) {
      this.setData({ page: this.data.page + 1 })
      this.loadOrders()
    }
  },

  onPullDownRefresh() {
    this.setData({ page: 1, orders: [], noMore: false })
    this.loadOrders()
    wx.stopPullDownRefresh()
  },

  getStatusText(status) {
    const map = { 0: '待支付', 1: '待接单', 2: '已接单', 3: '已完成', 4: '已取消' }
    return map[status] || '未知'
  }
})
