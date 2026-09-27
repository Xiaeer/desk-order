import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '../utils/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/login/LoginView.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/',
    component: () => import('../components/Layout/AppLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/dashboard/DashboardView.vue'),
        meta: { title: '数据概览' }
      },
      {
        path: 'merchants',
        name: 'MerchantList',
        component: () => import('../views/merchant/MerchantList.vue'),
        meta: { title: '商家列表' }
      },
      {
        path: 'users',
        name: 'UserList',
        component: () => import('../views/user/UserList.vue'),
        meta: { title: '用户列表' }
      },
      {
        path: 'merchant/:id',
        name: 'MerchantDetail',
        component: () => import('../views/merchant/MerchantDetail.vue'),
        meta: { title: '商家详情' }
      },
      {
        path: 'user/:id',
        name: 'UserDetail',
        component: () => import('../views/user/UserDetail.vue'),
        meta: { title: '用户详情' }
      },
      {
        path: 'shops',
        name: 'ShopList',
        component: () => import('../views/shop/ShopList.vue'),
        meta: { title: '店铺列表' }
      },
      {
        path: 'shop/:id',
        name: 'ShopDetail',
        component: () => import('../views/shop/ShopDetail.vue'),
        meta: { title: '店铺详情' }
      },
      {
        path: 'orders',
        name: 'OrderList',
        component: () => import('../views/order/OrderList.vue'),
        meta: { title: '订单列表' }
      },
      {
		path: 'settings',
		name: 'SystemConfigView',
		component: () => import('../views/settings/SystemConfigView.vue'),
		meta: { title: '系统设置' }
	  },
      {
    path: 'recharge-orders',
    name: 'RechargeOrderList',
    component: () => import('../views/recharge/RechargeOrderList.vue'),
    meta: { title: '充值订单' }
    },
      {
    path: 'recharge-refunds',
    name: 'RechargeRefundRequestList',
    component: () => import('../views/refund/RechargeRefundRequestList.vue'),
    meta: { title: '充值退款' }
    },
      {
		path: 'recharge-activities',
		name: 'RechargeActivityList',
		component: () => import('../views/recharge/RechargeActivityList.vue'),
		meta: { title: '充值活动' }
	  },
      {
    path: 'recharge-order/:id',
    name: 'RechargeOrderDetail',
    component: () => import('../views/recharge/RechargeOrderDetail.vue'),
    meta: { title: '充值订单详情' }
    },
      {
        path: 'order/:id',
        name: 'OrderDetail',
        component: () => import('../views/order/OrderDetail.vue'),
        meta: { title: '订单详情' }
      }
    ]
  }
]


const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  document.title = (to.meta.title || 'DeskOrder') + ' - 后台管理'
  if (to.path === '/login') {
    next()
  } else {
    const token = getToken()
    if (!token) {
      next('/login')
    } else {
      next()
    }
  }
})

export default router
