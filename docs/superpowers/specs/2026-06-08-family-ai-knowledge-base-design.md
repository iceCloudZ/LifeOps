# 家庭AI知识库 + 多Agent架构设计

**日期**: 2026-06-08
**状态**: 草稿
**范围**: MVP — 将 LifeOps 从"家庭消息收件箱"升级为"家庭AI知识库 + 多Agent问答系统"

## 问题背景

LifeOps 目前实现了 家庭消息 → AI提取 → 结构化草稿 的流程。用户需要更广泛的能力：一个能理解家庭全貌（财务、健康、工作、日程、习惯）的AI系统，覆盖5-6名家庭成员（父母、夫妻、子女），能以对话方式回答问题。

## 架构设计

### 多Agent + 管家路由

```
用户提问
  │
  ▼
┌──────────────────────────────────────┐
│        管家 Agent (路由 + LLM)        │
│  1. 分析用户问题（LLM调用）            │
│  2. 路由到专业Agent(s)               │
│  3. 综合跨领域的回答                  │
│  4. 确保回答质量和语气                │
└──┬──────┬──────┬──────┬──────────────┘
   │      │      │      │
   ▼      ▼      ▼      ▼
┌─────┐┌─────┐┌─────┐┌────────┐
│财务  ││健康  ││工作  ││家庭事务 │
│Agent ││Agent ││Agent ││Agent   │
└──┬──┘└──┬──┘└──┬──┘└───┬────┘
   │      │      │       │
   ▼      ▼      ▼       ▼
┌────────────────────────────────────┐
│   三层数据模型 (SQLite)             │
│  状态表(现状) + 流水表(变化)         │
│  + 知识笔记(不变的约束和偏好)        │
└────────────────────────────────────┘
```

### Agent 职责划分

| Agent | 领域 | Prompt 侧重点 | 示例问题 |
|-------|------|-------------|---------|
| **管家** | 全部 | "家庭管家，路由问题，综合回答，温暖简洁的中文" | "最近家里怎么样" |
| **财务** | finance | "家庭财务顾问，分析收支、资产负债，给出建议" | "这个月花了多少" |
| **健康** | health | "家庭健康助手，关注体检指标、用药、运动、饮食" | "妈的血压最近怎么样" |
| **工作** | work | "职业助手，关注项目进度、重要节点" | "最近有什么deadline" |
| **家庭事务** | family | "家庭事务管家，日程、育儿、家务、活动安排" | "周末有什么安排" |

### 路由逻辑

管家 Agent 通过 LLM（不是关键词匹配）来：
1. 判断用户问题涉及哪些领域
2. 单领域问题：调用对应的一个专业 Agent
3. 跨领域问题：并发调用多个专业 Agent（goroutine），然后综合
4. 始终审核专业 Agent 的回答后再返回给用户

每个专业 Agent 的工作流程：
1. 接收用户问题 + 领域上下文
2. 查状态表获取当前快照
3. 查流水表获取近期趋势和详细数据
4. 查知识笔记表获取相关约束和偏好
5. 用以上三层上下文构建 prompt，调用 LLM
6. 将结构化回答返回给管家

### LLM 调用链路

```
用户提问
  → 管家 Agent (LLM调用1: 分析 + 路由)
  → 专业 Agent(s) (LLM调用2+: 带三层知识上下文的领域回答)
  → 管家 Agent (多领域时综合)
  → 返回用户
```

## 数据模型：三层架构

数据按**状态（现状）→ 流水（变化）→ 笔记（知识）**三层组织。

```
┌─────────────────────────────────────┐
│  状态/账户表 — 现在怎么样             │  ← 当前快照，频繁更新
│  账户余额、健康状态、项目进度         │
├─────────────────────────────────────┤
│  流水/记录表 — 发生了什么             │  ← 历史事件，追加写入
│  收支流水、健康指标、工作任务、日程   │
├─────────────────────────────────────┤
│  知识笔记表 — 持续有效的事            │  ← 不变的约束和偏好
│  饮食禁忌、家庭习惯、过敏信息         │
└─────────────────────────────────────┘
```

### 第一层：状态/账户表（现在怎么样）

**财务账户表 (finance_accounts)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PK | |
| member_id | TEXT FK | 归属成员，可空（家庭共用账户） |
| name | TEXT | 账户名称（工商银行储蓄卡、支付宝等） |
| type | TEXT | bank/ewallet/investment/insurance/loan |
| balance | REAL | 当前余额（负债为负数） |
| note | TEXT | 备注 |
| updated_at | DATETIME | 最后更新时间 |

示例数据：

| id | member_id | name | type | balance | updated_at |
|----|-----------|------|------|---------|------------|
| 1 | dad | 工商银行储蓄卡 | bank | 85000 | 2026-06-07 |
| 2 | mom | 支付宝 | ewallet | 12000 | 2026-06-07 |
| 3 | — | 家庭基金 | investment | 500000 | 2026-06-01 |
| 4 | dad | 房贷 | loan | -1200000 | 2026-06-05 |

→ "现在家里有多少钱？" → `SELECT SUM(balance) FROM finance_accounts`

**健康状态表 (health_profiles)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PK | |
| member_id | TEXT FK NOT NULL | 每个成员一条 |
| summary | TEXT | 当前健康状态的自然语言摘要 |
| updated_at | DATETIME | 最后更新时间 |

示例数据：

| id | member_id | summary | updated_at |
|----|-----------|---------|------------|
| 1 | dad | 整体健康，血压略高（每日服降压药），血糖正常，体重75kg | 2026-06-07 |
| 2 | mom | 血压偏高(140/90)，近期控制饮食中，无慢性病 | 2026-06-07 |
| 3 | son | 健康，花生过敏，体重35kg | 2026-06-01 |

**工作状态表 (work_status)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PK | |
| member_id | TEXT FK NOT NULL | 每个成员一条 |
| summary | TEXT | 当前工作状态摘要 |
| updated_at | DATETIME | 最后更新时间 |

示例数据：

| id | member_id | summary | updated_at |
|----|-----------|---------|------------|
| 1 | dad | Q3系统迁移进行中，7/15方案评审，每周一例会 | 2026-06-07 |
| 2 | mom | 季度汇报6/30截止，常规工作 | 2026-06-07 |

**家庭状态表 (family_status)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PK | 全家只有一条 |
| summary | TEXT | 当前家庭事务状态摘要 |
| updated_at | DATETIME | 最后更新时间 |

示例数据：`"下周三家长会待确认参加人，空调需清洗，暑假看护待安排"`

### 第二层：流水/记录表（发生了什么）

**财务流水表 (finance_records)**：

```sql
CREATE TABLE finance_records (
    id          TEXT PRIMARY KEY,
    member_id   TEXT,
    type        TEXT NOT NULL,       -- income/expense/transfer
    amount      REAL NOT NULL,
    currency    TEXT DEFAULT 'CNY',
    category    TEXT NOT NULL,       -- salary/food/housing/transport/medical/education/entertainment/other
    note        TEXT DEFAULT '',
    record_date TEXT NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    FOREIGN KEY (member_id) REFERENCES family_members(id)
);
```

示例数据：

| id | member_id | type | amount | category | note | record_date |
|----|-----------|------|--------|----------|------|-------------|
| a1 | dad | income | 30000 | salary | 6月工资 | 2026-06-01 |
| a2 | mom | income | 15000 | salary | 6月工资 | 2026-06-01 |
| a3 | dad | expense | 5200 | housing | 房贷月供 | 2026-06-05 |
| a4 | mom | expense | 380 | food | 周末超市采购 | 2026-06-07 |
| a5 | dad | expense | 2000 | education | 孩子暑期班 | 2026-06-03 |

→ "这个月花了多少" → `SELECT SUM(amount) FROM finance_records WHERE type='expense' AND strftime('%Y-%m', record_date)='2026-06'`
→ "最大的开支类别" → `SELECT category, SUM(amount) FROM finance_records WHERE type='expense' GROUP BY category ORDER BY SUM(amount) DESC`

**健康记录表 (health_records)**：

```sql
CREATE TABLE health_records (
    id          TEXT PRIMARY KEY,
    member_id   TEXT NOT NULL,
    type        TEXT NOT NULL,       -- vitals/checkup/medication/exercise/diet/sleep
    metric      TEXT,                -- blood_pressure/weight/heart_rate/blood_sugar/...
    value       TEXT,                -- "140/90", "72.5", "5.2"
    unit        TEXT,                -- mmHg/kg/bpm/mmol-L
    note        TEXT DEFAULT '',
    record_date TEXT NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    FOREIGN KEY (member_id) REFERENCES family_members(id)
);
```

示例数据：

| id | member_id | type | metric | value | unit | note | record_date |
|----|-----------|------|--------|-------|------|------|-------------|
| b1 | mom | vitals | blood_pressure | 140/90 | mmHg | 略偏高 | 2026-06-05 |
| b2 | mom | vitals | blood_pressure | 135/85 | mmHg | 比上次好 | 2026-06-01 |
| b3 | dad | checkup | blood_sugar | 5.2 | mmol/L | 正常范围 | 2026-05-20 |
| b4 | dad | medication | — | — | — | 降压药每天1粒 | 2026-06-01 |
| b5 | son | exercise | running | 3 | km | 晨跑 | 2026-06-07 |

→ "妈血压趋势" → `SELECT value, record_date FROM health_records WHERE member_id='mom' AND metric='blood_pressure' ORDER BY record_date`

**工作记录表 (work_records)**：

```sql
CREATE TABLE work_records (
    id          TEXT PRIMARY KEY,
    member_id   TEXT NOT NULL,
    type        TEXT NOT NULL,       -- project/deadline/meeting/business_trip/milestone
    title       TEXT NOT NULL,
    status      TEXT DEFAULT 'active', -- active/completed/cancelled
    priority    TEXT DEFAULT 'medium',
    project     TEXT DEFAULT '',
    due_date    TEXT,
    note        TEXT DEFAULT '',
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    FOREIGN KEY (member_id) REFERENCES family_members(id)
);
```

**家庭事务表 (family_records)**：

```sql
CREATE TABLE family_records (
    id              TEXT PRIMARY KEY,
    member_id       TEXT,
    type            TEXT NOT NULL,   -- schedule/chore/childcare/activity/shopping
    title           TEXT NOT NULL,
    status          TEXT DEFAULT 'pending', -- pending/done/cancelled
    location        TEXT DEFAULT '',
    participants    TEXT DEFAULT '[]', -- JSON array of member_ids
    scheduled_date  TEXT,
    note            TEXT DEFAULT '',
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL,
    FOREIGN KEY (member_id) REFERENCES family_members(id)
);
```

### 第三层：知识笔记表（持续有效的事）

```sql
CREATE TABLE knowledge_notes (
    id          TEXT PRIMARY KEY,
    domain      TEXT,               -- finance/health/work/family，可空
    member_id   TEXT,               -- 可空
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    tags        TEXT DEFAULT '[]',  -- JSON 数组
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    FOREIGN KEY (member_id) REFERENCES family_members(id)
);
```

存不适合放进流水表的非结构化信息：

| id | domain | member_id | title | content | tags |
|----|--------|-----------|-------|---------|------|
| e1 | health | dad | 饮食禁忌 | 不能吃太咸，少吃辣，医生嘱咐过 | ["饮食","医嘱"] |
| e2 | family | — | 家庭习惯 | 周末一般去外婆家吃饭 | ["习惯","周末"] |
| e3 | finance | — | 理财偏好 | 妈偏好稳健型基金，不碰股票 | ["理财","偏好"] |
| e4 | health | son | 过敏信息 | 对花生过敏 | ["过敏","重要"] |

### 通用表

```sql
-- 家庭成员
CREATE TABLE family_members (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    role        TEXT NOT NULL,   -- dad/mom/son/daughter 等
    birth_date  TEXT,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);
```

### 对话历史

```sql
-- AI 对话
CREATE TABLE conversations (
    id          TEXT PRIMARY KEY,
    title       TEXT DEFAULT '',
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE TABLE messages (
    id              TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    role            TEXT NOT NULL,   -- user/assistant/system
    content         TEXT NOT NULL,
    agent_used      TEXT,            -- butler/finance/health/work/family
    tokens_used     INTEGER DEFAULT 0,
    created_at      DATETIME NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
```

### 状态更新机制

两种方式并存：
1. **用户手动更新**：直接编辑状态表的 summary / balance 字段
2. **从流水自动推导**：录入新流水/记录后，系统调用 LLM 自动刷新对应的状态摘要

例如：录入 "妈血压 140/90" 的健康记录后，系统自动更新 `health_profiles` 中 mom 的 summary。

财务账户余额：可从流水聚合计算，也可手动维护（取决于用户记账习惯）。

### AI Agent 查询策略

每个专业 Agent 回答问题时查三层：

```
1. 状态表   → "现在怎么样"（当前快照）
2. 流水表   → "最近发生了什么"（趋势和详细数据）
3. 笔记表   → "有什么约束和偏好"（背景知识）
```

三层结果拼接成上下文，一起发给 LLM。

## API 接口

### 家庭成员

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/members` | 列出家庭成员 |
| POST | `/api/members` | 添加成员 |
| PUT | `/api/members/{id}` | 更新成员 |
| DELETE | `/api/members/{id}` | 删除成员 |

### 财务

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/finance/accounts` | 列出账户（含余额） |
| POST | `/api/finance/accounts` | 添加账户 |
| PUT | `/api/finance/accounts/{id}` | 更新账户 |
| GET | `/api/finance/records` | 列出流水（按日期、类别、成员过滤） |
| POST | `/api/finance/records` | 添加流水 |
| PUT | `/api/finance/records/{id}` | 更新流水 |
| DELETE | `/api/finance/records/{id}` | 删除流水 |

### 健康

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health/profiles` | 获取所有成员健康状态 |
| PUT | `/api/health/profiles/{member_id}` | 更新成员健康状态 |
| GET | `/api/health/records` | 列出健康记录 |
| POST | `/api/health/records` | 添加健康记录 |
| PUT | `/api/health/records/{id}` | 更新记录 |
| DELETE | `/api/health/records/{id}` | 删除记录 |

### 工作

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/work/status` | 获取工作状态 |
| PUT | `/api/work/status/{member_id}` | 更新工作状态 |
| GET | `/api/work/records` | 列出工作记录 |
| POST | `/api/work/records` | 添加工作记录 |
| PUT | `/api/work/records/{id}` | 更新记录 |
| DELETE | `/api/work/records/{id}` | 删除记录 |

### 家庭事务

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/family/status` | 获取家庭事务状态 |
| PUT | `/api/family/status` | 更新家庭事务状态 |
| GET | `/api/family/records` | 列出家庭事务记录 |
| POST | `/api/family/records` | 添加记录 |
| PUT | `/api/family/records/{id}` | 更新记录 |
| DELETE | `/api/family/records/{id}` | 删除记录 |

### 知识笔记

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/notes` | 列出笔记（按领域、成员、标签过滤） |
| POST | `/api/notes` | 添加笔记 |
| PUT | `/api/notes/{id}` | 更新笔记 |
| DELETE | `/api/notes/{id}` | 删除笔记 |

### AI 对话

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/conversations` | 列出对话 |
| POST | `/api/conversations` | 创建新对话 |
| GET | `/api/conversations/{id}/messages` | 获取消息列表 |
| POST | `/api/conversations/{id}/messages` | 发送消息，获取AI回答 |
| DELETE | `/api/conversations/{id}` | 删除对话 |

### 快速录入（自然语言）

`POST /api/quick-entry` 请求体：
```json
{
    "input": "6月工资到账3万，妈血压140/90，下周三家长会",
    "mode": "ai_parse"
}
```
响应：AI 拆分为多条跨领域的记录草稿，返回给用户确认后写入对应流水表和状态表。

### 配置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/config/ai` | 获取AI配置（API key 脱敏） |
| PUT | `/api/config/ai` | 更新AI配置（endpoint、model、API key） |

## 前端（Vue 3 + Vite）

### 页面

1. **仪表盘** (`/`) — 概览卡片（状态快照）+ 快速录入 + 最近记录
2. **对话** (`/chat`) — 与AI管家对话
3. **财务** (`/finance`) — 账户余额、收支流水
4. **健康** (`/health`) — 成员健康状态、指标趋势
5. **工作** (`/work`) — 项目、deadline
6. **家庭** (`/family`) — 日程、家务、育儿
7. **笔记** (`/notes`) — 知识笔记浏览/搜索
8. **成员** (`/members`) — 管理家庭成员
9. **设置** (`/settings`) — AI配置、通用设置

### 仪表盘布局

```
┌─────────────────────────────────────────┐
│  LifeOps                    [成员] [设置] │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐    │
│  │  问问管家...                      │    │  ← 主交互：对话输入框
│  └─────────────────────────────────┘    │
│                                         │
│  ┌──────────┐ ┌──────────┐ ┌────────┐  │
│  │ 财务概览  │ │ 健康概览  │ │ 本周日程 │  │  ← 来自状态表
│  │ 总资产    │ │ 爸: 血压略 │ │ 家长会   │  │
│  │ 39.7w    │ │    高      │ │ 买蛋糕   │  │
│  │ 本月支出  │ │ 妈: 血压偏 │ │ 空调清洗 │  │
│  │ 7580     │ │    高      │ │          │  │
│  └──────────┘ └──────────┘ └────────┘  │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │ + 添加记录                        │    │  ← 快速录入
│  │   [领域▼] [成员▼] [内容____]      │    │
│  └─────────────────────────────────┘    │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │ 最近记录                          │    │  ← 来自流水表
│  │ • 爸 6月工资到账 3w        财务    │    │
│  │ • 妈 血压 140/90          健康    │    │
│  │ • 爸 Q3项目deadline 7/15   工作    │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

## 实现策略

**渐进式**：先用裸 Go（goroutine + HTTP 调 LLM API）跑通，复杂度上升后再迁移到框架（如 go-agent）。

### Agent 实现（MVP）

每个 Agent 是一个实现通用接口的 Go struct：

```go
type Agent interface {
    Name() string
    Domain() string
    SystemPrompt() string
    RetrieveContext(query string, memberID string) (*AgentContext, error)
}

type AgentContext struct {
    Status   string          // 来自状态表
    RecentRecords []Record   // 来自流水表（近10条）
    Notes    []Note          // 来自知识笔记表
}
```

管家调用 `Agent.RetrieveContext()` 后，用 system prompt + AgentContext + 用户问题构建 LLM 请求。

### LLM 集成

复用现有 `ai_extractor.go` 的模式（OpenAI 兼容 API），扩展支持：
- 每个 Agent 可配置的 system prompt
- 从三层知识注入上下文
- 多轮对话历史
- 流式响应（SSE 给前端）

## MVP 范围

### 在范围内

- 家庭成员 CRUD
- 四领域（财务/健康/工作/家庭）的状态表 + 流水表 CRUD
- 知识笔记 CRUD
- 自然语言快速录入（AI拆分到对应领域）
- 多Agent对话（管家 + 4个专业Agent）
- 对话历史
- AI配置（endpoint、model、API key）
- Vue 3 前端：仪表盘、对话、各领域页面

### 明确不在范围内（后续迭代）

- 向量搜索 / 嵌入式 RAG
- 聊天记录导入（微信等）
- 数据导入（Excel/CSV）
- 重构现有收件箱/草稿模块
- 多用户 / 超出单token的鉴权
- 移动端 App
- 通知 / 提醒
- Agent 框架迁移

## 技术栈

| 层 | 选择 | 原因 |
|---|------|------|
| 后端 | Go（现有） | 复用现有代码，内存占用低 |
| 数据库 | SQLite（现有） | 家庭级数据量，免运维 |
| 全文搜索 | SQLite FTS5 | 内置，无额外依赖 |
| AI | OpenAI 兼容 API | BYOK，支持 DeepSeek/Qwen/Ollama 等 |
| 前端 | Vue 3 + Vite + PWA | 现代、轻量、移动端友好 |
| 部署 | Docker（现有） | 云服务器部署 |

## 环境变量（新增）

- `LIFEOPS_AI_ENDPOINT` — LLM API 地址（默认: `https://api.openai.com/v1`）
- `LIFEOPS_AI_MODEL` — 模型名称（默认: `gpt-4o-mini`）
- `LIFEOPS_AI_API_KEY` — LLM 服务的 API key
- `LIFEOPS_AI_MAX_TOKENS` — 最大响应 token 数（默认: `2048`）

已有变量不变：
- `LIFEOPS_ADDR` — 监听地址
- `LIFEOPS_WEBHOOK_TOKEN` — 认证 token
