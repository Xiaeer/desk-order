<template>
  <section class="page-stack">
    <div class="card page-stack">
      <div class="page-title">设置 / 重置 H5 密码</div>
      <div class="page-subtitle">
        已登录商户可在这里修改 merchant-h5 登录密码。网页端修改密码需要先验证当前密码；若忘记旧密码，请回到商户小程序重置。
      </div>
      <div class="field-grid">
        <div class="field">
          <label>当前密码</label>
          <input v-model.trim="form.currentPassword" type="password" placeholder="请输入当前密码" />
        </div>
        <div class="field">
          <label>新密码</label>
          <input v-model.trim="form.newPassword" type="password" placeholder="至少 6 位" />
        </div>
        <div class="field">
          <label>确认新密码</label>
          <input v-model.trim="form.confirmPassword" type="password" placeholder="请再次输入新密码" />
        </div>
      </div>
      <button class="primary-btn" :disabled="submitting" @click="submit">{{ submitting ? '保存中...' : '保存密码' }}</button>
    </div>
  </section>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { updateMerchantWebPassword } from '../api/auth'
import { getApiErrorMessage } from '../utils/request'

const router = useRouter()
const submitting = ref(false)
const form = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

async function submit() {
  if (!form.currentPassword) {
    window.alert('请输入当前密码')
    return
  }
  if (!form.newPassword || !form.confirmPassword) {
    window.alert('请填写完整的新密码信息')
    return
  }
  if (form.newPassword !== form.confirmPassword) {
    window.alert('两次输入的新密码不一致')
    return
  }

  submitting.value = true
  try {
    await updateMerchantWebPassword({
      current_password: form.currentPassword,
      new_password: form.newPassword,
      confirm_password: form.confirmPassword
    })
    window.alert('H5 登录密码已更新')
    router.back()
  } catch (error) {
    window.alert(getApiErrorMessage(error, '保存密码失败'))
  } finally {
    submitting.value = false
  }
}
</script>
