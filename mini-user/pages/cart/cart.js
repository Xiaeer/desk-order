const cartUtils = require('../../utils/cart')

function normalizeCartItem(item) {
	return Object.assign({}, item, {
		price_text: (item.price / 100).toFixed(2),
		subtotal_text: (item.subtotal / 100).toFixed(2),
		option_summary: item.option_summary || '',
		cart_item_key: item.cart_item_key || ''
	})
}

Page({
	data: {
		shopName: '',
		items: [],
		totalCount: 0,
		totalAmount: 0
	},

	onShow() {
		this.loadCart()
	},

	loadCart() {
		const cart = cartUtils.getCartState()
		this.setData({
			shopName: cart.shopName,
			items: (cart.items || []).map(normalizeCartItem),
			totalCount: cartUtils.getTotalCount(cart),
			totalAmount: cartUtils.getTotalAmount(cart),
			totalAmountText: (cartUtils.getTotalAmount(cart) / 100).toFixed(2)
		})
	},

	increase(e) {
		const item = e.currentTarget.dataset.item
		cartUtils.updateItemQuantity(item.product_id, item.quantity + 1, item.cart_item_key)
		this.loadCart()
	},

	decrease(e) {
		const item = e.currentTarget.dataset.item
		cartUtils.updateItemQuantity(item.product_id, item.quantity - 1, item.cart_item_key)
		this.loadCart()
	},

	clearCart() {
		wx.showModal({
			title: '清空购物车',
			content: '确定清空当前店铺已选商品吗？',
			success: res => {
				if (res.confirm) {
					cartUtils.clearCart()
					this.loadCart()
				}
			}
		})
	},

	goConfirm() {
		if (!this.data.totalCount) {
			wx.showToast({ title: '购物车为空', icon: 'none' })
			return
		}
		wx.navigateTo({ url: '/pages/order-confirm/order-confirm' })
	}
})
