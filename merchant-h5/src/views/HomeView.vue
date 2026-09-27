<template>
  <section class="page-stack">
    <div class="card page-stack">
      <div class="row-between">
        <div>
          <div class="page-title">{{ shop.name || '商户首页' }}</div>
          <div class="page-subtitle">{{ shop.address || '正在加载店铺信息...' }}</div>
        </div>
        <span class="chip">{{ shop.is_open ? '营业中' : '已打烊' }}</span>
      </div>
      <div class="row-between">
        <button class="secondary-btn" @click="toggleOpen">{{ shop.is_open ? '暂停营业' : '开始营业' }}</button>
        <button class="primary-btn" @click="loadPage">刷新</button>
      </div>
    </div>

    <div class="card row-between">
      <div>
        <div class="page-subtitle">待接单数</div>
        <div class="page-title">{{ orders.length }}</div>
      </div>
      <div style="text-align: right;">
        <div class="page-subtitle">待接金额</div>
        <div class="page-title">¥{{ todayAmountText }}</div>
      </div>
    </div>

    <div class="card page-stack">
      <div class="row-between">
        <div class="page-title" style="font-size: 18px;">待接订单</div>
        <button class="secondary-btn" @click="router.push('/orders')">查看全部</button>
      </div>
      <div v-if="orders.length" class="list">
        <div v-for="order in orders" :key="order.id" class="card" style="padding: 14px; box-shadow: none; background: #fafbfc;">
          <div class="row-between">
            <strong>{{ order.order_no }}</strong>
            <span class="chip">¥{{ (order.total_amount / 100).toFixed(2) }}</span>
          </div>
          <div class="page-subtitle">{{ formatTime(order.created_at) }}</div>
          <div class="page-subtitle">备注：{{ order.remark || '无' }}</div>
          <div class="row-between">
            <button class="secondary-btn" @click="router.push(`/order/${order.id}`)">详情</button>
            <button class="primary-btn" @click="handleAccept(order.id)">接单</button>
          </div>
        </div>
      </div>
      <div v-else class="empty-text">当前没有待接订单</div>
    </div>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getOrders, acceptOrder } from '../api/order'
import { getShop, toggleShopOpen } from '../api/shop'
import { setupMerchantOrderRealtime } from '../utils/merchant-ws'
import { getApiErrorMessage, isShopNotFoundError } from '../utils/request'

const router = useRouter()
const shop = ref({})
const orders = ref([])
const todayAmountText = computed(() => (orders.value.reduce((sum, order) => sum + order.total_amount, 0) / 100).toFixed(2))
let disposeRealtime = null

function formatTime(value) {
  return String(value || '').replace('T', ' ').slice(0, 19) || '-'
}

async function loadPage(options = {}) {
  const { silent = false } = options
  try {
    const shopResponse = await getShop()
    if (shopResponse.data.status !== 1) {
      router.replace(`/audit-status?status=${shopResponse.data.status || 0}`)
      return
    }
    shop.value = shopResponse.data
    const orderResponse = await getOrders({ page: 1, size: 50, status: 1 })
    orders.value = orderResponse.data.list || []
  } catch (error) {
    if (isShopNotFoundError(error)) {
      router.replace('/register')
      return
    }
    if (!silent) {
      window.alert(getApiErrorMessage(error, '加载首页失败'))
    }
  }
}

async function toggleOpen() {
  try {
    await toggleShopOpen(!shop.value.is_open)
    shop.value.is_open = !shop.value.is_open
  } catch (error) {
    window.alert(getApiErrorMessage(error, '切换营业状态失败'))
  }
}

async function handleAccept(id) {
  try {
    await acceptOrder(id)
    await loadPage()
  } catch (error) {
    window.alert(getApiErrorMessage(error, '接单失败'))
  }
}

onMounted(async () => {
  await loadPage()
  disposeRealtime = setupMerchantOrderRealtime(() => loadPage({ silent: true }))
})

onBeforeUnmount(() => {
  disposeRealtime?.()
})
</script>
