<template>
  <div>
    <h1 class="page-title">工作管理</h1>

    <!-- Work status cards -->
    <div class="card" style="margin-bottom: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <div class="card-title" style="margin-bottom: 0;">工作状态</div>
      </div>
      <div class="grid-3" v-if="statuses.length > 0">
        <div class="card overview-card" v-for="s in statuses" :key="s.member_id" style="margin-bottom: 0;">
          <div style="font-size: 16px; font-weight: 600;">{{ s.member_name || '成员' }}</div>
          <div style="margin-top: 8px;">
            <div v-if="s.current_project" style="font-size: 14px;">当前项目: {{ s.current_project }}</div>
            <div v-if="s.deadline" style="font-size: 14px; color: var(--text-secondary);">截止日期: {{ formatDate(s.deadline) }}</div>
            <div v-if="s.status" style="margin-top: 6px;">
              <span class="badge" :class="statusBadge(s.status)">{{ statusLabel(s.status) }}</span>
            </div>
          </div>
          <button class="btn btn-secondary btn-sm" style="margin-top: 12px;" @click="editStatus(s)">编辑</button>
        </div>
      </div>
      <div v-else class="empty-state">
        <p>暂无工作状态</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="card" style="margin-bottom: 20px;">
      <div class="filter-bar">
        <div class="form-group">
          <label>成员</label>
          <select v-model="filters.member_id" @change="loadRecords">
            <option value="">全部</option>
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>状态</label>
          <select v-model="filters.status" @change="loadRecords">
            <option value="">全部</option>
            <option value="active">进行中</option>
            <option value="completed">已完成</option>
            <option value="pending">待开始</option>
          </select>
        </div>
        <button class="btn btn-primary btn-sm" @click="showRecordModal = true">+ 添加记录</button>
      </div>
    </div>

    <!-- Work records -->
    <div class="card">
      <div class="card-title">工作记录</div>
      <div v-if="records.length === 0" class="empty-state">
        <p>暂无记录</p>
      </div>
      <div class="table-container" v-else>
        <table>
          <thead>
            <tr>
              <th>日期</th>
              <th>成员</th>
              <th>项目</th>
              <th>状态</th>
              <th>内容</th>
              <th>截止日期</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id">
              <td>{{ formatDate(r.date || r.created_at) }}</td>
              <td>{{ r.member_name || '-' }}</td>
              <td>{{ r.project || '-' }}</td>
              <td><span class="badge" :class="statusBadge(r.status)">{{ statusLabel(r.status) }}</span></td>
              <td>{{ r.content || r.title || '-' }}</td>
              <td>{{ formatDate(r.deadline) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Edit Status Modal -->
    <div class="modal-overlay" v-if="showStatusModal" @click.self="showStatusModal = false">
      <div class="modal">
        <h3>编辑工作状态</h3>
        <div class="form-group">
          <label>当前项目</label>
          <input v-model="statusForm.current_project" type="text" placeholder="项目名称" />
        </div>
        <div class="form-group">
          <label>状态</label>
          <select v-model="statusForm.status">
            <option value="active">进行中</option>
            <option value="completed">已完成</option>
            <option value="pending">待开始</option>
          </select>
        </div>
        <div class="form-group">
          <label>截止日期</label>
          <input v-model="statusForm.deadline" type="date" />
        </div>
        <div class="form-group">
          <label>备注</label>
          <textarea v-model="statusForm.note" placeholder="可选备注"></textarea>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showStatusModal = false">取消</button>
          <button class="btn btn-primary" @click="saveStatus">保存</button>
        </div>
      </div>
    </div>

    <!-- Add Record Modal -->
    <div class="modal-overlay" v-if="showRecordModal" @click.self="showRecordModal = false">
      <div class="modal">
        <h3>添加工作记录</h3>
        <div class="form-group">
          <label>成员</label>
          <select v-model="recordForm.member_id">
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>项目</label>
          <input v-model="recordForm.project" type="text" placeholder="项目名称" />
        </div>
        <div class="form-group">
          <label>状态</label>
          <select v-model="recordForm.status">
            <option value="active">进行中</option>
            <option value="completed">已完成</option>
            <option value="pending">待开始</option>
          </select>
        </div>
        <div class="form-group">
          <label>内容</label>
          <input v-model="recordForm.content" type="text" placeholder="工作内容" />
        </div>
        <div class="form-group">
          <label>截止日期</label>
          <input v-model="recordForm.deadline" type="date" />
        </div>
        <div class="form-group">
          <label>备注</label>
          <textarea v-model="recordForm.note" placeholder="可选备注"></textarea>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showRecordModal = false">取消</button>
          <button class="btn btn-primary" @click="addRecord">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../api.js'

export default {
  name: 'Work',
  data() {
    return {
      statuses: [],
      records: [],
      members: [],
      showStatusModal: false,
      showRecordModal: false,
      filters: { member_id: '', status: '' },
      statusForm: { member_id: '', current_project: '', status: 'active', deadline: '', note: '' },
      recordForm: { member_id: '', project: '', status: 'active', content: '', deadline: '', note: '' },
    }
  },
  async mounted() {
    await Promise.all([this.loadStatuses(), this.loadRecords(), this.loadMembers()])
  },
  methods: {
    async loadStatuses() {
      try {
        const res = await api.get('/api/work/status')
        this.statuses = res.data.statuses || res.data || []
      } catch {
        this.statuses = []
      }
    },
    async loadRecords() {
      try {
        const params = new URLSearchParams()
        if (this.filters.member_id) params.set('member_id', this.filters.member_id)
        if (this.filters.status) params.set('status', this.filters.status)
        const res = await api.get('/api/work/records', { params })
        this.records = res.data.records || res.data || []
      } catch {
        this.records = []
      }
    },
    async loadMembers() {
      try {
        const res = await api.get('/api/members')
        this.members = res.data.members || res.data || []
        if (this.members.length > 0 && !this.recordForm.member_id) {
          this.recordForm.member_id = this.members[0].id
        }
      } catch {
        this.members = []
      }
    },
    editStatus(s) {
      this.statusForm = {
        member_id: s.member_id,
        current_project: s.current_project || '',
        status: s.status || 'active',
        deadline: s.deadline || '',
        note: s.note || '',
      }
      this.showStatusModal = true
    },
    async saveStatus() {
      try {
        await api.put(`/api/work/status/${this.statusForm.member_id}`, this.statusForm)
        this.showStatusModal = false
        await this.loadStatuses()
      } catch {
        // ignore
      }
    },
    async addRecord() {
      if (!this.recordForm.content) return
      try {
        await api.post('/api/work/records', this.recordForm)
        this.showRecordModal = false
        this.recordForm = { member_id: this.members.length > 0 ? this.members[0].id : '', project: '', status: 'active', content: '', deadline: '', note: '' }
        await this.loadRecords()
      } catch {
        // ignore
      }
    },
    statusLabel(s) {
      const map = { active: '进行中', completed: '已完成', pending: '待开始' }
      return map[s] || s || '-'
    },
    statusBadge(s) {
      const map = { active: 'badge-yellow', completed: 'badge-green', pending: 'badge-blue' }
      return map[s] || 'badge-primary'
    },
    formatDate(d) {
      if (!d) return '-'
      return new Date(d).toLocaleDateString('zh-CN')
    },
  },
}
</script>
