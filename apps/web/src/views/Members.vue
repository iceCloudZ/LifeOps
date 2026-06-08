<template>
  <div>
    <h1 class="page-title">家庭成员</h1>

    <div class="card" style="margin-bottom: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <div class="card-title" style="margin-bottom: 0;">成员列表</div>
        <button class="btn btn-primary btn-sm" @click="openAdd">+ 添加成员</button>
      </div>
      <div v-if="members.length === 0" class="empty-state">
        <p>暂无成员</p>
      </div>
      <div class="table-container" v-else>
        <table>
          <thead>
            <tr>
              <th>姓名</th>
              <th>角色</th>
              <th>生日</th>
              <th>电话</th>
              <th>邮箱</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in members" :key="m.id">
              <td style="font-weight: 600;">{{ m.name }}</td>
              <td><span class="badge" :class="roleBadge(m.role)">{{ roleLabel(m.role) }}</span></td>
              <td>{{ m.birthday || '-' }}</td>
              <td>{{ m.phone || '-' }}</td>
              <td>{{ m.email || '-' }}</td>
              <td>
                <div style="display: flex; gap: 6px;">
                  <button class="btn btn-secondary btn-sm" @click="openEdit(m)">编辑</button>
                  <button class="btn btn-danger btn-sm" @click="deleteMember(m.id)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
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
            <option value="admin">管理员</option>
            <option value="member">普通成员</option>
            <option value="child">孩子</option>
          </select>
        </div>
        <div class="form-group">
          <label>生日</label>
          <input v-model="form.birthday" type="date" />
        </div>
        <div class="form-group">
          <label>电话</label>
          <input v-model="form.phone" type="text" placeholder="手机号" />
        </div>
        <div class="form-group">
          <label>邮箱</label>
          <input v-model="form.email" type="text" placeholder="邮箱地址" />
        </div>
        <div class="form-group">
          <label>头像URL</label>
          <input v-model="form.avatar" type="url" placeholder="可选头像链接" />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showModal = false">取消</button>
          <button class="btn btn-primary" @click="saveMember">{{ isEditing ? '保存' : '确定' }}</button>
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
      showModal: false,
      isEditing: false,
      editingId: null,
      form: { name: '', role: 'member', birthday: '', phone: '', email: '', avatar: '' },
    }
  },
  async mounted() {
    await this.loadMembers()
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
    openAdd() {
      this.isEditing = false
      this.editingId = null
      this.form = { name: '', role: 'member', birthday: '', phone: '', email: '', avatar: '' }
      this.showModal = true
    },
    openEdit(m) {
      this.isEditing = true
      this.editingId = m.id
      this.form = {
        name: m.name || '',
        role: m.role || 'member',
        birthday: m.birthday || '',
        phone: m.phone || '',
        email: m.email || '',
        avatar: m.avatar || '',
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
      } catch {
        // ignore
      }
    },
    roleLabel(r) {
      const map = { admin: '管理员', member: '成员', child: '孩子' }
      return map[r] || r || '-'
    },
    roleBadge(r) {
      const map = { admin: 'badge-primary', member: 'badge-blue', child: 'badge-green' }
      return map[r] || 'badge-primary'
    },
  },
}
</script>
