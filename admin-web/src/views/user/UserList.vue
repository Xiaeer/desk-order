<template>
  <el-card>
    <div class="toolbar">
      <el-input
        v-model="keyword"
        class="keyword-input"
        clearable
        placeholder="用户ID / 昵称 / 手机号 / OpenID尾号"
        @clear="handleReset"
      />
      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>

    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column label="用户信息" min-width="220">
        <template #default="{ row }">
          <div class="user-cell">
            <el-avatar :src="row.avatar">{{ getDisplayName(row).slice(0, 1) }}</el-avatar>
            <div>
              <div class="user-name-row">
                <span class="user-name">{{ getDisplayName(row) }}</span>
                <el-tag size="small" effect="plain">ID {{ row.id }}</el-tag>
              </div>
              <div class="user-phone">{{ getIdentityLabel(row) }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="当前余额" width="120">
        <template #default="{ row }">¥{{ formatAmount(row.balance_amount) }}</template>
      </el-table-column>
      <el-table-column prop="order_count" label="订单数" width="100" />
      <el-table-column prop="recharge_order_count" label="充值单数" width="110" />
      <el-table-column prop="created_at" label="注册时间" width="180" />
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button type="primary" link @click="$router.push(`/user/${row.id}`)">详情</el-button>
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
import { onMounted, ref } from 'vue'
import { getUserList } from '../../api/user'

const list = ref([])
const total = ref(0)
const page = ref(1)
const size = 20
const loading = ref(false)
const keyword = ref('')

function formatAmount(value) {
  return (Number(value || 0) / 100).toFixed(2)
}

function getDisplayName(row) {
  return row.display_name || row.nickname || `用户#${row.id}`
}

function getIdentityLabel(row) {
  return row.identity_label || row.phone || `用户ID ${row.id}`
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getUserList({ page: page.value, size, keyword: keyword.value.trim() })
    list.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchList()
}

function handleReset() {
  if (!keyword.value && page.value === 1) {
    fetchList()
    return
  }
  keyword.value = ''
  page.value = 1
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.keyword-input {
  width: 320px;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-name {
  font-weight: 600;
  color: #303133;
}

.user-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-phone {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>