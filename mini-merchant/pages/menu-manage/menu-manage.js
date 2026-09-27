const http = require('../../utils/request')
const { getUserApiBase } = require('../../utils/config')

Page({
  data: {
    categories: [],
    activeCategory: 0,
    products: [],
    showCatModal: false,
    catName: '',
    editCatId: null
  },

  onShow() {
    this.loadCategories()
  },

  // 加载分类
  loadCategories() {
    http.get('/shop').then(shop => {
      // 通过菜单接口获取分类和商品
      // 用户端接口可以获取完整菜单
      const token = wx.getStorageSync('token')
      return new Promise((resolve, reject) => {
        wx.request({
          url: getUserApiBase() + '/shop/' + shop.id + '/menu',
          method: 'GET',
          header: { 'Authorization': 'Bearer ' + token },
          success(res) {
            if (res.data.code === 0) resolve(res.data.data)
            else reject(res.data)
          },
          fail: reject
        })
      })
    }).then(menu => {
      const categories = (menu.categories || []).map(c => ({
        id: c.id,
        name: c.name,
        sort: c.sort,
        products: c.products || []
      }))
      const activeCategory = this.data.activeCategory
      const products = categories[activeCategory] ? categories[activeCategory].products : []
      this.setData({ categories, products })
    }).catch(() => {})
  },

  // 切换分类
  switchCategory(e) {
    const idx = e.currentTarget.dataset.index
    const products = this.data.categories[idx] ? this.data.categories[idx].products : []
    this.setData({ activeCategory: idx, products })
  },

  // 显示新增分类弹窗
  showAddCategory() {
    this.setData({ showCatModal: true, catName: '', editCatId: null })
  },

  // 显示编辑分类弹窗
  editCategory(e) {
    const cat = this.data.categories[e.currentTarget.dataset.index]
    this.setData({ showCatModal: true, catName: cat.name, editCatId: cat.id })
  },

  onCatNameInput(e) {
    this.setData({ catName: e.detail.value })
  },

  cancelCatModal() {
    this.setData({ showCatModal: false })
  },

  // 保存分类
  saveCat() {
    const name = this.data.catName.trim()
    if (!name) return wx.showToast({ title: '请输入分类名', icon: 'none' })

    const promise = this.data.editCatId
      ? http.put('/category/' + this.data.editCatId, { name })
      : http.post('/category', { name })

    promise.then(() => {
      this.setData({ showCatModal: false })
      wx.showToast({ title: '保存成功', icon: 'success' })
      this.loadCategories()
    })
  },

  // 删除分类
  deleteCategory(e) {
    const id = this.data.categories[e.currentTarget.dataset.index].id
    wx.showModal({
      title: '确认删除',
      content: '删除分类后，该分类下的商品也会受影响',
      success: (res) => {
        if (res.confirm) {
          http.del('/category/' + id).then(() => {
            wx.showToast({ title: '已删除', icon: 'success' })
            this.setData({ activeCategory: 0 })
            this.loadCategories()
          })
        }
      }
    })
  },

  // 新增商品
  addProduct() {
    const cat = this.data.categories[this.data.activeCategory]
    if (!cat) return wx.showToast({ title: '请先创建分类', icon: 'none' })
    wx.navigateTo({ url: '/pages/product-edit/product-edit?category_id=' + cat.id })
  },

  // 编辑商品
  editProduct(e) {
    const id = Number(e.currentTarget.dataset.id)
    const product = this.data.products.find(item => Number(item.id) === id)
    let url = '/pages/product-edit/product-edit?id=' + id
    if (product) {
      url += '&product=' + encodeURIComponent(JSON.stringify(product))
    }
    wx.navigateTo({ url })
  },

  // 删除商品
  deleteProduct(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认删除',
      content: '确定删除该商品吗？',
      success: (res) => {
        if (res.confirm) {
          http.del('/product/' + id).then(() => {
            wx.showToast({ title: '已删除', icon: 'success' })
            this.loadCategories()
          })
        }
      }
    })
  },

  onPullDownRefresh() {
    this.loadCategories()
    wx.stopPullDownRefresh()
  }
})
