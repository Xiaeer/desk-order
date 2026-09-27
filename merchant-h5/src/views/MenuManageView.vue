<template>
  <section class="page-stack">
    <div class="card page-stack">
      <div class="row-between">
        <div>
          <div class="page-title">菜单管理</div>
          <div class="page-subtitle">按分类维护商品，保留与 mini-merchant 一致的经营能力。</div>
        </div>
        <button class="primary-btn" @click="openCategoryModal()">新增分类</button>
      </div>
    </div>

    <div class="card" v-if="categories.length">
      <div class="category-grid">
        <button
          v-for="category in categories"
          :key="category.id"
          class="secondary-btn category-button"
          :style="activeCategoryId === Number(category.id) ? activeStyle : ''"
          @click="switchCategory(category.id)"
        >
          {{ category.name }}
        </button>
      </div>
      <div style="margin-top: 12px;" class="row-between">
        <button class="secondary-btn" @click="openCategoryModal(activeCategory)">编辑分类</button>
        <button class="danger-btn" @click="removeCategory(activeCategory)">删除分类</button>
      </div>
    </div>

    <div class="card page-stack">
      <div class="row-between">
        <div class="page-title" style="font-size: 18px;">商品列表</div>
        <button class="primary-btn" :disabled="!activeCategory" @click="goCreateProduct">新增商品</button>
      </div>
      <div v-if="products.length" class="list">
        <div v-for="product in products" :key="product.id" class="card" style="padding: 14px; box-shadow: none; background: #fafbfc;">
          <div class="product-card-head">
            <img v-if="product.image" :src="product.image" alt="商品图" class="product-thumb" />
            <div style="flex: 1; min-width: 0;">
              <strong>{{ product.name }}</strong>
              <div class="page-subtitle">¥{{ (product.price / 100).toFixed(2) }}</div>
              <div v-if="product.options?.length" class="page-subtitle">{{ product.options.length }} 组选项</div>
            </div>
            <span class="chip">{{ product.is_on_sale ? '上架中' : '已下架' }}</span>
          </div>
          <div class="page-subtitle">{{ product.description || '暂无描述' }}</div>
          <div class="row-between">
            <button class="secondary-btn" @click="goEditProduct(product)">编辑</button>
            <button class="danger-btn" @click="removeProduct(product.id)">删除</button>
          </div>
        </div>
      </div>
      <div v-else class="empty-text">当前分类暂无商品</div>
    </div>

    <div v-if="showCategoryModal" class="dialog-mask" @click.self="closeCategoryModal">
      <div class="dialog-card card page-stack">
        <div class="page-title" style="font-size: 18px;">{{ editingCategory ? '编辑分类' : '新增分类' }}</div>
        <div class="field">
          <label>分类名称</label>
          <input v-model.trim="categoryForm.name" placeholder="请输入分类名称" />
        </div>
        <div class="field">
          <label>排序值</label>
          <input v-model.number="categoryForm.sort" type="number" placeholder="默认 0" />
        </div>
        <div class="row-between">
          <button class="secondary-btn" @click="closeCategoryModal">取消</button>
          <button class="primary-btn" @click="submitCategory">保存</button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onActivated, onDeactivated, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createCategory, deleteCategory, deleteProduct, getShopMenu, updateCategory } from '../api/menu'
import { getShop } from '../api/shop'
import { getApiErrorMessage, isShopNotFoundError } from '../utils/request'

defineOptions({ name: 'MenuManageView' })

const route = useRoute()
const router = useRouter()
const shop = ref(null)
const categories = ref([])
const activeCategoryId = ref(0)
const showCategoryModal = ref(false)
const editingCategory = ref(null)
const lastRouteCategoryId = ref(0)
const categoryForm = reactive({ name: '', sort: 0 })
const activeStyle = 'background:#ff6a3d;color:#fff;'
const productDraftStorageKey = 'merchant-product-edit-draft'

const activeCategory = computed(() => categories.value.find(category => Number(category.id) === activeCategoryId.value) || null)
const products = computed(() => activeCategory.value?.products || [])

function getRouteCategoryId() {
  return Number(route.query.categoryId || 0)
}

function restoreActiveCategory(preferredCategoryId = 0) {
  const targetCategoryId = Number(preferredCategoryId || activeCategoryId.value || 0)
  if (targetCategoryId) {
    const targetCategory = categories.value.find(category => Number(category.id) === targetCategoryId)
    if (targetCategory) {
      activeCategoryId.value = Number(targetCategory.id)
      return
    }
  }

  activeCategoryId.value = Number(categories.value[0]?.id || 0)
}

async function loadMenu(preferredCategoryId = 0) {
  try {
    const shopResponse = await getShop()
    shop.value = shopResponse.data
    if (shop.value.status !== 1) {
      router.replace(`/audit-status?status=${shop.value.status || 0}`)
      return
    }
    const menuResponse = await getShopMenu(shop.value.id)
    categories.value = menuResponse.data.categories || []
    restoreActiveCategory(preferredCategoryId)
  } catch (error) {
    if (isShopNotFoundError(error)) {
      router.replace('/register')
      return
    }
    window.alert(getApiErrorMessage(error, '加载菜单失败'))
  }
}

function switchCategory(categoryId) {
  activeCategoryId.value = Number(categoryId)
}

function openCategoryModal(category = null) {
  editingCategory.value = category
  categoryForm.name = category?.name || ''
  categoryForm.sort = category?.sort || 0
  showCategoryModal.value = true
}

function closeCategoryModal() {
  showCategoryModal.value = false
  editingCategory.value = null
}

async function submitCategory() {
  if (!categoryForm.name) {
    window.alert('请输入分类名称')
    return
  }
  try {
    const targetCategoryId = Number(editingCategory.value?.id || activeCategoryId.value || 0)
    if (editingCategory.value) {
      await updateCategory(editingCategory.value.id, { name: categoryForm.name, sort: categoryForm.sort })
    } else {
      await createCategory({ name: categoryForm.name, sort: categoryForm.sort })
    }
    closeCategoryModal()
    await loadMenu(targetCategoryId)
  } catch (error) {
    window.alert(getApiErrorMessage(error, '保存分类失败'))
  }
}

async function removeCategory(category) {
  if (!category) {
    window.alert('请先选择分类')
    return
  }
  if (!window.confirm('删除分类后，该分类下商品将一并删除，确认删除吗？')) {
    return
  }
  try {
    await deleteCategory(category.id)
    await loadMenu(activeCategoryId.value)
  } catch (error) {
    window.alert(getApiErrorMessage(error, '删除分类失败'))
  }
}

function goCreateProduct() {
  if (!activeCategory.value) {
    window.alert('请先创建分类')
    return
  }
  router.push({ path: '/product-edit', query: { categoryId: String(activeCategory.value.id) } })
}

function goEditProduct(product) {
  sessionStorage.setItem(productDraftStorageKey, JSON.stringify(product || null))
  router.push({ path: '/product-edit', query: { productId: String(product.id), categoryId: String(product.category_id) } })
}

async function removeProduct(id) {
  if (!window.confirm('确认删除该商品吗？')) {
    return
  }
  try {
    await deleteProduct(id)
    await loadMenu(activeCategoryId.value)
  } catch (error) {
    window.alert(getApiErrorMessage(error, '删除商品失败'))
  }
}

onMounted(async () => {
  const routeCategoryId = getRouteCategoryId()
  lastRouteCategoryId.value = routeCategoryId
  await loadMenu(routeCategoryId)
})

onActivated(async () => {
  const routeCategoryId = getRouteCategoryId()
  if (!routeCategoryId) {
    return
  }
  if (routeCategoryId === lastRouteCategoryId.value) {
    return
  }
  lastRouteCategoryId.value = routeCategoryId
  await loadMenu(routeCategoryId)
})

onDeactivated(() => {
  lastRouteCategoryId.value = 0
})
</script>

<style scoped>
.dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(17, 24, 39, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.dialog-card {
  width: 100%;
  max-width: 420px;
}

.category-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(96px, 1fr));
  gap: 10px;
}

.category-button {
  width: 100%;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.product-card-head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.product-thumb {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  object-fit: cover;
  background: #fff;
  flex-shrink: 0;
}
</style>
