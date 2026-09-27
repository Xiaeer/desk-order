import { defineStore } from 'pinia'
import { getToken, setToken, removeToken } from '../utils/auth'
import { login } from '../api/auth'

const VIEW_TAB_STORAGE_KEY = 'admin_view_tabs'
const SINGLE_INSTANCE_ROUTE_NAMES = new Set([
  'MerchantDetail',
  'UserDetail',
  'ShopDetail',
  'OrderDetail',
  'RechargeOrderDetail'
])

function createDashboardTab() {
  return {
    key: '/dashboard',
    fullPath: '/dashboard',
    title: '数据概览',
    name: 'Dashboard',
    closable: false
  }
}

function canUseSessionStorage() {
  return typeof window !== 'undefined' && typeof window.sessionStorage !== 'undefined'
}

function normalizeVisitedTabs(tabs) {
  const dashboardTab = createDashboardTab()
  const normalized = []
  const seen = new Set()
  const candidates = Array.isArray(tabs) ? [dashboardTab, ...tabs] : [dashboardTab]

  for (const item of candidates) {
    if (!item) {
      continue
    }
    const fullPath = typeof item.fullPath === 'string' && item.fullPath ? item.fullPath : item.key
    const key = typeof item.key === 'string' && item.key ? item.key : fullPath
    if (!fullPath || !key || seen.has(key)) {
      continue
    }
    const tab = {
      key,
      fullPath,
      title: typeof item.title === 'string' && item.title ? item.title : '未命名页面',
      name: item.name || fullPath,
      closable: fullPath !== '/dashboard' && item.closable !== false
    }
    if (fullPath === '/dashboard') {
      normalized.push(dashboardTab)
      seen.add(dashboardTab.key)
      continue
    }
    normalized.push(tab)
    seen.add(tab.key)
  }

  return normalized.length > 0 ? normalized : [dashboardTab]
}

function loadVisitedTabs() {
  if (!canUseSessionStorage()) {
    return [createDashboardTab()]
  }
  const raw = window.sessionStorage.getItem(VIEW_TAB_STORAGE_KEY)
  if (!raw) {
    return [createDashboardTab()]
  }
  try {
    return normalizeVisitedTabs(JSON.parse(raw))
  } catch {
    return [createDashboardTab()]
  }
}

function persistVisitedTabs(tabs) {
  if (!canUseSessionStorage()) {
    return
  }
  const payload = normalizeVisitedTabs(tabs).map(item => ({
    key: item.key,
    fullPath: item.fullPath,
    title: item.title,
    name: item.name,
    closable: item.closable
  }))
  window.sessionStorage.setItem(VIEW_TAB_STORAGE_KEY, JSON.stringify(payload))
}

function buildTabTitle(route) {
  const baseTitle = route?.meta?.title || '未命名页面'
  const rawID = Array.isArray(route?.params?.id) ? route.params.id[0] : route?.params?.id
  if (!rawID) {
    return baseTitle
  }
  return `${baseTitle} #${rawID}`
}

export function resolveAppViewTabKey(route) {
  const routeName = typeof route?.name === 'string' ? route.name : ''
  if (routeName && SINGLE_INSTANCE_ROUTE_NAMES.has(routeName)) {
    return routeName
  }
  return route?.fullPath || route?.path || routeName
}

export const useUserStore = defineStore('user', {
  state: () => ({
    token: getToken() || '',
    nickname: localStorage.getItem('admin_nickname') || ''
  }),
  actions: {
    async loginAction(loginForm) {
      const res = await login(loginForm)
      this.token = res.data.token
      this.nickname = res.data.nickname
      setToken(res.data.token)
      localStorage.setItem('admin_nickname', res.data.nickname)
      return res
    },
    logout() {
      this.token = ''
      this.nickname = ''
      removeToken()
      localStorage.removeItem('admin_nickname')
    }
  }
})

export const useAppViewStore = defineStore('app-view', {
  state: () => ({
    visitedTabs: loadVisitedTabs(),
    cacheVersions: {}
  }),
  actions: {
    bumpCacheVersion(keys) {
      for (const key of keys) {
        if (!key) {
          continue
        }
        this.cacheVersions[key] = (this.cacheVersions[key] || 0) + 1
      }
    },
    ensureTab(route) {
      if (!route || route.path === '/login') {
        return
      }
      const tab = {
        key: resolveAppViewTabKey(route),
        fullPath: route.fullPath,
        title: buildTabTitle(route),
        name: route.name || route.fullPath,
        closable: route.path !== '/dashboard'
      }
      const existing = this.visitedTabs.find(item => item.key === tab.key)
      if (existing) {
        Object.assign(existing, tab)
        persistVisitedTabs(this.visitedTabs)
        return
      }
      this.visitedTabs.push(tab)
      persistVisitedTabs(this.visitedTabs)
    },
    removeTab(key) {
      const removedKeys = this.visitedTabs.filter(item => item.key === key && item.closable).map(item => item.key)
      if (removedKeys.length === 0) {
        return
      }
      this.visitedTabs = this.visitedTabs.filter(item => item.key !== key || !item.closable)
      this.bumpCacheVersion(removedKeys)
      persistVisitedTabs(this.visitedTabs)
    },
    closeOtherTabs(activeKey) {
      const removedKeys = this.visitedTabs.filter(item => item.key !== activeKey && item.closable).map(item => item.key)
      this.visitedTabs = this.visitedTabs.filter(item => item.key === activeKey || !item.closable)
      this.bumpCacheVersion(removedKeys)
      persistVisitedTabs(this.visitedTabs)
    },
    closeRightTabs(activeKey) {
      const activeIndex = this.visitedTabs.findIndex(item => item.key === activeKey)
      if (activeIndex < 0) {
        return
      }
      const removedKeys = this.visitedTabs.slice(activeIndex + 1).filter(item => item.closable).map(item => item.key)
      this.visitedTabs = this.visitedTabs.filter((item, index) => index <= activeIndex || !item.closable)
      this.bumpCacheVersion(removedKeys)
      persistVisitedTabs(this.visitedTabs)
    },
    refreshTab(key) {
      this.bumpCacheVersion([key])
    },
    resetTabs() {
      this.visitedTabs = [createDashboardTab()]
      this.cacheVersions = {}
      persistVisitedTabs(this.visitedTabs)
    }
  }
})
