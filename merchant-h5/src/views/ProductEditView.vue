<template>
  <section class="page-stack">
    <div class="card page-stack">
      <div class="page-title">{{ isEdit ? '编辑商品' : '新增商品' }}</div>
      <div class="field-grid">
        <div class="field">
          <label>商品名称</label>
          <input v-model.trim="form.name" placeholder="请输入商品名称" />
        </div>
        <div class="field">
          <label>商品图片</label>
          <div class="upload-panel">
            <img v-if="form.image" :src="form.image" alt="商品图片" class="product-preview" />
            <div v-else class="upload-placeholder">暂未上传商品图片</div>
            <div class="upload-actions">
              <label class="secondary-btn upload-trigger" :class="{ disabled: uploadingImage }">
                {{ uploadingImage ? '上传中...' : form.image ? '重新上传' : '选择图片' }}
                <input
                  class="hidden-file-input"
                  type="file"
                  accept="image/png,image/jpeg,image/webp,image/gif"
                  :disabled="uploadingImage"
                  @change="handleImageSelect"
                />
              </label>
              <button v-if="form.image" class="secondary-btn" type="button" @click="clearImage">移除图片</button>
            </div>
            <div class="page-subtitle upload-tip">支持 jpg、jpeg、png、webp、gif，单张不超过 5MB。</div>
          </div>
        </div>
        <div class="field-grid" style="grid-template-columns: 1fr 1fr;">
          <div class="field">
            <label>价格（元）</label>
            <input v-model.trim="form.price" placeholder="请输入价格" />
          </div>
          <div class="field">
            <label>排序值</label>
            <input v-model.number="form.sort" type="number" placeholder="0" />
          </div>
        </div>
        <div class="field">
          <label>商品描述</label>
          <textarea v-model.trim="form.description" rows="3" placeholder="可选"></textarea>
        </div>
        <div class="field page-stack">
          <label>子选项组</label>
          <div class="page-subtitle">支持口味、规格、加料等动态选项。单选组适合辣度/规格，多选组适合加料。</div>
          <div v-if="form.options.length" class="option-group-list">
            <div v-for="(group, groupIndex) in form.options" :key="group.id" class="option-group-card">
              <div class="row-between option-group-head">
                <strong>选项组 {{ groupIndex + 1 }}</strong>
                <button class="danger-btn" type="button" @click="removeOptionGroup(groupIndex)">删除选项组</button>
              </div>
              <div class="field-grid option-group-grid">
                <div class="field">
                  <label>组名称</label>
                  <input v-model.trim="group.name" placeholder="例如：辣度、规格、加料" />
                </div>
                <div class="field">
                  <label>选择方式</label>
                  <select v-model="group.selectType">
                    <option value="single">单选</option>
                    <option value="multi">多选</option>
                  </select>
                </div>
                <div class="field">
                  <label>排序值</label>
                  <input v-model.number="group.sort" type="number" placeholder="0" />
                </div>
              </div>
              <label class="row-between" style="justify-content:flex-start; gap: 10px;">
                <input v-model="group.required" type="checkbox" style="width:auto;" />
                必选组
              </label>
              <div v-if="group.selectType === 'multi'" class="field-grid option-group-grid">
                <div class="field">
                  <label>至少选择</label>
                  <input v-model.number="group.minSelect" type="number" min="0" placeholder="0" />
                </div>
                <div class="field">
                  <label>最多选择</label>
                  <input v-model.number="group.maxSelect" type="number" min="1" placeholder="默认不限" />
                </div>
              </div>
              <div class="page-subtitle" style="margin-top: 4px;">选项值</div>
              <div class="option-value-list">
                <div v-for="(value, valueIndex) in group.values" :key="value.id" class="option-value-row">
                  <input v-model.trim="value.name" placeholder="选项值名称，例如：微辣" />
                  <input v-model.trim="value.priceDeltaYuan" placeholder="加价，单位元" />
                  <input v-model.number="value.sort" type="number" placeholder="排序" />
                  <button class="danger-btn" type="button" @click="removeOptionValue(groupIndex, valueIndex)">删除</button>
                </div>
              </div>
              <button class="secondary-btn" type="button" @click="addOptionValue(groupIndex)">新增选项值</button>
            </div>
          </div>
          <div v-else class="page-subtitle">暂未配置子选项。未配置时，用户可直接加购。</div>
          <button class="secondary-btn" type="button" @click="addOptionGroup">新增选项组</button>
        </div>
        <label class="row-between" style="justify-content:flex-start; gap: 10px;">
          <input v-model="form.isOnSale" type="checkbox" style="width:auto;" />
          上架销售
        </label>
      </div>
      <button class="primary-btn" :disabled="submitting" @click="submit">{{ submitting ? '保存中...' : '保存商品' }}</button>
    </div>
  </section>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createProduct, updateProduct, uploadProductImage } from '../api/menu'
import { getApiErrorMessage } from '../utils/request'

const productDraftStorageKey = 'merchant-product-edit-draft'

function createLocalId(prefix) {
  if (window.crypto && typeof window.crypto.randomUUID === 'function') {
    return `${prefix}${window.crypto.randomUUID().replace(/-/g, '').slice(0, 12)}`
  }
  return `${prefix}${Date.now().toString(16)}${Math.random().toString(16).slice(2, 8)}`
}

function formatFenToCompactYuan(value) {
  return (Number(value || 0) / 100).toFixed(2).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
}

function createEmptyOptionValue() {
  return {
    id: createLocalId('pov_'),
    name: '',
    priceDeltaYuan: '0',
    sort: 0
  }
}

function createEmptyOptionGroup() {
  return {
    id: createLocalId('pog_'),
    name: '',
    selectType: 'single',
    required: false,
    minSelect: 0,
    maxSelect: 1,
    sort: 0,
    values: [createEmptyOptionValue()]
  }
}

function parseDraftProduct(route) {
  const productId = Number(route.query.productId || 0)
  if (!productId) {
    try {
      return route.query.product ? JSON.parse(route.query.product) : null
    } catch (error) {
      return null
    }
  }
  try {
    const rawDraft = sessionStorage.getItem(productDraftStorageKey)
    const parsedDraft = rawDraft ? JSON.parse(rawDraft) : null
    if (parsedDraft && Number(parsedDraft.id) === productId) {
      return parsedDraft
    }
  } catch (error) {
    return null
  }
  return null
}

function toFormOptionGroups(optionGroups) {
  return Array.isArray(optionGroups)
    ? optionGroups.map(group => ({
        id: group.id || createLocalId('pog_'),
        name: group.name || '',
        selectType: group.select_type || 'single',
        required: !!group.required,
        minSelect: Number(group.min_select || 0),
        maxSelect: Number(group.max_select || (group.select_type === 'multi' ? 0 : 1)),
        sort: Number(group.sort || 0),
        values: Array.isArray(group.values) && group.values.length
          ? group.values.map(value => ({
              id: value.id || createLocalId('pov_'),
              name: value.name || '',
              priceDeltaYuan: formatFenToCompactYuan(value.price_delta),
              sort: Number(value.sort || 0)
            }))
          : [createEmptyOptionValue()]
      }))
    : []
}

function parseYuanToFen(value) {
  const numeric = Number(String(value || '').trim())
  if (!Number.isFinite(numeric)) {
    return null
  }
  return Math.round(numeric * 100)
}

const route = useRoute()
const router = useRouter()
const submitting = ref(false)
const uploadingImage = ref(false)
const initialProduct = parseDraftProduct(route)
const isEdit = computed(() => !!initialProduct)

const form = reactive({
  id: initialProduct?.id || null,
  categoryId: Number(initialProduct?.category_id || route.query.categoryId || 0),
  name: initialProduct?.name || '',
  image: initialProduct?.image || '',
  price: initialProduct ? String((initialProduct.price / 100).toFixed(2)) : '',
  description: initialProduct?.description || '',
  options: toFormOptionGroups(initialProduct?.options),
  isOnSale: initialProduct?.is_on_sale ?? true,
  sort: initialProduct?.sort || 0
})

function addOptionGroup() {
  form.options.push(createEmptyOptionGroup())
}

function removeOptionGroup(groupIndex) {
  form.options.splice(groupIndex, 1)
}

function addOptionValue(groupIndex) {
  form.options[groupIndex]?.values.push(createEmptyOptionValue())
}

function removeOptionValue(groupIndex, valueIndex) {
  const group = form.options[groupIndex]
  if (!group) {
    return
  }
  if (group.values.length === 1) {
    window.alert('每个选项组至少保留一个选项值')
    return
  }
  group.values.splice(valueIndex, 1)
}

function buildOptionsPayload() {
  return form.options.map((group, groupIndex) => {
    if (!group.name) {
      throw new Error(`请填写第 ${groupIndex + 1} 组选项名称`)
    }
    const values = group.values.map((value, valueIndex) => {
      if (!value.name) {
        throw new Error(`请填写选项组 ${group.name || groupIndex + 1} 的第 ${valueIndex + 1} 个选项值名称`)
      }
      const priceDelta = parseYuanToFen(value.priceDeltaYuan)
      if (priceDelta === null || priceDelta < 0) {
        throw new Error(`选项值 ${value.name} 的加价格式不正确`)
      }
      return {
        id: value.id,
        name: value.name,
        price_delta: priceDelta,
        sort: Number(value.sort) || 0
      }
    })

    const isMulti = group.selectType === 'multi'
    const minSelect = group.required ? Math.max(1, Number(group.minSelect) || 0) : Math.max(0, Number(group.minSelect) || 0)
    const maxSelect = isMulti ? Math.max(Number(group.maxSelect) || values.length, 1) : 1
    if (isMulti && minSelect > maxSelect) {
      throw new Error(`选项组 ${group.name} 的最少选择数量不能大于最多选择数量`)
    }

    return {
      id: group.id,
      name: group.name,
      select_type: isMulti ? 'multi' : 'single',
      required: !!group.required,
      min_select: isMulti ? minSelect : (group.required ? 1 : 0),
      max_select: maxSelect,
      sort: Number(group.sort) || 0,
      values
    }
  })
}

async function handleImageSelect(event) {
  const [file] = event.target.files || []
  event.target.value = ''
  if (!file) {
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    window.alert('图片不能超过 5MB')
    return
  }

  uploadingImage.value = true
  try {
    const response = await uploadProductImage(file)
    form.image = response.data.url || ''
  } catch (error) {
    window.alert(getApiErrorMessage(error, '上传图片失败'))
  } finally {
    uploadingImage.value = false
  }
}

function clearImage() {
  form.image = ''
}

async function submit() {
  if (uploadingImage.value) {
    window.alert('图片上传中，请稍后再保存商品')
    return
  }
  if (!form.name || !form.price) {
    window.alert('请填写商品名称和价格')
    return
  }
  const price = Math.round(Number(form.price) * 100)
  if (!price || price <= 0) {
    window.alert('请输入正确的价格')
    return
  }

  submitting.value = true
  try {
    const options = buildOptionsPayload()
    const payload = {
      name: form.name,
      image: form.image,
      price,
      description: form.description,
      options,
      is_on_sale: form.isOnSale,
      sort: Number(form.sort) || 0
    }
    if (isEdit.value) {
      payload.category_id = form.categoryId
      await updateProduct(form.id, payload)
    } else {
      payload.category_id = form.categoryId
      await createProduct(payload)
    }
    sessionStorage.removeItem(productDraftStorageKey)
    router.replace({ path: '/menu', query: { categoryId: String(form.categoryId) } })
  } catch (error) {
    window.alert(getApiErrorMessage(error, '保存商品失败'))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.option-group-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.option-group-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  border-radius: 16px;
  border: 1px solid #e4e7ec;
  background: #fcfcfd;
}

.option-group-head {
  align-items: center;
}

.option-group-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.option-value-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.option-value-row {
  display: grid;
  grid-template-columns: minmax(0, 1.3fr) minmax(120px, 0.8fr) minmax(80px, 0.5fr) auto;
  gap: 10px;
  align-items: center;
}

.upload-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  border: 1px dashed #d7dbe0;
  border-radius: 16px;
  background: #fafbfc;
}

.product-preview {
  width: 100%;
  max-height: 220px;
  object-fit: cover;
  border-radius: 12px;
  background: #fff;
}

.upload-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 180px;
  border-radius: 12px;
  background: #fff;
  color: #98a2b3;
  font-size: 14px;
}

.upload-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.upload-trigger {
  position: relative;
  overflow: hidden;
  cursor: pointer;
}

.upload-trigger.disabled {
  pointer-events: none;
  opacity: 0.7;
}

.hidden-file-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.upload-tip {
  margin: 0;
}

select {
  width: 100%;
  min-height: 44px;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid #d0d5dd;
  background: #fff;
  color: #101828;
}

@media (max-width: 720px) {
  .option-group-grid,
  .option-value-row {
    grid-template-columns: 1fr;
  }
}
</style>
