<template>
  <div>
    <h1 class="page-title">运动记录</h1>

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
        <button class="btn btn-primary btn-sm" @click="showAddModal = true">+ 添加记录</button>
      </div>
    </div>

    <!-- Movement records -->
    <div class="card">
      <div class="card-title">运动记录</div>
      <div v-if="records.length === 0" class="empty-state">
        <p>暂无记录</p>
      </div>
      <div class="table-container" v-else>
        <table>
          <thead>
            <tr>
              <th>日期</th>
              <th>成员</th>
              <th>指标</th>
              <th>数值</th>
              <th>备注</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id">
              <td>{{ formatDate(r.record_date || r.created_at) }}</td>
              <td>{{ r.member_name || '-' }}</td>
              <td><span class="badge badge-blue">{{ metricLabel(r.metric) }}</span></td>
              <td>{{ r.value ? r.value + (r.unit ? ' ' + r.unit : '') : '-' }}</td>
              <td>{{ r.note || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Add Record Modal -->
    <div class="modal-overlay" v-if="showAddModal" @click.self="showAddModal = false">
      <div class="modal">
        <h3>添加运动记录</h3>
        <div class="form-group">
          <label>成员</label>
          <select v-model="form.member_id">
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>指标</label>
          <select v-model="form.metric">
            <option value="steps">步数</option>
            <option value="distance">距离</option>
            <option value="calories">消耗热量</option>
            <option value="duration">运动时长</option>
            <option value="heart_rate">心率</option>
            <option value="weight">体重</option>
          </select>
        </div>
        <div class="form-group">
          <label>数值</label>
          <input v-model="form.value" type="number" step="any" placeholder="数值" />
        </div>
        <div class="form-group">
          <label>单位</label>
          <select v-model="form.unit">
            <option value="步">步</option>
            <option value="公里">公里</option>
            <option value="米">米</option>
            <option value="千卡">千卡</option>
            <option value="分钟">分钟</option>
            <option value="小时">小时</option>
            <option value="bpm">bpm</option>
            <option value="kg">kg</option>
          </select>
        </div>
        <div class="form-group">
          <label>日期</label>
          <input v-model="form.record_date" type="date" />
        </div>
        <div class="form-group">
          <label>备注</label>
          <textarea v-model="form.note" placeholder="可选备注"></textarea>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showAddModal = false">取消</button>
          <button class="btn btn-primary" @click="addRecord">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../api.js'

export default {
  name: 'Movement',
  data() {
    return {
      records: [],
      members: [],
      showAddModal: false,
      filters: { member_id: '' },
      form: { member_id: '', metric: 'steps', value: '', unit: '步', record_date: '', note: '' },
    }
  },
  async mounted() {
    await Promise.all([this.loadMembers(), this.loadRecords()])
  },
  methods: {
    async loadRecords() {
      try {
        const params = new URLSearchParams()
        if (this.filters.member_id) params.set('member_id', this.filters.member_id)
        const res = await api.get('/api/movement/records', { params })
        this.records = res.data.records || res.data || []
      } catch {
        this.records = []
      }
    },
    async loadMembers() {
      try {
        const res = await api.get('/api/members')
        this.members = res.data.members || res.data || []
        if (this.members.length > 0 && !this.form.member_id) {
          this.form.member_id = this.members[0].id
        }
      } catch {
        this.members = []
      }
    },
    async addRecord() {
      if (!this.form.value) return
      try {
        await api.post('/api/movement/records', this.form)
        this.showAddModal = false
        this.form = { member_id: this.members.length > 0 ? this.members[0].id : '', metric: 'steps', value: '', unit: '步', record_date: '', note: '' }
        await this.loadRecords()
      } catch {
        // ignore
      }
    },
    metricLabel(m) {
      const map = { steps: '步数', distance: '距离', calories: '消耗热量', duration: '运动时长', heart_rate: '心率', weight: '体重' }
      return map[m] || m || '-'
    },
    formatDate(d) {
      if (!d) return '-'
      return new Date(d).toLocaleDateString('zh-CN')
    },
  },
}
</script>
