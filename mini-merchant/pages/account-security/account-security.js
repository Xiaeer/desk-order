const http = require('../../utils/request')

Page({
  data: {
    newPassword: '',
    confirmPassword: '',
    loading: false
  },

  onInputNewPassword(e) {
    this.setData({ newPassword: e.detail.value })
  },

  onInputConfirmPassword(e) {
    this.setData({ confirmPassword: e.detail.value })
  },

  handleSubmit() {
    const { newPassword, confirmPassword, loading } = this.data
    if (loading) return
    if (!newPassword.trim() || !confirmPassword.trim()) {
      wx.showToast({ title: '请填写完整的新密码信息', icon: 'none' })
      return
    }
    if (newPassword.trim().length < 6) {
      wx.showToast({ title: '新密码至少 6 位', icon: 'none' })
      return
    }
    if (newPassword.trim() !== confirmPassword.trim()) {
      wx.showToast({ title: '两次输入的新密码不一致', icon: 'none' })
      return
    }

    this.setData({ loading: true })
    http.put('/password', {
      new_password: newPassword.trim(),
      confirm_password: confirmPassword.trim()
    }).then(() => {
      wx.showToast({ title: 'H5 密码已更新', icon: 'success' })
      this.setData({
        newPassword: '',
        confirmPassword: ''
      })
      setTimeout(() => {
        wx.navigateBack()
      }, 1000)
    }).catch(() => {}).finally(() => {
      this.setData({ loading: false })
    })
  }
})
