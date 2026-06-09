<template>
  <div>
    <h1 class="page-title">家庭成员</h1>

    <div style="display: flex; justify-content: flex-end; margin-bottom: 16px;">
      <button class="btn btn-primary btn-sm" @click="openAdd">+ 添加成员</button>
    </div>

    <div v-if="members.length === 0" class="empty-state card">
      <p>暂无成员</p>
    </div>

    <div class="grid-2" v-else>
      <div class="card" v-for="m in members" :key="m.id" style="margin-bottom: 0;">
        <!-- 基本信息 -->
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
          <div>
            <router-link :to="'/members/' + m.id" style="font-size: 18px; font-weight: 600; color: var(--primary); text-decoration: none;">{{ m.name }}</router-link>
            <span class="badge" :class="roleBadge(m.role)" style="margin-left: 8px;">{{ roleLabel(m.role) }}</span>
          </div>
          <div style="display: flex; gap: 6px;">
            <button class="btn btn-secondary btn-sm" @click="openEdit(m)">编辑</button>
            <button class="btn btn-danger btn-sm" @click="deleteMember(m.id)">删除</button>
          </div>
        </div>
        <div style="font-size: 14px; color: var(--text-secondary);">
          生日：{{ m.birth_date || '未填写' }}
        </div>

        <!-- 工作档案 -->
        <div style="margin-top: 12px; border-top: 1px solid var(--border); padding-top: 12px;">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
            <span style="font-size: 14px; font-weight: 600; color: var(--text-secondary);">工作档案</span>
            <button class="btn btn-secondary btn-sm" @click="openWorkProfile(m)">编辑</button>
          </div>
          <div v-if="getProfile(m.id)">
            <div class="profile-grid">
              <div v-if="getProfile(m.id).employment_status" class="profile-item">
                <span class="profile-label">状态</span>
                <span>{{ getProfile(m.id).employment_status }}</span>
              </div>
              <div v-if="getProfile(m.id).company" class="profile-item">
                <span class="profile-label">公司</span>
                <span>{{ getProfile(m.id).company }}</span>
              </div>
              <div v-if="getProfile(m.id).position" class="profile-item">
                <span class="profile-label">职位</span>
                <span>{{ getProfile(m.id).position }}</span>
              </div>
              <div v-if="getProfile(m.id).industry" class="profile-item">
                <span class="profile-label">行业</span>
                <span>{{ getProfile(m.id).industry }}</span>
              </div>
              <div v-if="getProfile(m.id).work_location" class="profile-item">
                <span class="profile-label">地点</span>
                <span>{{ getProfile(m.id).work_location }}</span>
              </div>
              <div v-if="getProfile(m.id).income_range" class="profile-item">
                <span class="profile-label">收入</span>
                <span>{{ getProfile(m.id).income_range }}</span>
              </div>
              <div v-if="getProfile(m.id).work_schedule" class="profile-item">
                <span class="profile-label">班制</span>
                <span>{{ getProfile(m.id).work_schedule }}</span>
              </div>
              <div v-if="getProfile(m.id).commute_minutes" class="profile-item">
                <span class="profile-label">通勤</span>
                <span>{{ getProfile(m.id).commute_minutes }}分钟</span>
              </div>
              <div v-if="getProfile(m.id).started_at" class="profile-item">
                <span class="profile-label">入职</span>
                <span>{{ getProfile(m.id).started_at }}</span>
              </div>
            </div>
          </div>
          <div v-else style="font-size: 13px; color: var(--text-secondary);">未填写</div>
        </div>
      </div>
    </div>

    <!-- Member Modal -->
    <div class="modal-overlay" v-if="showModal" @click.self="showModal = false">
      <div class="modal">
        <h3>{{ isEditing ? '编辑成员' : '添加成员' }}</h3>
        <div class="form-group">
          <label>姓名</label>
          <input v-model="form.name" type="text" placeholder="成员姓名" />
        </div>
        <div class="form-group">
          <label>角色</label>
          <select v-model="form.role">
            <option value="dad">爸爸</option>
            <option value="mom">妈妈</option>
            <option value="son">儿子</option>
            <option value="daughter">女儿</option>
            <option value="grandpa">爷爷/外公</option>
            <option value="grandma">奶奶/外婆</option>
          </select>
        </div>
        <div class="form-group">
          <label>生日</label>
          <input v-model="form.birth_date" type="date" />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showModal = false">取消</button>
          <button class="btn btn-primary" @click="saveMember">{{ isEditing ? '保存' : '确定' }}</button>
        </div>
      </div>
    </div>

    <!-- Work Profile Modal -->
    <div class="modal-overlay" v-if="showWorkModal" @click.self="showWorkModal = false">
      <div class="modal">
        <h3>编辑工作档案 - {{ workForm.memberName }}</h3>
        <div class="form-group">
          <label>就业状态</label>
          <select v-model="workForm.employment_status">
            <option value="">未填写</option>
            <option value="全职">全职</option>
            <option value="兼职">兼职</option>
            <option value="自由职业">自由职业</option>
            <option value="待业">待业</option>
            <option value="退休">退休</option>
            <option value="学生">学生</option>
          </select>
        </div>
        <div class="form-group">
          <label>公司</label>
          <input v-model="workForm.company" type="text" placeholder="公司名称" />
        </div>
        <div class="form-group">
          <label>职位</label>
          <input v-model="workForm.position" type="text" placeholder="如：高级工程师" />
        </div>
        <div class="form-group">
          <label>行业</label>
          <input v-model="workForm.industry" type="text" placeholder="如：互联网/软件" />
        </div>
        <div class="form-group">
          <label>工作地点</label>
          <input v-model="workForm.work_location" type="text" placeholder="如：上海" />
        </div>
        <div class="form-group">
          <label>收入范围</label>
          <select v-model="workForm.income_range">
            <option value="">未填写</option>
            <option value="10万以下">10万以下</option>
            <option value="10-20万">10-20万</option>
            <option value="20-30万">20-30万</option>
            <option value="30-50万">30-50万</option>
            <option value="50-100万">50-100万</option>
            <option value="100万以上">100万以上</option>
          </select>
        </div>
        <div class="form-group">
          <label>班制</label>
          <select v-model="workForm.work_schedule">
            <option value="">未填写</option>
            <option value="双休">双休</option>
            <option value="单休">单休</option>
            <option value="大小周">大小周</option>
            <option value="弹性">弹性</option>
            <option value="排班制">排班制</option>
          </select>
        </div>
        <div class="form-group">
          <label>通勤时间（分钟）</label>
          <input v-model.number="workForm.commute_minutes" type="number" placeholder="0" />
        </div>
        <div class="form-group">
          <label>入职时间</label>
          <input v-model="workForm.started_at" type="text" placeholder="如：2018-07" />
        </div>
        <div class="form-group">
          <label>备注</label>
          <textarea v-model="workForm.note" placeholder="其他补充"></textarea>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showWorkModal = false">取消</button>
          <button class="btn btn-primary" @click="saveWorkProfile">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../api.js'

export default {
  name: 'Members',
  data() {
    return {
      members: [],
      profiles: {},
      showModal: false,
      showWorkModal: false,
      isEditing: false,
      editingId: null,
      form: { name: '', role: 'dad', birth_date: '' },
      workForm: { member_id: '', memberName: '', employment_status: '', company: '', position: '', industry: '', work_location: '', income_range: '', work_schedule: '', commute_minutes: 0, started_at: '', note: '' },
    }
  },
  async mounted() {
    await Promise.all([this.loadMembers(), this.loadProfiles()])
  },
  methods: {
    async loadMembers() {
      try {
        const res = await api.get('/api/members')
        this.members = res.data.members || res.data || []
      } catch {
        this.members = []
      }
    },
    async loadProfiles() {
      try {
        const res = await api.get('/api/work/profiles')
        const list = res.data.profiles || res.data || []
        const map = {}
        for (const p of list) {
          map[p.member_id] = p
        }
        this.profiles = map
      } catch {
        this.profiles = {}
      }
    },
    getProfile(memberId) {
      const p = this.profiles[memberId]
      if (!p) return null
      const hasData = p.employment_status || p.company || p.position || p.industry || p.work_location || p.income_range
      return hasData ? p : null
    },
    openAdd() {
      this.isEditing = false
      this.editingId = null
      this.form = { name: '', role: 'dad', birth_date: '' }
      this.showModal = true
    },
    openEdit(m) {
      this.isEditing = true
      this.editingId = m.id
      this.form = {
        name: m.name || '',
        role: m.role || 'dad',
        birth_date: m.birth_date || '',
      }
      this.showModal = true
    },
    async saveMember() {
      if (!this.form.name) return
      try {
        if (this.isEditing) {
          await api.put(`/api/members/${this.editingId}`, this.form)
        } else {
          await api.post('/api/members', this.form)
        }
        this.showModal = false
        await this.loadMembers()
      } catch {
        // ignore
      }
    },
    async deleteMember(id) {
      if (!confirm('确定删除此成员？')) return
      try {
        await api.delete(`/api/members/${id}`)
        await this.loadMembers()
        await this.loadProfiles()
      } catch {
        // ignore
      }
    },
    openWorkProfile(m) {
      const existing = this.profiles[m.id] || {}
      this.workForm = {
        member_id: m.id,
        memberName: m.name,
        employment_status: existing.employment_status || '',
        company: existing.company || '',
        position: existing.position || '',
        industry: existing.industry || '',
        work_location: existing.work_location || '',
        income_range: existing.income_range || '',
        work_schedule: existing.work_schedule || '',
        commute_minutes: existing.commute_minutes || 0,
        started_at: existing.started_at || '',
        note: existing.note || '',
      }
      this.showWorkModal = true
    },
    async saveWorkProfile() {
      try {
        const data = { ...this.workForm }
        delete data.memberName
        await api.put(`/api/work/profiles/${data.member_id}`, data)
        this.showWorkModal = false
        await this.loadProfiles()
      } catch {
        // ignore
      }
    },
    roleLabel(r) {
      const map = { dad: '爸爸', mom: '妈妈', son: '儿子', daughter: '女儿', grandpa: '爷爷/外公', grandma: '奶奶/外婆' }
      return map[r] || r || '-'
    },
    roleBadge(r) {
      const map = { dad: 'badge-blue', mom: 'badge-pink', son: 'badge-green', daughter: 'badge-green', grandpa: 'badge-primary', grandma: 'badge-primary' }
      return map[r] || 'badge-primary'
    },
  },
}
</script>

<style scoped>
.grid-2 {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
.profile-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px 16px;
  font-size: 13px;
}
.profile-item {
  display: flex;
  gap: 4px;
}
.profile-label {
  color: var(--text-secondary);
  white-space: nowrap;
}
.profile-label::after {
  content: '：';
}
@media (max-width: 768px) {
  .grid-2 { grid-template-columns: 1fr; }
  .profile-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
