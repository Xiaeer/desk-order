import { merchantRequest } from '../utils/request'

export function createShop(data) {
  return merchantRequest.post('/shop', data)
}

export function getShop() {
  return merchantRequest.get('/shop')
}

export function updateShop(data) {
  return merchantRequest.put('/shop', data)
}

export function rotateShopPOSBindToken() {
  return merchantRequest.post('/shop/pos-bind-token/rotate')
}

export function toggleShopOpen(isOpen) {
  return merchantRequest.put('/shop/status', { is_open: isOpen })
}

export function toggleShopAutoAccept(autoAcceptOrders) {
  return merchantRequest.put('/shop/auto-accept', { auto_accept_orders: autoAcceptOrders })
}

export function listShopTables() {
  return merchantRequest.get('/shop/tables')
}

export function createShopTable(data) {
  return merchantRequest.post('/shop/table', data)
}

export function updateShopTable(id, data) {
  return merchantRequest.put(`/shop/table/${id}`, data)
}

export function deleteShopTable(id) {
  return merchantRequest.delete(`/shop/table/${id}`)
}
