import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '../utils/auth'

const routes = [
  {
    path: '/login',
    component: () => import('../views/LoginView.vue'),
    meta: { title: '商户登录', public: true }
  },
  {
    path: '/register',
    component: () => import('../views/RegisterView.vue'),
    meta: { title: '商户注册', public: true }
  },
  {
    path: '/audit-status',
    component: () => import('../views/AuditStatusView.vue'),
    meta: { title: '审核状态' }
  },
  {
    path: '/',
    component: () => import('../components/MobileLayout.vue'),
    children: [
      {
        path: '',
        redirect: '/home'
      },
      {
        path: '/home',
        component: () => import('../views/HomeView.vue'),
        meta: { title: '首页', tab: 'home' }
      },
      {
        path: '/orders',
        component: () => import('../views/OrderListView.vue'),
        meta: { title: '订单', tab: 'orders' }
      },
      {
        path: '/menu',
        component: () => import('../views/MenuManageView.vue'),
        meta: { title: '菜单', tab: 'menu', keepAlive: true }
      },
      {
        path: '/shop',
        component: () => import('../views/ShopInfoView.vue'),
        meta: { title: '店铺', tab: 'shop' }
      },
      {
        path: 'audit-shop',
        component: () => import('../views/ShopInfoView.vue'),
        meta: { title: '店铺信息', hideTab: true }
      },
      {
        path: '/password',
        component: () => import('../views/PasswordView.vue'),
        meta: { title: '设置密码', hideTab: true }
      },
      {
        path: '/order/:id',
        component: () => import('../views/OrderDetailView.vue'),
        meta: { title: '订单详情', hideTab: true }
      },
      {
        path: '/product-edit',
        component: () => import('../views/ProductEditView.vue'),
        meta: { title: '商品编辑', hideTab: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  document.title = `${to.meta.title || 'DeskOrder'} - 商户端`
  if (to.meta.public) {
    next()
    return
  }
  if (!getToken()) {
    next('/login')
    return
  }
  next()
})

export default router
