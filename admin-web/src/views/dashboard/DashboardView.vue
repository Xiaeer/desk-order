<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ data.merchant_count }}</div>
          <div class="stat-label">商家总数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ data.shop_count }}</div>
          <div class="stat-label">店铺总数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ data.order_count }}</div>
          <div class="stat-label">订单总数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ data.user_count }}</div>
          <div class="stat-label">用户总数</div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">¥{{ (data.total_amount / 100).toFixed(2) }}</div>
          <div class="stat-label">累计交易金额</div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover" class="stat-card pending">
          <div class="stat-value">{{ data.pending_shops }}</div>
          <div class="stat-label">待审核店铺</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getDashboard } from '../../api/order'

const data = ref({
  merchant_count: 0,
  shop_count: 0,
  order_count: 0,
  user_count: 0,
  total_amount: 0,
  pending_shops: 0
})

onMounted(async () => {
  const res = await getDashboard()
  data.value = res.data
})
</script>

<style scoped>
.stat-card {
  text-align: center;
  padding: 20px 0;
}
.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
}
.stat-label {
  margin-top: 8px;
  font-size: 14px;
  color: #909399;
}
.pending .stat-value {
  color: #E6A23C;
}
</style>
