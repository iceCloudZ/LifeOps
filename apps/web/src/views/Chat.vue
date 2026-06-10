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
                <MarkdownRender mode="chat" :content="msg.content" :final="true" :fade="false" />
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
            <MarkdownRender
              mode="chat"
              :content="streamingContent"
              :final="false"
              smooth-streaming="auto"
              :fade="false"
              :max-live-nodes="0"
              :typewriter="true"
            />
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
import 'markstream-vue/index.css'

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

      this.isStreaming = true
      this.streamingContent = ''
      this.streamingTools = []

      const assistantMsgId = 'a-' + Date.now()
      const streamingConfirms = []

      try {
        const response = await fetch(`${API_BASE}/api/chat/conversations/${this.currentId}/chat`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + API_TOKEN,
          },
          body: JSON.stringify({
            content: text,
            currentMemberId: this.currentMemberId,
          }),
        })

        if (!response.ok) throw new Error('Chat request failed: ' + response.status)

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
              } else if (currentEvent === 'route') {
                // Route result — show lens confirmation if applicable
                try {
                  const route = JSON.parse(data)
                  if (route.lens) {
                    this.routingResult = {
                      _for: userMsg.id,
                      domain: route.domain || [],
                      lens: route.lens,
                      reason: route.reason,
                      needsWebSearch: route.needsWebSearch || false,
                    }
                    this.$nextTick(() => this.scrollToBottom())
                  }
                } catch { /* ignore parse error */ }
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
                } catch { /* not JSON or not a confirmation */ }
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
        lens_name: '',
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
    async executeWithLens() {
      const r = this.routingResult
      this.routingResult = null
      this.isStreaming = true
      this.streamingContent = ''
      this.streamingTools = []
      this.$nextTick(() => this.scrollToBottom())

      try {
        const response = await fetch(`${API_BASE}/api/chat/conversations/${this.currentId}/execute`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + API_TOKEN,
          },
          body: JSON.stringify({
            content: this.messages[this.messages.length - 2].content,
            currentMemberId: this.currentMemberId,
            lens: r.lens,
            domain: r.domain,
          }),
        })
        if (!response.ok) throw new Error('Execute failed')
        await this.consumeSSE(response, 'a-' + Date.now())
      } catch {
        this.streamingContent = this.streamingContent || '执行失败'
      }
    },
    async executeWithoutLens() {
      this.routingResult = null
    },
    async consumeSSE(response, assistantMsgId) {
      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
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
                this.streamingTools.push({ name: info.tool || '', label: '正在执行 ' + (info.tool || '') })
              } catch {
                this.streamingTools.push({ name: data, label: '正在查询...' })
              }
            } else if (currentEvent === 'error') {
              if (!this.streamingContent) this.streamingContent = '请求出错：' + data
            }
          }
        }
      }
      this.messages.push({
        id: assistantMsgId, role: 'assistant', content: this.streamingContent,
        confirms: [], tools: this.streamingTools,
      })
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
/* ── Chat CSS Variables ── */
:root {
  --chat-bubble-assistant: #f3f4f6;
  --chat-bubble-user: var(--primary);
  --chat-bubble-user-text: #fff;
  --chat-avatar-size: 32px;
  --chat-content-max: 75%;
  --chat-bubble-radius: 18px;
}

/* ── Lens badge ── */
.lens-info {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.lens-badge {
  display: inline-block;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--primary-bg);
  color: var(--primary);
  font-weight: 500;
}

.lens-badge.lens-default {
  background: transparent;
  color: var(--text-secondary);
  padding: 0;
  font-weight: 400;
}

.lens-reason {
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.7;
}

/* ── Chat cards (lens / confirm) ── */
.chat-card {
  margin: 4px 0 12px;
  margin-left: calc(var(--chat-avatar-size) + 10px);
  padding: 12px 16px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  max-width: var(--chat-content-max);
  box-shadow: 0 1px 2px rgba(0,0,0,0.04);
}

.chat-card-header {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
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
  background: #eef2ff;
  border-color: #c7d2fe;
}
.lens-card .chat-card-header { color: var(--primary); }

.confirm-card {
  background: #fffbeb;
  border-color: #fde68a;
}
.confirm-card .chat-card-header { color: #b45309; }

/* ── Streaming cursor ── */
.streaming-cursor {
  display: inline-block;
  width: 2px;
  height: 1.1em;
  background: var(--primary);
  vertical-align: text-bottom;
  margin-left: 2px;
  border-radius: 1px;
  animation: blink 1s step-end infinite;
}
@keyframes blink {
  50% { opacity: 0; }
}

/* ── Tool indicators ── */
.tool-indicators {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tool-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  padding: 3px 10px;
  border-radius: 12px;
  background: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
  transition: opacity 0.3s;
}

.tool-pill::before {
  content: '';
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--primary);
  animation: pulse-dot 1.4s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

/* ── Input area ── */
.chat-input-area {
  padding: 14px 20px;
  border-top: 1px solid var(--border);
  display: flex;
  gap: 10px;
  background: var(--card-bg);
}

.chat-input-area input {
  flex: 1;
  padding: 10px 16px;
  border: 1px solid var(--border);
  border-radius: 24px;
  font-size: 14px;
  font-family: var(--font);
  color: var(--text);
  background: var(--bg);
  transition: border-color 0.2s, box-shadow 0.2s;
}
.chat-input-area input:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(79,70,229,0.1);
}

.chat-input-area button {
  padding: 10px 20px;
  border-radius: 24px;
  font-size: 14px;
  font-weight: 500;
}

.chat-input-area input:disabled,
.chat-input-area button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* ── Message bubbles ── */
.chat-message {
  margin-bottom: 20px;
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.chat-message.user {
  flex-direction: row-reverse;
}

.chat-message .avatar {
  width: var(--chat-avatar-size);
  height: var(--chat-avatar-size);
  min-width: var(--chat-avatar-size);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
}

.chat-message.assistant .avatar {
  background: var(--primary);
  color: #fff;
  font-size: 12px;
}

.chat-message.user .avatar {
  background: #dcfce7;
  color: #16a34a;
}

.chat-message .content {
  max-width: var(--chat-content-max);
  padding: 10px 16px;
  border-radius: var(--chat-bubble-radius);
  font-size: 14px;
  line-height: 1.65;
}

.chat-message.assistant .content {
  background: var(--chat-bubble-assistant);
  color: var(--text);
  border-bottom-left-radius: 6px;
}

.chat-message.user .content {
  background: var(--chat-bubble-user);
  color: var(--chat-bubble-user-text);
  border-bottom-right-radius: 6px;
}

/* ── Messages container ── */
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 24px 20px;
  background: var(--bg);
}
</style>
