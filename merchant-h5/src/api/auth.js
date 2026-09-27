import { merchantRequest } from '../utils/request'

export function merchantWebLogin(data) {
  return merchantRequest.post('/web/login', data)
}

export function merchantWebRegister(data) {
  return merchantRequest.post('/web/register', data)
}

export function updateMerchantWebPassword(data) {
  return merchantRequest.put('/web/password', data)
}
