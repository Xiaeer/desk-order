<template>
  <el-menu
    :default-active="activeMenu"
    :collapse="collapsed"
    router
    background-color="#304156"
    text-color="#bfcbd9"
    active-text-color="#409EFF"
    class="side-menu"
  >
    <div class="logo">
      <span v-show="!collapsed">DeskOrder</span>
      <span v-show="collapsed">DO</span>
    </div>
    <el-menu-item index="/dashboard">
      <el-icon><DataAnalysis /></el-icon>
      <template #title>数据概览</template>
    </el-menu-item>
    <el-menu-item index="/merchants">
      <el-icon><User /></el-icon>
      <template #title>商家管理</template>
    </el-menu-item>
    <el-menu-item index="/users">
      <el-icon><UserFilled /></el-icon>
      <template #title>用户管理</template>
    </el-menu-item>
    <el-menu-item index="/shops">
      <el-icon><Shop /></el-icon>
      <template #title>店铺管理</template>
    </el-menu-item>
    <el-menu-item index="/orders">
      <el-icon><List /></el-icon>
      <template #title>订单管理</template>
    </el-menu-item>
    <el-menu-item index="/recharge-activities">
      <el-icon><CreditCard /></el-icon>
      <template #title>充值活动</template>
    </el-menu-item>
    <el-menu-item index="/recharge-orders">
      <el-icon><CreditCard /></el-icon>
      <template #title>充值订单</template>
    </el-menu-item>
    <el-menu-item index="/recharge-refunds">
      <el-icon><Money /></el-icon>
      <template #title>充值退款</template>
    </el-menu-item>
    <el-menu-item index="/settings">
      <el-icon><Setting /></el-icon>
      <template #title>系统设置</template>
    </el-menu-item>
  </el-menu>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

defineProps({
  collapsed: { type: Boolean, default: false }
})

const route = useRoute()
const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/merchant')) return '/merchants'
  if (path.startsWith('/user')) return '/users'
  if (path.startsWith('/shop')) return '/shops'
  if (path.startsWith('/order')) return '/orders'
  if (path.startsWith('/recharge-order')) return '/recharge-orders'
  if (path.startsWith('/recharge-refund')) return '/recharge-refunds'
  if (path.startsWith('/recharge-activity')) return '/recharge-activities'
  if (path.startsWith('/settings')) return '/settings'
  return path
})
</script>

<style scoped>
.side-menu {
  height: 100%;
  border-right: none;
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.28) transparent;
}

.side-menu::-webkit-scrollbar {
  width: 6px;
}

.side-menu::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.28);
  border-radius: 999px;
}

.side-menu::-webkit-scrollbar-track {
  background: transparent;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  border-bottom: 1px solid #3a4a5f;
}
</style>
