const http = require('../../utils/request')

Page({
  data: {
    id: null,
    categoryId: null,
    name: '',
    image: '',
    price: '',
    description: '',
    isOnSale: true,
    sort: 0,
    loading: false,
    uploadingImage: false
  },

  onLoad(options) {
    if (options.id) {
      // 编辑模式，加载商品信息
      this.setData({ id: parseInt(options.id) })
      wx.setNavigationBarTitle({ title: '编辑商品' })
      this.loadProduct(options)
    } else if (options.category_id) {
      this.setData({ categoryId: parseInt(options.category_id) })
      wx.setNavigationBarTitle({ title: '新增商品' })
    }
  },

  loadProduct(options) {
    if (!options || !options.product) {
      return
    }
    try {
      const product = JSON.parse(decodeURIComponent(options.product))
      this.setData({
        id: product.id,
        categoryId: product.category_id,
        name: product.name || '',
        image: product.image || '',
        price: product.price ? (product.price / 100).toFixed(2) : '',
        description: product.description || '',
        isOnSale: typeof product.is_on_sale === 'boolean' ? product.is_on_sale : true,
        sort: product.sort || 0
      })
    } catch (err) {
      wx.showToast({ title: '商品信息加载失败', icon: 'none' })
    }
  },

  onInputName(e) { this.setData({ name: e.detail.value }) },
  onInputPrice(e) { this.setData({ price: e.detail.value }) },
  onInputDesc(e) { this.setData({ description: e.detail.value }) },
  onInputSort(e) { this.setData({ sort: parseInt(e.detail.value) || 0 }) },

  toggleOnSale() {
    this.setData({ isOnSale: !this.data.isOnSale })
  },

  // 选择图片
  chooseImage() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      success: (res) => {
        const filePath = res.tempFiles[0].tempFilePath
        const previousImage = this.data.image
        this.setData({ uploadingImage: true })
        wx.showLoading({ title: '上传图片中', mask: true })
        http.upload('/upload/product-image', filePath).then((data) => {
          this.setData({ image: data.url || '' })
          wx.showToast({ title: '图片上传成功', icon: 'success' })
        }).catch(() => {
          this.setData({ image: previousImage })
        }).finally(() => {
          wx.hideLoading()
          this.setData({ uploadingImage: false })
        })
      }
    })
  },

  // 保存商品
  handleSave() {
    const { name, price, categoryId, id, image, description, isOnSale, sort, uploadingImage } = this.data
    if (!name.trim()) return wx.showToast({ title: '请输入商品名', icon: 'none' })
    if (!price || parseFloat(price) <= 0) return wx.showToast({ title: '请输入正确价格', icon: 'none' })
    if (uploadingImage) return wx.showToast({ title: '图片上传中，请稍后', icon: 'none' })

    const priceInCents = Math.round(parseFloat(price) * 100)
    this.setData({ loading: true })

    const payload = {
      name: name.trim(),
      price: priceInCents,
      image,
      description: description.trim(),
      is_on_sale: isOnSale,
      sort
    }

    let promise
    if (id) {
      promise = http.put('/product/' + id, payload)
    } else {
      payload.category_id = categoryId
      promise = http.post('/product', payload)
    }

    promise.then(() => {
      wx.showToast({ title: '保存成功', icon: 'success' })
      setTimeout(() => wx.navigateBack(), 1000)
    }).catch(() => {}).finally(() => {
      this.setData({ loading: false })
    })
  }
})
