import request from '../utils/request'

export function getShopList(params) {
  return request.get('/shops', { params })
}

export function getShopDetail(id) {
  return request.get(`/shop/${id}`)
}

export function auditShop(id, data) {
  return request.put(`/shop/${id}/audit`, data)
}

export function getShopAuditRecords(id) {
  return request.get(`/shop/${id}/audit-records`)
}

export function deleteShop(id) {
  return request.delete(`/shop/${id}`)
}
