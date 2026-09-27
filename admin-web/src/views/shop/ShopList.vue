<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>店铺列表</span>
        <el-select v-model="statusFilter" placeholder="筛选状态" clearable style="width: 150px" @change="handleFilter">
          <el-option label="待审核" :value="0" />
          <el-option label="已通过" :value="1" />
          <el-option label="已拒绝" :value="2" />
        </el-select>
      </div>
    </template>
    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="店铺名称" />
      <el-table-column prop="merchant_name" label="所属商家" />
      <el-table-column prop="address" label="地址" show-overflow-tooltip />
      <el-table-column prop="phone" label="电话" width="130" />
      <el-table-column label="审核状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="营业状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_open ? 'success' : 'info'">{{ row.is_open ? '营业中' : '休息中' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="240">
        <template #default="{ row }">
          <el-button type="primary" link @click="$router.push(`/shop/${row.id}`)">详情</el-button>
          <el-button v-if="row.status === 0" type="success" link @click="audit(row.id, 1)">通过</el-button>
          <el-button v-if="row.status === 0" type="danger" link @click="audit(row.id, 2)">拒绝</el-button>
          <el-button type="danger" link @click="removeShop(row.id)">删除</el-button>
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
import { ref, onMounted } from 'vue'
import { getShopList, auditShop, deleteShop } from '../../api/shop'
import { ElMessage, ElMessageBox } from 'element-plus'

const list = ref([])
const total = ref(0)
const page = ref(1)
const size = 20
const loading = ref(false)
const statusFilter = ref(undefined)

function statusText(s) {
  return ['待审核', '已通过', '已拒绝'][s] || '未知'
}
function statusType(s) {
  return ['warning', 'success', 'danger'][s] || 'info'
}

async function fetchList() {
  loading.value = true
  try {
    const params = { page: page.value, size }
    if (statusFilter.value !== undefined && statusFilter.value !== null && statusFilter.value !== '') {
      params.status = statusFilter.value
    }
    const res = await getShopList(params)
    list.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

function handleFilter() {
  page.value = 1
  fetchList()
}

async function audit(id, status) {
  const action = status === 1 ? '通过' : '拒绝'
  await ElMessageBox.confirm(`确认${action}该店铺？`, '审核确认', { type: 'warning' })
  await auditShop(id, { status })
  ElMessage.success(`已${action}`)
  fetchList()
}

async function removeShop(id) {
  await ElMessageBox.confirm('确认删除该店铺？删除后商户可重新发起入驻申请。若店铺已有关联订单，将禁止删除。', '删除确认', { type: 'warning' })
  await deleteShop(id)
  ElMessage.success('店铺已删除')
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
