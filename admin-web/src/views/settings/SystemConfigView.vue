<template>
  <div class="settings-page">
    <el-card class="settings-card">
      <template #header>
        <div class="card-header">
          <div>
            <div class="title">用户端退款功能</div>
            <div class="subtitle">控制小程序是否允许用户侧发起自助退款申请；退款入口显隐会自动跟随这个开关。</div>
          </div>
          <el-button type="primary" :loading="saving" @click="saveConfig">保存设置</el-button>
        </div>
      </template>

      <div v-loading="loading" class="settings-body">
        <div class="setting-row">
          <div>
            <div class="setting-label">允许用户自助退款</div>
            <div class="setting-desc">关闭后，小程序不再展示用户自助退款入口；即使用户端通过旧入口或缓存页面触发退款，后端也会拒绝用户侧退款申请。</div>
          </div>
          <el-switch v-model="form.mini_user_refund_apply_enabled" />
        </div>

        <div class="setting-field">
          <div class="setting-label">用户端品牌名</div>
          <div class="setting-desc">用于小程序首页品牌标识与“关于”文案。留空时默认显示 DeskOrder。</div>
          <el-input v-model="form.mini_user_brand_name" maxlength="24" placeholder="例如：幸福火锅" />
        </div>

        <div class="setting-group">
      <div class="group-title">退款订阅消息</div>
      <div class="group-desc">这里配置小程序退款通知模板。未配置模板 ID 时，站内退款通知仍可用，但不会发微信订阅消息。</div>

      <div class="setting-field">
        <div class="setting-label">跳转页面</div>
        <el-input v-model="form.refund_subscribe_page" placeholder="例如：pages/refund-center/refund-center" />
      </div>

      <div class="setting-split-title">退款审核通知</div>
      <div class="setting-grid">
        <div class="setting-field">
          <div class="setting-label">模板 ID</div>
          <el-input v-model="form.refund_review_subscribe_template_id" placeholder="微信订阅消息模板 ID" />
        </div>
        <div class="setting-field">
          <div class="setting-label">订单号关键词 key</div>
          <el-input v-model="form.refund_review_subscribe_order_no_key" placeholder="如 character_string1" />
        </div>
        <div class="setting-field">
          <div class="setting-label">金额关键词 key</div>
          <el-input v-model="form.refund_review_subscribe_amount_key" placeholder="如 amount2" />
        </div>
        <div class="setting-field">
          <div class="setting-label">状态关键词 key</div>
          <el-input v-model="form.refund_review_subscribe_status_key" placeholder="如 phrase3" />
        </div>
        <div class="setting-field field-span-2">
          <div class="setting-label">备注关键词 key</div>
          <el-input v-model="form.refund_review_subscribe_remark_key" placeholder="如 thing4" />
        </div>
      </div>

      <div class="setting-split-title">退款结果通知</div>
      <div class="setting-grid">
        <div class="setting-field">
          <div class="setting-label">模板 ID</div>
          <el-input v-model="form.refund_result_subscribe_template_id" placeholder="微信订阅消息模板 ID" />
        </div>
        <div class="setting-field">
          <div class="setting-label">订单号关键词 key</div>
          <el-input v-model="form.refund_result_subscribe_order_no_key" placeholder="如 character_string1" />
        </div>
        <div class="setting-field">
          <div class="setting-label">金额关键词 key</div>
          <el-input v-model="form.refund_result_subscribe_amount_key" placeholder="如 amount2" />
        </div>
        <div class="setting-field">
          <div class="setting-label">状态关键词 key</div>
          <el-input v-model="form.refund_result_subscribe_status_key" placeholder="如 phrase3" />
        </div>
        <div class="setting-field field-span-2">
          <div class="setting-label">备注关键词 key</div>
          <el-input v-model="form.refund_result_subscribe_remark_key" placeholder="如 thing4" />
        </div>
      </div>
    </div>

        <el-alert
          title="当前版本采用全量开关"
          type="warning"
          :closable="false"
          description="开关一旦打开，同一后端下的所有小程序用户都会看到并可使用相应退款能力；如果后续要做到仅测试账号可见，需要再补白名单机制。"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getSystemConfigs, updateSystemConfig } from '../../api/system-config'

const loading = ref(false)
const saving = ref(false)

const form = reactive({
  mini_user_refund_visible: false,
  mini_user_refund_apply_enabled: false,
  mini_user_brand_name: '',
  refund_subscribe_page: '',
  refund_review_subscribe_template_id: '',
  refund_review_subscribe_order_no_key: '',
  refund_review_subscribe_amount_key: '',
  refund_review_subscribe_status_key: '',
  refund_review_subscribe_remark_key: '',
  refund_result_subscribe_template_id: '',
  refund_result_subscribe_order_no_key: '',
  refund_result_subscribe_amount_key: '',
  refund_result_subscribe_status_key: '',
  refund_result_subscribe_remark_key: ''
})

async function fetchConfig() {
  loading.value = true
  try {
    const res = await getSystemConfigs()
    const refundApplyEnabled = !!res.data.mini_user_refund_apply_enabled
    Object.assign(form, {
      mini_user_refund_visible: refundApplyEnabled,
      mini_user_refund_apply_enabled: refundApplyEnabled,
      mini_user_brand_name: res.data.mini_user_brand_name || 'DeskOrder',
      refund_subscribe_page: res.data.refund_subscribe_page || '',
      refund_review_subscribe_template_id: res.data.refund_review_subscribe_template_id || '',
      refund_review_subscribe_order_no_key: res.data.refund_review_subscribe_order_no_key || '',
      refund_review_subscribe_amount_key: res.data.refund_review_subscribe_amount_key || '',
      refund_review_subscribe_status_key: res.data.refund_review_subscribe_status_key || '',
      refund_review_subscribe_remark_key: res.data.refund_review_subscribe_remark_key || '',
      refund_result_subscribe_template_id: res.data.refund_result_subscribe_template_id || '',
      refund_result_subscribe_order_no_key: res.data.refund_result_subscribe_order_no_key || '',
      refund_result_subscribe_amount_key: res.data.refund_result_subscribe_amount_key || '',
      refund_result_subscribe_status_key: res.data.refund_result_subscribe_status_key || '',
      refund_result_subscribe_remark_key: res.data.refund_result_subscribe_remark_key || ''
    })
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  try {
    const refundApplyEnabled = !!form.mini_user_refund_apply_enabled
    const res = await updateSystemConfig({
      mini_user_refund_visible: refundApplyEnabled,
      mini_user_refund_apply_enabled: refundApplyEnabled,
      mini_user_brand_name: form.mini_user_brand_name,
      refund_subscribe_page: form.refund_subscribe_page,
      refund_review_subscribe_template_id: form.refund_review_subscribe_template_id,
      refund_review_subscribe_order_no_key: form.refund_review_subscribe_order_no_key,
      refund_review_subscribe_amount_key: form.refund_review_subscribe_amount_key,
      refund_review_subscribe_status_key: form.refund_review_subscribe_status_key,
      refund_review_subscribe_remark_key: form.refund_review_subscribe_remark_key,
      refund_result_subscribe_template_id: form.refund_result_subscribe_template_id,
      refund_result_subscribe_order_no_key: form.refund_result_subscribe_order_no_key,
      refund_result_subscribe_amount_key: form.refund_result_subscribe_amount_key,
      refund_result_subscribe_status_key: form.refund_result_subscribe_status_key,
      refund_result_subscribe_remark_key: form.refund_result_subscribe_remark_key
    })
    const nextRefundApplyEnabled = !!res.data.mini_user_refund_apply_enabled
    Object.assign(form, {
      mini_user_refund_visible: nextRefundApplyEnabled,
      mini_user_refund_apply_enabled: nextRefundApplyEnabled,
      mini_user_brand_name: res.data.mini_user_brand_name || 'DeskOrder',
      refund_subscribe_page: res.data.refund_subscribe_page || '',
      refund_review_subscribe_template_id: res.data.refund_review_subscribe_template_id || '',
      refund_review_subscribe_order_no_key: res.data.refund_review_subscribe_order_no_key || '',
      refund_review_subscribe_amount_key: res.data.refund_review_subscribe_amount_key || '',
      refund_review_subscribe_status_key: res.data.refund_review_subscribe_status_key || '',
      refund_review_subscribe_remark_key: res.data.refund_review_subscribe_remark_key || '',
      refund_result_subscribe_template_id: res.data.refund_result_subscribe_template_id || '',
      refund_result_subscribe_order_no_key: res.data.refund_result_subscribe_order_no_key || '',
      refund_result_subscribe_amount_key: res.data.refund_result_subscribe_amount_key || '',
      refund_result_subscribe_status_key: res.data.refund_result_subscribe_status_key || '',
      refund_result_subscribe_remark_key: res.data.refund_result_subscribe_remark_key || ''
    })
    ElMessage.success('设置已保存')
  } finally {
    saving.value = false
  }
}

onMounted(fetchConfig)
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settings-card {
  max-width: 860px;
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.title {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.subtitle {
  margin-top: 6px;
  color: #6b7280;
  line-height: 1.6;
}

.settings-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.setting-group {
  padding-top: 8px;
  border-top: 1px solid #eef2f7;
}

.group-title {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
}

.group-desc {
  margin-top: 6px;
  color: #6b7280;
  line-height: 1.6;
}

.setting-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.setting-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-span-2 {
  grid-column: span 2;
}

.setting-split-title {
  margin-top: 20px;
  font-size: 14px;
  font-weight: 700;
  color: #374151;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 18px 0;
  border-bottom: 1px solid #eef2f7;
}

.setting-row:last-of-type {
  border-bottom: none;
}

.setting-label {
  font-size: 15px;
  font-weight: 600;
  color: #111827;
}

.setting-desc {
  margin-top: 6px;
  color: #6b7280;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .card-header,
  .setting-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .setting-grid {
    grid-template-columns: 1fr;
  }

  .field-span-2 {
    grid-column: span 1;
  }
}
</style>