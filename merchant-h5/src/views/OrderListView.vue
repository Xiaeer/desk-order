<template>
  <section class="page-stack">
    <div class="card page-stack">
      <div class="page-title">订单管理</div>
      <div class="row-between" style="flex-wrap: wrap;">
        <button
          v-for="item in tabs"
          :key="item.key"
          class="secondary-btn"
          :style="activeTab === item.key ? activeStyle : ''"
          @click="switchTab(item.key)"
        >
          {{ item.label }}
        </button>
      </div>
    </div>

    <div class="list">
      <div v-for="order in orders" :key="order.id" class="card page-stack">
        <div class="row-between">
          <strong>{{ order.order_no }}</strong>
          <span class="chip">{{ statusText(order.status) }}</span>
        </div>
        <div class="page-subtitle">{{ order.shop_name }} · ¥{{ (order.total_amount / 100).toFixed(2) }}</div>
        <div class="page-subtitle">{{ formatTime(order.created_at) }}</div>
        <div v-if="order.items?.length" class="page-subtitle order-item-summary">
          {{ summarizeItems(order.items) }}
        </div>
        <div class="row-between">
          <button class="secondary-btn" @click="router.push(`/order/${order.id}`)">详情</button>
          <button v-if="order.status === 1" class="primary-btn" @click="changeOrder(order.id, 'accept')">接单</button>
          <button v-else-if="order.status === 2" class="primary-btn" @click="changeOrder(order.id, 'complete')">完成</button>
        </div>
      </div>
      <div v-if="!orders.length" class="card empty-text">当前没有符合条件的订单</div>
    </div>
  </section>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { acceptOrder, completeOrder, getOrders } from '../api/order'
import { setupMerchantOrderRealtime } from '../utils/merchant-ws'
import { getApiErrorMessage } from '../utils/request'

const router = useRouter()
const activeTab = ref('')
const orders = ref([])
let disposeRealtime = null
const tabs = [
  { key: '', label: '全部' },
  { key: '1', label: '待接单' },
  { key: '2', label: '已接单' },
  { key: '3', label: '已完成' }
]
const activeStyle = 'background:#ff6a3d;color:#fff;'

function formatTime(value) {
  return String(value || '').replace('T', ' ').slice(0, 19) || '-'
}

function statusText(status) {
  return ['待支付', '待接单', '已接单', '已完成', '已取消'][status] || '未知'
}

function summarizeItems(items) {
  return (items || [])
    .slice(0, 2)
    .map(item => item.option_summary ? `${item.name}（${item.option_summary}）` : item.name)
    .join(' · ')
}

async function loadOrders(options = {}) {
  const { silent = false } = options
  try {
    const params = { page: 1, size: 50 }
    if (activeTab.value) {
      params.status = activeTab.value
    }
    const response = await getOrders(params)
    orders.value = response.data.list || []
  } catch (error) {
    if (!silent) {
      window.alert(getApiErrorMessage(error, '加载订单失败'))
    }
  }
}

function switchTab(key) {
  activeTab.value = key
  loadOrders()
}

async function changeOrder(id, action) {
  try {
    if (action === 'accept') {
      await acceptOrder(id)
    } else {
      await completeOrder(id)
    }
    await loadOrders()
  } catch (error) {
    window.alert(getApiErrorMessage(error, '更新订单状态失败'))
  }
}

onMounted(async () => {
  await loadOrders()
  disposeRealtime = setupMerchantOrderRealtime(() => loadOrders({ silent: true }))
})

onBeforeUnmount(() => {
  disposeRealtime?.()
})
</script>

<style scoped>
.order-item-summary {
  line-height: 1.6;
}
</style>
