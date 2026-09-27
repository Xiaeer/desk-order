<template>
  <el-card v-loading="loading">
    <template #header>
      <div class="card-header">
        <span>订单详情</span>
        <el-button @click="$router.back()">返回</el-button>
      </div>
    </template>
    <el-descriptions :column="2" border>
      <el-descriptions-item label="订单号">{{ order.order_no }}</el-descriptions-item>
      <el-descriptions-item label="店铺">{{ order.shop_name }}</el-descriptions-item>
      <el-descriptions-item label="用户ID">{{ order.user_id }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="orderStatusType(order.status)">{{ orderStatusText(order.status) }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="总金额">¥{{ (order.total_amount / 100).toFixed(2) }}</el-descriptions-item>
      <el-descriptions-item label="备注">{{ order.remark || '-' }}</el-descriptions-item>
      <el-descriptions-item label="下单时间">{{ formatTime(order.created_at) }}</el-descriptions-item>
      <el-descriptions-item label="支付时间">{{ order.paid_at ? formatTime(order.paid_at) : '-' }}</el-descriptions-item>
    </el-descriptions>

    <h3 style="margin-top: 20px">订单商品</h3>
    <el-table :data="order.items || []" border>
      <el-table-column prop="name" label="商品名称" />
      <el-table-column label="子选项" min-width="220">
        <template #default="{ row }">{{ row.option_summary || '-' }}</template>
      </el-table-column>
      <el-table-column label="图片" width="80">
        <template #default="{ row }">
          <el-image v-if="row.image" :src="row.image" style="width: 40px; height: 40px" fit="cover" />
        </template>
      </el-table-column>
      <el-table-column label="单价" width="120">
        <template #default="{ row }">¥{{ (row.price / 100).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column prop="quantity" label="数量" width="80" />
      <el-table-column label="小计" width="120">
        <template #default="{ row }">¥{{ (row.subtotal / 100).toFixed(2) }}</template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getOrderDetail } from '../../api/order'

const route = useRoute()
const loading = ref(false)
const order = ref({})

function orderStatusText(s) {
  return ['待支付', '已支付', '已接单', '已完成', '已取消'][s] || '未知'
}
function orderStatusType(s) {
  return ['info', 'warning', 'primary', 'success', 'danger'][s] || 'info'
}
function formatTime(t) {
  if (!t) return '-'
  return String(t).replace('T', ' ').substring(0, 19)
}

onMounted(async () => {
  loading.value = true
  try {
    const res = await getOrderDetail(route.params.id)
    order.value = res.data
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
