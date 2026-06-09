<template>
  <div v-if="member">
    <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 20px;">
      <button class="btn btn-secondary btn-sm" @click="$router.push('/members')">← 返回</button>
      <h1 class="page-title" style="margin: 0;">{{ member.name }}</h1>
      <span class="badge" :class="roleBadge(member.role)">{{ roleLabel(member.role) }}</span>
    </div>

    <!-- 基本信息 -->
    <div class="card" style="margin-bottom: 16px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
        <div class="card-title" style="margin: 0;">基本信息</div>
        <button class="btn btn-secondary btn-sm" @click="editBasic">编辑</button>
      </div>
      <div class="info-row">
        <span class="info-label">生日</span>
        <span>{{ member.birth_date || '未填写' }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">角色</span>
        <span>{{ roleLabel(member.role) }}</span>
      </div>
    </div>

    <!-- 工作档案 -->
    <div class="card" style="margin-bottom: 16px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
        <div class="card-title" style="margin: 0;">工作档案</div>
        <button class="btn btn-secondary btn-sm" @click="editWork">编辑</button>
      </div>
      <div v-if="hasProfile">
        <div class="info-row"><span class="info-label">就业状态</span><span>{{ profile.employment_status || '-' }}</span></div>
        <div class="info-row"><span class="info-label">公司</span><span>{{ profile.company || '-' }}</span></div>
        <div class="info-row"><span class="info-label">职位</span><span>{{ profile.position || '-' }}</span></div>
        <div class="info-row"><span class="info-label">行业</span><span>{{ profile.industry || '-' }}</span></div>
        <div class="info-row"><span class="info-label">工作地点</span><span>{{ profile.work_location || '-' }}</span></div>
        <div class="info-row"><span class="info-label">收入范围</span><span>{{ profile.income_range || '-' }}</span></div>
        <div class="info-row"><span class="info-label">班制</span><span>{{ profile.work_schedule || '-' }}</span></div>
        <div class="info-row"><span class="info-label">通勤</span><span>{{ profile.commute_minutes ? profile.commute_minutes + '分钟' : '-' }}</span></div>
        <div class="info-row"><span class="info-label">入职时间</span><span>{{ profile.started_at || '-' }}</span></div>
        <div v-if="profile.note" class="info-row"><span class="info-label">备注</span><span>{{ profile.note }}</span></div>
      </div>
      <div v-else class="empty-hint">未填写，点击编辑添加</div>
    </div>

    <!-- 健康记录 -->
    <div class="card" style="margin-bottom: 16px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
        <div class="card-title" style="margin: 0;">健康记录 <span class="count-badge">{{ healthRecords.length }}</span></div>
        <button class="btn btn-secondary btn-sm" @click="openAddRecord('health')">+ 添加</button>
      </div>
      <div v-if="healthRecords.length === 0" class="empty-hint">暂无记录</div>
      <div class="table-container" v-else>
        <table>
          <thead><tr><th>日期</th><th>指标</th><th>值</th><th>备注</th></tr></thead>
          <tbody>
            <tr v-for="r in healthRecords" :key="r.id">
              <td>{{ formatDate(r.record_date) }}</td>
              <td>{{ r.metric || r.type || '-' }}</td>
              <td>{{ (r.value || '-') + (r.unit ? r.unit : '') }}</td>
              <td>{{ r.note || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 运动记录 -->
    <div class="card" style="margin-bottom: 16px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
        <div class="card-title" style="margin: 0;">运动记录 <span class="count-badge">{{ movementRecords.length }}</span></div>
        <button class="btn btn-secondary btn-sm" @click="openAddRecord('movement')">+ 添加</button>
      </div>
      <div v-if="movementRecords.length === 0" class="empty-hint">暂无记录</div>
      <div class="table-container" v-else>
        <table>
          <thead><tr><th>日期</th><th>指标</th><th>值</th><th>备注</th></tr></thead>
          <tbody>
            <tr v-for="r in movementRecords" :key="r.id">
              <td>{{ formatDate(r.record_date) }}</td>
              <td>{{ r.metric || '-' }}</td>
              <td>{{ (r.value || '-') + (r.unit ? r.unit : '') }}</td>
              <td>{{ r.note || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 财务记录 -->
    <div class="card" style="margin-bottom: 16px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
        <div class="card-title" style="margin: 0;">财务记录 <span class="count-badge">{{ financeRecords.length }}</span></div>
      </div>
      <div v-if="financeRecords.length === 0" class="empty-hint">暂无记录</div>
      <div class="table-container" v-else>
        <table>
          <thead><tr><th>日期</th><th>类型</th><th>分类</th><th>金额</th><th>备注</th></tr></thead>
          <tbody>
            <tr v-for="r in financeRecords" :key="r.id">
              <td>{{ formatDate(r.record_date) }}</td>
              <td><span class="badge" :class="r.type === 'income' ? 'badge-green' : 'badge-red'">{{ r.type === 'income' ? '收入' : '支出' }}</span></td>
              <td>{{ r.category || '-' }}</td>
              <td style="font-weight: 600;">¥{{ parseFloat(r.amount || 0).toFixed(2) }}</td>
              <td>{{ r.note || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Edit Basic Info Modal -->
    <div class="modal-overlay" v-if="showBasicModal" @click.self="showBasicModal = false">
      <div class="modal">
        <h3>编辑基本信息</h3>
        <div class="form-group"><label>姓名</label><input v-model="basicForm.name" type="text" /></div>
        <div class="form-group"><label>角色</label>
          <select v-model="basicForm.role">
            <option value="dad">爸爸</option><option value="mom">妈妈</option>
            <option value="son">儿子</option><option value="daughter">女儿</option>
            <option value="grandpa">爷爷/外公</option><option value="grandma">奶奶/外婆</option>
          </select>
        </div>
        <div class="form-group"><label>生日</label><input v-model="basicForm.birth_date" type="date" /></div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showBasicModal = false">取消</button>
          <button class="btn btn-primary" @click="saveBasic">保存</button>
        </div>
      </div>
    </div>

    <!-- Edit Work Profile Modal -->
    <div class="modal-overlay" v-if="showWorkModal" @click.self="showWorkModal = false">
      <div class="modal">
        <h3>编辑工作档案</h3>
        <div class="form-group"><label>就业状态</label>
          <select v-model="workForm.employment_status">
            <option value="">未填写</option><option value="全职">全职</option><option value="兼职">兼职</option>
            <option value="自由职业">自由职业</option><option value="待业">待业</option><option value="退休">退休</option><option value="学生">学生</option>
          </select>
        </div>
        <div class="form-group"><label>公司</label><input v-model="workForm.company" type="text" placeholder="公司名称" /></div>
        <div class="form-group"><label>职位</label><input v-model="workForm.position" type="text" /></div>
        <div class="form-group"><label>行业</label><input v-model="workForm.industry" type="text" /></div>
        <div class="form-group"><label>工作地点</label><input v-model="workForm.work_location" type="text" /></div>
        <div class="form-group"><label>收入范围</label>
          <select v-model="workForm.income_range">
            <option value="">未填写</option><option value="10万以下">10万以下</option><option value="10-20万">10-20万</option>
            <option value="20-30万">20-30万</option><option value="30-50万">30-50万</option><option value="50-100万">50-100万</option><option value="100万以上">100万以上</option>
          </select>
        </div>
        <div class="form-group"><label>班制</label>
          <select v-model="workForm.work_schedule">
            <option value="">未填写</option><option value="双休">双休</option><option value="单休">单休</option>
            <option value="大小周">大小周</option><option value="弹性">弹性</option><option value="排班制">排班制</option>
          </select>
        </div>
        <div class="form-group"><label>通勤（分钟）</label><input v-model.number="workForm.commute_minutes" type="number" /></div>
        <div class="form-group"><label>入职时间</label><input v-model="workForm.started_at" type="text" placeholder="如 2018-07" /></div>
        <div class="form-group"><label>备注</label><textarea v-model="workForm.note"></textarea></div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showWorkModal = false">取消</button>
          <button class="btn btn-primary" @click="saveWork">保存</button>
        </div>
      </div>
    </div>

    <!-- Add Record Modal -->
    <div class="modal-overlay" v-if="showRecordModal" @click.self="showRecordModal = false">
      <div class="modal">
        <h3>添加{{ recordType === 'health' ? '健康' : '运动' }}记录</h3>
        <div class="form-group"><label>日期</label><input v-model="recordForm.record_date" type="date" /></div>
        <div class="form-group"><label>指标</label><input v-model="recordForm.metric" type="text" :placeholder="recordType === 'health' ? '如：血压、血糖、体重' : '如：步数、距离、时长'" /></div>
        <div class="form-group"><label>数值</label><input v-model="recordForm.value" type="text" placeholder="如：120/80、5.2" /></div>
        <div class="form-group"><label>单位</label><input v-model="recordForm.unit" type="text" placeholder="如：mmHg、km、分钟" /></div>
        <div v-if="recordType === 'health'" class="form-group"><label>类型</label>
          <select v-model="recordForm.type"><option value="checkup">体检</option><option value="daily">日常</option><option value="medication">用药</option></select>
        </div>
        <div class="form-group"><label>备注</label><textarea v-model="recordForm.note"></textarea></div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showRecordModal = false">取消</button>
          <button class="btn btn-primary" @click="saveRecord">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../api.js'

export default {
  name: 'MemberDetail',
  data() {
    return {
      member: null,
      profile: {},
      healthRecords: [],
      movementRecords: [],
      financeRecords: [],
      showBasicModal: false,
      showWorkModal: false,
      showRecordModal: false,
      recordType: 'health',
      basicForm: { name: '', role: 'dad', birth_date: '' },
      workForm: { employment_status: '', company: '', position: '', industry: '', work_location: '', income_range: '', work_schedule: '', commute_minutes: 0, started_at: '', note: '' },
      recordForm: { record_date: '', metric: '', value: '', unit: '', type: 'daily', note: '' },
    }
  },
  computed: {
    hasProfile() {
      if (!this.profile) return false
      return this.profile.employment_status || this.profile.company || this.profile.position || this.profile.industry || this.profile.work_location || this.profile.income_range
    },
  },
  async mounted() {
    await this.loadAll()
  },
  methods: {
    async loadAll() {
      const id = this.$route.params.id
      await Promise.all([
        this.loadMember(id),
        this.loadProfile(id),
        this.loadHealth(id),
        this.loadMovement(id),
        this.loadFinance(id),
      ])
    },
    async loadMember(id) {
      try {
        const res = await api.get(`/api/members/${id}`)
        this.member = res.data
      } catch { this.member = null }
    },
    async loadProfile(id) {
      try {
        const res = await api.get(`/api/work/profiles/${id}`)
        this.profile = res.data || {}
      } catch { this.profile = {} }
    },
    async loadHealth(id) {
      try {
        const res = await api.get('/api/health/records', { params: { member_id: id } })
        this.healthRecords = res.data.records || res.data || []
      } catch { this.healthRecords = [] }
    },
    async loadMovement(id) {
      try {
        const res = await api.get('/api/movement/records', { params: { member_id: id } })
        this.movementRecords = res.data.records || res.data || []
      } catch { this.movementRecords = [] }
    },
    async loadFinance(id) {
      try {
        const res = await api.get('/api/finance/records', { params: { member_id: id } })
        this.financeRecords = res.data.records || res.data || []
      } catch { this.financeRecords = [] }
    },
    editBasic() {
      this.basicForm = { name: this.member.name, role: this.member.role, birth_date: this.member.birth_date || '' }
      this.showBasicModal = true
    },
    async saveBasic() {
      try {
        await api.put(`/api/members/${this.member.id}`, this.basicForm)
        this.showBasicModal = false
        await this.loadMember(this.member.id)
      } catch {}
    },
    editWork() {
      const p = this.profile || {}
      this.workForm = {
        employment_status: p.employment_status || '', company: p.company || '', position: p.position || '',
        industry: p.industry || '', work_location: p.work_location || '', income_range: p.income_range || '',
        work_schedule: p.work_schedule || '', commute_minutes: p.commute_minutes || 0,
        started_at: p.started_at || '', note: p.note || '',
      }
      this.showWorkModal = true
    },
    async saveWork() {
      try {
        await api.put(`/api/work/profiles/${this.member.id}`, this.workForm)
        this.showWorkModal = false
        await this.loadProfile(this.member.id)
      } catch {}
    },
    openAddRecord(type) {
      this.recordType = type
      this.recordForm = { record_date: new Date().toISOString().slice(0, 10), metric: '', value: '', unit: '', type: 'daily', note: '' }
      this.showRecordModal = true
    },
    async saveRecord() {
      try {
        const data = {
          member_id: this.member.id,
          metric: this.recordForm.metric || null,
          value: this.recordForm.value || null,
          unit: this.recordForm.unit || null,
          note: this.recordForm.note || '',
          record_date: this.recordForm.record_date,
        }
        if (this.recordType === 'health') {
          data.type = this.recordForm.type
          await api.post('/api/health/records', data)
          await this.loadHealth(this.member.id)
        } else {
          await api.post('/api/movement/records', data)
          await this.loadMovement(this.member.id)
        }
        this.showRecordModal = false
      } catch {}
    },
    roleLabel(r) {
      const map = { dad: '爸爸', mom: '妈妈', son: '儿子', daughter: '女儿', grandpa: '爷爷/外公', grandma: '奶奶/外婆' }
      return map[r] || r || '-'
    },
    roleBadge(r) {
      const map = { dad: 'badge-blue', mom: 'badge-pink', son: 'badge-green', daughter: 'badge-green', grandpa: 'badge-primary', grandma: 'badge-primary' }
      return map[r] || 'badge-primary'
    },
    formatDate(d) {
      if (!d) return '-'
      return new Date(d).toLocaleDateString('zh-CN')
    },
  },
}
</script>

<style scoped>
.info-row {
  display: flex;
  padding: 6px 0;
  font-size: 14px;
  border-bottom: 1px solid var(--border);
}
.info-row:last-child { border-bottom: none; }
.info-label {
  width: 80px;
  color: var(--text-secondary);
  flex-shrink: 0;
}
.empty-hint {
  font-size: 13px;
  color: var(--text-secondary);
  padding: 8px 0;
}
.count-badge {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 400;
}
</style>
