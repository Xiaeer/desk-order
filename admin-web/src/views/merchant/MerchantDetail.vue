<template>
  <el-card v-loading="loading">
    <template #header>
      <div class="card-header">
        <span>商家详情</span>
        <div class="header-actions">
          <el-button type="danger" @click="removeMerchant">删除商家</el-button>
          <el-button @click="$router.back()">返回</el-button>
        </div>
      </div>
    </template>
    <el-descriptions :column="2" border>
      <el-descriptions-item label="ID">{{ merchant.id }}</el-descriptions-item>
      <el-descriptions-item label="商家名称">{{ merchant.name }}</el-descriptions-item>
      <el-descriptions-item label="手机号">{{ merchant.phone }}</el-descriptions-item>
      <el-descriptions-item label="注册时间">{{ merchant.created_at }}</el-descriptions-item>
      <el-descriptions-item label="是否开店">
        <el-tag :type="merchant.has_shop ? 'success' : 'info'">{{ merchant.has_shop ? '已开店' : '未开店' }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="店铺名称">{{ merchant.shop_name || '-' }}</el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMerchantDetail, deleteMerchant } from '../../api/merchant'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const merchant = ref({})

async function fetchDetail() {
  loading.value = true
  try {
    const res = await getMerchantDetail(route.params.id)
    merchant.value = res.data
  } finally {
    loading.value = false
  }
}

async function removeMerchant() {
  const shopTip = merchant.value.has_shop ? '该商家关联的店铺也会一并删除。' : ''
  await ElMessageBox.confirm(`确认删除商家“${merchant.value.name || merchant.value.id}”？${shopTip}`, '删除确认', { type: 'warning' })
  await deleteMerchant(route.params.id)
  ElMessage.success('商家已删除')
  router.push('/merchants')
}

onMounted(fetchDetail)
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 12px;
}
</style>
