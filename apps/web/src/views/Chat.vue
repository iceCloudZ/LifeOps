<template>
  <div class="chat-layout">
    <!-- Sidebar -->
    <div class="chat-sidebar">
      <button class="btn btn-primary" style="width: 100%; margin-bottom: 12px;" @click="createConversation">
        + 新对话
      </button>
      <div
        v-for="c in conversations"
        :key="c.id"
        class="chat-sidebar-item"
        :class="{ active: currentId === c.id }"
        @click="selectConversation(c.id)"
      >
        {{ c.title || '对话 ' + c.id }}
      </div>
      <div v-if="conversations.length === 0" class="empty-state" style="padding: 20px;">
        <p>暂无对话</p>
      </div>
    </div>

    <!-- Chat main area -->
    <div class="chat-main">
      <div class="chat-messages" ref="messagesContainer">
        <div v-if="messages.length === 0" class="empty-state" style="padding-top: 80px;">
          <p>选择或创建一个对话开始聊天</p>
        </div>
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="chat-message"
          :class="msg.role"
        >
          <div class="avatar">{{ msg.role === 'user' ? '我' : 'AI' }}</div>
          <div class="content">{{ msg.content }}</div>
        </div>
      </div>
      <div class="chat-input-area">
        <input
          v-model="inputText"
          type="text"
          placeholder="输入消息..."
          @keydown.enter="sendMessage"
        />
        <button class="btn btn-primary" @click="sendMessage">发送</button>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../api.js'

export default {
  name: 'Chat',
  data() {
    return {
      conversations: [],
      currentId: null,
      messages: [],
      inputText: '',
    }
  },
  async mounted() {
    await this.loadConversations()
  },
  methods: {
    async loadConversations() {
      try {
        const res = await api.get('/api/chat/conversations')
        this.conversations = res.data.conversations || res.data || []
        if (this.conversations.length > 0 && !this.currentId) {
          await this.selectConversation(this.conversations[0].id)
        }
      } catch {
        this.conversations = []
      }
    },
    async selectConversation(id) {
      this.currentId = id
      try {
        const res = await api.get(`/api/chat/conversations/${id}/messages`)
        this.messages = res.data.messages || res.data || []
        this.$nextTick(() => this.scrollToBottom())
      } catch {
        this.messages = []
      }
    },
    async createConversation() {
      try {
        const res = await api.post('/api/chat/conversations', { title: '新对话' })
        const conv = res.data
        this.conversations.unshift(conv)
        await this.selectConversation(conv.id)
      } catch {
        // ignore
      }
    },
    async sendMessage() {
      if (!this.inputText.trim() || !this.currentId) return
      const text = this.inputText
      this.inputText = ''
      this.messages.push({ id: Date.now(), role: 'user', content: text })
      this.$nextTick(() => this.scrollToBottom())
      try {
        const res = await api.post(`/api/chat/conversations/${this.currentId}/messages`, {
          content: text,
        })
        const reply = res.data
        if (reply.content || reply.message) {
          this.messages.push({
            id: Date.now() + 1,
            role: 'assistant',
            content: reply.content || reply.message,
          })
        }
        this.$nextTick(() => this.scrollToBottom())
      } catch {
        this.messages.push({
          id: Date.now() + 1,
          role: 'assistant',
          content: '请求失败，请稍后重试。',
        })
        this.$nextTick(() => this.scrollToBottom())
      }
    },
    scrollToBottom() {
      const el = this.$refs.messagesContainer
      if (el) el.scrollTop = el.scrollHeight
    },
  },
}
</script>
