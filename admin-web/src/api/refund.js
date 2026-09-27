import request from '../utils/request'

export function getRechargeRefundRequestList(params) {
  return request.get('/recharge-refund-requests', { params })
}

export function getRechargeRefundRequestDetail(id) {
  return request.get(`/recharge-refund-request/${id}`)
}

export function createRechargeRefundRequest(data) {
  return request.post('/recharge-refund-request', data)
}

export function reviewRechargeRefundRequest(id, data) {
  return request.put(`/recharge-refund-request/${id}/review`, data)
}

export function syncRechargeRefundRequestStatus(id) {
  return request.put(`/recharge-refund-request/${id}/sync-status`)
}