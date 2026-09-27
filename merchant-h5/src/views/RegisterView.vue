<template>
  <div class="mobile-main">
    <section class="page-stack">
      <div class="card page-stack">
        <div class="page-title">商户注册</div>
        <div class="page-subtitle">先创建商户账号，再提交店铺入驻信息。审核通过后即可营业。</div>
        <div class="field-grid">
          <div class="field">
            <label>联系人姓名</label>
            <input v-model.trim="form.name" placeholder="请输入联系人姓名" />
          </div>
          <div class="field">
            <label>手机号</label>
            <input v-model.trim="form.phone" placeholder="请输入手机号" />
          </div>
          <div class="field">
            <label>登录密码</label>
            <input v-model.trim="form.password" type="password" placeholder="至少 6 位" />
          </div>
          <div class="field">
            <label>店铺名称</label>
            <input v-model.trim="form.shopName" placeholder="请输入店铺名称" />
          </div>
          <div class="field">
            <label>店铺地址</label>
            <input v-model.trim="form.address" placeholder="请输入店铺地址" />
          </div>
          <div class="field">
            <label>店铺描述</label>
            <textarea v-model.trim="form.description" rows="3" placeholder="可选"></textarea>
          </div>
          <div class="field-grid" style="grid-template-columns: 1fr 1fr;">
            <div class="field">
              <label>纬度</label>
              <input v-model.number="form.latitude" placeholder="纬度" />
            </div>
            <div class="field">
              <label>经度</label>
              <input v-model.number="form.longitude" placeholder="经度" />
            </div>
          </div>
          <button class="secondary-btn" :disabled="locating" @click="fillCurrentLocation">
            {{ locating ? '定位中...' : '获取当前位置' }}
          </button>
        </div>
        <button class="primary-btn" :disabled="submitting" @click="handleSubmit">{{ submitting ? '提交中...' : '注册并提交入驻' }}</button>
        <button class="secondary-btn" @click="router.push('/login')">已有账号，返回登录</button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { merchantWebRegister } from '../api/auth'
import { createShop } from '../api/shop'
import { setHasShop, setToken } from '../utils/auth'
import { wgs84ToGcj02 } from '../utils/coordTransform'
import { getApiErrorMessage } from '../utils/request'

const router = useRouter()
const submitting = ref(false)
const locating = ref(false)
const form = reactive({
  name: '',
  phone: '',
  password: '',
  shopName: '',
  address: '',
  description: '',
  latitude: '',
  longitude: ''
})

function getCurrentPosition(options) {
  return new Promise((resolve, reject) => {
    navigator.geolocation.getCurrentPosition(resolve, reject, options)
  })
}

function getLocationErrorMessage(error) {
  if (!error) {
    return '定位失败，请稍后重试或手动填写经纬度'
  }
  if (error.code === 1) {
    return '浏览器已拒绝定位，请在系统或浏览器设置中允许定位权限'
  }
  if (error.code === 2) {
    return '定位服务暂时不可用，请确认手机系统定位服务已开启'
  }
  if (error.code === 3) {
    return '定位超时，请在开阔区域重试；若仍失败，请手动填写经纬度'
  }
  return error.message || '定位失败，请稍后重试或手动填写经纬度'
}

function getLocationAccuracy(position) {
  const accuracy = position?.coords?.accuracy
  return typeof accuracy === 'number' && accuracy >= 0 ? accuracy : null
}

function getLocationQualityMessage(position, usedFallback) {
  const accuracy = getLocationAccuracy(position)
  if (!usedFallback && (accuracy === null || accuracy <= 100)) {
    return ''
  }
  if (accuracy !== null) {
    return `当前定位精度约 ${accuracy.toFixed(0)} 米，可能为模糊位置。建议到室外或靠窗重试；如店铺位置确认无误，也可继续使用当前坐标。`
  }
  return '当前定位可能已回退到模糊位置。建议到室外或靠窗重试；如店铺位置确认无误，也可继续使用当前坐标。'
}

async function fillCurrentLocation() {
  if (locating.value) {
    return
  }
  if (!window.isSecureContext) {
    window.alert('当前页面不是安全上下文，手机浏览器通常只允许 HTTPS 页面获取定位')
    return
  }
  if (!navigator.geolocation) {
    window.alert('当前浏览器不支持定位，请手动填写经纬度')
    return
  }

  locating.value = true
  try {
    let position
    let usedFallback = false
    try {
      position = await getCurrentPosition({
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 0
      })
    } catch (error) {
      if (error?.code === 1) {
        throw error
      }
      usedFallback = true
      position = await getCurrentPosition({
        enableHighAccuracy: false,
        timeout: 15000,
        maximumAge: 300000
      })
    }

    const converted = wgs84ToGcj02(position.coords.latitude, position.coords.longitude)
    form.latitude = Number(converted.latitude.toFixed(6))
    form.longitude = Number(converted.longitude.toFixed(6))

    const qualityMessage = getLocationQualityMessage(position, usedFallback)
    if (qualityMessage) {
      window.alert(qualityMessage)
    }
  } catch (error) {
    window.alert(getLocationErrorMessage(error))
  } finally {
    locating.value = false
  }
}

async function handleSubmit() {
  if (!form.name || !form.phone || !form.password || !form.shopName || !form.address) {
    window.alert('请完整填写注册信息')
    return
  }
  if (!form.latitude || !form.longitude) {
    window.alert('请先获取定位或手动填写经纬度')
    return
  }

  submitting.value = true
  try {
    const registerResponse = await merchantWebRegister({
      name: form.name,
      phone: form.phone,
      password: form.password
    })
    setToken(registerResponse.data.token)
    setHasShop(false)

    await createShop({
      name: form.shopName,
      address: form.address,
      phone: form.phone,
      latitude: Number(form.latitude),
      longitude: Number(form.longitude),
      description: form.description
    })

    setHasShop(true)
    router.replace('/audit-status?status=0')
  } catch (error) {
    window.alert(getApiErrorMessage(error, '注册失败'))
  } finally {
    submitting.value = false
  }
}
</script>
