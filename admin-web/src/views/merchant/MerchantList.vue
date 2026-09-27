<template>
  <el-card>
    <el-table :data="list" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="商家名称" />
      <el-table-column prop="phone" label="手机号" />
      <el-table-column label="店铺">
        <template #default="{ row }">
          <el-tag v-if="row.has_shop" type="success">{{ row.shop_name }}</el-tag>
          <el-tag v-else type="info">未创建</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="注册时间" width="180" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button type="primary" link @click="$router.push(`/merchant/${row.id}`)">详情</el-button>
          <el-button type="danger" link @click="removeMerchant(row)">删除</el-button>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMerchantList, deleteMerchant } from '../../api/merchant'

const list = ref([])
const total = ref(0)
const page = ref(1)
const size = 20
const loading = ref(false)

async function fetchList() {
  loading.value = true
  try {
    const res = await getMerchantList({ page: page.value, size })
    list.value = res.data.list || []
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

async function removeMerchant(row) {
  const shopTip = row.has_shop ? '该商家关联的店铺也会一并删除。' : ''
  await ElMessageBox.confirm(`确认删除商家“${row.name || row.id}”？${shopTip}`, '删除确认', { type: 'warning' })
  await deleteMerchant(row.id)
  ElMessage.success('商家已删除')
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
