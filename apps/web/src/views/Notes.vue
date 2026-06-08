<template>
  <div>
    <h1 class="page-title">笔记</h1>

    <!-- Filters & add -->
    <div class="card" style="margin-bottom: 20px;">
      <div class="filter-bar">
        <div class="form-group">
          <label>领域</label>
          <select v-model="filters.domain" @change="loadNotes">
            <option value="">全部</option>
            <option value="finance">财务</option>
            <option value="health">健康</option>
            <option value="work">工作</option>
            <option value="family">家庭</option>
            <option value="general">通用</option>
          </select>
        </div>
        <div class="form-group">
          <label>成员</label>
          <select v-model="filters.member_id" @change="loadNotes">
            <option value="">全部</option>
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <button class="btn btn-primary btn-sm" @click="openAddNote">+ 添加笔记</button>
      </div>
    </div>

    <!-- Notes list -->
    <div v-if="notes.length === 0" class="card">
      <div class="empty-state">
        <p>暂无笔记</p>
      </div>
    </div>
    <div v-else>
      <div class="card" v-for="n in notes" :key="n.id" style="margin-bottom: 12px;">
        <div style="display: flex; justify-content: space-between; align-items: flex-start;">
          <div style="flex: 1;">
            <div style="display: flex; gap: 8px; align-items: center; margin-bottom: 8px;">
              <span style="font-size: 16px; font-weight: 600;">{{ n.title || '无标题' }}</span>
              <span v-if="n.domain" class="badge badge-primary">{{ domainLabel(n.domain) }}</span>
              <span v-for="tag in parseTags(n.tags)" :key="tag" class="badge badge-blue">{{ tag }}</span>
            </div>
            <div style="font-size: 14px; color: var(--text-secondary); margin-bottom: 8px;">
              {{ n.content || '' }}
            </div>
            <div style="font-size: 12px; color: var(--text-secondary);">
              {{ formatDate(n.created_at) }}
              <span v-if="n.member_name"> - {{ n.member_name }}</span>
            </div>
          </div>
          <div style="display: flex; gap: 6px; margin-left: 12px;">
            <button class="btn btn-secondary btn-sm" @click="openEditNote(n)">编辑</button>
            <button class="btn btn-danger btn-sm" @click="deleteNote(n.id)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Note Modal -->
    <div class="modal-overlay" v-if="showModal" @click.self="showModal = false">
      <div class="modal">
        <h3>{{ isEditing ? '编辑笔记' : '添加笔记' }}</h3>
        <div class="form-group">
          <label>标题</label>
          <input v-model="noteForm.title" type="text" placeholder="笔记标题" />
        </div>
        <div class="form-group">
          <label>内容</label>
          <textarea v-model="noteForm.content" placeholder="笔记内容" style="min-height: 120px;"></textarea>
        </div>
        <div class="form-group">
          <label>领域</label>
          <select v-model="noteForm.domain">
            <option value="general">通用</option>
            <option value="finance">财务</option>
            <option value="health">健康</option>
            <option value="work">工作</option>
            <option value="family">家庭</option>
          </select>
        </div>
        <div class="form-group">
          <label>成员</label>
          <select v-model="noteForm.member_id">
            <option value="">不指定</option>
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>标签 (逗号分隔)</label>
          <input v-model="noteForm.tags_input" type="text" placeholder="标签1, 标签2" />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showModal = false">取消</button>
          <button class="btn btn-primary" @click="saveNote">{{ isEditing ? '保存' : '确定' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../api.js'

export default {
  name: 'Notes',
  data() {
    return {
      notes: [],
      members: [],
      showModal: false,
      isEditing: false,
      editingId: null,
      filters: { domain: '', member_id: '' },
      noteForm: { title: '', content: '', domain: 'general', member_id: '', tags_input: '' },
    }
  },
  async mounted() {
    await Promise.all([this.loadNotes(), this.loadMembers()])
  },
  methods: {
    async loadNotes() {
      try {
        const params = new URLSearchParams()
        if (this.filters.domain) params.set('domain', this.filters.domain)
        if (this.filters.member_id) params.set('member_id', this.filters.member_id)
        const res = await api.get('/api/notes', { params })
        this.notes = res.data.notes || res.data || []
      } catch {
        this.notes = []
      }
    },
    async loadMembers() {
      try {
        const res = await api.get('/api/members')
        this.members = res.data.members || res.data || []
      } catch {
        this.members = []
      }
    },
    openAddNote() {
      this.isEditing = false
      this.editingId = null
      this.noteForm = { title: '', content: '', domain: 'general', member_id: '', tags_input: '' }
      this.showModal = true
    },
    openEditNote(n) {
      this.isEditing = true
      this.editingId = n.id
      this.noteForm = {
        title: n.title || '',
        content: n.content || '',
        domain: n.domain || 'general',
        member_id: n.member_id || '',
        tags_input: Array.isArray(n.tags) ? n.tags.join(', ') : (n.tags || ''),
      }
      this.showModal = true
    },
    async saveNote() {
      if (!this.noteForm.title && !this.noteForm.content) return
      const payload = {
        title: this.noteForm.title,
        content: this.noteForm.content,
        domain: this.noteForm.domain,
        member_id: this.noteForm.member_id,
        tags: this.noteForm.tags_input
          ? this.noteForm.tags_input.split(',').map(t => t.trim()).filter(Boolean)
          : [],
      }
      try {
        if (this.isEditing) {
          await api.put(`/api/notes/${this.editingId}`, payload)
        } else {
          await api.post('/api/notes', payload)
        }
        this.showModal = false
        await this.loadNotes()
      } catch {
        // ignore
      }
    },
    async deleteNote(id) {
      if (!confirm('确定删除此笔记？')) return
      try {
        await api.delete(`/api/notes/${id}`)
        await this.loadNotes()
      } catch {
        // ignore
      }
    },
    parseTags(tags) {
      if (Array.isArray(tags)) return tags
      if (typeof tags === 'string' && tags) return tags.split(',').map(t => t.trim()).filter(Boolean)
      return []
    },
    domainLabel(d) {
      const map = { finance: '财务', health: '健康', work: '工作', family: '家庭', general: '通用' }
      return map[d] || d || '-'
    },
    formatDate(d) {
      if (!d) return '-'
      return new Date(d).toLocaleString('zh-CN')
    },
  },
}
</script>
