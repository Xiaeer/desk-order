import { merchantRequest, userRequest } from '../utils/request'

export function getShopMenu(shopId) {
  return userRequest.get(`/shop/${shopId}/menu`)
}

export function createCategory(data) {
  return merchantRequest.post('/category', data)
}

export function updateCategory(id, data) {
  return merchantRequest.put(`/category/${id}`, data)
}

export function deleteCategory(id) {
  return merchantRequest.delete(`/category/${id}`)
}

export function createProduct(data) {
  return merchantRequest.post('/product', data)
}

export function uploadProductImage(file) {
  const formData = new FormData()
  formData.append('file', file)
  return merchantRequest.post('/upload/product-image', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function updateProduct(id, data) {
  return merchantRequest.put(`/product/${id}`, data)
}

export function deleteProduct(id) {
  return merchantRequest.delete(`/product/${id}`)
}
