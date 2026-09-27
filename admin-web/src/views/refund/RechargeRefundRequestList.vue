<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>充值退款管理</span>
        <div class="toolbar">
          <el-select v-model="statusFilter" placeholder="筛选状态" clearable style="width: 160px" @change="handleFilter">
            <el-option label="待审核" value="pending_review" />
            <el-option label="处理中" value="processing" />
            <el-option label="已完成" value="success" />
            <el-option label="已驳回" value="rejected" />
            <el-option label="已失败" value="failed" />
            <el-option label="已撤回" value="cancelled" />
          </el-select>
          <el-select v-model="channelFilter" placeholder="筛选发起方" clearable style="width: 160px" @change="handleFilter">
            <el-option label="用户发起" value="user" />
            <el-option label="后台发起" value="admin" />
            <el-option label="商家发起" value="merchant" />
          </el-select>
          <el-button type="primary" @click="openCreate">创建退款申请</el-button>
        </div>
      </div>
    </template>

    <el-alert
      class="page-tip"
      title="支持自动微信退款，也保留手工兜底"
      type="warning"
      :closable="false"
      description="自动微信退款会先把本地账本标记为处理中并预扣余额，之后需要在列表里同步微信状态；如果你已经在外部渠道手工退款，则填写外部退款单号后再完成。"
    />

    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="request_no" label="申请单号" width="210" />
      <el-table-column label="充值单号" width="210">
        <template #default="{ row }">
          <el-button v-if="row.recharge_order_id" type="primary" link @click="openRechargeOrder(row.recharge_order_id)">
            {{ row.recharge_order_no || `#${row.recharge_order_id}` }}
          </el-button>
          <span v-else>{{ row.recharge_order_no || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="用户" min-width="180">
        <template #default="{ row }">
          <div>{{ row.user_nickname || '未命名用户' }}</div>
          <div class="cell-sub">ID {{ row.user_id }} / {{ row.user_phone || '未绑定手机号' }}</div>
        </template>
      </el-table-column>
      <el-table-column label="发起方" width="110">
        <template #default="{ row }">{{ row.request_channel_text || '-' }}</template>
      </el-table-column>
      <el-table-column label="申请备注" min-width="180">
        <template #default="{ row }">
          <el-tooltip v-if="row.request_note" :content="row.request_note" placement="top-start">
            <div class="ellipsis-text">{{ row.request_note }}</div>
          </el-tooltip>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="申请退款" width="140">
        <template #default="{ row }">¥{{ formatAmount(row.requested_principal_amount) }}</template>
      </el-table-column>
      <el-table-column label="作废赠送" width="140">
        <template #default="{ row }">¥{{ formatAmount(row.requested_gift_void_amount) }}</template>
      </el-table-column>
      <el-table-column label="当前批次剩余" width="160">
        <template #default="{ row }">
          <div>本 ¥{{ formatAmount(row.current_principal_remaining) }}</div>
          <div class="cell-sub">赠 ¥{{ formatAmount(row.current_gift_remaining) }}</div>
        </template>
      </el-table-column>
      <el-table-column label="当前余额" width="120">
        <template #default="{ row }">¥{{ formatAmount(row.current_balance_amount) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="申请时间" width="180">
        <template #default="{ row }">{{ row.created_at || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="openDetail(row.id)">详情</el-button>
          <el-button v-if="row.status === 'pending_review'" type="success" link @click="openComplete(row.id)">完成退款</el-button>
          <el-button v-if="row.status === 'processing'" type="warning" link @click="syncStatus(row.id)">同步微信状态</el-button>
          <el-button v-if="row.status === 'pending_review'" type="danger" link @click="openReject(row.id)">驳回</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        :page-size="size"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="fetchList"
      />
    </div>
  </el-card>

  <el-dialog v-model="createVisible" title="创建充值退款申请" width="520px">
    <el-form label-width="120px">
      <el-form-item label="充值订单ID">
        <el-input-number v-model="createForm.recharge_order_id" :min="1" style="width: 100%" />
      </el-form-item>
      <el-form-item label="充值单号">
        <el-input v-model="createForm.recharge_order_no" maxlength="64" placeholder="可选；不知道内部ID时填充值单号" />
      </el-form-item>
      <el-form-item label="申请备注">
        <el-input v-model="createForm.request_note" type="textarea" :rows="3" maxlength="120" placeholder="申请说明，可选" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="createVisible = false">取消</el-button>
      <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="detailVisible" title="退款申请详情" width="640px">
    <el-descriptions v-if="detail" :column="2" border>
      <el-descriptions-item label="申请单号">{{ detail.request_no }}</el-descriptions-item>
      <el-descriptions-item label="充值单号">
        <el-button v-if="detail.recharge_order_id" type="primary" link @click="openRechargeOrder(detail.recharge_order_id)">
          {{ detail.recharge_order_no || `#${detail.recharge_order_id}` }}
        </el-button>
        <span v-else>{{ detail.recharge_order_no || '-' }}</span>
      </el-descriptions-item>
      <el-descriptions-item label="用户">{{ detail.user_nickname || '-' }}</el-descriptions-item>
      <el-descriptions-item label="手机号">{{ detail.user_phone || '-' }}</el-descriptions-item>
      <el-descriptions-item label="申请退款本金">¥{{ formatAmount(detail.requested_principal_amount) }}</el-descriptions-item>
      <el-descriptions-item label="申请作废赠送">¥{{ formatAmount(detail.requested_gift_void_amount) }}</el-descriptions-item>
      <el-descriptions-item label="当前剩余本金">¥{{ formatAmount(detail.current_principal_remaining) }}</el-descriptions-item>
      <el-descriptions-item label="当前剩余赠送">¥{{ formatAmount(detail.current_gift_remaining) }}</el-descriptions-item>
      <el-descriptions-item label="当前用户余额">¥{{ formatAmount(detail.current_balance_amount) }}</el-descriptions-item>
      <el-descriptions-item label="发起方式">{{ detail.request_channel_text || '-' }}</el-descriptions-item>
      <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
      <el-descriptions-item label="执行状态">{{ executionStatusText(detail.execution_status) }}</el-descriptions-item>
      <el-descriptions-item label="已同步次数">{{ detail.sync_retry_count || 0 }}</el-descriptions-item>
      <el-descriptions-item label="最近同步时间">{{ detail.last_sync_at || '-' }}</el-descriptions-item>
      <el-descriptions-item label="下次自动同步">{{ detail.next_auto_sync_at || '-' }}</el-descriptions-item>
      <el-descriptions-item label="自动同步状态">{{ detail.auto_sync_paused ? '已暂停，需人工同步' : '正常' }}</el-descriptions-item>
      <el-descriptions-item label="外部退款单号">{{ detail.external_refund_no || '-' }}</el-descriptions-item>
      <el-descriptions-item label="外部状态码">{{ detail.external_response_code || '-' }}</el-descriptions-item>
      <el-descriptions-item label="申请时间">{{ detail.created_at || '-' }}</el-descriptions-item>
      <el-descriptions-item label="审核时间">{{ detail.reviewed_at || '-' }}</el-descriptions-item>
      <el-descriptions-item label="申请备注" :span="2">{{ detail.request_note || '-' }}</el-descriptions-item>
      <el-descriptions-item label="审核备注" :span="2">{{ detail.review_note || '-' }}</el-descriptions-item>
      <el-descriptions-item label="驳回原因" :span="2">{{ detail.reject_reason || '-' }}</el-descriptions-item>
      <el-descriptions-item label="外部返回信息" :span="2">{{ detail.external_response_msg || '-' }}</el-descriptions-item>
    </el-descriptions>
  </el-dialog>

  <el-dialog v-model="reviewVisible" :title="reviewMode === 'complete' ? '完成退款' : '驳回退款申请'" width="560px">
    <el-form label-width="120px">
      <template v-if="reviewMode === 'complete'">
        <el-alert
          class="dialog-tip"
          title="留空则尝试自动退款"
          type="info"
          :closable="false"
          description="留空时会按当前后端支付配置尝试自动发起微信退款，并先进入处理中；后续请用列表里的“同步微信状态”确认最终结果。"
        />
        <el-form-item label="外部退款单号">
          <el-input v-model="reviewForm.external_refund_no" maxlength="64" placeholder="可选；手工退款时填写，例如微信商户平台退款单号" />
        </el-form-item>
        <el-form-item label="外部返回信息">
          <el-input v-model="reviewForm.external_response_msg" type="textarea" :rows="3" maxlength="120" placeholder="可选；手工退款时记录外部渠道返回说明" />
        </el-form-item>
      </template>
      <template v-else>
        <el-form-item label="驳回原因">
          <el-input v-model="reviewForm.reject_reason" type="textarea" :rows="3" maxlength="120" placeholder="请输入驳回原因" />
        </el-form-item>
      </template>
      <el-form-item label="审核备注">
        <el-input v-model="reviewForm.review_note" type="textarea" :rows="3" maxlength="120" placeholder="审核备注，可选" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="reviewVisible = false">取消</el-button>
      <el-button :type="reviewMode === 'complete' ? 'success' : 'danger'" :loading="reviewing" @click="submitReview">
        {{ reviewMode === 'complete' ? '确认完成' : '确认驳回' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import {
  createRechargeRefundRequest,
  getRechargeRefundRequestDetail,
  getRechargeRefundRequestList,
  reviewRechargeRefundRequest,
  syncRechargeRefundRequestStatus
} from '../../api/refund'

const router = useRouter()

const list = ref([])
const total = ref(0)
const page = ref(1)
const size = 20
const loading = ref(false)
const creating = ref(false)
const reviewing = ref(false)
const statusFilter = ref('pending_review')
const channelFilter = ref('')

const createVisible = ref(false)
const detailVisible = ref(false)
const reviewVisible = ref(false)
const reviewMode = ref('complete')
const currentReviewId = ref(0)
const detail = ref(null)

const createForm = reactive({
  recharge_order_id: undefined,
  recharge_order_no: '',
  request_note: ''
})

const reviewForm = reactive({
  external_refund_no: '',
  external_response_msg: '',
  reject_reason: '',
  review_note: ''
})

function formatAmount(value) {
  return (Number(value || 0) / 100).toFixed(2)
}

function statusText(status) {
  return {
    pending_review: '待审核',
    processing: '处理中',
    success: '已完成',
    rejected: '已驳回',
    failed: '已失败',
    cancelled: '已撤回'
  }[status] || status || '-'
}

function executionStatusText(status) {
  return {
    processing: '处理中',
    success: '已完成',
    failed: '已失败'
  }[status] || status || '-'
}

function statusType(status) {
  return {
    pending_review: 'warning',
    processing: 'primary',
    success: 'success',
    rejected: 'danger',
    failed: 'danger',
    cancelled: 'info'
  }[status] || 'info'
}

function openRechargeOrder(rechargeOrderID) {
  if (!rechargeOrderID) {
    return
  }
  router.push(`/recharge-order/${rechargeOrderID}`)
}

async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, size }
    if (statusFilter.value) {
      params.status = statusFilter.value
    }
    if (channelFilter.value) {
      params.request_channel = channelFilter.value
    }
    const res = await getRechargeRefundRequestList(params)
    list.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

function handleFilter() {
  page.value = 1
  fetchList()
}

function resetCreateForm() {
  createForm.recharge_order_id = undefined
  createForm.recharge_order_no = ''
  createForm.request_note = ''
}

function resetReviewForm() {
  reviewForm.external_refund_no = ''
  reviewForm.external_response_msg = ''
  reviewForm.reject_reason = ''
  reviewForm.review_note = ''
}

function openCreate() {
  resetCreateForm()
  createVisible.value = true
}

async function submitCreate() {
  if (!createForm.recharge_order_id && !createForm.recharge_order_no.trim()) {
    ElMessage.error('请输入充值订单ID或充值单号')
    return
  }
  creating.value = true
  try {
    await createRechargeRefundRequest({
      recharge_order_id: createForm.recharge_order_id,
      recharge_order_no: createForm.recharge_order_no.trim(),
      request_note: createForm.request_note.trim()
    })
    createVisible.value = false
    ElMessage.success('退款申请已创建')
    fetchList()
  } finally {
    creating.value = false
  }
}

async function openDetail(id) {
  const res = await getRechargeRefundRequestDetail(id)
  detail.value = res.data || null
  detailVisible.value = true
}

function openComplete(id) {
  currentReviewId.value = id
  reviewMode.value = 'complete'
  resetReviewForm()
  reviewVisible.value = true
}

function openReject(id) {
  currentReviewId.value = id
  reviewMode.value = 'reject'
  resetReviewForm()
  reviewVisible.value = true
}

async function submitReview() {
  if (!currentReviewId.value) {
    return
  }
  if (reviewMode.value === 'reject' && !reviewForm.reject_reason.trim()) {
    ElMessage.error('请输入驳回原因')
    return
  }
  reviewing.value = true
  try {
    const res = await reviewRechargeRefundRequest(currentReviewId.value, {
      action: reviewMode.value,
      review_note: reviewForm.review_note.trim(),
      reject_reason: reviewForm.reject_reason.trim(),
      external_refund_no: reviewForm.external_refund_no.trim(),
      external_response_msg: reviewForm.external_response_msg.trim()
    })
    reviewVisible.value = false
    const nextStatus = res.data?.status
    if (reviewMode.value === 'complete') {
      ElMessage.success(nextStatus === 'processing' ? '退款已受理，请稍后同步微信状态' : '退款申请已完成')
    } else {
      ElMessage.success('退款申请已驳回')
    }
    fetchList()
  } finally {
    reviewing.value = false
  }
}

async function syncStatus(id) {
  const res = await syncRechargeRefundRequestStatus(id)
  const status = res.data?.status
  if (status === 'success') {
    ElMessage.success('微信退款已确认成功')
  } else if (status === 'failed') {
    ElMessage.error('微信退款确认失败，已回滚本地预扣余额')
  } else {
    ElMessage.info('微信退款仍在处理中')
  }
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-tip {
  margin-bottom: 16px;
}

.dialog-tip {
  margin-bottom: 16px;
}

.cell-sub {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.ellipsis-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .card-header,
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>