<template>
  <div class="mobile-main">
    <section class="page-stack">
      <div class="card page-stack">
        <span class="chip">DeskOrder 商户端</span>
        <div class="page-title">手机号登录</div>
        <div class="page-subtitle">保留现有 mini-merchant 代码不动，新 H5 使用手机号 + 密码登录。</div>
        <div v-if="loginNotice" class="login-notice">{{ loginNotice }}</div>
        <div class="field-grid">
          <div class="field">
            <label>手机号</label>
            <input v-model.trim="form.phone" placeholder="请输入手机号" />
          </div>
          <div class="field">
            <label>密码</label>
            <input v-model.trim="form.password" type="password" placeholder="请输入密码" />
          </div>
        </div>
        <button class="primary-btn" :disabled="submitting" @click="handleLogin">{{ submitting ? '登录中...' : '登录' }}</button>
        <button class="secondary-btn" @click="router.push('/register')">没有账号，去注册</button>
        <div class="card page-stack" style="padding:14px;background:#fff7ef;box-shadow:none;border:1px solid #ffd7c7;">
          <div style="font-size:16px;font-weight:600;color:#8a3d1f;">首次登录 / 忘记密码</div>
          <div class="page-subtitle" style="margin:0;color:#8a3d1f;">
            老商户迁移到 H5，请先在原商户小程序进入“店铺 - 账号安全”设置密码；如果忘记 H5 旧密码，也请回到小程序同一入口重置。
          </div>
          <div class="page-subtitle" style="margin:0;color:#8a3d1f;">
            操作路径：商户小程序登录 -> 店铺 -> 账号安全 -> 设置 / 重置 H5 密码。
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { merchantWebLogin } from '../api/auth'
import { setHasShop, setToken, getToken } from '../utils/auth'
import { resolveMerchantEntry } from '../utils/merchant'
import { getApiErrorMessage } from '../utils/request'

const router = useRouter()
const submitting = ref(false)
const loginNotice = ref('')
const form = reactive({
  phone: '',
  password: ''
})

onMounted(() => {
  loginNotice.value = sessionStorage.getItem('merchant_h5_auth_expired_notice') || ''
  if (loginNotice.value) {
    sessionStorage.removeItem('merchant_h5_auth_expired_notice')
  }
  if (getToken()) {
    resolveMerchantEntry(router).catch(error => {
      window.alert(getApiErrorMessage(error, '进入商户端失败'))
    })
  }
})

async function handleLogin() {
  if (!form.phone || !form.password) {
    window.alert('请输入手机号和密码')
    return
  }

  submitting.value = true
  try {
    const response = await merchantWebLogin(form)
    setToken(response.data.token)
    setHasShop(response.data.has_shop)
    await resolveMerchantEntry(router, response.data.has_shop)
  } catch (error) {
    window.alert(getApiErrorMessage(error, '登录失败'))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.login-notice {
  margin: 8px 0 2px;
  padding: 10px 12px;
  border-radius: 12px;
  background: #fff1f0;
  color: #b42318;
  font-size: 13px;
  line-height: 1.5;
  border: 1px solid #fecdca;
}
</style>
