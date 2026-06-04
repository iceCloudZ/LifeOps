# LifeOps / FamilyOps MVP 背景、规格与技术架构

> 评审目的：判断 LifeOps 这个长期方向是否值得启动，以及 FamilyOps 作为第一个 MVP 模块时，范围、架构、隐私边界和自托管路线是否合理。

## 1. 背景

当前讨论的长期方向是一个面向个人生活、家庭协作和工作事务的开源、自托管、隐私优先智能中枢。

暂定总项目名：**LifeOps**

第一阶段 MVP 模块名：**FamilyOps**

LifeOps 一句话定位：

> LifeOps 是一个开源、自托管、隐私优先的个人与家庭智能中枢，帮助用户管理工作、家庭、健康、财务、育儿、饮食和长期生活记录。

FamilyOps 一句话定位：

> FamilyOps 是一个开源、自托管、可接入 Home Assistant 的家庭 AI 收件箱，帮助夫妻把家庭碎片信息整理成日程、待办、购物清单和每日简报。

LifeOps 不和腾讯元宝、豆包、通义、ChatGPT 这类通用 AI 助理正面竞争。大厂会占据通用入口、模型能力和生态连接，但它们很难天然获得用户对个人和家庭长期敏感数据的信任，例如工作上下文、孩子信息、夫妻沟通、家庭财务、健康记录、保单、家庭分工等。

选择先做 FamilyOps，是因为家庭事务同时具备高频、真实痛点、隐私敏感和自托管差异化：

- 家庭事务每天都发生，但信息散落在微信、学校通知、账单、口头沟通、截图、日历和脑子里。
- 夫妻协作里最痛的不是“不会做计划”，而是“没人统一接住信息、没人确认负责人、没人持续提醒和复盘”。
- Home Assistant / NAS / 自托管用户已经接受本地优先、Docker、BYOK、自建服务和家庭自动化。
- 开源能增强信任，降低“家庭数据进平台服务器”的顾虑。

LifeOps 的长期模块可以包括：

- FamilyOps：家庭事务、夫妻协作、家庭简报、Home Assistant 集成。
- WorkOps：工作收件箱、项目任务、会议纪要、日报周报、工作复盘。
- FinanceOps：家庭账单、预算、保单、投资研究记录。
- HealthOps：健康记录、饮食、运动、体检和习惯。
- ParentingOps：育儿记录、学校通知、亲子计划。

但 MVP 只实现 FamilyOps，不实现其他模块。

## 2. 目标用户

MVP 不面向普通大众家庭，也不覆盖完整个人工作流，而先面向以下用户：

- 技术人家庭：愿意 Docker 部署、配置 API Key 或 Ollama。
- Home Assistant 用户：已有家庭自动化基础，理解本地优先。
- NAS / Homelab 用户：愿意把家庭服务部署在自己的设备上。
- 隐私敏感家庭：不希望孩子、财务、健康、夫妻沟通数据默认进入第三方平台。
- 双职工夫妻：日程、孩子、家务、购物、账单和临时事项密集。

第一使用场景是“夫妻共同使用”，不是单人效率工具。

## 3. 产品原则

### 3.1 本地优先

家庭和个人数据默认存在用户自己的设备、NAS 或私有服务器中。LifeOps 官方不默认托管用户明文数据。

### 3.2 AI 只做管家，不做裁判

AI 的角色是整理、提醒、复盘和沟通辅助，不评价夫妻谁对谁错，不替用户做高风险决策。

### 3.3 先确认，再入库

AI 从收件箱中抽取日程、任务、购物项和记录后，只生成“待确认事项”。用户确认后才正式写入家庭数据。

### 3.4 小而稳

MVP 不做完整 LifeOps，也不做完整家庭操作系统，只做 FamilyOps 里的家庭收件箱、今日面板、每日简报和 Home Assistant 初步集成。

### 3.5 用户掌控模型

MVP 支持 BYOK，即用户自带模型 API Key 或本地模型服务。LifeOps 不承担默认大模型成本。

BYOK 指 Bring Your Own Key，用户可以配置 OpenAI-compatible endpoint、DeepSeek、通义、OpenRouter、Ollama、LM Studio 等。

## 4. MVP 范围

### 4.1 核心目标

FamilyOps MVP 要验证一个问题：

> 夫妻是否愿意把家庭碎片信息丢进一个自托管 AI 收件箱，并持续使用它生成家庭待办、日程、购物清单和每日简报。

### 4.2 MVP 功能

#### 家庭空间

- 创建一个家庭空间。
- 创建家庭成员。
- 第一版至少支持两个成人账号。
- 成员角色先简化为 owner 和 member。

#### 家庭收件箱

用户可以输入家庭碎片信息，例如：

- “周五孩子要带彩笔，别忘了交水费。”
- “下周二 18:30 家长会，我可能来不及，你看能不能去。”
- “周日买牛奶、鸡蛋、纸巾，晚上少油少辣。”
- “保险缴费 6 月 20 日到期，提醒我看一下。”

MVP 支持文本输入。图片、语音、OCR、微信 Bot 先不作为核心能力。

但 MVP 应预留一个轻量的 Webhook Inbox 接口，用于接收外部系统转发的文本消息。这样国内技术用户可以自行通过 WeChatFerry、企业微信机器人、Telegram、Home Assistant 自动化或其他工具，把消息推入收件箱，而项目本身不需要在第一版维护复杂的微信登录、微信 Bot 或平台合规。

建议接口：

```text
POST /api/inbox/webhook
```

最小请求体：

```json
{
  "source": "wechat",
  "sender": "partner",
  "content": "周五孩子要带彩笔，别忘了交水费。"
}
```

#### AI 整理

AI 将收件箱内容整理成结构化草稿：

- event：日程，例如家长会、体检、兴趣班。
- task：待办，例如交水费、修理、接送、准备材料。
- shopping_item：购物项，例如牛奶、鸡蛋、纸巾。
- note：普通记录，例如孩子最近不爱吃辣、某个事项背景。

AI 输出必须可解释，并进入待确认状态。

#### 输出守卫

MVP 需要把 AI Extractor 设计成“模型调用 + 输出守卫”的两层结构，而不是直接信任模型返回。

输出守卫负责：

- 优先使用 OpenAI-compatible 的 JSON mode 或 structured output。
- 对模型返回做 JSON Schema 校验。
- 允许模型返回多个 draft，但每个 draft 必须有明确 draft_type。
- 校验失败时进行一次结构化重试。
- 重试仍失败时，把 inbox_item 标记为 extraction_failed，并提示用户手动整理。
- 不允许格式错误的 AI 输出直接进入待确认事项。

这对本地小模型尤其重要。本地 7B 级模型可能会输出解释性文本、缺字段或混入 Markdown，必须通过守卫层兜底。

#### 待确认事项

- 展示 AI 解析出的事项。
- 用户可以确认、编辑、删除。
- 确认后写入正式数据表。
- 未确认事项在今日面板中提示。

#### 今日面板

展示：

- 今天和明天的日程。
- 未完成待办。
- 待确认事项。
- 购物清单。
- 重要提醒。

#### 家庭简报

支持生成每日简报：

- 今天需要关注的日程。
- 谁负责哪些待办。
- 哪些事项还未确认。
- 明天需要提前准备什么。
- 简短风险提示，例如时间冲突、任务无人负责。

MVP 可以先手动点击生成，后续再做定时生成。

#### Home Assistant 初步集成

第一版不做完整 Home Assistant Add-on，先做低成本集成：

- 提供 REST API，供 Home Assistant 拉取今日简报和待办统计。
- 提供 Webhook，在重要提醒变化时推送。
- 可选 MQTT 发布状态。

示例 MQTT topic：

- `lifeops/family/today/briefing`
- `lifeops/family/tasks/pending_count`
- `lifeops/family/inbox/unconfirmed_count`
- `lifeops/family/shopping/count`
- `lifeops/family/reminders/urgent`

### 4.3 MVP 不做

以下能力不进入 MVP：

- 医疗诊断。
- 投资建议。
- 保险销售推荐。
- 心理咨询或婚姻关系评判。
- 儿童直接聊天。
- 位置追踪。
- 照片人脸识别。
- 语音识别。
- OCR。
- 多家庭复杂权限。
- 端到端加密家庭同步。
- 移动原生 App。
- 官方云托管。
- 完整 Home Assistant Add-on。

这些能力后续可以做，但不能影响 MVP 验证。

## 5. 核心数据模型

MVP 只实现 FamilyOps 数据模型。为了给 LifeOps 长期演进留空间，可以在设计上预留 workspace 概念，但第一版不需要实现多 workspace UI。

### workspace

长期用于表达不同生活域，例如 family、work、finance、health。

MVP 可以暂不建表，或只在 family 表中隐含一个默认 family workspace。

字段建议：

- id
- owner_user_id
- workspace_type：family、work、finance、health、personal
- name
- created_at
- updated_at

MVP 需要的数据对象：

### family

FamilyOps 的家庭空间。

字段建议：

- id
- name
- timezone
- created_at
- updated_at

### family_member

家庭成员，包括夫妻和孩子。MVP 中只有成人账号能登录，孩子只是家庭档案对象。

字段建议：

- id
- family_id
- display_name
- role
- member_type：adult 或 child
- created_at
- updated_at

### user_account

登录账号。

字段建议：

- id
- family_id
- member_id
- username
- password_hash
- role：owner 或 member
- created_at
- updated_at

### inbox_item

家庭收件箱条目。

字段建议：

- id
- family_id
- created_by
- source_type：text、image、voice、webhook
- content
- status：new、extracted、confirmed、archived
- created_at
- updated_at

### extracted_draft

AI 解析出来的待确认草稿。

字段建议：

- id
- family_id
- inbox_item_id
- draft_type：event、task、shopping_item、note
- title
- description
- due_at
- assignee_member_id
- confidence
- raw_json
- status：pending、confirmed、dismissed
- created_at
- updated_at

### family_event

正式日程。

字段建议：

- id
- family_id
- title
- description
- starts_at
- ends_at
- participant_member_ids
- source_inbox_item_id
- created_at
- updated_at

### family_task

正式待办。

字段建议：

- id
- family_id
- title
- description
- assignee_member_id
- due_at
- status：open、done、cancelled
- source_inbox_item_id
- created_at
- updated_at

### shopping_item

购物项。

字段建议：

- id
- family_id
- name
- quantity
- status：open、bought、cancelled
- source_inbox_item_id
- created_at
- updated_at

### family_note

家庭记录。

字段建议：

- id
- family_id
- title
- content
- topic：general、child、health、finance、home
- source_inbox_item_id
- created_at
- updated_at

### briefing

家庭简报。

字段建议：

- id
- family_id
- briefing_type：daily、weekly
- period_start
- period_end
- content
- generated_by_model
- created_at

### ai_audit_log

AI 调用审计。

字段建议：

- id
- family_id
- provider
- model
- purpose：extract、briefing、rewrite
- input_ref_type
- input_ref_id
- token_usage
- created_at

审计日志只记录调用元信息，不记录完整敏感 prompt，避免二次泄露。

## 6. 技术架构

### 6.1 总体架构

```text
------------------+
| Web / PWA        |
| Mobile browser   |
+--------+---------+
         |
         | REST / SSE
         v
+---------------------------+
| LifeOps Server            |
|                           |
| - Auth                    |
| - Workspace               |
| - FamilyOps Module        |
| - Inbox Service           |
| - AI Extractor            |
| - Review Workflow         |
| - Briefing Service        |
| - Integration Service     |
| - Scheduler               |
+-------------+-------------+
              |
              v
+---------------------------+
| Local Data Store          |
| SQLite for MVP            |
| PostgreSQL optional       |
+---------------------------+
              |
              v
+---------------------------+
| External / Local LLM      |
| OpenAI-compatible API     |
| Ollama / LM Studio        |
| DeepSeek / Qwen / etc.    |
+---------------------------+
              |
              v
+---------------------------+
| Home Assistant            |
| REST / Webhook / MQTT     |
+---------------------------+
```

### 6.2 推荐技术栈

后端推荐：**Go + 标准库优先 + 少量必要依赖**

前端推荐：**Vue 3 + Vite + PWA**

首批目标用户是 Home Assistant、NAS、Homelab 和技术家庭。这类用户常部署在低功耗设备上，例如 Intel J4125、N100、ARM NAS、树莓派或旧 mini PC。因此 MVP 后端应优先考虑低资源占用、镜像小、启动快和多架构构建容易。

已完成同等功能 spike：

- Quarkus JVM
- Quarkus Native
- Go 标准库

同一核心能力包括：

- `POST /api/inbox/webhook`
- 内存收件箱写入
- AI draft JSON 输出守卫
- 自动化测试

实测摘要：

| 指标 | Quarkus JVM | Quarkus Native | Go |
|---|---:|---:|---:|
| 首次成功请求 | 3.1s | 57ms | 25ms |
| 运行内存 | 168.6 MB RSS | 48.4 MB RSS | 6.2 MB RSS |
| 产物大小 | 18.1 MB `quarkus-app` | 60 MB executable | 8.4 MB executable |
| 构建复杂度 | Maven/JVM | GraalVM + gcc + reflection hints | 直接 cross-build |

详细数据见：

- `docs/backend-spike-comparison.md`
- `docs/quarkus-graalvm-spike.md`
- `docs/go-spike.md`

结论：

> LifeOps MVP 后端推荐选择 Go。Quarkus Native 可作为 Java 备选，但不作为开源首发默认路线。

选择 Go 的原因：

- 单二进制，部署简单。
- 内存占用极低，适合 NAS 和 Homelab。
- Docker 镜像可以非常小。
- linux/amd64 和 linux/arm64 cross-build 简单。
- 适合 REST API、Webhook、MQTT、定时任务、SQLite 和 OpenAI-compatible HTTP 调用。
- 更符合 Home Assistant / NAS 用户对自托管工具的预期。

Go 代码组织要求：

- 不使用复杂框架作为默认方案。
- 标准库优先，必要时引入小而稳的依赖。
- 代码结构按 Java 工程师也能读懂的方式组织：handler、service、store、model、config。
- 避免 Go 黑魔法和过度抽象。
- 所有 AI 输出必须经过结构化解析和守卫，不直接信任模型文本。

Quarkus Native 备选说明：

- Native runtime 表现不错，启动和内存都明显优于 JVM。
- 但 native build 需要 GraalVM/gcc 或平台构建工具，构建链更重。
- DTO/JSON 序列化需要维护 native reflection metadata。
- 当前 spike 中 native build 峰值 RSS 约 3GB，native executable 约 60MB。
- 若后续必须复用 Java Agent 生态，可重新评估 Quarkus Native。

前端：

- Vue 3
- Vite
- Element Plus
- PWA 支持

数据库：

- MVP 默认 SQLite，降低部署门槛。
- 长期支持 PostgreSQL，适合 NAS 或长期家庭数据。
- Go 后端需要对 SQLite 写操作做进程内串行化，例如单写队列或写互斥锁。
- MVP 的家庭并发很低，但夫妻多设备同时写入时，串行化写入能避免偶发 `database is locked`，用几毫秒延迟换稳定性。

AI 接入：

- OpenAI-compatible API 优先。
- 支持 base_url、model、api_key 配置。
- 支持 Ollama / LM Studio 本地模型。
- 支持 DeepSeek、通义千问、OpenRouter 等预设 endpoint，降低国内用户配置门槛。
- AI Extractor 必须包含输出守卫，确保 JSON Schema 校验、失败重试和降级。
- Prompt 策略需要文档化，见 `docs/prompt-strategy.md`。重点包括 few-shot、严格 JSON 输出、失败重试提示词、本地小模型适配和敏感信息边界。

部署：

- Docker Compose。
- 单容器优先，降低复杂度。
- 可选 PostgreSQL compose。
- 提供多架构镜像，至少覆盖 linux/amd64 和 linux/arm64。
- 面向国内用户提供镜像拉取替代说明，例如阿里云镜像源、GitHub Container Registry 或手动构建方式。
- MVP 稳定后补充 1Panel / CasaOS 安装配置，作为 Homelab 圈子的传播入口。

Home Assistant 集成：

- REST API。
- Webhook。
- MQTT。
- 后续再做 Add-on 和 Lovelace Card。

## 7. 部署方案

### 7.1 MVP 默认部署

用户通过 Docker Compose 部署：

```yaml
services:
  lifeops:
    image: lifeops/lifeops:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - LIFEOPS_DB=sqlite
      - LIFEOPS_DB_PATH=/data/lifeops.db
      - LIFEOPS_LLM_BASE_URL=http://host.docker.internal:11434/v1
      - LIFEOPS_LLM_MODEL=qwen2.5:7b
```

访问地址：

```text
http://localhost:8080
```

或家庭内网：

```text
http://nas-ip:8080
```

### 7.2 手机访问

自托管服务天然部署在内网。MVP 文档应明确推荐：

首选：

- Tailscale
- WireGuard
- ZeroTier

进阶：

- Caddy / Nginx Proxy Manager 反向代理。
- Cloudflare Tunnel。
- Tailscale Funnel。

安全警告：

> 不要把 LifeOps 的 8080 端口直接裸露到公网。若需要公网访问，必须使用 HTTPS 反向代理、强认证、限流和合理的安全配置。

默认安全建议：

> LifeOps is designed for private home networks. For mobile access outside home, use Tailscale or WireGuard. Do not expose LifeOps directly to the public internet unless you understand reverse proxy and security configuration.

### 7.3 后续商业化远程访问

后续可以参考 Home Assistant Cloud / Nabu Casa，提供官方远程访问中继：

- 用户家庭服务仍在本地运行。
- 官方云只做加密隧道和访问入口。
- 可作为订阅服务。

该能力不进入 MVP。

### 7.4 国内部署注意事项

国内用户拉取 Docker Hub 镜像可能超时。开源项目应提供以下备选方式：

- GitHub Container Registry 镜像地址。
- 国内镜像源配置说明。
- 从源码本地构建镜像的命令。

示例：

```bash
docker build -t lifeops/lifeops:local .
docker compose up -d
```

### 7.5 Homelab 面板安装

Docker Compose 是 MVP 的首选部署方式。等核心功能稳定后，应补充以下部署入口：

- CasaOS 一键安装配置。
- 1Panel 应用商店配置或手动安装说明。
- 群晖 Container Manager 示例。

这些不是 MVP 核心功能，但会影响 Home Assistant / NAS / Homelab 用户的真实采用率。

## 8. 隐私与安全边界

### 8.1 默认隐私承诺

- 家庭和个人数据默认存储在用户自己的设备或服务器。
- API Key 存储在本地服务中，不上传到 LifeOps 官方服务器。
- MVP 不提供官方云端 AI 转发。
- AI 调用只发给用户配置的模型服务。
- AI 调用记录写入本地审计日志。

### 8.2 敏感数据策略

MVP 避免处理高风险建议，只做信息整理。

敏感主题包括：

- 孩子信息。
- 健康信息。
- 财务信息。
- 保险信息。
- 夫妻沟通。

AI 可以整理这些信息，但不能输出诊断、投资、保险销售或心理咨询结论。

### 8.3 认证

MVP 需要基础账号密码登录。

最低要求：

- 密码 hash 存储。
- Session 或 JWT。
- CSRF / CORS 合理配置。
- 默认不允许匿名访问家庭数据。

后续增强：

- 双因素认证。
- 家庭邀请。
- 细粒度权限。

## 9. Home Assistant 集成设计

### 9.1 REST API

示例接口：

```text
GET /api/ha/today
GET /api/ha/summary
GET /api/ha/tasks/count
GET /api/ha/reminders/urgent
```

返回内容应适合 Home Assistant template sensor 消费。

### 9.2 Webhook

LifeOps / FamilyOps 可以在以下事件触发 webhook：

- 新的紧急提醒。
- 每日简报生成。
- 待确认事项数量变化。
- 今日任务全部完成。

### 9.3 Webhook Inbox

MVP 应提供一个入站 Webhook，把外部文本消息写入家庭收件箱。

示例接口：

```text
POST /api/inbox/webhook
```

用途：

- 技术用户自行桥接微信消息。
- 接入 Home Assistant 自动化。
- 接入 Telegram、企业微信、飞书或自定义脚本。
- 从其他自托管系统推送家庭事项。

安全要求：

- Webhook 必须配置 token。
- 默认关闭，用户显式启用。
- Webhook 只创建 inbox_item，不直接创建正式任务或日程。

### 9.3 MQTT

LifeOps / FamilyOps 可以发布状态到 MQTT broker：

```text
lifeops/family/today/briefing
lifeops/family/tasks/pending_count
lifeops/family/inbox/unconfirmed_count
lifeops/family/shopping/count
lifeops/family/reminders/urgent
```

MVP 可以先实现 Webhook，MQTT 作为可选。

## 10. 开源项目形态

建议项目不是从 App 开始，而是从开源自托管项目开始。

建议仓库结构：

```text
lifeops/
  apps/
    api/                      # Go backend, official MVP API service
    web/                      # Vue 3 + Vite + PWA frontend
  internal/
    family/
      inbox/                  # family inbox domain
      drafts/                 # AI extracted draft review queue
      tasks/                  # family tasks
      events/                 # family events
      briefing/               # daily/weekly briefings
    ai/
      extractor/              # OpenAI-compatible extraction flow
      guardrails/             # JSON/schema output guard
    config/
    storage/
      sqlite/
    integrations/
      homeassistant/
      mqtt/
      webhook/
  deploy/
    docker-compose.yml
  docs/
    vision.md
    architecture.md
    mvp-familyops.md
    privacy.md
    home-assistant.md
    remote-access.md
    backend-spike-comparison.md
  examples/
    inbox-samples.md
```

当前仓库中的 `apps/api` 已由 Go spike 迁移为正式 MVP API 起点。Quarkus spike 已移入 `experiments/quarkus`，仅作为备选技术验证材料保留。

README 需要突出：

- Self-hosted。
- Local-first。
- BYOK。
- Home Assistant friendly。
- FamilyOps as first module。
- Family inbox。
- No cloud required。

## 11. 商业化方向

MVP 阶段先不商业化，但架构应保留后续空间。

可能路径：

### 开源核心 + 付费托管

技术用户自托管免费，普通用户购买托管版。

### 开源核心 + 官方远程访问

类似 Home Assistant Cloud，提供加密远程访问中继。

### 开源核心 + Pro 功能

可能 Pro 功能：

- 家庭高级权限。
- 自动备份。
- 多家庭空间。
- Home Assistant Add-on 一键安装。
- 移动 App。
- 高级简报模板。

### 部署服务

为 NAS / Homelab 用户提供一次性部署服务。

## 12. 主要风险

### 12.1 用户群窄

自托管、BYOK、Home Assistant 用户不是大众家庭。MVP 应接受这是小众切口。

### 12.2 家庭协作留存难

夫妻共同使用需要非常低摩擦。收件箱和今日面板必须比复杂配置更重要。

### 12.3 AI 解析不稳定

必须通过“待确认事项”降低 AI 出错风险。

同时需要把 prompt 策略文档化。AI Extractor 的效果高度依赖提示词，尤其是本地 7B 级小模型。缺少 prompt 文档会让社区难以复现、评审和优化解析效果。

### 12.4 远程访问复杂

NAS 在内网，手机外网访问需要 VPN、Tunnel 或反向代理。MVP 需要文档引导，不能假设所有用户都会。

### 12.5 范围膨胀

家庭、健康、饮食、育儿、财务、关系都很诱人，但 MVP 只能做整理和提醒。

### 12.6 大厂竞争

大厂会做通用 AI 助理，但 LifeOps 通过开源、自托管、Home Assistant 集成和个人/家庭事务深度避开正面竞争。

### 12.7 Go 学习与维护成本

Go 后端更适合开源自托管分发，但主要开发者已有 Java 经验。MVP 需要控制 Go 代码风格：

- 标准库优先。
- 少量稳定依赖。
- 清晰分层。
- 不使用复杂框架和过度抽象。
- 测试先行，避免边学边堆出难维护结构。

### 12.8 SQLite 写锁

SQLite 适合家庭级自托管，但多设备同时写入时可能出现写锁竞争。Go 后端应将 SQLite 写操作集中到 store 层，并通过写互斥锁或单写队列串行化。

这不需要引入复杂队列系统。MVP 可以用进程内 `sync.Mutex` 或一个单 worker channel 处理写事务。

## 13. 需要评审的问题

请重点评审以下问题：

1. LifeOps 是否应该先作为开源自托管项目，而不是 App 或小程序？
2. MVP 是否应该只保留“家庭收件箱 + 今日面板 + 每日简报”？
3. SQLite 默认、PostgreSQL 可选是否合理？
4. 基于 spike 数据，Go + Vue 是否应作为正式 MVP 技术栈？
5. BYOK 是否足以支撑 MVP 的 AI 能力？
6. Home Assistant 集成应从 REST/Webhook 开始，还是一开始就做 Add-on？
7. 是否需要在 MVP 中加入 OCR 或语音输入？
8. 手机远程访问是否应默认推荐 Tailscale？
9. 隐私边界是否足够清楚？
10. 这个项目最早的目标用户是否应该锁定 Home Assistant / NAS / 技术家庭？
11. LifeOps 总项目 + FamilyOps 首模块的命名和边界是否清晰？
12. 是否需要在 MVP 中预留 WorkOps 数据结构，还是只在文档中表达长期愿景？
13. Quarkus spike 是否应保留在仓库中，还是移入 `experiments/` 或删除以减少噪音？
14. SQLite 写入应使用简单互斥锁，还是从一开始使用单 worker 写队列？
15. CasaOS / 1Panel 部署配置应放在 MVP 后第几个里程碑？

## 14. 建议结论

建议进入下一阶段，但保持 MVP 克制。

下一阶段不应直接开发完整 LifeOps 或 App，而应先建立开源项目骨架和 FamilyOps 可运行 demo：

1. 将 Go spike 迁移为正式 `apps/api`。
2. Docker Compose 一键启动。
3. Web/PWA 登录和家庭初始化。
4. SQLite 持久化文本收件箱。
5. AI 结构化抽取。
6. 待确认事项。
7. 今日面板。
8. 手动生成每日简报。
9. Home Assistant REST/Webhook 初步集成。
10. 补充 prompt 策略文档和 SQLite 写入串行化测试。

若该 demo 能让少量技术家庭持续使用，再扩展 OCR、语音、移动 App、官方远程访问和商业化能力。
