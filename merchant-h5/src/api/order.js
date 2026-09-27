import { merchantRequest } from '../utils/request'

export function getOrders(params) {
  return merchantRequest.get('/orders', { params })
}

export function getOrderDetail(id) {
  return merchantRequest.get(`/order/${id}`)
}

export function acceptOrder(id) {
  return merchantRequest.put(`/order/${id}/accept`)
}

export function completeOrder(id) {
  return merchantRequest.put(`/order/${id}/complete`)
}
