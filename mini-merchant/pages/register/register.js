const http = require('../../utils/request')

const AUDIT_STATUS_PAGE = '/pages/audit-status/audit-status'

Page({
  data: {
    name: '',
    phone: '',
    shopName: '',
    address: '',
    latitude: 0,
    longitude: 0,
    description: '',
    loading: false,
    located: false,
    locationDenied: false
  },

  onInputName(e) { this.setData({ name: e.detail.value }) },
  onInputPhone(e) { this.setData({ phone: e.detail.value }) },
  onInputShopName(e) { this.setData({ shopName: e.detail.value }) },
  onInputAddress(e) { this.setData({ address: e.detail.value }) },
  onInputDesc(e) { this.setData({ description: e.detail.value }) },

  isLocationPermissionDenied(error) {
    const errMsg = error && error.errMsg ? error.errMsg : ''
    return /auth deny|auth denied|authorize|scope.userLocation/i.test(errMsg)
  },

  requestLocation() {
    wx.getLocation({
      type: 'gcj02',
      success: (res) => {
        this.setData({
          latitude: res.latitude,
          longitude: res.longitude,
          located: true,
          locationDenied: false
        })
        wx.showToast({ title: '定位成功', icon: 'success' })
      },
      fail: (error) => {
        if (this.isLocationPermissionDenied(error)) {
          this.setData({ locationDenied: true })
          this.showLocationPermissionGuide()
          return
        }
        wx.showToast({ title: '定位失败，请稍后重试', icon: 'none' })
      }
    })
  },

  showLocationPermissionGuide() {
    wx.showModal({
      title: '需要定位权限',
      content: '提交入驻申请需要获取店铺坐标，请在设置中开启“位置信息”权限。',
      confirmText: '去开启',
      success: (res) => {
        if (!res.confirm) return

        wx.openSetting({
          success: (settingRes) => {
            if (settingRes.authSetting['scope.userLocation']) {
              this.requestLocation()
              return
            }

            wx.showToast({ title: '未开启定位权限', icon: 'none' })
          },
          fail: () => {
            wx.showToast({ title: '无法打开设置页', icon: 'none' })
          }
        })
      }
    })
  },

  // 获取当前位置
  getLocation() {
    wx.getSetting({
      success: (res) => {
        if (res.authSetting['scope.userLocation'] === false) {
          this.setData({ locationDenied: true })
          this.showLocationPermissionGuide()
          return
        }

        this.requestLocation()
      },
      fail: () => {
        this.requestLocation()
      }
    })
  },

  // 提交注册
  handleSubmit() {
    const { name, phone, shopName, address, latitude, longitude, description } = this.data
    if (!name.trim()) return wx.showToast({ title: '请输入姓名', icon: 'none' })
    if (!phone.trim()) return wx.showToast({ title: '请输入手机号', icon: 'none' })
    if (!shopName.trim()) return wx.showToast({ title: '请输入店铺名称', icon: 'none' })
    if (!address.trim()) return wx.showToast({ title: '请输入地址', icon: 'none' })
    if (!this.data.located) return wx.showToast({ title: '请先获取定位', icon: 'none' })

    this.setData({ loading: true })

    // 先注册商家信息
    wx.login({
      success: (loginRes) => {
        http.post('/register', {
          code: loginRes.code,
          name: name.trim(),
          phone: phone.trim()
        }).then((data) => {
          wx.setStorageSync('token', data.token)
          wx.setStorageSync('is_new', data.is_new)
          wx.setStorageSync('has_shop', data.has_shop)

          // 再创建店铺
          return http.post('/shop', {
            name: shopName.trim(),
            address: address.trim(),
            phone: phone.trim(),
            latitude,
            longitude,
            description: description.trim()
          })
        }).then(() => {
          wx.setStorageSync('has_shop', true)
          wx.showToast({ title: '注册成功，待审核', icon: 'success' })
          setTimeout(() => {
            wx.reLaunch({ url: AUDIT_STATUS_PAGE + '?status=0' })
          }, 1500)
        }).catch(() => {}).finally(() => {
          this.setData({ loading: false })
        })
      }
    })
  }
})
