<template>
  <div>
    <h1 class="page-title">仪表盘</h1>

    <!-- Chat input -->
    <div class="card" style="margin-bottom: 20px;">
      <form @submit.prevent="sendChat" style="display: flex; gap: 8px;">
        <input
          v-model="chatInput"
          type="text"
          placeholder="问问管家..."
          style="flex: 1;"
        />
        <button type="submit" class="btn btn-primary">发送</button>
      </form>
    </div>

    <!-- Overview cards -->
    <div class="grid-4" style="margin-bottom: 20px;">
      <div class="card overview-card">
        <div class="card-title">财务概览</div>
        <div class="stat-number">{{ financeTotal }}</div>
        <div class="stat-label">总余额</div>
      </div>
      <div class="card overview-card">
        <div class="card-title">健康概览</div>
        <div class="stat-number">{{ healthCount }}</div>
        <div class="stat-label">成员数</div>
      </div>
      <div class="card overview-card">
        <div class="card-title">工作状态</div>
        <div class="stat-number">{{ workCount }}</div>
        <div class="stat-label">进行中</div>
      </div>
      <div class="card overview-card">
        <div class="card-title">家庭事务</div>
        <div class="stat-number">{{ familyCount }}</div>
        <div class="stat-label">待处理</div>
      </div>
    </div>

    <!-- Quick entry -->
    <div class="card" style="margin-bottom: 20px;">
      <div class="card-title">快速记录</div>
      <form @submit.prevent="quickEntry" class="inline-form">
        <div class="form-group">
          <label>领域</label>
          <select v-model="entry.domain">
            <option value="finance">财务</option>
            <option value="health">健康</option>
            <option value="work">工作</option>
            <option value="family">家庭</option>
            <option value="note">笔记</option>
          </select>
        </div>
        <div class="form-group">
          <label>成员</label>
          <select v-model="entry.member_id">
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="form-group" style="flex: 2;">
          <label>内容</label>
          <input v-model="entry.content" type="text" placeholder="输入内容..." />
        </div>
        <button type="submit" class="btn btn-primary">提交</button>
      </form>
    </div>

    <!-- Recent records -->
    <div class="card">
      <div class="card-title">最近记录</div>
      <div v-if="recentRecords.length === 0" class="empty-state">
        <p>暂无记录</p>
      </div>
      <div class="table-container" v-else>
        <table>
          <thead>
            <tr>
              <th>时间</th>
              <th>领域</th>
              <th>成员</th>
              <th>内容</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in recentRecords" :key="r.id">
              <td>{{ formatTime(r.created_at) }}</td>
              <td><span class="badge badge-primary">{{ r.domain || '-' }}</span></td>
              <td>{{ r.member_name || '-' }}</td>
              <td>{{ r.content || r.title || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../api.js'

export default {
  name: 'Dashboard',
  data() {
    return {
      chatInput: '',
      financeTotal: '¥0.00',
      healthCount: 0,
      workCount: 0,
      familyCount: 0,
      members: [],
      recentRecords: [],
      entry: {
        domain: 'family',
        member_id: '',
        content: '',
      },
    }
  },
  async mounted() {
    await Promise.all([
      this.loadFinance(),
      this.loadHealth(),
      this.loadWork(),
      this.loadFamily(),
      this.loadMembers(),
    ])
  },
  methods: {
    async sendChat() {
      if (!this.chatInput.trim()) return
      const text = this.chatInput
      this.chatInput = ''
      try {
        const res = await api.post('/api/chat/conversations', { title: text })
        const convId = res.data.id || res.data.conversation_id
        if (convId) {
          this.$router.push({ name: 'Chat', query: { conv: convId, msg: text } })
        }
      } catch {
        this.chatInput = text
      }
    },
    async loadFinance() {
      try {
        const res = await api.get('/api/finance/accounts')
        const accounts = res.data.accounts || res.data || []
        const total = accounts.reduce((sum, a) => sum + (parseFloat(a.balance) || 0), 0)
        this.financeTotal = '¥' + total.toFixed(2)
      } catch {
        this.financeTotal = '¥0.00'
      }
    },
    async loadHealth() {
      try {
        const res = await api.get('/api/health/profiles')
        const profiles = res.data.profiles || res.data || []
        this.healthCount = profiles.length
      } catch {
        this.healthCount = 0
      }
    },
    async loadWork() {
      try {
        const res = await api.get('/api/work/status')
        const items = res.data.statuses || res.data || []
        this.workCount = Array.isArray(items) ? items.length : 0
      } catch {
        this.workCount = 0
      }
    },
    async loadFamily() {
      try {
        const res = await api.get('/api/family/status')
        const items = res.data.records || res.data || []
        this.familyCount = Array.isArray(items) ? items.length : 0
      } catch {
        this.familyCount = 0
      }
    },
    async loadMembers() {
      try {
        const res = await api.get('/api/members')
        this.members = res.data.members || res.data || []
        if (this.members.length > 0 && !this.entry.member_id) {
          this.entry.member_id = this.members[0].id
        }
      } catch {
        this.members = []
      }
    },
    async quickEntry() {
      if (!this.entry.content.trim()) return
      try {
        const domain = this.entry.domain
        if (domain === 'note') {
          await api.post('/api/notes', {
            title: this.entry.content,
            content: this.entry.content,
            member_id: this.entry.member_id,
          })
        } else {
          await api.post(`/api/${domain}/records`, {
            member_id: this.entry.member_id,
            content: this.entry.content,
            title: this.entry.content,
          })
        }
        this.entry.content = ''
      } catch {
        // ignore
      }
    },
    formatTime(t) {
      if (!t) return '-'
      return new Date(t).toLocaleString('zh-CN')
    },
  },
}
</script>
