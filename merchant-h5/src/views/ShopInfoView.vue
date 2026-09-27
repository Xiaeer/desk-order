<template>
  <section class="page-stack">
    <div class="card page-stack" v-if="shop">
      <div class="row-between">
        <div>
          <div class="page-title">{{ pageTitle }}</div>
          <div class="page-subtitle">{{ pageSubtitle }}</div>
        </div>
        <span class="chip">{{ headerChipText }}</span>
      </div>
      <div class="field-grid">
        <div class="field">
          <label>店铺名称</label>
          <input v-model.trim="form.name" :disabled="!editing" />
        </div>
        <div class="field">
          <label>店铺地址</label>
          <input v-model.trim="form.address" :disabled="!editing" />
        </div>
        <div class="field">
          <label>联系电话</label>
          <input v-model.trim="form.phone" :disabled="!editing" />
        </div>
        <div class="field">
          <label>店铺描述</label>
          <textarea v-model.trim="form.description" rows="3" :disabled="!editing"></textarea>
        </div>
        <div class="field-grid" style="grid-template-columns: 1fr 1fr;">
          <div class="field">
            <label>纬度</label>
            <input v-model.number="form.latitude" :disabled="!editing" placeholder="纬度" />
          </div>
          <div class="field">
            <label>经度</label>
            <input v-model.number="form.longitude" :disabled="!editing" placeholder="经度" />
          </div>
        </div>
        <button v-if="editing" class="secondary-btn" :disabled="locating" @click="fillCurrentLocation">
          {{ locating ? '定位中...' : '获取当前位置' }}
        </button>
      </div>
      <div class="row-between">
        <button v-if="canToggleOpen" class="secondary-btn" @click="toggleOpen">{{ shop.is_open ? '暂停营业' : '开始营业' }}</button>
        <button v-if="editing" class="primary-btn" @click="saveShop">{{ submitButtonText }}</button>
        <button v-else-if="canEdit" class="primary-btn" @click="startEdit">{{ editButtonText }}</button>
      </div>
      <button class="secondary-btn" @click="router.push('/password')">设置 / 重置 H5 密码</button>
      <button class="danger-btn" @click="logout">退出登录</button>
    </div>

    <div class="card page-stack" v-if="shop">
      <div class="row-between">
        <div>
          <div class="page-title section-title">POS 绑定</div>
          <div class="page-subtitle">把 shop_id 和 POS 密钥填到 Windows POS 终端。POS 现在不能再只靠 shop_id 连接。</div>
        </div>
        <span class="chip">安全绑定</span>
      </div>
      <div class="pos-binding-grid">
        <div class="pos-binding-item">
          <label>shop_id</label>
          <div class="pos-binding-value">{{ shop.id || '-' }}</div>
          <div class="page-subtitle">POS 配置里的门店 ID</div>
        </div>
        <div class="pos-binding-item">
          <label>POS 密钥</label>
          <div class="pos-binding-value token-value">{{ shop.pos_bind_token || '-' }}</div>
          <div class="page-subtitle">POS 登录和推单 WebSocket 都会校验这个密钥</div>
        </div>
      </div>
      <div class="pos-actions">
        <button class="secondary-btn" @click="copyText(String(shop.id || ''), 'shop_id')">复制 shop_id</button>
        <button class="secondary-btn" @click="copyText(shop.pos_bind_token, 'POS 密钥')">复制 POS 密钥</button>
        <button class="danger-btn" :disabled="rotatingPosToken" @click="rotatePosBindTokenAction">
          {{ rotatingPosToken ? '重置中...' : '重置 POS 密钥' }}
        </button>
      </div>
    </div>

    <div class="card page-stack" v-if="shop">
      <div class="row-between">
        <div>
          <div class="page-title section-title">自动接单</div>
          <div class="page-subtitle">用户完成支付后可自动进入已接单，减少商家手动确认操作。</div>
        </div>
        <span class="chip">{{ shop.auto_accept_orders ? '已开启' : '已关闭' }}</span>
      </div>
      <div class="row-between operation-row">
        <div class="setting-value">
          {{ canManageOperationalSettings ? (shop.auto_accept_orders ? '当前支付成功后将自动接单' : '当前仍需商家手动接单') : '店铺审核通过后才可开启自动接单' }}
        </div>
        <button class="secondary-btn" :disabled="!canManageOperationalSettings || savingAutoAccept" @click="toggleAutoAccept">
          {{ savingAutoAccept ? '保存中...' : (shop.auto_accept_orders ? '关闭自动接单' : '开启自动接单') }}
        </button>
      </div>
    </div>

    <div class="card page-stack" v-if="shop">
      <div class="row-between">
        <div>
          <div class="page-title section-title">桌号管理</div>
          <div class="page-subtitle">给每张桌生成独立桌码，用户扫码可直接带桌号进店下单。</div>
        </div>
        <span class="chip">{{ tables.length }} 个桌号</span>
      </div>
      <div class="field-grid table-form-grid">
        <div class="field">
          <label>桌号</label>
          <input v-model.trim="tableDraft.tableNo" placeholder="例如：01、A1、包厢2" />
        </div>
        <div class="field">
          <label>排序</label>
          <input v-model.number="tableDraft.sort" type="number" placeholder="0" />
        </div>
        <button class="primary-btn table-create-btn" :disabled="creatingTable" @click="createTable">
          {{ creatingTable ? '创建中...' : '新增桌号' }}
        </button>
      </div>

      <div v-if="loadingTables" class="page-subtitle">桌号加载中...</div>
      <div v-else-if="tables.length" class="table-list">
        <div v-for="table in tables" :key="table.id" class="table-item">
          <div class="row-between">
            <div>
              <div class="table-title">桌号 {{ table.table_no }}</div>
              <div class="table-meta">scene：{{ table.scene_token }} · 排序 {{ table.sort || 0 }}</div>
            </div>
            <span class="chip">{{ table.status === 0 ? '启用中' : '已停用' }}</span>
          </div>
          <div class="table-qr-row">
            <img v-if="table.qrcode_url" :src="table.qrcode_url" class="table-qr" alt="桌码二维码" />
            <div class="table-actions">
              <a class="secondary-link" :href="table.qrcode_url" target="_blank" rel="noreferrer">查看二维码</a>
              <button class="secondary-btn" @click="copyQRCode(table.qrcode_url)">复制二维码链接</button>
              <button class="secondary-btn" :disabled="updatingTableId === table.id" @click="toggleTableStatus(table)">
                {{ updatingTableId === table.id ? '保存中...' : (table.status === 0 ? '停用桌号' : '启用桌号') }}
              </button>
              <button class="danger-btn" :disabled="deletingTableId === table.id" @click="deleteTableItem(table)">
                {{ deletingTableId === table.id ? '删除中...' : '删除桌号' }}
              </button>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="page-subtitle">暂无桌号。创建后即可打印二维码贴到桌面。</div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { clearAuth } from '../utils/auth'
import { createShopTable, deleteShopTable, getShop, listShopTables, rotateShopPOSBindToken, toggleShopAutoAccept, toggleShopOpen, updateShop, updateShopTable } from '../api/shop'
import { wgs84ToGcj02 } from '../utils/coordTransform'
import { getApiErrorMessage } from '../utils/request'

const route = useRoute()
const router = useRouter()
const shop = ref(null)
const tables = ref([])
const editing = ref(false)
const locating = ref(false)
const loadingTables = ref(false)
const creatingTable = ref(false)
const savingAutoAccept = ref(false)
const rotatingPosToken = ref(false)
const updatingTableId = ref(0)
const deletingTableId = ref(0)
const form = reactive({
  name: '',
  address: '',
  phone: '',
  description: '',
  latitude: '',
  longitude: ''
})
const tableDraft = reactive({
  tableNo: '',
  sort: 0
})
const canEdit = computed(() => [1, 2].includes(Number(shop.value?.status)))
const canToggleOpen = computed(() => Number(shop.value?.status) === 1)
const canManageOperationalSettings = computed(() => Number(shop.value?.status) === 1)
const isRejected = computed(() => Number(shop.value?.status) === 2)
const isResubmitMode = computed(() => route.query.mode === 'resubmit' && isRejected.value)
const headerChipText = computed(() => {
  if (!shop.value) {
    return '-'
  }
  return canToggleOpen.value ? (shop.value.is_open ? '营业中' : '已打烊') : statusText(shop.value.status)
})
const pageTitle = computed(() => (isResubmitMode.value ? '重新提交入驻' : '店铺信息'))
const pageSubtitle = computed(() => {
  if (isResubmitMode.value) {
    return '审核未通过，请核对店铺资料和当前位置后重新提交审核。'
  }
  return `审核状态：${statusText(shop.value?.status)}`
})
const editButtonText = computed(() => (isRejected.value ? '修改并重新提交' : '编辑信息'))
const submitButtonText = computed(() => (isRejected.value ? '重新提交审核' : '保存信息'))

function statusText(status) {
  return ['待审核', '已通过', '已拒绝'][status] || '未知'
}

function hasCoordinateValue(value) {
  return value !== '' && value !== null && value !== undefined && !Number.isNaN(Number(value))
}

function getCurrentPosition(options) {
  return new Promise((resolve, reject) => {
    navigator.geolocation.getCurrentPosition(resolve, reject, options)
  })
}

function getLocationErrorMessage(error) {
  if (!error) {
    return '定位失败，请稍后重试或手动填写经纬度'
  }
  if (error.code === 1) {
    return '浏览器已拒绝定位，请在系统或浏览器设置中允许定位权限'
  }
  if (error.code === 2) {
    return '定位服务暂时不可用，请确认手机系统定位服务已开启'
  }
  if (error.code === 3) {
    return '定位超时，请在开阔区域重试；若仍失败，请手动填写经纬度'
  }
  return error.message || '定位失败，请稍后重试或手动填写经纬度'
}

function getLocationAccuracy(position) {
  const accuracy = position?.coords?.accuracy
  return typeof accuracy === 'number' && accuracy >= 0 ? accuracy : null
}

function getLocationQualityMessage(position, usedFallback) {
  const accuracy = getLocationAccuracy(position)
  if (!usedFallback && (accuracy === null || accuracy <= 100)) {
    return ''
  }
  if (accuracy !== null) {
    return `当前定位精度约 ${accuracy.toFixed(0)} 米，可能为模糊位置。建议到室外或靠窗重试；如店铺位置确认无误，也可继续使用当前坐标。`
  }
  return '当前定位可能已回退到模糊位置。建议到室外或靠窗重试；如店铺位置确认无误，也可继续使用当前坐标。'
}

function syncForm() {
  form.name = shop.value?.name || ''
  form.address = shop.value?.address || ''
  form.phone = shop.value?.phone || ''
  form.description = shop.value?.description || ''
  form.latitude = hasCoordinateValue(shop.value?.latitude) ? Number(shop.value.latitude) : ''
  form.longitude = hasCoordinateValue(shop.value?.longitude) ? Number(shop.value.longitude) : ''
}

async function loadShop() {
  try {
    const response = await getShop()
    shop.value = response.data
    syncForm()
    editing.value = route.query.mode === 'resubmit' && shop.value.status === 2
    await loadTables(false)
  } catch (error) {
    if (error.msg === '店铺不存在') {
      router.replace('/register')
      return
    }
    window.alert(getApiErrorMessage(error, '加载店铺信息失败'))
  }
}

async function loadTables(showAlert = true) {
  if (!shop.value) {
    tables.value = []
    return
  }
  loadingTables.value = true
  try {
    const response = await listShopTables()
    tables.value = response.data || []
  } catch (error) {
    tables.value = []
    if (showAlert) {
      window.alert(getApiErrorMessage(error, '加载桌号失败'))
    }
  } finally {
    loadingTables.value = false
  }
}

function startEdit() {
  if (!canEdit.value) {
    window.alert('当前状态不可编辑店铺信息')
    return
  }
  editing.value = true
}

async function fillCurrentLocation() {
  if (locating.value) {
    return
  }
  if (!window.isSecureContext) {
    window.alert('当前页面不是安全上下文，手机浏览器通常只允许 HTTPS 页面获取定位')
    return
  }
  if (!navigator.geolocation) {
    window.alert('当前浏览器不支持定位，请手动填写经纬度')
    return
  }

  locating.value = true
  try {
    let position
    let usedFallback = false
    try {
      position = await getCurrentPosition({
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 0
      })
    } catch (error) {
      if (error?.code === 1) {
        throw error
      }
      usedFallback = true
      position = await getCurrentPosition({
        enableHighAccuracy: false,
        timeout: 15000,
        maximumAge: 300000
      })
    }

    const converted = wgs84ToGcj02(position.coords.latitude, position.coords.longitude)
    form.latitude = Number(converted.latitude.toFixed(6))
    form.longitude = Number(converted.longitude.toFixed(6))

    const qualityMessage = getLocationQualityMessage(position, usedFallback)
    if (qualityMessage) {
      window.alert(qualityMessage)
    }
  } catch (error) {
    window.alert(getLocationErrorMessage(error))
  } finally {
    locating.value = false
  }
}

async function saveShop() {
  if (!form.name || !form.address || !form.phone) {
    window.alert('请完整填写店铺信息')
    return
  }
  if (!hasCoordinateValue(form.latitude) || !hasCoordinateValue(form.longitude)) {
    window.alert('请先获取定位或手动填写经纬度')
    return
  }

  try {
    const response = await updateShop({
      name: form.name,
      address: form.address,
      phone: form.phone,
      description: form.description,
      latitude: Number(form.latitude),
      longitude: Number(form.longitude)
    })
    shop.value = response.data
    editing.value = false
    syncForm()
    if (shop.value.status !== 1) {
      router.replace(`/audit-status?status=${shop.value.status || 0}`)
    }
  } catch (error) {
    window.alert(getApiErrorMessage(error, '保存店铺信息失败'))
  }
}

async function toggleOpen() {
  try {
    await toggleShopOpen(!shop.value.is_open)
    shop.value.is_open = !shop.value.is_open
  } catch (error) {
    window.alert(getApiErrorMessage(error, '切换营业状态失败'))
  }
}

async function toggleAutoAccept() {
  savingAutoAccept.value = true
  try {
    const response = await toggleShopAutoAccept(!shop.value.auto_accept_orders)
    shop.value = response.data
    syncForm()
  } catch (error) {
    window.alert(getApiErrorMessage(error, '切换自动接单失败'))
  } finally {
    savingAutoAccept.value = false
  }
}

async function createTable() {
  if (!tableDraft.tableNo) {
    window.alert('请先填写桌号')
    return
  }
  creatingTable.value = true
  try {
    await createShopTable({
      table_no: tableDraft.tableNo,
      sort: Number(tableDraft.sort) || 0
    })
    tableDraft.tableNo = ''
    tableDraft.sort = 0
    await loadTables(false)
  } catch (error) {
    window.alert(getApiErrorMessage(error, '创建桌号失败'))
  } finally {
    creatingTable.value = false
  }
}

async function toggleTableStatus(table) {
  if (!table) {
    return
  }
  updatingTableId.value = table.id
  try {
    await updateShopTable(table.id, { status: table.status === 0 ? 1 : 0 })
    await loadTables(false)
  } catch (error) {
    window.alert(getApiErrorMessage(error, '更新桌号失败'))
  } finally {
    updatingTableId.value = 0
  }
}

async function deleteTableItem(table) {
  if (!table) {
    return
  }
  if (!window.confirm(`确认删除桌号 ${table.table_no} 吗？`)) {
    return
  }
  deletingTableId.value = table.id
  try {
    await deleteShopTable(table.id)
    await loadTables(false)
  } catch (error) {
    window.alert(getApiErrorMessage(error, '删除桌号失败'))
  } finally {
    deletingTableId.value = 0
  }
}

async function copyQRCode(url) {
  await copyText(url, '二维码链接')
}

async function copyText(value, label) {
  if (!value) {
    window.alert(`${label}为空`)
    return
  }
  try {
    await navigator.clipboard.writeText(String(value))
    window.alert(`${label}已复制`)
  } catch (error) {
    window.alert(`复制${label}失败，请手动复制`)
  }
}

async function rotatePosBindTokenAction() {
  if (!window.confirm('重置后，旧的 POS 密钥会立即失效，已连接的 POS 终端需要重新填写新密钥并重连。确认继续吗？')) {
    return
  }
  rotatingPosToken.value = true
  try {
    const response = await rotateShopPOSBindToken()
    shop.value = response.data
    syncForm()
    window.alert('POS 密钥已重置，请同步更新所有 POS 终端配置')
  } catch (error) {
    window.alert(getApiErrorMessage(error, '重置 POS 密钥失败'))
  } finally {
    rotatingPosToken.value = false
  }
}

function logout() {
  clearAuth()
  router.replace('/login')
}

onMounted(loadShop)
</script>

<style scoped>
.section-title {
  font-size: 20px;
}

.operation-row {
  gap: 12px;
}

.setting-value {
  color: #667085;
  line-height: 1.6;
}

.pos-binding-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.pos-binding-item {
  padding: 16px;
  border-radius: 16px;
  border: 1px solid #e4e7ec;
  background: #fafafa;
}

.pos-binding-item label {
  display: block;
  font-size: 13px;
  color: #667085;
}

.pos-binding-value {
  margin-top: 8px;
  font-size: 18px;
  font-weight: 700;
  color: #101828;
  word-break: break-all;
}

.token-value {
  font-family: 'SFMono-Regular', 'Consolas', monospace;
  font-size: 15px;
}

.pos-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.table-form-grid {
  grid-template-columns: minmax(0, 1.2fr) minmax(120px, 0.8fr) auto;
  align-items: end;
}

.table-create-btn {
  align-self: end;
}

.table-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.table-item {
  padding: 16px;
  border-radius: 16px;
  border: 1px solid #e4e7ec;
  background: #fafafa;
}

.table-title {
  font-size: 18px;
  font-weight: 700;
  color: #101828;
}

.table-meta {
  margin-top: 6px;
  font-size: 12px;
  color: #667085;
  word-break: break-all;
}

.table-qr-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
  flex-wrap: wrap;
}

.table-qr {
  width: 132px;
  height: 132px;
  object-fit: contain;
  border-radius: 16px;
  border: 1px solid #d0d5dd;
  background: #fff;
}

.table-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.secondary-link {
  color: #f97316;
  font-weight: 600;
  text-decoration: none;
}

@media (max-width: 720px) {
  .pos-binding-grid {
    grid-template-columns: 1fr;
  }

  .table-form-grid {
    grid-template-columns: 1fr;
  }

  .table-create-btn {
    width: 100%;
  }
}
</style>
