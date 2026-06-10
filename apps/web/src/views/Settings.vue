<template>
  <div>
    <h1 class="page-title">设置</h1>

    <div class="card" style="max-width: 600px;">
      <div class="card-title">AI 配置</div>
      <div v-if="loading" class="loading">加载中...</div>
      <form v-else @submit.prevent="saveConfig">
        <div class="form-group">
          <label>AI 接口地址</label>
          <input v-model="config.endpoint" type="url" placeholder="如: http://localhost:11434/v1" />
        </div>
        <div class="form-group">
          <label>模型名称</label>
          <input v-model="config.model" type="text" placeholder="如: qwen2.5, deepseek-chat" />
        </div>
        <div class="form-group">
          <label>API Key</label>
          <input v-model="config.api_key" type="password" placeholder="输入API Key (已保存的将显示为*** )" />
        </div>
        <div class="form-group">
          <label>最大 Tokens</label>
          <input v-model.number="config.max_tokens" type="number" placeholder="4096" />
        </div>
        <div class="form-group">
          <label>温度</label>
          <input v-model.number="config.temperature" type="number" step="0.1" min="0" max="2" placeholder="0.7" />
        </div>

        <div v-if="currentConfig" style="margin-bottom: 16px; padding: 12px; background: var(--bg); border-radius: var(--radius);">
          <div style="font-size: 13px; font-weight: 600; margin-bottom: 8px;">当前配置</div>
          <div style="font-size: 13px; color: var(--text-secondary);">
            <div>接口: {{ currentConfig.endpoint || '-' }}</div>
            <div>模型: {{ currentConfig.model || '-' }}</div>
            <div>Key: {{ maskedKey }}</div>
            <div>最大 Tokens: {{ currentConfig.max_tokens || '-' }}</div>
          </div>
        </div>

        <div v-if="message" :style="{ color: messageColor, fontSize: '14px', marginBottom: '12px' }">{{ message }}</div>

        <button type="submit" class="btn btn-primary" :disabled="saving">
          {{ saving ? '保存中...' : '保存配置' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script>
import api from '../api.js'

export default {
  name: 'Settings',
  data() {
    return {
      loading: true,
      saving: false,
      message: '',
      messageColor: '#16a34a',
      currentConfig: null,
      config: {
        endpoint: '',
        model: '',
        api_key: '',
        max_tokens: 4096,
        temperature: 0.7,
      },
    }
  },
  computed: {
    maskedKey() {
      if (!this.currentConfig || !this.currentConfig.api_key) return '-'
      const key = this.currentConfig.api_key
      if (key.length <= 8) return '***'
      return key.substring(0, 4) + '***' + key.substring(key.length - 4)
    },
  },
  async mounted() {
    await this.loadConfig()
  },
  methods: {
    async loadConfig() {
      this.loading = true
      try {
        const res = await api.get('/api/config/ai')
        this.currentConfig = res.data.config || res.data || null
        if (this.currentConfig) {
          this.config.endpoint = this.currentConfig.endpoint || ''
          this.config.model = this.currentConfig.model || ''
          this.config.api_key = this.currentConfig.api_key || ''
          this.config.max_tokens = this.currentConfig.max_tokens || 4096
          this.config.temperature = this.currentConfig.temperature || 0.7
        }
      } catch {
        this.currentConfig = null
      }
      this.loading = false
    },
    async saveConfig() {
      this.saving = true
      this.message = ''
      try {
        await api.put('/api/config/ai', this.config)
        this.message = '配置保存成功'
        this.messageColor = '#16a34a'
        await this.loadConfig()
      } catch {
        this.message = '保存失败，请检查配置'
        this.messageColor = '#dc2626'
      }
      this.saving = false
    },
  },
}
</script>
