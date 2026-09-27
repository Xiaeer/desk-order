import request from '../utils/request'

export function getUserList(params) {
  return request.get('/users', { params })
}

export function getUserDetail(id) {
  return request.get(`/user/${id}`)
}

export function cleanupUserHistoricalBalance(id, data) {
  return request.post(`/user/${id}/cleanup-historical-balance`, data)
}