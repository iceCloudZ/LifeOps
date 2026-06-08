<template>
  <div>
    <h1 class="page-title">家庭事务</h1>

    <!-- Family status -->
    <div class="card" style="margin-bottom: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <div class="card-title" style="margin-bottom: 0;">家庭状态</div>
        <button class="btn btn-secondary btn-sm" @click="showStatusModal = true">编辑状态</button>
      </div>
      <div class="grid-2" v-if="familyStatus">
        <div>
          <div style="font-size: 14px; color: var(--text-secondary);">本周待办</div>
          <div class="stat-number">{{ familyStatus.pending_count || 0 }}</div>
        </div>
        <div>
          <div style="font-size: 14px; color: var(--text-secondary);">本周已完成</div>
          <div class="stat-number" style="color: #16a34a;">{{ familyStatus.completed_count || 0 }}</div>
        </div>
      </div>
      <div v-else class="empty-state">
        <p>暂无家庭状态</p>
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
          <label>类型</label>
          <select v-model="filters.type" @change="loadRecords">
            <option value="">全部</option>
            <option value="task">任务</option>
            <option value="event">事件</option>
            <option value="shopping">购物</option>
            <option value="chore">家务</option>
          </select>
        </div>
        <div class="form-group">
          <label>状态</label>
          <select v-model="filters.status" @change="loadRecords">
            <option value="">全部</option>
            <option value="pending">待处理</option>
            <option value="in_progress">进行中</option>
            <option value="done">已完成</option>
          </select>
        </div>
        <button class="btn btn-primary btn-sm" @click="showRecordModal = true">+ 添加记录</button>
      </div>
    </div>

    <!-- Family records -->
    <div class="card">
      <div class="card-title">家庭记录</div>
      <div v-if="records.length === 0" class="empty-state">
        <p>暂无记录</p>
      </div>
      <div class="table-container" v-else>
        <table>
          <thead>
            <tr>
              <th>日期</th>
              <th>类型</th>
              <th>内容</th>
              <th>成员</th>
              <th>状态</th>
              <th>截止日期</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id">
              <td>{{ formatDate(r.date || r.created_at) }}</td>
              <td><span class="badge badge-primary">{{ typeLabel(r.type) }}</span></td>
              <td>{{ r.content || r.title || '-' }}</td>
              <td>{{ r.member_name || '-' }}</td>
              <td><span class="badge" :class="statusBadge(r.status)">{{ statusLabel(r.status) }}</span></td>
              <td>{{ formatDate(r.deadline) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Edit Status Modal -->
    <div class="modal-overlay" v-if="showStatusModal" @click.self="showStatusModal = false">
      <div class="modal">
        <h3>编辑家庭状态</h3>
        <div class="form-group">
          <label>本周待办数</label>
          <input v-model.number="statusForm.pending_count" type="number" />
        </div>
        <div class="form-group">
          <label>本周已完成数</label>
          <input v-model.number="statusForm.completed_count" type="number" />
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
        <h3>添加家庭记录</h3>
        <div class="form-group">
          <label>成员</label>
          <select v-model="recordForm.member_id">
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>类型</label>
          <select v-model="recordForm.type">
            <option value="task">任务</option>
            <option value="event">事件</option>
            <option value="shopping">购物</option>
            <option value="chore">家务</option>
          </select>
        </div>
        <div class="form-group">
          <label>内容</label>
          <input v-model="recordForm.content" type="text" placeholder="记录内容" />
        </div>
        <div class="form-group">
          <label>状态</label>
          <select v-model="recordForm.status">
            <option value="pending">待处理</option>
            <option value="in_progress">进行中</option>
            <option value="done">已完成</option>
          </select>
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
  name: 'Family',
  data() {
    return {
      familyStatus: null,
      records: [],
      members: [],
      showStatusModal: false,
      showRecordModal: false,
      filters: { member_id: '', type: '', status: '' },
      statusForm: { pending_count: 0, completed_count: 0, note: '' },
      recordForm: { member_id: '', type: 'task', content: '', status: 'pending', deadline: '', note: '' },
    }
  },
  async mounted() {
    await Promise.all([this.loadStatus(), this.loadRecords(), this.loadMembers()])
  },
  methods: {
    async loadStatus() {
      try {
        const res = await api.get('/api/family/status')
        this.familyStatus = res.data.status || res.data || null
      } catch {
        this.familyStatus = null
      }
    },
    async loadRecords() {
      try {
        const params = new URLSearchParams()
        if (this.filters.member_id) params.set('member_id', this.filters.member_id)
        if (this.filters.type) params.set('type', this.filters.type)
        if (this.filters.status) params.set('status', this.filters.status)
        const res = await api.get('/api/family/records', { params })
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
    async saveStatus() {
      try {
        await api.put('/api/family/status', this.statusForm)
        this.showStatusModal = false
        await this.loadStatus()
      } catch {
        // ignore
      }
    },
    async addRecord() {
      if (!this.recordForm.content) return
      try {
        await api.post('/api/family/records', this.recordForm)
        this.showRecordModal = false
        this.recordForm = { member_id: this.members.length > 0 ? this.members[0].id : '', type: 'task', content: '', status: 'pending', deadline: '', note: '' }
        await this.loadRecords()
      } catch {
        // ignore
      }
    },
    typeLabel(t) {
      const map = { task: '任务', event: '事件', shopping: '购物', chore: '家务' }
      return map[t] || t || '-'
    },
    statusLabel(s) {
      const map = { pending: '待处理', in_progress: '进行中', done: '已完成' }
      return map[s] || s || '-'
    },
    statusBadge(s) {
      const map = { pending: 'badge-yellow', in_progress: 'badge-blue', done: 'badge-green' }
      return map[s] || 'badge-primary'
    },
    formatDate(d) {
      if (!d) return '-'
      return new Date(d).toLocaleDateString('zh-CN')
    },
  },
}
</script>
