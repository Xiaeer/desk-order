const { routeMerchantEntry } = require('../../utils/navigation')

Page({
  onLoad() {
    routeMerchantEntry().catch(() => {})
  }
})