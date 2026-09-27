<template>
  <div class="mobile-layout">
    <header class="mobile-header">
      <button v-if="showBack" class="header-back" @click="router.back()">返回</button>
      <span>{{ route.meta.title || 'DeskOrder' }}</span>
      <span class="header-placeholder"></span>
    </header>

    <main class="mobile-main" :class="{ 'with-tabbar': !route.meta.hideTab }">
      <router-view v-slot="{ Component, route: currentRoute }">
        <keep-alive>
          <component :is="Component" v-if="currentRoute.meta.keepAlive" />
        </keep-alive>
        <component :is="Component" v-if="!currentRoute.meta.keepAlive" />
      </router-view>
    </main>

    <nav v-if="!route.meta.hideTab" class="mobile-tabbar">
      <button
        v-for="item in tabs"
        :key="item.path"
        class="tabbar-item"
        :class="{ active: route.path === item.path }"
        @click="router.push(item.path)"
      >
        {{ item.label }}
      </button>
    </nav>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const tabs = [
  { path: '/home', label: '首页' },
  { path: '/orders', label: '订单' },
  { path: '/menu', label: '菜单' },
  { path: '/shop', label: '店铺' }
]

const showBack = computed(() => route.meta.hideTab)
</script>
