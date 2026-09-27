const http = require('../../utils/request')
const auth = require('../../utils/auth')
const { goAuditStatus } = require('../../utils/navigation')

Page({
  data: {
    shop: null,
    editing: false,
    loading: false,
    form: {
      name: '',
      address: '',
      phone: '',
      description: ''
    }
  },

  onShow() {
    this.loadShop()
  },

  loadShop() {
    http.get('/shop').then(data => {
      this.setData({
        shop: data,
        form: {
          name: data.name,
          address: data.address,
          phone: data.phone,
          description: data.description || ''
        }
      })
    }).catch(() => {})
  },

  // 切换编辑模式
  toggleEdit() {
    if (!this.data.shop || this.data.loading) return

    if (this.data.editing) {
      const { name, address, phone, description } = this.data.form
      if (!name.trim()) return wx.showToast({ title: '店名不能为空', icon: 'none' })
      if (!address.trim()) return wx.showToast({ title: '地址不能为空', icon: 'none' })
      if (!phone.trim()) return wx.showToast({ title: '电话不能为空', icon: 'none' })

      const payload = {
        name: name.trim(),
        address: address.trim(),
        phone: phone.trim(),
        description: description.trim()
      }

      if (!this.hasAuditFieldChanges(payload)) {
        this.setData({ editing: false })
        return
      }

      wx.showModal({
        title: '提交变更审核',
        content: '修改店铺名称、地址、电话、描述等审核字段后，将重新进入审核，审核期间无法营业，确认提交吗？',
        success: (res) => {
          if (!res.confirm) return
          this.submitAuditChange(payload)
        }
      })
    } else {
      if (this.data.shop.status !== 1) {
        wx.showToast({ title: '当前状态不可修改店铺信息', icon: 'none' })
        return
      }
      this.setData({ editing: true })
    }
  },

  hasAuditFieldChanges(payload) {
    const shop = this.data.shop || {}
    return payload.name !== (shop.name || '') ||
      payload.address !== (shop.address || '') ||
      payload.phone !== (shop.phone || '') ||
      payload.description !== (shop.description || '')
  },

  submitAuditChange(payload) {
    this.setData({ loading: true })
    http.put('/shop', payload).then((data) => {
      this.setData({
        editing: false,
        loading: false,
        shop: data,
        form: {
          name: data.name,
          address: data.address,
          phone: data.phone,
          description: data.description || ''
        }
      })
      wx.showToast({ title: '已提交变更，等待审核', icon: 'success' })
      setTimeout(() => {
        goAuditStatus(data.status || 0)
      }, 1200)
    }).catch(() => {
      this.setData({ loading: false })
    })
  },

  cancelEdit() {
    this.setData({
      editing: false,
      form: {
        name: this.data.shop.name,
        address: this.data.shop.address,
        phone: this.data.shop.phone,
        description: this.data.shop.description || ''
      }
    })
  },

  onInputName(e) { this.setData({ 'form.name': e.detail.value }) },
  onInputAddress(e) { this.setData({ 'form.address': e.detail.value }) },
  onInputPhone(e) { this.setData({ 'form.phone': e.detail.value }) },
  onInputDesc(e) { this.setData({ 'form.description': e.detail.value }) },

  goAccountSecurity() {
    wx.navigateTo({ url: '/pages/account-security/account-security' })
  },

  // 切换营业状态
  toggleOpen() {
    if (this.data.loading) return
    const isOpen = !this.data.shop.is_open
    http.put('/shop/status', { is_open: isOpen }).then(() => {
      this.setData({ 'shop.is_open': isOpen })
      wx.showToast({ title: isOpen ? '已开始营业' : '已休息', icon: 'success' })
    })
  },

  // 退出登录
  handleLogout() {
    wx.showModal({
      title: '提示',
      content: '确定退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          auth.logout()
          wx.reLaunch({ url: '/pages/login/login' })
        }
      }
    })
  }
})
