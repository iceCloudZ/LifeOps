<template>
  <div>
    <h1 class="page-title">健康管理</h1>

    <!-- Health profiles -->
    <div class="card" style="margin-bottom: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <div class="card-title" style="margin-bottom: 0;">成员健康档案</div>
      </div>
      <div class="grid-3" v-if="profiles.length > 0">
        <div class="card overview-card" v-for="p in profiles" :key="p.member_id" style="margin-bottom: 0;">
          <div style="font-size: 16px; font-weight: 600;">{{ p.member_name || '成员' }}</div>
          <div style="margin-top: 8px; font-size: 14px; color: var(--text-secondary);">
            <div v-if="p.age">年龄: {{ p.age }}</div>
            <div v-if="p.blood_type">血型: {{ p.blood_type }}</div>
            <div v-if="p.allergies">过敏: {{ p.allergies }}</div>
            <div v-if="p.conditions">病史: {{ p.conditions }}</div>
          </div>
          <button class="btn btn-secondary btn-sm" style="margin-top: 12px;" @click="editProfile(p)">编辑</button>
        </div>
      </div>
      <div v-else class="empty-state">
        <p>暂无健康档案</p>
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
            <option value="checkup">体检</option>
            <option value="medication">用药</option>
            <option value="exercise">运动</option>
            <option value="diet">饮食</option>
            <option value="sleep">睡眠</option>
          </select>
        </div>
        <button class="btn btn-primary btn-sm" @click="showRecordModal = true">+ 添加记录</button>
      </div>
    </div>

    <!-- Health records -->
    <div class="card">
      <div class="card-title">健康记录</div>
      <div v-if="records.length === 0" class="empty-state">
        <p>暂无记录</p>
      </div>
      <div class="table-container" v-else>
        <table>
          <thead>
            <tr>
              <th>日期</th>
              <th>成员</th>
              <th>类型</th>
              <th>内容</th>
              <th>备注</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id">
              <td>{{ formatDate(r.date || r.created_at) }}</td>
              <td>{{ r.member_name || '-' }}</td>
              <td><span class="badge badge-blue">{{ typeLabel(r.type) }}</span></td>
              <td>{{ r.content || r.title || '-' }}</td>
              <td>{{ r.note || r.description || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Edit Profile Modal -->
    <div class="modal-overlay" v-if="showProfileModal" @click.self="showProfileModal = false">
      <div class="modal">
        <h3>编辑健康档案</h3>
        <div class="form-group">
          <label>年龄</label>
          <input v-model="profileForm.age" type="number" placeholder="年龄" />
        </div>
        <div class="form-group">
          <label>血型</label>
          <select v-model="profileForm.blood_type">
            <option value="">未知</option>
            <option value="A">A</option>
            <option value="B">B</option>
            <option value="AB">AB</option>
            <option value="O">O</option>
          </select>
        </div>
        <div class="form-group">
          <label>过敏信息</label>
          <textarea v-model="profileForm.allergies" placeholder="如无过敏可留空"></textarea>
        </div>
        <div class="form-group">
          <label>既往病史</label>
          <textarea v-model="profileForm.conditions" placeholder="如无病史可留空"></textarea>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showProfileModal = false">取消</button>
          <button class="btn btn-primary" @click="saveProfile">保存</button>
        </div>
      </div>
    </div>

    <!-- Add Record Modal -->
    <div class="modal-overlay" v-if="showRecordModal" @click.self="showRecordModal = false">
      <div class="modal">
        <h3>添加健康记录</h3>
        <div class="form-group">
          <label>成员</label>
          <select v-model="recordForm.member_id">
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>类型</label>
          <select v-model="recordForm.type">
            <option value="checkup">体检</option>
            <option value="medication">用药</option>
            <option value="exercise">运动</option>
            <option value="diet">饮食</option>
            <option value="sleep">睡眠</option>
          </select>
        </div>
        <div class="form-group">
          <label>内容</label>
          <input v-model="recordForm.content" type="text" placeholder="记录内容" />
        </div>
        <div class="form-group">
          <label>日期</label>
          <input v-model="recordForm.date" type="date" />
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
  name: 'Health',
  data() {
    return {
      profiles: [],
      records: [],
      members: [],
      showProfileModal: false,
      showRecordModal: false,
      filters: { member_id: '', type: '' },
      profileForm: { member_id: '', age: '', blood_type: '', allergies: '', conditions: '' },
      recordForm: { member_id: '', type: 'checkup', content: '', date: '', note: '' },
    }
  },
  async mounted() {
    await Promise.all([this.loadProfiles(), this.loadRecords(), this.loadMembers()])
  },
  methods: {
    async loadProfiles() {
      try {
        const res = await api.get('/api/health/profiles')
        this.profiles = res.data.profiles || res.data || []
      } catch {
        this.profiles = []
      }
    },
    async loadRecords() {
      try {
        const params = new URLSearchParams()
        if (this.filters.member_id) params.set('member_id', this.filters.member_id)
        if (this.filters.type) params.set('type', this.filters.type)
        const res = await api.get('/api/health/records', { params })
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
    editProfile(p) {
      this.profileForm = {
        member_id: p.member_id,
        age: p.age || '',
        blood_type: p.blood_type || '',
        allergies: p.allergies || '',
        conditions: p.conditions || '',
      }
      this.showProfileModal = true
    },
    async saveProfile() {
      try {
        await api.put(`/api/health/profiles/${this.profileForm.member_id}`, this.profileForm)
        this.showProfileModal = false
        await this.loadProfiles()
      } catch {
        // ignore
      }
    },
    async addRecord() {
      if (!this.recordForm.content) return
      try {
        await api.post('/api/health/records', this.recordForm)
        this.showRecordModal = false
        this.recordForm = { member_id: this.members.length > 0 ? this.members[0].id : '', type: 'checkup', content: '', date: '', note: '' }
        await this.loadRecords()
      } catch {
        // ignore
      }
    },
    typeLabel(t) {
      const map = { checkup: '体检', medication: '用药', exercise: '运动', diet: '饮食', sleep: '睡眠' }
      return map[t] || t || '-'
    },
    formatDate(d) {
      if (!d) return '-'
      return new Date(d).toLocaleDateString('zh-CN')
    },
  },
}
</script>
