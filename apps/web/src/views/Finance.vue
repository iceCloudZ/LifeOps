<template>
  <div>
    <h1 class="page-title">财务管理</h1>

    <!-- Account cards -->
    <div class="card" style="margin-bottom: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <div class="card-title" style="margin-bottom: 0;">账户列表</div>
        <button class="btn btn-primary btn-sm" @click="showAccountModal = true">+ 添加账户</button>
      </div>
      <div class="grid-3" v-if="accounts.length > 0">
        <div class="card overview-card" v-for="a in accounts" :key="a.id" style="margin-bottom: 0;">
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div style="font-size: 14px; color: var(--text-secondary);">{{ a.name }}</div>
            <span v-if="memberName(a.member_id)" class="badge badge-blue" style="font-size: 11px;">{{ memberName(a.member_id) }}</span>
          </div>
          <div style="font-size: 24px; font-weight: 700; color: var(--primary); margin-top: 8px;">
            ¥{{ parseFloat(a.balance || 0).toFixed(2) }}
          </div>
          <div style="font-size: 12px; color: var(--text-secondary); margin-top: 4px;">
            {{ accountTypeLabel(a.type) }}
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <p>暂无账户</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="card" style="margin-bottom: 20px;">
      <div class="filter-bar">
        <div class="form-group">
          <label>开始日期</label>
          <input v-model="filters.from_date" type="date" @change="loadRecords" />
        </div>
        <div class="form-group">
          <label>结束日期</label>
          <input v-model="filters.to_date" type="date" @change="loadRecords" />
        </div>
        <div class="form-group">
          <label>分类</label>
          <input v-model="filters.category" type="text" placeholder="分类" @change="loadRecords" />
        </div>
        <div class="form-group">
          <label>成员</label>
          <select v-model="filters.member_id" @change="loadRecords">
            <option value="">全部</option>
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <button class="btn btn-primary btn-sm" @click="showRecordModal = true">+ 添加记录</button>
      </div>
    </div>

    <!-- Transaction records -->
    <div class="card">
      <div class="card-title">交易记录</div>
      <div v-if="records.length === 0" class="empty-state">
        <p>暂无记录</p>
      </div>
      <div class="table-container" v-else>
        <table>
          <thead>
            <tr>
              <th>日期</th>
              <th>类型</th>
              <th>分类</th>
              <th>金额</th>
              <th>成员</th>
              <th>备注</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in records" :key="r.id">
              <td>{{ formatDate(r.date || r.created_at) }}</td>
              <td><span class="badge" :class="r.type === 'income' ? 'badge-green' : 'badge-red'">{{ r.type === 'income' ? '收入' : '支出' }}</span></td>
              <td>{{ r.category || '-' }}</td>
              <td style="font-weight: 600;">¥{{ parseFloat(r.amount || 0).toFixed(2) }}</td>
              <td>{{ memberName(r.member_id) || '-' }}</td>
              <td>{{ r.note || r.description || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Add Account Modal -->
    <div class="modal-overlay" v-if="showAccountModal" @click.self="showAccountModal = false">
      <div class="modal">
        <h3>添加账户</h3>
        <div class="form-group">
          <label>账户名称</label>
          <input v-model="accountForm.name" type="text" placeholder="如: 工商银行储蓄卡" />
        </div>
        <div class="form-group">
          <label>账户类型</label>
          <select v-model="accountForm.type">
            <option value="savings">储蓄卡</option>
            <option value="credit">信用卡</option>
            <option value="cash">现金</option>
            <option value="investment">投资</option>
          </select>
        </div>
        <div class="form-group">
          <label>初始余额</label>
          <input v-model="accountForm.balance" type="number" step="0.01" placeholder="0.00" />
        </div>
        <div class="form-group">
          <label>归属成员</label>
          <select v-model="accountForm.member_id">
            <option value="">家庭共有</option>
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showAccountModal = false">取消</button>
          <button class="btn btn-primary" @click="addAccount">确定</button>
        </div>
      </div>
    </div>

    <!-- Add Record Modal -->
    <div class="modal-overlay" v-if="showRecordModal" @click.self="showRecordModal = false">
      <div class="modal">
        <h3>添加记录</h3>
        <div class="form-group">
          <label>类型</label>
          <select v-model="recordForm.type">
            <option value="expense">支出</option>
            <option value="income">收入</option>
          </select>
        </div>
        <div class="form-group">
          <label>金额</label>
          <input v-model="recordForm.amount" type="number" step="0.01" placeholder="0.00" />
        </div>
        <div class="form-group">
          <label>分类</label>
          <input v-model="recordForm.category" type="text" placeholder="如: 餐饮、交通" />
        </div>
        <div class="form-group">
          <label>成员</label>
          <select v-model="recordForm.member_id">
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
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
  name: 'Finance',
  data() {
    return {
      accounts: [],
      records: [],
      members: [],
      showAccountModal: false,
      showRecordModal: false,
      filters: {
        from_date: '',
        to_date: '',
        category: '',
        member_id: '',
      },
      accountForm: { name: '', type: 'savings', balance: '', member_id: '' },
      recordForm: { type: 'expense', amount: '', category: '', member_id: '', date: '', note: '' },
    }
  },
  async mounted() {
    await Promise.all([this.loadAccounts(), this.loadRecords(), this.loadMembers()])
  },
  methods: {
    async loadAccounts() {
      try {
        const res = await api.get('/api/finance/accounts')
        this.accounts = res.data.accounts || res.data || []
      } catch {
        this.accounts = []
      }
    },
    async loadRecords() {
      try {
        const params = new URLSearchParams()
        if (this.filters.from_date) params.set('from_date', this.filters.from_date)
        if (this.filters.to_date) params.set('to_date', this.filters.to_date)
        if (this.filters.category) params.set('category', this.filters.category)
        if (this.filters.member_id) params.set('member_id', this.filters.member_id)
        const res = await api.get('/api/finance/records', { params })
        this.records = res.data.records || res.data || []
      } catch {
        this.records = []
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
    async addAccount() {
      if (!this.accountForm.name) return
      try {
        await api.post('/api/finance/accounts', {
          name: this.accountForm.name,
          type: this.accountForm.type,
          balance: parseFloat(this.accountForm.balance) || 0,
          member_id: this.accountForm.member_id || null,
        })
        this.showAccountModal = false
        this.accountForm = { name: '', type: 'savings', balance: '', member_id: '' }
        await this.loadAccounts()
      } catch {
        // ignore
      }
    },
    async addRecord() {
      if (!this.recordForm.amount) return
      try {
        await api.post('/api/finance/records', {
          type: this.recordForm.type,
          amount: parseFloat(this.recordForm.amount),
          category: this.recordForm.category,
          member_id: this.recordForm.member_id,
          date: this.recordForm.date,
          note: this.recordForm.note,
        })
        this.showRecordModal = false
        this.recordForm = { type: 'expense', amount: '', category: '', member_id: '', date: '', note: '' }
        await this.loadRecords()
      } catch {
        // ignore
      }
    },
    formatDate(d) {
      if (!d) return '-'
      return new Date(d).toLocaleDateString('zh-CN')
    },
    memberName(id) {
      if (!id) return ''
      const m = this.members.find(m => m.id === id)
      return m ? m.name : ''
    },
    accountTypeLabel(t) {
      const map = { savings: '储蓄卡', credit: '信用卡', cash: '现金', investment: '投资' }
      return map[t] || t || '默认'
    },
  },
}
</script>
