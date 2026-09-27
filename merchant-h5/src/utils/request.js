import axios from 'axios'
import { clearAuth, getToken } from './auth'

const apiOrigin = (import.meta.env.VITE_API_ORIGIN || '').trim().replace(/\/$/, '')
const AUTH_EXPIRED_NOTICE_KEY = 'merchant_h5_auth_expired_notice'

function buildApiBase(path) {
  return apiOrigin ? `${apiOrigin}${path}` : path
}

function createApi(baseURL) {
  const instance = axios.create({
    baseURL,
    timeout: 10000
  })

  instance.interceptors.request.use(config => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })

  instance.interceptors.response.use(
    response => {
      const result = response.data || {}
      if (result.code === 0) {
        return result
      }
      return Promise.reject(result)
    },
    error => {
      if (error.response?.status === 401) {
        const expiredMessage = error.response?.data?.msg || '登录已失效，请重新登录'
        clearAuth()
        sessionStorage.setItem(AUTH_EXPIRED_NOTICE_KEY, expiredMessage)
        window.location.href = '/login'
        return Promise.reject({ msg: expiredMessage, code: 401 })
      }
      return Promise.reject({ msg: error.message || '网络错误' })
    }
  )

  return instance
}

export const merchantRequest = createApi(buildApiBase('/api/v1/merchant'))
export const userRequest = createApi(buildApiBase('/api/v1/user'))

export function getApiErrorMessage(error, fallbackMessage) {
  if (!error) {
    return fallbackMessage
  }
  if (typeof error === 'string') {
    return error || fallbackMessage
  }
  if (error.msg) {
    return error.msg
  }
  if (error.message) {
    return error.message
  }
  return fallbackMessage
}

export function isShopNotFoundError(error) {
  return Boolean(error && error.msg === '店铺不存在')
}
