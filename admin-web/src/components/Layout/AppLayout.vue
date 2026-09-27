<template>
  <el-container class="app-layout">
    <el-aside :width="collapsed ? '64px' : '210px'" class="aside">
      <SideMenu :collapsed="collapsed" />
    </el-aside>
    <el-container class="content-container">
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="collapsed = !collapsed">
            <Fold v-if="!collapsed" />
            <Expand v-else />
          </el-icon>
          <span class="page-title">{{ route.meta.title }}</span>
        </div>
        <div class="header-right">
          <span class="nickname">{{ userStore.nickname }}</span>
          <el-button type="danger" text @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <div class="tags-bar">
        <el-tabs v-model="activeTabKey" type="card" class="tags-tabs" @tab-remove="handleTabRemove">
          <el-tab-pane
            v-for="tab in appViewStore.visitedTabs"
            :key="tab.key"
            :label="tab.title"
            :name="tab.key"
            :closable="tab.closable"
          />
        </el-tabs>
        <el-dropdown class="tags-actions" trigger="click" @command="handleTabCommand">
          <el-button class="tags-action-btn" text>
            标签操作
            <el-icon class="action-icon"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="refreshCurrent">刷新当前页</el-dropdown-item>
              <el-dropdown-item command="closeOthers" :disabled="!hasOtherClosableTabs">关闭其他</el-dropdown-item>
              <el-dropdown-item command="closeRight" :disabled="!hasRightClosableTabs">关闭右侧</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <el-main class="main">
        <router-view v-slot="{ Component, route: currentRoute }">
          <keep-alive :max="16">
            <component :is="Component" :key="getRouteRenderKey(currentRoute)" />
          </keep-alive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { resolveAppViewTabKey, useAppViewStore, useUserStore } from '../../store'
import SideMenu from './SideMenu.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appViewStore = useAppViewStore()
const collapsed = ref(false)

const activeTabKey = computed({
  get: () => resolveAppViewTabKey(route),
  set: value => {
    const targetTab = appViewStore.visitedTabs.find(item => item.key === value)
    if (targetTab && targetTab.fullPath !== route.fullPath) {
      router.push(targetTab.fullPath)
      return
    }
    if (value && value !== route.fullPath && value.startsWith('/')) {
      router.push(value)
    }
  }
})

const hasOtherClosableTabs = computed(() => appViewStore.visitedTabs.some(item => item.key !== activeTabKey.value && item.closable))

const hasRightClosableTabs = computed(() => {
  const activeIndex = appViewStore.visitedTabs.findIndex(item => item.key === activeTabKey.value)
  if (activeIndex < 0) {
    return false
  }
  return appViewStore.visitedTabs.slice(activeIndex + 1).some(item => item.closable)
})

watch(
  () => route.fullPath,
  () => {
    appViewStore.ensureTab(route)
  },
  { immediate: true }
)

function getRouteRenderKey(currentRoute) {
  const tabKey = resolveAppViewTabKey(currentRoute)
  const version = appViewStore.cacheVersions[tabKey] || 0
  return `${tabKey}::${currentRoute.fullPath}::${version}`
}

function handleTabRemove(targetKey) {
  const currentTabs = [...appViewStore.visitedTabs]
  const currentIndex = currentTabs.findIndex(item => item.key === targetKey)
  const wasActive = activeTabKey.value === targetKey
  appViewStore.removeTab(targetKey)
  if (!wasActive) {
    return
  }
  const fallback = appViewStore.visitedTabs[currentIndex] || appViewStore.visitedTabs[currentIndex - 1] || appViewStore.visitedTabs[0]
  if (fallback && fallback.fullPath !== route.fullPath) {
    router.push(fallback.fullPath)
  }
}

function handleTabCommand(command) {
  if (command === 'refreshCurrent') {
    appViewStore.refreshTab(activeTabKey.value)
    return
  }
  if (command === 'closeOthers') {
    appViewStore.closeOtherTabs(activeTabKey.value)
    return
  }
  if (command === 'closeRight') {
    appViewStore.closeRightTabs(activeTabKey.value)
  }
}

function handleLogout() {
  appViewStore.resetTabs()
  userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.app-layout {
  height: 100vh;
}

.aside {
  background-color: #304156;
  transition: width 0.3s;
  overflow: hidden;
}

.content-container {
  min-width: 0;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  padding: 0 20px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.collapse-btn {
  cursor: pointer;
  font-size: 20px;
}
.page-title {
  font-size: 16px;
  font-weight: 500;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.nickname {
  color: #606266;
}

.tags-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
}

.tags-tabs {
  flex: 1;
  min-width: 0;
}

.tags-tabs :deep(.el-tabs__header) {
  margin: 0;
}

.tags-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
}

.tags-tabs :deep(.el-tabs__item) {
  height: 34px;
  line-height: 34px;
}

.tags-actions {
  flex-shrink: 0;
}

.tags-action-btn {
  color: #606266;
}

.action-icon {
  margin-left: 4px;
}

.main {
  background: #f0f2f5;
  padding: 20px;
  overflow: auto;
}
</style>
