<template>
  <div class="chat-layout">
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

    <div class="chat-main">
      <div class="chat-messages" ref="messagesContainer">
        <div v-if="messages.length === 0 && !routingResult && !isStreaming" class="empty-state" style="padding-top: 80px;">
          <p>选择或创建一个对话开始聊天</p>
        </div>

        <template v-for="msg in messages" :key="msg.id">
          <div class="chat-message" :class="msg.role">
            <div class="avatar">{{ msg.role === 'user' ? '我' : 'AI' }}</div>
            <div class="content">
              <template v-if="msg.role === 'assistant'">
                <MarkdownRender :content="msg.content" />
              </template>
              <template v-else>{{ msg.content }}</template>
              <div v-if="msg.role === 'assistant'" class="lens-info">
                <span v-if="msg.lens_name" class="lens-badge" :title="msg.lens_reason">
                  {{ msg.lens_name }}
                </span>
                <span v-else class="lens-badge lens-default">默认回答</span>
                <span v-if="msg.lens_reason" class="lens-reason">{{ msg.lens_reason }}</span>
              </div>
            </div>
          </div>

          <div v-if="msg.role === 'user' && routingResult && routingResult._for === msg.id" class="chat-card lens-card">
            <div class="chat-card-header">建议使用 {{ routingResult.lens }} 方式分析</div>
            <div class="chat-card-body">原因：{{ routingResult.reason }}</div>
            <div class="chat-card-actions">
              <button class="btn btn-primary btn-sm" @click="executeWithLens">确认使用</button>
              <button class="btn btn-secondary btn-sm" @click="executeWithoutLens">跳过</button>
            </div>
          </div>

          <div v-for="confirm in msg.confirms || []" :key="confirm.id" class="chat-card confirm-card">
            <div class="chat-card-header">操作确认</div>
            <div class="chat-card-body">{{ confirm.summary }}</div>
            <div class="chat-card-actions">
              <button
                class="btn btn-primary btn-sm"
                :disabled="confirm.resolved"
                @click="resolveConfirm(confirm, 'approve')"
              >确认执行</button>
              <button
                class="btn btn-secondary btn-sm"
                :disabled="confirm.resolved"
                @click="resolveConfirm(confirm, 'cancel')"
              >取消</button>
            </div>
          </div>
        </template>

        <div v-if="isStreaming" class="chat-message assistant">
          <div class="avatar">AI</div>
          <div class="content">
            <MarkdownRender :content="streamingContent" :typewriter="true" :max-live-nodes="0" :fade="false" />
            <span v-if="streamingContent" class="streaming-cursor">|</span>
            <div v-if="streamingTools.length > 0" class="tool-indicators">
              <span v-for="tool in streamingTools" :key="tool.name" class="tool-pill">
                {{ tool.label || '正在查询...' }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="chat-input-area">
        <input
          v-model="inputText"
          type="text"
          placeholder="输入消息..."
          :disabled="isStreaming || !!routingResult"
          @keydown.enter="sendMessage"
        />
        <button class="btn btn-primary" :disabled="isStreaming || !!routingResult" @click="sendMessage">发送</button>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../api.js'
import { MarkdownRender } from 'markstream-vue'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const API_TOKEN = import.meta.env.VITE_API_TOKEN || 'dev-token'

export default {
  name: 'Chat',
  components: { MarkdownRender },
  data() {
    return {
      conversations: [],
      currentId: null,
      messages: [],
      inputText: '',
      routingResult: null,
      isStreaming: false,
      streamingContent: '',
      streamingTools: [],
      currentMemberId: null,
    }
  },
  async mounted() {
    await this.loadConversations()
    const convId = this.$route.query.conv
    const msg = this.$route.query.msg
    if (convId) {
      await this.selectConversation(convId)
      if (msg) {
        this.inputText = msg
        await this.sendMessage()
      }
    }
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
      this.routingResult = null
      this.isStreaming = false
      this.streamingContent = ''
      this.streamingTools = []
      try {
        const res = await api.get(`/api/chat/conversations/${id}/messages`)
        this.messages = (res.data.messages || res.data || []).map(m => ({
          ...m,
          confirms: m.confirms || [],
        }))
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
      if (this.isStreaming || this.routingResult) return
      const text = this.inputText
      this.inputText = ''
      const userMsg = { id: 'u-' + Date.now(), role: 'user', content: text }
      this.messages.push(userMsg)
      this.$nextTick(() => this.scrollToBottom())
      try {
        const res = await api.post(`/api/chat/conversations/${this.currentId}/route`, {
          content: text,
          currentMemberId: this.currentMemberId,
        })
        const route = res.data
        this.routingResult = {
          _for: userMsg.id,
          domain: route.domain || [],
          lens: route.lens,
          reason: route.reason,
          needsWebSearch: route.needsWebSearch || false,
        }
        this.$nextTick(() => this.scrollToBottom())
        if (!route.lens) {
          this.routingResult = null
          await this.executeStream(route.lens, route.domain)
        }
      } catch {
        this.messages.push({
          id: 'e-' + Date.now(),
          role: 'assistant',
          content: '路由失败，请稍后重试。',
          confirms: [],
        })
        this.$nextTick(() => this.scrollToBottom())
      }
    },
    async executeWithLens() {
      const r = this.routingResult
      this.routingResult = null
      await this.executeStream(r.lens, r.domain)
    },
    async executeWithoutLens() {
      const r = this.routingResult
      this.routingResult = null
      await this.executeStream(null, r.domain)
    },
    async executeStream(lens, domain) {
      const userMsg = this.messages[this.messages.length - 1]
      this.isStreaming = true
      this.streamingContent = ''
      this.streamingTools = []
      this.$nextTick(() => this.scrollToBottom())

      const assistantMsgId = 'a-' + Date.now()
      const streamingConfirms = []
      const payload = {
        content: userMsg.content,
        lens: lens,
        domain: domain || [],
        currentMemberId: this.currentMemberId,
      }

      try {
        const response = await fetch(`${API_BASE}/api/chat/conversations/${this.currentId}/execute`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + API_TOKEN,
          },
          body: JSON.stringify(payload),
        })

        if (!response.ok) {
          throw new Error('Execute request failed: ' + response.status)
        }

        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''
        const toolNameMap = {
          listMembers: '正在获取家庭成员...',
          listFinanceSummary: '正在查询财务概览...',
          queryFinanceRecords: '正在查询财务记录...',
          listHealthProfiles: '正在查询健康数据...',
          listMovementRecords: '正在查询运动记录...',
          listWorkStatus: '正在查询工作状态...',
          listFamilyRecords: '正在查询家庭事务...',
          queryNotes: '正在查询笔记...',
          webSearch: '正在搜索网络...',
          createFinanceRecord: '正在记账...',
          createHealthRecord: '正在记录健康数据...',
          createMovementRecord: '正在记录运动...',
          createWorkRecord: '正在添加工作任务...',
          createFamilyRecord: '正在添加家庭事务...',
          createNote: '正在保存笔记...',
        }

        while (true) {
          const { done, value } = await reader.read()
          if (done) break
          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop() || ''

          let currentEvent = ''
          for (const line of lines) {
            if (line.startsWith('event:')) {
              currentEvent = line.slice(6).trim()
            } else if (line.startsWith('data:')) {
              const data = line.slice(5)
              if (currentEvent === 'token') {
                this.streamingContent += data
                this.$nextTick(() => this.scrollToBottom())
              } else if (currentEvent === 'tool_start') {
                try {
                  const info = JSON.parse(data)
                  const tName = info.tool || info.name || ''
                  this.streamingTools.push({
                    name: tName,
                    label: toolNameMap[tName] || ('正在执行 ' + tName),
                  })
                } catch {
                  this.streamingTools.push({ name: data, label: '正在查询...' })
                }
                this.$nextTick(() => this.scrollToBottom())
              } else if (currentEvent === 'tool_result') {
                try {
                  const result = JSON.parse(data)
                  if (result.status === 'pending_confirmation') {
                    streamingConfirms.push({
                      id: 'cf-' + Date.now() + '-' + streamingConfirms.length,
                      msgId: assistantMsgId,
                      action: result.action,
                      summary: result.summary,
                      data: result.data,
                      resolved: false,
                    })
                  }
                } catch {
                  // not JSON or not a confirmation, ignore
                }
              } else if (currentEvent === 'confirm') {
                // informational confirm event
              } else if (currentEvent === 'done') {
                // stream finished
              } else if (currentEvent === 'error') {
                if (!this.streamingContent) {
                  this.streamingContent = '请求出错：' + data
                }
              }
            }
          }
        }
      } catch (err) {
        if (!this.streamingContent) {
          this.streamingContent = '请求失败，请稍后重试。'
        }
      }

      const assistantMsg = {
        id: assistantMsgId,
        role: 'assistant',
        content: this.streamingContent,
        lens_name: lens || '',
        lens_reason: '',
        confirms: streamingConfirms,
        tools: this.streamingTools,
      }
      this.messages.push(assistantMsg)
      this.isStreaming = false
      this.streamingContent = ''
      this.streamingTools = []
      this.$nextTick(() => this.scrollToBottom())
    },
    async resolveConfirm(confirm, action) {
      try {
        if (action === 'approve') {
          await api.post(`/api/chat/conversations/${this.currentId}/confirm/${confirm.msgId}`, {
            action: confirm.action,
            data: confirm.data || {},
          })
        } else {
          await api.post(`/api/chat/conversations/${this.currentId}/confirm/${confirm.msgId}`, {
            action: 'cancel',
            data: {},
          })
        }
        confirm.resolved = true
      } catch {
        confirm.resolved = true
      }
    },
    scrollToBottom() {
      const el = this.$refs.messagesContainer
      if (el) el.scrollTop = el.scrollHeight
    },
  },
}
</script>

<style scoped>
.lens-info {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.lens-badge {
  display: inline-block;
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 10px;
  background: #e0f0ff;
  color: #1a73e8;
  font-weight: 500;
}

.lens-badge.lens-default {
  background: #f0f0f0;
  color: #888;
}

.lens-reason {
  font-size: 11px;
  color: #999;
}

.chat-card {
  margin: 8px 0 12px 46px;
  padding: 12px 16px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  max-width: 70%;
}

.chat-card-header {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 6px;
}

.chat-card-body {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 10px;
  line-height: 1.5;
}

.chat-card-actions {
  display: flex;
  gap: 8px;
}

.lens-card {
  background: #f0f4ff;
  border-color: #c7d2fe;
}

.lens-card .chat-card-header {
  color: var(--primary);
}

.confirm-card {
  background: #fffbeb;
  border-color: #fde68a;
}

.confirm-card .chat-card-header {
  color: #b45309;
}

.streaming-cursor {
  display: inline-block;
  animation: blink 0.8s step-end infinite;
  color: var(--primary);
  font-weight: 300;
  margin-left: 1px;
}

@keyframes blink {
  50% { opacity: 0; }
}

.tool-indicators {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tool-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  padding: 2px 10px;
  border-radius: 12px;
  background: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
}

.tool-pill::before {
  content: '';
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #9ca3af;
  animation: pulse-dot 1.4s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}

.chat-input-area input:disabled,
.chat-input-area button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
