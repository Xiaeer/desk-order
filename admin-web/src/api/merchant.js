import request from '../utils/request'

export function getMerchantList(params) {
  return request.get('/merchants', { params })
}

export function getMerchantDetail(id) {
  return request.get(`/merchant/${id}`)
}

export function deleteMerchant(id) {
  return request.delete(`/merchant/${id}`)
}
