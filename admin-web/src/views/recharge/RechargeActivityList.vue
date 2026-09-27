<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>充值活动管理</span>
        <div class="toolbar">
          <el-select v-model="statusFilter" placeholder="筛选状态" clearable style="width: 150px" @change="handleFilter">
            <el-option label="已下线" :value="0" />
            <el-option label="已上线" :value="1" />
          </el-select>
          <el-button type="primary" @click="openCreate">新增活动</el-button>
        </div>
      </div>
    </template>

    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="name" label="活动名称" min-width="180" />
      <el-table-column label="充值金额" width="110">
        <template #default="{ row }">¥{{ formatAmount(row.recharge_amount) }}</template>
      </el-table-column>
      <el-table-column label="赠送金额" width="110">
        <template #default="{ row }">¥{{ formatAmount(row.gift_amount) }}</template>
      </el-table-column>
      <el-table-column label="到账金额" width="110">
        <template #default="{ row }">¥{{ formatAmount(row.arrival_amount) }}</template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '已上线' : '已下线' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="活动时间" min-width="220">
        <template #default="{ row }">
          <div>{{ row.start_at || '不限开始' }}</div>
          <div>{{ row.end_at || '不限结束' }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="180" show-overflow-tooltip />
      <el-table-column label="操作" width="340">
        <template #default="{ row }">
          <el-button type="primary" link @click="openEdit(row)">编辑</el-button>
          <el-button type="warning" link @click="toggleStatus(row)">{{ row.status === 1 ? '下线' : '上线' }}</el-button>
          <el-button v-if="row.status === 0" type="danger" plain size="small" @click="removeActivity(row)">删除活动</el-button>
          <el-button v-else type="danger" plain size="small" disabled>先下线后删除</el-button>
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

  <el-dialog v-model="dialogVisible" :title="editingId ? '编辑充值活动' : '新增充值活动'" width="560px">
    <el-form label-width="110px" class="activity-form">
      <el-form-item label="活动名称">
        <el-input v-model="form.name" maxlength="40" placeholder="例如：新客储值礼" />
      </el-form-item>
      <el-form-item label="充值金额">
        <el-input-number v-model="form.rechargeAmountYuan" :min="0.01" :precision="2" :step="10" style="width: 100%" />
      </el-form-item>
      <el-form-item label="赠送金额">
        <el-input-number v-model="form.giftAmountYuan" :min="0" :precision="2" :step="5" style="width: 100%" />
      </el-form-item>
      <el-form-item label="状态">
        <el-radio-group v-model="form.status">
          <el-radio :value="1">上线</el-radio>
          <el-radio :value="0">下线</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="排序">
        <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
      </el-form-item>
      <el-form-item label="开始时间">
        <el-date-picker v-model="form.startAt" type="datetime" placeholder="不限制" style="width: 100%" clearable />
      </el-form-item>
      <el-form-item label="结束时间">
        <el-date-picker v-model="form.endAt" type="datetime" placeholder="不限制" style="width: 100%" clearable />
      </el-form-item>
      <el-form-item label="活动说明">
        <el-input v-model="form.description" type="textarea" :rows="3" maxlength="120" placeholder="活动展示说明，可选" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submitForm">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createRechargeActivity, deleteRechargeActivity, getRechargeActivityList, updateRechargeActivity } from '../../api/recharge'

const list = ref([])
const total = ref(0)
const page = ref(1)
const size = 20
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const editingId = ref(0)
const statusFilter = ref(undefined)

const defaultForm = () => ({
  name: '',
  rechargeAmountYuan: 100,
  giftAmountYuan: 0,
  status: 1,
  sort: 0,
  startAt: null,
  endAt: null,
  description: ''
})

const form = reactive(defaultForm())

function formatAmount(value) {
  return (Number(value || 0) / 100).toFixed(2)
}

function resetForm() {
  Object.assign(form, defaultForm())
}

async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, size }
    if (statusFilter.value !== undefined && statusFilter.value !== null && statusFilter.value !== '') {
      params.status = statusFilter.value
    }
    const res = await getRechargeActivityList(params)
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

function openCreate() {
  editingId.value = 0
  resetForm()
  dialogVisible.value = true
}

function openEdit(row) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    rechargeAmountYuan: Number(row.recharge_amount || 0) / 100,
    giftAmountYuan: Number(row.gift_amount || 0) / 100,
    status: row.status,
    sort: row.sort,
    startAt: row.start_at ? new Date(row.start_at) : null,
    endAt: row.end_at ? new Date(row.end_at) : null,
    description: row.description || ''
  })
  dialogVisible.value = true
}

function buildPayload() {
  return {
    name: form.name.trim(),
    recharge_amount: Math.round(Number(form.rechargeAmountYuan || 0) * 100),
    gift_amount: Math.round(Number(form.giftAmountYuan || 0) * 100),
    status: Number(form.status),
    sort: Number(form.sort) || 0,
    start_at: form.startAt || null,
    end_at: form.endAt || null,
    description: form.description.trim()
  }
}

async function submitForm() {
  const payload = buildPayload()
  if (!payload.name) {
    ElMessage.error('请输入活动名称')
    return
  }
  if (!payload.recharge_amount || payload.recharge_amount <= 0) {
    ElMessage.error('请输入正确的充值金额')
    return
  }
  submitting.value = true
  try {
    if (editingId.value) {
      await updateRechargeActivity(editingId.value, payload)
      ElMessage.success('活动已更新')
    } else {
      await createRechargeActivity(payload)
      ElMessage.success('活动已创建')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

async function toggleStatus(row) {
  const targetStatus = row.status === 1 ? 0 : 1
  await ElMessageBox.confirm(`确认${targetStatus === 1 ? '上线' : '下线'}该充值活动？`, '操作确认', { type: 'warning' })
  await updateRechargeActivity(row.id, {
    name: row.name,
    recharge_amount: row.recharge_amount,
    gift_amount: row.gift_amount,
    status: targetStatus,
    sort: row.sort,
    start_at: row.start_at ? new Date(row.start_at) : null,
    end_at: row.end_at ? new Date(row.end_at) : null,
    description: row.description || ''
  })
  ElMessage.success(`活动已${targetStatus === 1 ? '上线' : '下线'}`)
  fetchList()
}

async function removeActivity(row) {
  if (row.status !== 0) {
    ElMessage.warning('仅允许删除已下线活动，请先下线')
    return
  }
  await ElMessageBox.confirm(
    `活动“${row.name}”删除后无法恢复。仅已下线且无关联充值订单的活动可以删除，确认继续吗？`,
    '高风险删除确认',
    {
      type: 'error',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消'
    }
  )
  await deleteRechargeActivity(row.id)
  ElMessage.success('活动已删除')
  if (list.value.length === 1 && page.value > 1) {
    page.value -= 1
  }
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-button--danger.is-plain) {
  font-weight: 600;
}
</style>