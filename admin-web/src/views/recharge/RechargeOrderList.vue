<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span class="card-title">充值订单</span>
        <div class="toolbar">
          <el-input
            v-model="filters.order_no"
            clearable
            placeholder="搜索充值单号"
            style="width: 220px"
            @keyup.enter="handleSearch"
          />
          <el-input
            v-model="filters.user_keyword"
            clearable
            placeholder="搜索用户ID/昵称/手机号/OpenID尾号"
            style="width: 260px"
            @keyup.enter="handleSearch"
          />
          <el-select v-model="filters.status" clearable placeholder="筛选状态" style="width: 140px" @change="handleSearch">
            <el-option label="待支付" :value="0" />
            <el-option label="已支付" :value="1" />
            <el-option label="已取消" :value="2" />
          </el-select>
          <el-select v-model="filters.suggested_action" clearable placeholder="建议下一步" style="width: 190px" @change="handleSearch">
            <el-option label="去用户详情处理历史余额" value="handle_historical_balance" />
            <el-option label="先处理退款申请" value="handle_refund" />
            <el-option label="申请退款或作废此批次剩余" value="refund_or_cleanup" />
            <el-option label="仅作废此批次剩余" value="cleanup_only" />
            <el-option label="无需处理" value="no_action" />
          </el-select>
          <el-date-picker
            v-model="filters.created_range"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            range-separator="至"
            start-placeholder="创建开始时间"
            end-placeholder="创建结束时间"
            style="width: 360px"
          />
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button :loading="exporting" @click="handleExport">导出 CSV</el-button>
        </div>
      </div>
    </template>

    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="order_no" label="充值单号" min-width="210" />
      <el-table-column label="用户" min-width="180">
        <template #default="{ row }">
          <el-button type="primary" link @click="router.push(`/user/${row.user_id}`)">
            {{ row.user_display_name || row.user_nickname || `用户#${row.user_id}` }}
          </el-button>
          <div class="cell-sub">{{ row.user_identity_label || row.user_phone || `用户ID ${row.user_id}` }}</div>
        </template>
      </el-table-column>
      <el-table-column label="充值活动" min-width="180">
        <template #default="{ row }">
          <div>{{ row.activity_name || '-' }}</div>
          <div class="cell-sub">活动ID {{ row.activity_id }} / 规则 v{{ row.activity_version || 1 }}</div>
        </template>
      </el-table-column>
      <el-table-column label="到账构成" width="150">
        <template #default="{ row }">
          <div>充 ¥{{ formatAmount(row.recharge_amount) }}</div>
          <div class="cell-sub">赠 ¥{{ formatAmount(row.gift_amount) }}</div>
        </template>
      </el-table-column>
      <el-table-column label="支付金额" width="110">
        <template #default="{ row }">¥{{ formatAmount(row.pay_amount) }}</template>
      </el-table-column>
      <el-table-column label="到账金额" width="110">
        <template #default="{ row }">¥{{ formatAmount(row.total_arrival_amount) }}</template>
      </el-table-column>
      <el-table-column label="充值支付方式" width="120">
        <template #default="{ row }">{{ row.pay_channel_text || row.pay_channel || '-' }}</template>
      </el-table-column>
      <el-table-column label="订单状态" width="110">
        <template #default="{ row }">
          <el-tag :type="orderStatusType(row.status)">{{ row.status_text || '-' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最新退款" width="130">
        <template #default="{ row }">
          <el-tag v-if="row.refund_status" :type="refundStatusType(row.refund_status)">{{ row.refund_status_text || row.refund_status }}</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="当前剩余" width="120">
        <template #default="{ row }">{{ formatRemainingAmount(row.current_remaining_amount, row.has_remaining_snapshot) }}</template>
      </el-table-column>
      <el-table-column label="建议下一步" min-width="240">
        <template #default="{ row }">
          <div>{{ row.suggested_action_text || '-' }}</div>
          <el-button
            v-if="getSuggestedActionLink(row)"
            class="cell-link"
            type="primary"
            link
            @click="handleSuggestedAction(row)"
          >
            {{ getSuggestedActionLink(row).label }}
          </el-button>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ row.created_at || '-' }}</template>
      </el-table-column>
      <el-table-column label="支付时间" width="180">
        <template #default="{ row }">{{ row.paid_at || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="row.suggested_action === 'handle_refund'"
            type="warning"
            link
            @click="handleSuggestedAction(row)"
          >
            去退款管理
          </el-button>
          <el-button
            v-if="row.can_create_refund"
            type="danger"
            link
            :loading="creatingRefundId === row.id"
            @click="handleCreateRefund(row)"
          >
            申请退款
          </el-button>
          <el-button type="primary" link @click="router.push(`/recharge-order/${row.id}`)">详情</el-button>
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
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { getRechargeOrderList } from '../../api/recharge'
import { createRechargeRefundRequest } from '../../api/refund'

const router = useRouter()

const list = ref([])
const total = ref(0)
const page = ref(1)
const size = 20
const loading = ref(false)
const exporting = ref(false)
const creatingRefundId = ref(0)
const filters = reactive({
  order_no: '',
  user_keyword: '',
  status: undefined,
  suggested_action: '',
  created_range: []
})

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

function getSuggestedActionLink(row) {
  if (!row) {
    return null
  }
  if (row.suggested_action === 'handle_historical_balance') {
    return { label: '去用户详情', path: `/user/${row.user_id}` }
  }
  if (row.suggested_action === 'handle_refund') {
    return { label: '去退款管理', path: '/recharge-refunds' }
  }
  if (row.suggested_action === 'refund_or_cleanup' || row.suggested_action === 'cleanup_only') {
    return { label: '去订单详情', path: `/recharge-order/${row.id}` }
  }
  return null
}

function handleSuggestedAction(row) {
  const link = getSuggestedActionLink(row)
  if (!link) {
    return
  }
  router.push(link.path)
}

function buildQueryParams(currentPage = page.value, currentSize = size) {
  const params = { page: currentPage, size: currentSize }
  if (filters.order_no.trim()) {
    params.order_no = filters.order_no.trim()
  }
  if (filters.user_keyword.trim()) {
    params.user_keyword = filters.user_keyword.trim()
  }
  if (filters.status !== undefined && filters.status !== null && filters.status !== '') {
    params.status = filters.status
  }
  if (filters.suggested_action) {
    params.suggested_action = filters.suggested_action
  }
  if (Array.isArray(filters.created_range) && filters.created_range.length === 2) {
    params.created_start_at = filters.created_range[0]
    params.created_end_at = filters.created_range[1]
  }
  return params
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getRechargeOrderList(buildQueryParams())
    list.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

function csvEscape(value) {
  return `"${String(value ?? '').replace(/"/g, '""')}"`
}

function buildExportFileName() {
  const now = new Date()
  const parts = [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, '0'),
    String(now.getDate()).padStart(2, '0'),
    String(now.getHours()).padStart(2, '0'),
    String(now.getMinutes()).padStart(2, '0'),
    String(now.getSeconds()).padStart(2, '0')
  ]
  return `recharge-orders-${parts.join('')}.csv`
}

function downloadCsv(content) {
  const blob = new Blob([`\uFEFF${content}`], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.href = url
  link.download = buildExportFileName()
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function buildExportContent(rows) {
  const header = [
    '充值单号',
    '用户ID',
    '用户展示名',
    '识别标识',
    '充值活动',
    '规则版本',
    '充值金额(元)',
    '赠送金额(元)',
    '到账金额(元)',
    '支付金额(元)',
    '充值支付方式',
    '订单状态',
    '最新退款状态',
    '当前剩余(元)',
    '建议下一步',
    '退款申请单号',
    '退款申请金额(元)',
    '已批退款金额(元)',
    '创建时间',
    '支付时间'
  ]
  const body = rows.map((row) => [
    row.order_no || '',
    row.user_id || '',
    row.user_display_name || row.user_nickname || '',
    row.user_identity_label || row.user_phone || '',
    row.activity_name || '',
    row.activity_version || '',
    formatAmount(row.recharge_amount),
    formatAmount(row.gift_amount),
    formatAmount(row.total_arrival_amount),
    formatAmount(row.pay_amount),
    row.pay_channel_text || row.pay_channel || '',
    row.status_text || '',
    row.refund_status_text || '',
    row.has_remaining_snapshot ? formatAmount(row.current_remaining_amount) : '待核对',
    row.suggested_action_text || '',
    row.refund_request_no || '',
    formatAmount(row.refund_requested_amount),
    formatAmount(row.refund_approved_amount),
    row.created_at || '',
    row.paid_at || ''
  ].map(csvEscape).join(','))
  return [header.map(csvEscape).join(','), ...body].join('\n')
}

async function handleExport() {
  exporting.value = true
  try {
    const exportSize = 100
    const firstRes = await getRechargeOrderList(buildQueryParams(1, exportSize))
    const exportTotal = Number(firstRes.data.total || 0)
    if (!exportTotal) {
      ElMessage.warning('当前筛选条件下没有可导出的充值订单')
      return
    }
    const rows = [...(firstRes.data.list || [])]
    const totalPages = Math.ceil(exportTotal / exportSize)
    for (let currentPage = 2; currentPage <= totalPages; currentPage += 1) {
      const res = await getRechargeOrderList(buildQueryParams(currentPage, exportSize))
      rows.push(...(res.data.list || []))
    }
    downloadCsv(buildExportContent(rows.slice(0, exportTotal)))
    ElMessage.success(`已导出 ${Math.min(rows.length, exportTotal)} 条充值订单`)
  } finally {
    exporting.value = false
  }
}

async function handleCreateRefund(row) {
  if (!row || !row.id) {
    return
  }
  try {
    await ElMessageBox.confirm(`确认直接为充值单 ${row.order_no} 创建退款申请？`, '申请退款', {
      type: 'warning',
      confirmButtonText: '确认申请',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  creatingRefundId.value = row.id
  try {
    await createRechargeRefundRequest({ recharge_order_id: row.id })
    ElMessage.success('退款申请已创建，可到充值退款页继续审核')
    await fetchList()
  } finally {
    creatingRefundId.value = 0
  }
}

function handleSearch() {
  page.value = 1
  fetchList()
}

function handleReset() {
  filters.order_no = ''
  filters.user_keyword = ''
  filters.status = undefined
  filters.suggested_action = ''
  filters.created_range = []
  handleSearch()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.card-title {
  flex-shrink: 0;
  line-height: 32px;
  white-space: nowrap;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: flex-end;
  flex: 1 1 960px;
  min-width: 0;
}

.cell-sub {
  color: #909399;
  font-size: 12px;
  margin-top: 4px;
}

.cell-link {
  margin-top: 4px;
  padding: 0;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>