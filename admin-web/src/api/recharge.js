import request from '../utils/request'

export function getRechargeOrderList(params) {
  return request.get('/recharge-orders', { params })
}

export function getRechargeOrderDetail(id) {
  return request.get(`/recharge-order/${id}`)
}

export function cleanupRechargeOrderRemaining(id, data) {
  return request.post(`/recharge-order/${id}/cleanup-remaining`, data || {})
}

export function getRechargeActivityList(params) {
  return request.get('/recharge-activities', { params })
}

export function createRechargeActivity(data) {
  return request.post('/recharge-activity', data)
}

export function updateRechargeActivity(id, data) {
  return request.put(`/recharge-activity/${id}`, data)
}

export function deleteRechargeActivity(id) {
  return request.delete(`/recharge-activity/${id}`)
}