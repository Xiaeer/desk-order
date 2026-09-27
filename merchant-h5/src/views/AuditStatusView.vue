<template>
  <div class="page-stack">
    <div class="card page-stack">
      <span class="chip">审核状态</span>
      <div class="page-title">{{ title }}</div>
      <div class="page-subtitle">{{ description }}</div>
      <button class="secondary-btn" @click="goToAuditShop">{{ shopButtonText }}</button>
      <button class="primary-btn" :disabled="refreshing" @click="refreshStatus">
        {{ refreshing ? '刷新中...' : '刷新状态' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getShop } from '../api/shop'
import { getApiErrorMessage, isShopNotFoundError } from '../utils/request'

const route = useRoute()
const router = useRouter()
const refreshing = ref(false)
const status = ref(Number(route.query.status || 0))

watch(
  () => route.query.status,
  value => {
    status.value = Number(value || 0)
  },
  { immediate: true }
)

const title = computed(() => {
  if (status.value === 1) return '审核已通过'
  if (status.value === 2) return '审核未通过'
  return '等待平台审核'
})
const description = computed(() => {
  if (status.value === 1) return '店铺已审核通过，可以正常营业和接单。'
  if (status.value === 2) return '店铺审核未通过，请修改店铺信息后重新提交。'
  return '你已完成入驻提交，平台审核通过后即可营业。'
})
const shopButtonText = computed(() => (status.value === 2 ? '修改店铺并重新提交' : '查看店铺信息'))

function goToAuditShop() {
  router.push(status.value === 2 ? '/audit-shop?mode=resubmit' : '/audit-shop')
}

async function refreshStatus() {
  refreshing.value = true
  try {
    const response = await getShop()
    const nextStatus = Number(response.data?.status || 0)
    status.value = nextStatus
    if (nextStatus === 1) {
      router.replace('/home')
      return
    }
    router.replace(`/audit-status?status=${nextStatus}`)
  } catch (error) {
    if (isShopNotFoundError(error)) {
      router.replace('/register')
      return
    }
    window.alert(getApiErrorMessage(error, '刷新审核状态失败'))
  } finally {
    refreshing.value = false
  }
}
</script>
