# LifeOps Agent + Tool + Lens 设计

> 日期：2026-06-09
> 状态：已批准

## 目标

将 ButlerAgent 从"单次全量注入"升级为两阶段智能 agent：
- Phase 1：路由（domain 检测 + lens 推荐，用户确认）
- Phase 2：执行（流式 tool calling 循环，LLM 自主决策）

同时集成 lifestyle-skills 的 progressive disclosure lens 系统。

## 架构总览

```
用户提问
  │
  ▼
Phase 1: 路由（同步 ChatModel，< 2s）
  输入: 用户问题 + LensIndex + 对话历史
  输出: { domain, lens, reason, needsWebSearch }
  → 前端展示推荐卡片，用户确认
  │
  ▼
Phase 2: 执行（StreamingChatModel + Tools，首 token < 1s）
  System prompt: 选中 lens 完整内容 + domain systemPrompt
  Tools: 读工具（自动执行）+ 写工具（返回待确认）+ webSearch
  LLM 自主决定调哪些 tools，循环直到出最终回答
  → SSE 流式返回 token/tool_start/confirm/done
  │
  ▼
前端渲染回答 + 确认卡片
```

## Phase 1: 路由

### 输入

- 用户问题
- LensIndex（从 registry.json 提取的 triggers/blockers 摘要，~200-300 tokens）
- 最近对话历史

### LensIndex 格式

直接从 lifestyle-skills 的 `registry.json` 提取，在应用启动时加载到内存：

```
domain: finance
  bogleheads-style: 长期投资、资产配置 — blocker: 无应急基金
  conscious-spending: 消费方向、花钱没感觉 — blocker: 债务危机
  zero-based-budgeting: 现金流控制、每分钱有归属 — blocker: 不想细记账
domain: health
  longevity-medicine: 预防指标、体检趋势 — blocker: 急性症状
  blue-zones-style: 全生活改善、饮食运动社交 — blocker: 需要临床方案
  daily-dozen-style: 植物性饮食清单 — blocker: 进食障碍
domain: movement
  zone2-longevity: 有氧耐力、低损伤 — blocker: 急性损伤
  strength-baseline: 力量训练入门 — blocker: 无技术指导
  tiny-habits-movement: 微习惯、没时间没动力 — blocker: 高级训练者
domain: family
  gottman-style: 伴侣冲突修复 — blocker: 安全风险
  nvc-style: 沟通降级、需求表达 — blocker: 恶意操纵
  positive-discipline-style: 正面管教、温和坚定 — blocker: 儿童安全风险
```

### 输出 JSON

```json
{
  "domain": ["finance", "health"],
  "lens": "blue-zones-style",
  "reason": "你的问题涉及健康习惯和运动，Blue Zones 方式适合全生活改善",
  "needsWebSearch": false
}
```

domain 可以是数组（多域分析场景）。

### 路由 prompt

```
你是 LifeOps 家庭管家的路由模块。
根据用户问题和下面的 Lens 索引，推荐最合适的 domain 和 lens。

规则：
- 只选 1 个主 lens
- domain 可以多个（多域分析时）
- 如果问题不涉及生活方式建议（比如只是查数据），domain 选相关域但 lens 设为 null
- needsWebSearch: 如果问题涉及外部知识（市场行情、疾病信息等），设为 true

{lensIndex}

返回 JSON，不要解释。
```

## Phase 2: 执行

### System Prompt 构建

```
{选中 lens 的完整 .md 内容}

---

{选中 domain(s) 对应的 systemPrompt 合并。多 domain 时按顺序拼接，每个 domain 单独一段}

---

你是 LifeOps 家庭管家。基于提供的家庭数据和选定的思考方式回答问题。
家庭成员通过 memberId 引用。不确定是哪个成员时，先调 listMembers 查询或直接问用户。
写操作（记账、创建记录等）需要用户确认后才会真正执行。
```

### Tool 设计

#### 读 Tools（自动执行）

| Tool 名 | 参数 | 调用的 Service | 语义 |
|---------|------|---------------|------|
| `listMembers` | 无 | MemberService.listMembers | 家庭成员列表 |
| `listFinanceSummary` | memberId?（空=全家） | FinanceService.listAccounts + listRecords | 财务概览 |
| `queryFinanceRecords` | memberId?, type?, category?, fromDate?, toDate? | FinanceService.listRecords | 按条件查财务记录 |
| `listHealthProfiles` | memberId? | HealthService.listProfiles + listRecords | 健康概览 |
| `listMovementRecords` | memberId | MovementService.listRecords | 运动记录 |
| `listWorkStatus` | memberId? | WorkService.listStatuses + listProfiles + listRecords | 工作状态 |
| `listFamilyRecords` | memberId?, type?, status? | FamilyService.getStatus + listRecords | 家庭事务 |
| `queryNotes` | domain?, memberId? | NoteService.listNotes | 知识笔记 |
| `webSearch` | query | 外部搜索 API | 上网搜索 |

#### 写 Tools（需用户确认）

| Tool 名 | 参数 | 确认描述示例 |
|---------|------|------------|
| `createFinanceRecord` | memberId, amount, type, category, note? | "给张三记账：买菜 50 元" |
| `createHealthRecord` | memberId, metric, value, unit?, note? | "记录李四的血压：130/85" |
| `createMovementRecord` | memberId, metric, value, note? | "记录张三的运动：跑步 30 分钟" |
| `createWorkRecord` | memberId, title, project?, priority? | "给李四添加工作任务：完成报告" |
| `createFamilyRecord` | memberId, type, title, scheduledDate? | "添加待办：周三接孩子" |
| `createNote` | domain, title, content, memberId? | "保存笔记：体检注意空腹" |

写 tool 的返回值：

```json
{
  "status": "pending_confirmation",
  "action": "createFinanceRecord",
  "summary": "给张三记账：买菜 50 元，分类：食品",
  "data": { "memberId": "xxx", "amount": 50, "type": "expense", "category": "食品" }
}
```

#### 不暴露

- `delete*`：LLM 不应删数据
- `update*`：暂不暴露
- `getRecord(id)`：LLM 用不到按 ID 查单条

### Tool Calling 循环

使用 LangChain4j StreamingChatModel + AI Services：

1. LLM 流式输出 token
2. 遇到 tool call → 停止文字 → 逐 token 组装参数 → `onCompleteToolCall`
3. 读 tool：立即执行，结果喂回 LLM
4. 写 tool：返回 `pending_confirmation`，SSE 发送 `confirm` 事件，LLM 继续输出文字说明"我建议帮你做 X，请点确认"
5. LLM 决定继续调 tool 或出最终文字
6. 循环直到 LLM 返回纯文字（无 tool call）

写 tool 不阻塞循环——LLM 可以在返回 pending 后继续调其他读 tool 或输出更多分析。确认操作由前端异步处理，不影响当前流。

### Member 身份解析

所有读写 tool 的 memberId 参数：
- 前端在请求中传入 `currentMemberId`（最可靠）
- 未传时，LLM 调 `listMembers` 后根据对话上下文推断
- 推断不了时 LLM 反问用户

## 流式交互设计

### 两个 ChatModel Bean

```java
@Bean ChatModel chatModel;                // Phase 1 路由用（同步）
@Bean StreamingChatModel streamingModel;  // Phase 2 执行用（流式）
```

### SSE 事件类型

| 事件 | 数据 | 前端行为 |
|------|------|---------|
| `token` | 文本片段 | 逐字渲染回答 |
| `tool_start` | tool 名称 | 显示"正在查询..." |
| `tool_result` | 结果摘要 | 可选展示 |
| `confirm` | 待确认操作 JSON | 展示确认卡片 |
| `done` | 空 | 关闭 SSE 连接 |

### API 端点

```
POST /api/chat/{conversationId}
  body: { content, currentMemberId }
  response: { lens, domain, reason, needsWebSearch }

POST /api/chat/{conversationId}/execute
  body: { content, lens, domain, currentMemberId }
  response: SSE stream (token/tool_start/confirm/done)

POST /api/chat/{conversationId}/confirm/{msgId}
  body: { action, data }
  response: { success: true }
```

### SseEmitter 超时

设为 5 分钟（300s），因为多步 tool calling 可能耗时较长。

## lifestyle-skills 集成

### 打包方式

lifestyle-skills 的内容嵌入 api-java jar 的 classpath：

```
src/main/resources/skills/
  registry.json
  skills/life-butler/SKILL.md
  lenses/
    finance/bogleheads-style.md
    finance/conscious-spending.md
    finance/zero-based-budgeting.md
    health/blue-zones-style.md
    health/daily-dozen-style.md
    health/longevity-medicine.md
    movement/zone2-longevity.md
    movement/strength-baseline.md
    movement/tiny-habits-movement.md
    family/gottman-style.md
    family/nvc-style.md
    family/positive-discipline-style.md
```

### 加载方式

- 应用启动时读 `registry.json` → 构建内存 LensIndex
- Phase 1：用 LensIndex（不读 .md 文件）
- Phase 2：用户确认 lens 后，读对应 `.md` 注入 system prompt

### 更新方式

lifestyle-skills 更新后，重新复制到 resources/skills/，重新构建 jar。

## 前端交互时序

```
用户输入 "帮我分析我们家情况"
  │
  ▼ POST /api/chat/{id} { content, currentMemberId: "张三" }
  │
  ← { lens: "blue-zones-style", domain: ["health","movement","family"],
       reason: "..." }
  │
  前端展示: "建议用 Blue Zones 方式全面分析，涵盖健康+运动+家庭习惯"
  用户点确认
  │
  ▼ POST /api/chat/{id}/execute { content, lens: "blue-zones-style",
      domain: ["health","movement","family"], currentMemberId: "张三" }
  │
  ← SSE: tool_start: "listMembers"
  ← SSE: tool_start: "listFinanceSummary"
  ← SSE: tool_start: "listHealthProfiles"
  ← SSE: tool_start: "listMovementRecords"
  ← SSE: tool_start: "listWorkStatus"
  ← SSE: token: "根据"
  ← SSE: token: "你家"
  ← SSE: token: "数据..."
  ...（文字逐步渲染）
  ← SSE: confirm: { action: "createNote", summary: "保存体检提醒笔记" }
  ← SSE: done
  │
  前端渲染: 完整分析 + 确认卡片
  用户点确认
  │
  ▼ POST /api/chat/{id}/confirm/{msgId}
  ← { success: true }
```

## Web Search Tool

```java
@Tool("搜索互联网获取最新信息，用于回答需要外部知识的问题")
String webSearch(@P("搜索关键词") String query);
```

实现方式待定（DuckDuckGo API / Tavily / LangChain4j WebSearchTool），先定接口。
