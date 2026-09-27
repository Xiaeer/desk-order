const CART_KEY = 'user_cart'

function normalizeSelectedOptions(selectedOptions) {
  if (!Array.isArray(selectedOptions)) {
    return []
  }
  return selectedOptions
    .map(selection => {
      const groupId = String(selection.group_id || selection.groupId || '').trim()
      const valueIDs = Array.isArray(selection.value_ids || selection.valueIds)
        ? Array.from(new Set((selection.value_ids || selection.valueIds)
          .map(valueId => String(valueId || '').trim())
          .filter(Boolean)))
        : []
      return {
        group_id: groupId,
        value_ids: valueIDs
      }
    })
    .filter(selection => selection.group_id && selection.value_ids.length)
    .sort((left, right) => left.group_id.localeCompare(right.group_id))
}

function buildOptionSignature(selectedOptions) {
  const normalizedSelections = normalizeSelectedOptions(selectedOptions)
  if (!normalizedSelections.length) {
    return ''
  }
  return normalizedSelections
    .map(selection => `${selection.group_id}:${selection.value_ids.slice().sort().join(',')}`)
    .join('|')
}

function buildCartItemKey(productId, optionSignature) {
  return optionSignature ? `${productId}::${optionSignature}` : String(productId)
}

function normalizeCartItem(item) {
  const productId = Number(item.product_id || item.productId || 0)
  const quantity = Math.max(0, Number(item.quantity || 0))
  const price = Math.max(0, Number(item.price || 0))
  const selectedOptions = normalizeSelectedOptions(item.selected_options || item.selectedOptions)
  const optionSignature = String(item.option_signature || buildOptionSignature(selectedOptions))
  const cartItemKey = String(item.cart_item_key || buildCartItemKey(productId, optionSignature))
  return {
    product_id: productId,
    cart_item_key: cartItemKey,
    option_signature: optionSignature,
    selected_options: selectedOptions,
    option_summary: String(item.option_summary || item.optionSummary || '').trim(),
    name: item.name,
    image: item.image,
    price,
    quantity,
    subtotal: price * quantity
  }
}

function getCartState() {
	const rawCart = wx.getStorageSync(CART_KEY) || {}
	return {
		shopId: Number(rawCart.shopId || 0),
		shopName: rawCart.shopName || '',
		items: Array.isArray(rawCart.items) ? rawCart.items.map(normalizeCartItem).filter(item => item.quantity > 0) : []
	}
}

function saveCartState(cart) {
  wx.setStorageSync(CART_KEY, cart)
}

function clearCart() {
  saveCartState({
    shopId: 0,
    shopName: '',
    items: []
  })
}

function ensureShopCart(shopId, shopName) {
  const cart = getCartState()
  if (cart.shopId && cart.shopId !== shopId) {
    clearCart()
  }
  const nextCart = {
    shopId,
    shopName: shopName || cart.shopName || '',
    items: cart.shopId === shopId ? cart.items : []
  }
  saveCartState(nextCart)
  return nextCart
}

function addItem(shopId, shopName, product) {
  const extra = arguments[3] || {}
  const cart = ensureShopCart(shopId, shopName)
  const items = cart.items.slice()
  const selectedOptions = normalizeSelectedOptions(extra.selectedOptions || extra.selected_options)
  const optionSignature = String(extra.optionSignature || buildOptionSignature(selectedOptions))
  const cartItemKey = buildCartItemKey(product.id, optionSignature)
  const itemPrice = Math.max(0, Number(extra.price != null ? extra.price : product.price) || 0)
  const optionSummary = String(extra.optionSummary || extra.option_summary || '').trim()
  const index = items.findIndex(item => item.cart_item_key === cartItemKey)
  if (index >= 0) {
    items[index].price = itemPrice
    items[index].option_summary = optionSummary
    items[index].selected_options = selectedOptions
    items[index].option_signature = optionSignature
    items[index].quantity += 1
    items[index].subtotal = items[index].price * items[index].quantity
  } else {
    items.push({
      cart_item_key: cartItemKey,
      option_signature: optionSignature,
      selected_options: selectedOptions,
      option_summary: optionSummary,
      product_id: product.id,
      name: product.name,
      image: product.image,
      price: itemPrice,
      quantity: 1,
      subtotal: itemPrice
    })
  }
  const nextCart = {
    shopId,
    shopName,
    items
  }
  saveCartState(nextCart)
  return nextCart
}

function updateItemQuantity(productId, quantity, cartItemKey) {
  const cart = getCartState()
  const items = cart.items
    .map(item => {
      const matches = cartItemKey
        ? item.cart_item_key === cartItemKey
        : (item.product_id === productId && !item.option_signature)
      if (!matches) {
        return item
      }
      const nextQuantity = Math.max(0, quantity)
      return {
        cart_item_key: item.cart_item_key,
        option_signature: item.option_signature,
        selected_options: item.selected_options,
        option_summary: item.option_summary,
        product_id: item.product_id,
        name: item.name,
        image: item.image,
        price: item.price,
        quantity: nextQuantity,
        subtotal: item.price * nextQuantity
      }
    })
    .filter(item => item.quantity > 0)
  const nextCart = {
    shopId: cart.shopId,
    shopName: cart.shopName,
    items
  }
  saveCartState(nextCart)
  return nextCart
}

function getTotalCount(cart) {
  return (cart.items || []).reduce((sum, item) => sum + item.quantity, 0)
}

function getTotalAmount(cart) {
  return (cart.items || []).reduce((sum, item) => sum + item.subtotal, 0)
}

module.exports = {
  getCartState,
  saveCartState,
  clearCart,
  ensureShopCart,
  addItem,
  updateItemQuantity,
  getTotalCount,
  getTotalAmount,
  buildOptionSignature
}
