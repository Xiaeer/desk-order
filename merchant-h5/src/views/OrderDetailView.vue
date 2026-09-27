<template>
  <section class="page-stack">
    <div v-if="order" class="card page-stack">
      <div class="row-between">
        <div class="page-title">{{ order.order_no }}</div>
        <span class="chip">{{ statusText(order.status) }}</span>
      </div>
      <div class="page-subtitle">店铺：{{ order.shop_name }}</div>
      <div class="page-subtitle">下单时间：{{ formatTime(order.created_at) }}</div>
      <div class="page-subtitle">备注：{{ order.remark || '无' }}</div>
      <div class="page-title" style="font-size: 18px;">商品明细</div>
      <div class="list">
        <div v-for="item in order.items || []" :key="item.id" class="row-between">
          <div>
            <div>{{ item.name }}</div>
            <div v-if="item.option_summary" class="page-subtitle">{{ item.option_summary }}</div>
            <div class="page-subtitle">¥{{ (item.price / 100).toFixed(2) }} x {{ item.quantity }}</div>
          </div>
          <strong>¥{{ (item.subtotal / 100).toFixed(2) }}</strong>
        </div>
      </div>
      <div class="row-between">
        <button v-if="order.status === 1" class="primary-btn" @click="changeOrder('accept')">接单</button>
        <button v-else-if="order.status === 2" class="primary-btn" @click="changeOrder('complete')">完成订单</button>
      </div>
    </div>
  </section>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { acceptOrder, completeOrder, getOrderDetail } from '../api/order'
import { setupMerchantOrderRealtime } from '../utils/merchant-ws'
import { getApiErrorMessage } from '../utils/request'

const route = useRoute()
const router = useRouter()
const order = ref(null)
let disposeRealtime = null

function formatTime(value) {
  return String(value || '').replace('T', ' ').slice(0, 19) || '-'
}

function statusText(status) {
  return ['待支付', '待接单', '已接单', '已完成', '已取消'][status] || '未知'
}

async function loadDetail(options = {}) {
  const { silent = false } = options
  try {
    const response = await getOrderDetail(route.params.id)
    order.value = response.data
  } catch (error) {
    if (!silent) {
      window.alert(getApiErrorMessage(error, '加载订单详情失败'))
    }
  }
}

async function changeOrder(action) {
  try {
    if (action === 'accept') {
      await acceptOrder(route.params.id)
    } else {
      await completeOrder(route.params.id)
    }
    await loadDetail()
    router.push('/orders')
  } catch (error) {
    window.alert(getApiErrorMessage(error, '更新订单状态失败'))
  }
}

onMounted(async () => {
  await loadDetail()
  disposeRealtime = setupMerchantOrderRealtime(message => {
    const messageOrderID = Number(message?.data?.id || 0)
    if (messageOrderID && String(messageOrderID) !== String(route.params.id)) {
      return
    }
    return loadDetail({ silent: true })
  })
})

onBeforeUnmount(() => {
  disposeRealtime?.()
})
</script>
