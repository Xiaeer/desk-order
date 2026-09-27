<template>
  <el-card v-loading="loading">
    <template #header>
      <div class="card-header">
        <span>充值订单详情</span>
        <div class="header-actions">
          <el-button
            v-if="order?.can_cleanup_remaining"
            type="warning"
            :loading="cleaningRemaining"
            @click="handleCleanupRemaining"
          >
            作废此批次剩余
          </el-button>
          <el-button
            v-if="order?.can_create_refund"
            type="danger"
            :loading="creatingRefund"
            @click="handleCreateRefund"
          >
            一键申请退款
          </el-button>
          <el-button @click="$router.back()">返回</el-button>
        </div>
      </div>
    </template>

    <el-descriptions v-if="order" :column="2" border>
      <el-descriptions-item label="充值单号">{{ order.order_no }}</el-descriptions-item>
      <el-descriptions-item label="订单状态">
        <el-tag :type="orderStatusType(order.status)">{{ order.status_text || '-' }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="用户昵称">{{ order.user_nickname || '-' }}</el-descriptions-item>
      <el-descriptions-item label="用户手机号">{{ order.user_phone || '-' }}</el-descriptions-item>
      <el-descriptions-item label="用户ID">{{ order.user_id }}</el-descriptions-item>
      <el-descriptions-item label="充值活动">{{ order.activity_name || '-' }}</el-descriptions-item>
      <el-descriptions-item label="活动ID">{{ order.activity_id }}</el-descriptions-item>
      <el-descriptions-item label="规则版本">v{{ order.activity_version || 1 }}</el-descriptions-item>
      <el-descriptions-item label="充值金额">¥{{ formatAmount(order.recharge_amount) }}</el-descriptions-item>
      <el-descriptions-item label="赠送金额">¥{{ formatAmount(order.gift_amount) }}</el-descriptions-item>
      <el-descriptions-item label="到账金额">¥{{ formatAmount(order.total_arrival_amount) }}</el-descriptions-item>
      <el-descriptions-item label="支付金额">¥{{ formatAmount(order.pay_amount) }}</el-descriptions-item>
      <el-descriptions-item label="充值支付方式">{{ order.pay_channel_text || order.pay_channel || '-' }}</el-descriptions-item>
      <el-descriptions-item label="过期时间">{{ order.expires_at || '-' }}</el-descriptions-item>
      <el-descriptions-item label="支付时间">{{ order.paid_at || '-' }}</el-descriptions-item>
      <el-descriptions-item label="当前剩余本金">{{ formatRemainingAmount(order.current_principal_remaining, order.has_remaining_snapshot) }}</el-descriptions-item>
      <el-descriptions-item label="当前剩余赠送">{{ formatRemainingAmount(order.current_gift_remaining, order.has_remaining_snapshot) }}</el-descriptions-item>
      <el-descriptions-item label="当前剩余额度">{{ formatRemainingAmount(order.current_remaining_amount, order.has_remaining_snapshot) }}</el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ order.created_at || '-' }}</el-descriptions-item>
      <el-descriptions-item label="更新时间">{{ order.updated_at || '-' }}</el-descriptions-item>
    </el-descriptions>

    <el-alert
      v-if="order?.can_cleanup_remaining"
      class="cleanup-alert"
      type="warning"
      :closable="false"
      title="作废此批次剩余会直接作废当前充值单对应批次的剩余本金和赠送余额，不走微信退款。"
    />
    <el-alert
      v-else-if="order?.cleanup_unavailable_reason"
      class="cleanup-alert"
      type="info"
      :closable="false"
      :title="`未显示作废此批次剩余按钮：${order.cleanup_unavailable_reason}`"
    />
    <el-alert
      v-if="order?.refund_unavailable_reason"
      class="cleanup-alert"
      type="info"
      :closable="false"
      :title="`未显示一键申请退款按钮：${order.refund_unavailable_reason}`"
    />

    <h3 class="section-title">最新退款状态</h3>
    <el-descriptions v-if="order && order.refund_request_id" :column="2" border>
      <el-descriptions-item label="退款申请ID">{{ order.refund_request_id }}</el-descriptions-item>
      <el-descriptions-item label="退款申请单号">{{ order.refund_request_no || '-' }}</el-descriptions-item>
      <el-descriptions-item label="退款状态">
        <el-tag :type="refundStatusType(order.refund_status)">{{ order.refund_status_text || order.refund_status || '-' }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="发起方式">{{ order.refund_request_channel_text || order.refund_request_channel || '-' }}</el-descriptions-item>
      <el-descriptions-item label="申请退款金额">¥{{ formatAmount(order.refund_requested_amount) }}</el-descriptions-item>
      <el-descriptions-item label="已批退款金额">¥{{ formatAmount(order.refund_approved_amount) }}</el-descriptions-item>
      <el-descriptions-item label="退款更新时间" :span="2">{{ order.refund_updated_at || '-' }}</el-descriptions-item>
    </el-descriptions>
    <el-empty v-else description="暂无关联退款记录" />
  </el-card>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { cleanupRechargeOrderRemaining, getRechargeOrderDetail } from '../../api/recharge'
import { createRechargeRefundRequest } from '../../api/refund'

const route = useRoute()
const loading = ref(false)
const creatingRefund = ref(false)
const cleaningRemaining = ref(false)
const order = ref(null)

function formatAmount(value) {
  return (Number(value || 0) / 100).toFixed(2)
}

function formatRemainingAmount(value, hasSnapshot) {
  if (!hasSnapshot) {
    return '待核对'
  }
  return `¥${formatAmount(value)}`
}

function orderStatusType(status) {
  return {
    0: 'warning',
    1: 'success',
    2: 'info'
  }[status] || 'info'
}

function refundStatusType(status) {
  return {
    pending_review: 'warning',
    approved: 'primary',
    processing: 'primary',
    success: 'success',
    rejected: 'danger',
    failed: 'danger',
    cancelled: 'info',
    exception: 'danger'
  }[status] || 'info'
}

async function fetchDetail() {
  loading.value = true
  try {
    const res = await getRechargeOrderDetail(route.params.id)
    order.value = res.data || null
  } finally {
    loading.value = false
  }
}

async function handleCreateRefund() {
  if (!order.value?.id) {
    return
  }
  try {
    await ElMessageBox.confirm(`确认直接为充值单 ${order.value.order_no} 创建退款申请？`, '申请退款', {
      type: 'warning',
      confirmButtonText: '确认申请',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  creatingRefund.value = true
  try {
    await createRechargeRefundRequest({ recharge_order_id: order.value.id })
    ElMessage.success('退款申请已创建，可到充值退款页继续审核')
    await fetchDetail()
  } finally {
    creatingRefund.value = false
  }
}

async function handleCleanupRemaining() {
  if (!order.value?.id) {
    return
  }
  let promptResult
  try {
    promptResult = await ElMessageBox.prompt(
      `请输入备注。该操作会直接作废充值单 ${order.value.order_no} 对应批次的当前剩余余额，不走微信退款。`,
      '作废此批次剩余',
      {
        type: 'warning',
        confirmButtonText: '确认作废',
        cancelButtonText: '取消',
        inputValue: '作废此批次剩余',
        inputPlaceholder: '请输入备注'
      }
    )
  } catch {
    return
  }
  cleaningRemaining.value = true
  try {
    await cleanupRechargeOrderRemaining(order.value.id, { remark: promptResult.value || '' })
    ElMessage.success('该批次剩余已作废')
    await fetchDetail()
  } finally {
    cleaningRemaining.value = false
  }
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
  align-items: center;
  gap: 12px;
}

.section-title {
  margin: 24px 0 16px;
}

.cleanup-alert {
  margin-top: 16px;
}
</style>