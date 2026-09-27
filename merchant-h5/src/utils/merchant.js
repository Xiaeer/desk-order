import { getHasShop, setHasShop } from './auth'
import { getShop } from '../api/shop'
import { isShopNotFoundError } from './request'

export async function resolveMerchantEntry(router, hasShopOverride) {
  const hasShop = typeof hasShopOverride === 'boolean' ? hasShopOverride : getHasShop()
  if (!hasShop) {
    router.replace('/register')
    return
  }

  try {
    const response = await getShop()
    const shop = response.data
    setHasShop(true)
    if (shop.status === 1) {
      router.replace('/home')
      return
    }
    router.replace(`/audit-status?status=${shop.status || 0}`)
  } catch (error) {
    if (isShopNotFoundError(error)) {
      setHasShop(false)
      router.replace('/register')
      return
    }
    throw error
  }
}
