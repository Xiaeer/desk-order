const TOKEN_KEY = 'merchant_h5_token'
const HAS_SHOP_KEY = 'merchant_h5_has_shop'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function getHasShop() {
  return localStorage.getItem(HAS_SHOP_KEY) === 'true'
}

export function setHasShop(hasShop) {
  localStorage.setItem(HAS_SHOP_KEY, hasShop ? 'true' : 'false')
}

export function clearAuth() {
  clearToken()
  localStorage.removeItem(HAS_SHOP_KEY)
}
