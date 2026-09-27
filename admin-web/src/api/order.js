import request from '../utils/request'

export function getOrderList(params) {
  return request.get('/orders', { params })
}

export function getOrderDetail(id) {
  return request.get(`/order/${id}`)
}

export function getDashboard() {
  return request.get('/dashboard')
}
