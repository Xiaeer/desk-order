<template>
  <el-card v-loading="loading">
    <template #header>
      <div class="card-header">
        <span>店铺详情</span>
        <el-button @click="$router.back()">返回</el-button>
      </div>
    </template>
    <el-descriptions :column="2" border>
      <el-descriptions-item label="ID">{{ shop.id }}</el-descriptions-item>
      <el-descriptions-item label="店铺名称">{{ shop.name }}</el-descriptions-item>
      <el-descriptions-item label="所属商家">{{ shop.merchant_name }}</el-descriptions-item>
      <el-descriptions-item label="电话">{{ shop.phone }}</el-descriptions-item>
      <el-descriptions-item label="地址" :span="2">{{ shop.address }}</el-descriptions-item>
      <el-descriptions-item label="经度">{{ shop.longitude }}</el-descriptions-item>
      <el-descriptions-item label="纬度">{{ shop.latitude }}</el-descriptions-item>
      <el-descriptions-item label="描述" :span="2">{{ shop.description || '-' }}</el-descriptions-item>
      <el-descriptions-item label="审核状态">
        <el-tag :type="statusType(shop.status)">{{ statusText(shop.status) }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="营业状态">
        <el-tag :type="shop.is_open ? 'success' : 'info'">{{ shop.is_open ? '营业中' : '休息中' }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="Logo" :span="2">
        <el-image v-if="shop.logo" :src="shop.logo" style="width: 100px; height: 100px" fit="cover" />
        <span v-else>-</span>
      </el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ shop.created_at }}</el-descriptions-item>
    </el-descriptions>
    <div class="audit-actions">
      <el-button v-if="shop.status === 0" type="success" @click="audit(1)">通过审核</el-button>
      <el-button v-if="shop.status === 0" type="danger" @click="audit(2)">拒绝</el-button>
      <el-button type="danger" plain @click="removeShop">删除店铺</el-button>
    </div>

    <el-divider>审核记录</el-divider>
    <el-table :data="records" v-loading="historyLoading" stripe empty-text="暂无审核记录">
      <el-table-column prop="created_at" label="时间" width="180" />
      <el-table-column prop="action_text" label="操作" width="160" />
      <el-table-column label="状态变更" width="180">
        <template #default="{ row }">
          <span>{{ statusLabel(row.from_status) }} -> {{ statusLabel(row.to_status) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作者" width="180">
        <template #default="{ row }">
          <span>{{ operatorLabel(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="说明" min-width="220" />
    </el-table>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getShopDetail, getShopAuditRecords, auditShop, deleteShop } from '../../api/shop'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const historyLoading = ref(false)
const shop = ref({})
const records = ref([])

function statusText(s) {
  return ['待审核', '已通过', '已拒绝'][s] || '未知'
}
function statusType(s) {
  return ['warning', 'success', 'danger'][s] || 'info'
}
function statusLabel(s) {
  if (s === -1) return '无'
  return statusText(s)
}
function operatorLabel(row) {
  const roleText = row.operator_role === 'admin' ? '后台' : '商户'
  return row.operator_name ? `${roleText} / ${row.operator_name}` : roleText
}

async function fetchDetail() {
  loading.value = true
  try {
    const res = await getShopDetail(route.params.id)
    shop.value = res.data || {}
  } finally {
    loading.value = false
  }
}

async function fetchHistory() {
  historyLoading.value = true
  try {
    const res = await getShopAuditRecords(route.params.id)
    records.value = res.data || []
  } finally {
    historyLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([fetchDetail(), fetchHistory()])
})

async function audit(status) {
  const action = status === 1 ? '通过' : '拒绝'
  await ElMessageBox.confirm(`确认${action}该店铺？`, '审核确认', { type: 'warning' })
  await auditShop(route.params.id, { status })
  ElMessage.success(`已${action}`)
  await Promise.all([fetchDetail(), fetchHistory()])
}

async function removeShop() {
  await ElMessageBox.confirm('确认删除该店铺？删除后商户可以重新发起入驻申请。若店铺已有订单，将禁止删除。', '删除确认', { type: 'warning' })
  await deleteShop(route.params.id)
  ElMessage.success('店铺已删除')
  router.back()
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.audit-actions {
  margin-top: 20px;
  display: flex;
  gap: 12px;
}
</style>
