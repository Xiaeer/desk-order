<template>
  <div class="user-detail" v-loading="loading">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-user">
            <el-avatar :src="user.avatar" :size="48">{{ getDisplayName(user).slice(0, 1) }}</el-avatar>
            <div>
              <div class="header-name">{{ getDisplayName(user) }}</div>
              <div class="header-sub">{{ getIdentityLabel(user) }}</div>
            </div>
          </div>
          <el-button @click="$router.back()">返回</el-button>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户ID">{{ user.id }}</el-descriptions-item>
        <el-descriptions-item label="识别标识">{{ getIdentityLabel(user) }}</el-descriptions-item>
        <el-descriptions-item label="昵称">{{ user.nickname || '-' }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ user.phone || '未绑定手机号' }}</el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ user.created_at }}</el-descriptions-item>
        <el-descriptions-item label="最近更新">{{ user.updated_at }}</el-descriptions-item>
        <el-descriptions-item label="当前余额">¥{{ formatAmount(user.balance_amount) }}</el-descriptions-item>
        <el-descriptions-item label="累计充值">¥{{ formatAmount(user.total_recharge_amount) }}</el-descriptions-item>
        <el-descriptions-item label="累计赠送">¥{{ formatAmount(user.total_gift_amount) }}</el-descriptions-item>
        <el-descriptions-item label="累计消费">¥{{ formatAmount(user.total_consume_amount) }}</el-descriptions-item>
        <el-descriptions-item label="订单数">{{ user.order_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="充值单数">{{ user.recharge_order_count || 0 }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card v-if="showHistoricalBalanceCard" class="historical-balance-card" shadow="never">
      <template #header>
        <div class="section-header">
          <span>历史余额</span>
          <el-button
            v-if="user.can_cleanup_historical_balance"
            type="warning"
            :loading="cleaningHistoricalBalance"
            @click="handleCleanupHistoricalBalance"
          >
            清理历史余额
          </el-button>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="历史批次ID">{{ user.historical_balance_batch_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="待处理金额">¥{{ formatAmount(user.historical_balance_amount) }}</el-descriptions-item>
      </el-descriptions>

      <el-alert
        v-if="user.can_cleanup_historical_balance"
        class="historical-balance-alert"
        type="warning"
        :closable="false"
        title="该操作会直接扣减用户当前余额并清零历史 opening 批次，不会反向重算累计充值/累计赠送。"
      />
      <el-alert
        v-else-if="user.historical_balance_unavailable_reason"
        class="historical-balance-alert"
        type="info"
        :closable="false"
        :title="`当前不可清理历史余额：${user.historical_balance_unavailable_reason}`"
      />
    </el-card>

    <el-card class="recharge-card" shadow="never">
      <template #header>
        <div class="section-header">
          <span>最近充值订单</span>
          <el-tag size="small" effect="plain">近 10 条</el-tag>
        </div>
      </template>

      <el-table :data="recentRechargeOrders" stripe empty-text="暂无充值订单">
        <el-table-column prop="order_no" label="充值单号" min-width="180" show-overflow-tooltip />
        <el-table-column prop="activity_name" label="充值活动" min-width="160" show-overflow-tooltip />
        <el-table-column label="支付金额" width="120">
          <template #default="{ row }">¥{{ formatAmount(row.pay_amount) }}</template>
        </el-table-column>
        <el-table-column label="到账金额" width="120">
          <template #default="{ row }">¥{{ formatAmount(row.total_arrival_amount) }}</template>
        </el-table-column>
        <el-table-column label="充值支付方式" width="120">
          <template #default="{ row }">{{ row.pay_channel_text || '-' }}</template>
        </el-table-column>
        <el-table-column label="订单状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getRechargeOrderStatusType(row.status)">{{ row.status_text || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="退款状态" width="110">
          <template #default="{ row }">
            <el-tag :type="getRefundStatusType(row.refund_status)" effect="plain">{{ row.refund_status_text || '未发起' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="paid_at" label="支付时间" width="180" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="$router.push(`/recharge-order/${row.id}`)">查看订单</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { cleanupUserHistoricalBalance, getUserDetail } from '../../api/user'

const route = useRoute()
const loading = ref(false)
const cleaningHistoricalBalance = ref(false)
const user = ref({})
const recentRechargeOrders = computed(() => user.value.recent_recharge_orders || [])
const showHistoricalBalanceCard = computed(() => {
  return Number(user.value.historical_balance_amount || 0) > 0 || !!user.value.historical_balance_unavailable_reason
})

function formatAmount(value) {
  return (Number(value || 0) / 100).toFixed(2)
}

function getDisplayName(value) {
  return value.display_name || value.nickname || `用户#${value.id || '-'}`
}

function getIdentityLabel(value) {
  return value.identity_label || value.phone || `用户ID ${value.id || '-'}`
}

function getRechargeOrderStatusType(status) {
  if (status === 1) {
    return 'success'
  }
  if (status === 0) {
    return 'warning'
  }
  return 'info'
}

function getRefundStatusType(status) {
  if (status === 'success') {
    return 'success'
  }
  if (status === 'processing' || status === 'approved' || status === 'pending_review') {
    return 'warning'
  }
  if (status === 'failed' || status === 'exception' || status === 'rejected') {
    return 'danger'
  }
  return 'info'
}

async function fetchDetail() {
  loading.value = true
  try {
    const res = await getUserDetail(route.params.id)
    user.value = res.data || {}
  } finally {
    loading.value = false
  }
}

async function handleCleanupHistoricalBalance() {
  if (!user.value?.id || !user.value?.can_cleanup_historical_balance) {
    return
  }
  let promptResult
  try {
    promptResult = await ElMessageBox.prompt(
      `请输入备注。该操作会直接清理用户 ${getDisplayName(user.value)} 的历史余额 ¥${formatAmount(user.value.historical_balance_amount)}。`,
      '清理历史余额',
      {
        type: 'warning',
        confirmButtonText: '确认清理',
        cancelButtonText: '取消',
        inputValue: '历史余额清理',
        inputPlaceholder: '请输入备注'
      }
    )
  } catch {
    return
  }
  cleaningHistoricalBalance.value = true
  try {
    const res = await cleanupUserHistoricalBalance(user.value.id, { remark: promptResult.value || '' })
    user.value = res.data || {}
    ElMessage.success('用户历史余额已清理')
  } finally {
    cleaningHistoricalBalance.value = false
  }
}

onMounted(fetchDetail)
</script>

<style scoped>
.user-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-user {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.header-sub {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.historical-balance-alert {
  margin-top: 16px;
}

.recharge-card :deep(.el-card__body) {
  padding-top: 8px;
}
</style>