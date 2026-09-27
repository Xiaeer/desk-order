Component({
	properties: {
		count: {
			type: Number,
			value: 0
		},
		amountText: {
			type: String,
			value: '0.00'
		},
		checkoutText: {
			type: String,
			value: '去结算'
		},
		labelText: {
			type: String,
			value: '已选商品'
		}
	},
	methods: {
		handleCartTap() {
			this.triggerEvent('carttap')
		},
		handleCheckoutTap() {
			if (!this.properties.count) {
				return
			}
			this.triggerEvent('checkouttap')
		}
	}
})
