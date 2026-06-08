# Lifestyle Skills 开源路线图

**日期**: 2026-06-08
**状态**: 进行中
**范围**: 先建设独立的生活方式 skill 开源库，再集成到 LifeOps 的家庭知识库和管家 agent。

## 当前进度

截至 2026-06-08：

- Phase 1 已完成：独立 public GitHub 仓库、schema、registry、CI、README 和贡献约束已经建立。
- Phase 2 已完成第一版：1 个 `life-butler` 入口 skill，4 个领域共 12 个 lens。
- Phase 3 已达到第一阶段覆盖：`life-butler` 有 5 条首次使用 eval，每个 lens 有 3 条 synthetic eval，共 41 条。
- Phase 4 进行中：已有 metadata 校验、静态测试、单个/批量 eval prompt builder、baseline / with-skill / blind judge 三段式 runner 和 scoreboard builder；尚未发布真实 LLM 评分结果。
- Phase 5 尚未开始：LifeOps 还没有接入外部 `registry.json` 和 lens router。
- Weekly Review 暂不进入近期开发范围，只有在存在足够历史数据或用户主动要求复盘时才考虑。

## 背景

LifeOps 的家庭知识库和管家 agent 需要的不只是工具调用能力，而是能根据家庭成员、生活目标、约束和历史反馈，选择合适的思考流派。

这套能力不应该一开始就写死在 LifeOps 内部。更合适的边界是：

- `lifestyle-skills`: 独立开源的生活方式 skill 标准库，负责沉淀财务、健康、运动、家庭关系等领域的思考模式。
- `LifeOps`: 本地家庭数据产品，负责读取这些 skill，根据家庭上下文智能选择、组合和复盘。

这样可以让 skill 先作为可复用资产独立演进，也能避免 LifeOps 过早绑定某一批静态 prompt。

## 核心判断

这套 skill 适合单独开源，但它不应该只是 prompt 仓库。

它应该是一个可路由、可评估、可解释的生活方式思考库。每个 skill 需要声明自己适合谁、不适合谁、需要哪些上下文、如何推理、输出什么、有哪些安全边界，以及如何用场景评估它是否真的有用。

LifeOps 后续只把它当作外部方法库来消费。当前开源仓库采用 **一个入口 skill + 多个非触发 lens** 的结构：

```text
家庭成员画像 + 本地家庭数据 + 用户问题
  -> life-butler 入口 skill / SkillRouter 选择领域和候选 lens
  -> 用户确认或 LifeOps 根据历史反馈确认 lens
  -> DomainAgent 基于 lens 生成领域建议
  -> ButlerAgent 汇总跨领域冲突
  -> 有足够历史数据后，才进入 Review/WeeklyReview
```

## 非目标

第一阶段不做这些事：

- 不接 Claude Desktop、Cursor 或其他外部客户端。
- 不把 MCP/Tool Adapter 作为主线。
- 不让 skill 直接读写 LifeOps 本地数据。
- 不把 skill 写成某个真实人物的仿冒人格。
- 不承诺医疗、投资、法律、心理治疗等专业结论。

## 开源仓库形态

建议创建独立仓库：

```text
lifestyle-skills/
  README.md
  LICENSE
  registry.json
  schemas/
    lifestyle-skill.schema.json
  skills/
    life-butler/
      SKILL.md
      references/
        no-data-mode.md
        lens-catalog.md
        selection-rules.md
        scoring-model.md
  lenses/
    finance/
      conscious-spending.md
      bogleheads-style.md
      zero-based-budgeting.md
    health/
      longevity-medicine.md
      blue-zones-style.md
      daily-dozen-style.md
    movement/
      zone2-longevity.md
      strength-baseline.md
      tiny-habits-movement.md
    family/
      gottman-style.md
      nvc-style.md
      positive-discipline-style.md
  evals/
    life/
    finance/
    health/
    movement/
    family/
  examples/
    members/
    cases/
```

## Skill / Lens 标准格式

当前只有 `skills/life-butler/SKILL.md` 是用户安装后会触发的入口 skill。它负责：

- 首次使用时进入 No Data Mode。
- 判断领域。
- 根据 `registry.json` 和 selection rules 选择 2 到 3 个候选 lens。
- 推荐一个 lens，但让用户确认。
- 用户确认后再按 lens 输出建议。

每个 lens 在 `registry.json` 里至少包含：

```yaml
name: conscious-spending
skill_type: lens
domain: finance
description: Helps evaluate household spending through a values-first conscious spending lens.
best_for:
  - high-income-but-low-clarity
  - wants-more-enjoyment
  - dislikes-strict-budgeting
avoid_if:
  - debt-crisis
  - no-emergency-buffer
  - needs-regulatory-financial-advice
required_context:
  - income
  - fixed_costs
  - savings_goal
  - member_preferences
style:
  - values-first
  - low-shame
  - action-oriented
risk_level: low
```

每个 lens 正文建议包含：

- `Purpose`: 这个 skill 解决什么问题。
- `When to Use`: 什么场景适合触发。
- `When Not to Use`: 哪些场景应该避开。
- `Thinking Steps`: 推理步骤。
- `Output Format`: 输出结构。
- `Safety Boundaries`: 安全边界和免责声明。
- `Examples`: 典型输入输出或应用方式。
- `Evaluation Notes`: 评估时要看什么。

入口 skill 还需要维护：

- `no-data-mode.md`: 首次使用和无本地数据时怎么问问题。
- `lens-catalog.md`: 可选 lens 的人类可读目录。
- `selection-rules.md`: 常见用户 blocker 到 lens 的映射。
- `scoring-model.md`: 可解释选择协议，包括 `domain_match`、`intent_match`、`blocker_match`、`best_for_match`、`avoid_if_risk`、`missing_context`、`confirmation_required`。

## MVP 领域

第一版不要铺太大，先做 4 个领域，每个领域 3 个 skill。

### Finance

- `conscious-spending`: 价值优先、低羞耻感的支出规划。
- `bogleheads-style`: 长期、低成本、分散、少择时的投资思路。
- `zero-based-budgeting`: 每笔钱都有任务，适合现金流紧张或需要强约束的家庭。

### Health

- `longevity-medicine`: 长期健康、风险管理、指标趋势、预防优先。
- `blue-zones-style`: 饮食、活动、社交、压力、生活环境的综合改善。
- `daily-dozen-style`: 以日常饮食结构为核心的健康习惯检查。

### Movement

- `zone2-longevity`: 有氧基础、可持续训练、长期心肺能力。
- `strength-baseline`: 基础力量、渐进超负荷、动作安全。
- `tiny-habits-movement`: 低门槛、微习惯、适合刚开始或反复失败的人。

### Family

- `gottman-style`: 关系修复、冲突降温、积极互动。
- `nvc-style`: 观察、感受、需要、请求，适合沟通冲突。
- `positive-discipline-style`: 面向孩子的边界、鼓励和长期能力建设。

## 如何证明 Skill 有用

生活方式 skill 不能只靠“看起来很懂”证明价值。需要建立分层评估。

### Level 1: 格式正确

- `SKILL.md` 可被加载。
- metadata 完整。
- `best_for`、`avoid_if`、`required_context` 清晰。
- 没有危险指令或越界承诺。

### Level 2: 触发正确

- 该用这个 skill 时能选中。
- 不该用时不会误触发。
- 多个 skill 都可用时，能解释为什么选主 skill 和辅助 skill。

### Level 3: 单次建议更好

同一个家庭场景下，对比三种输出：

- base agent: 不使用 skill。
- wrong skill: 使用不匹配的 skill。
- matched skill: 使用正确匹配的 skill。

评价维度：

- 是否符合该流派原则。
- 是否考虑家庭成员画像和约束。
- 是否给出可执行下一步。
- 是否避免过度承诺。
- 是否没有冒充专业顾问。

### Level 4: 长期复盘有效

进入 LifeOps 后，用真实家庭反馈观察 4 到 8 周：

- 用户是否采纳建议。
- 用户是否修改建议。
- 用户是否忽略建议。
- 提醒是否减少噪音。
- 每周复盘是否帮助发现长期趋势。

## Eval Case 格式

每个 skill 配 5 到 10 个评估场景。示例：

```json
{
  "member_profile": {
    "age": 35,
    "constraints": ["sedentary", "limited_time"],
    "goals": ["fat_loss", "build_consistency"],
    "history": ["failed_three_previous_plans"]
  },
  "situation": "想减脂，但过去三次运动计划都坚持不下来。",
  "expected_skill": "tiny-habits-movement",
  "avoid_skills": ["advanced-strength-program"],
  "rubric": [
    "建议足够小",
    "不羞辱用户",
    "包含触发场景",
    "有明确下一步",
    "没有过度承诺"
  ]
}
```

## Scoreboard

开源仓库应维护一个简单评分表，哪怕第一版是人工评分。

```text
skill_name
domain
trigger_accuracy
fit_score
safety_score
actionability_score
notes
```

评分目的不是证明 skill 永远正确，而是让贡献者和使用者知道：

- 哪些 skill 已经比较稳定。
- 哪些 skill 还缺少测试场景。
- 哪些 skill 容易误触发。
- 哪些领域存在安全风险。

## LifeOps 集成边界

LifeOps 不负责定义所有生活流派。它负责用本地家庭数据智能选择和组合 skill。

建议新增三个核心模块：

```text
SkillRegistry
  读取本地或远程 lifestyle skills。
  校验 schema。
  暴露 skill metadata。

SkillRouter
  根据家庭成员画像、用户问题、历史采纳反馈选择 skill。
  输出 primary_skill、secondary_skills 和选择理由。

DomainAgent
  在 finance、health、movement、family 等领域内使用选中的 skill。
  生成符合家庭上下文的建议。
```

管家 agent 的职责是更高一层：

```text
ButlerAgent
  汇总多个领域 agent 的建议。
  识别冲突。
  按家庭优先级做取舍。
  生成最终建议和提醒草案。
```

## 每周复盘路径

每周主动复盘放在 LifeOps 集成后做。

```text
WeeklyReviewAgent
  1. 收集过去一周家庭数据。
  2. 调用各领域 DomainAgent。
  3. 每个 DomainAgent 自行选择合适 skill。
  4. ButlerAgent 汇总跨领域建议。
  5. 生成周报和提醒草案。
  6. 用户确认后再落地任务或提醒。
```

原则：

- 低频高信号。
- 不做每日唠叨。
- 只提醒真正影响家庭目标的事项。
- 用户保留最终确认权。

## 路线图

### Phase 1: 独立开源仓库

目标：创建 `lifestyle-skills` 仓库，定义目录结构、README、license、schema 和贡献规范。

交付物：

- 仓库骨架。
- `lifestyle-skill.schema.json`。
- `registry.json`。
- 贡献指南。
- 安全边界说明。

### Phase 2: MVP Skills

目标：完成 1 个入口 skill 和 4 个领域、共 12 个 MVP lens。

交付物：

- `life-butler` 入口 skill。
- finance 3 个。
- health 3 个。
- movement 3 个。
- family 3 个。
- 每个 lens 至少 2 个示例。

### Phase 3: Eval Cases

目标：`life-butler` 配 5 到 10 个首次使用评估场景；每个 lens 配 3 到 5 个评估场景，后续再扩到 5 到 10 个。

交付物：

- `evals/` 场景文件。
- first-use / no-data 场景。
- base agent / wrong lens / matched lens 对照方法。
- 人工评分 rubric。

### Phase 4: Local Skill Runner

目标：先在开源仓库内证明 skill 可加载、可选择、可评估。

交付物：

- 本地 CLI runner。
- metadata 校验。
- 单个 case 运行。
- 批量 eval 运行。
- scoreboard 输出。

### Phase 5: LifeOps SkillRegistry 和 SkillRouter

目标：LifeOps 开始消费外部 lifestyle skills。

交付物：

- `SkillRegistry`。
- `SkillRouter`。
- member profile 到 lens metadata 的匹配逻辑。
- agent decision 记录。
- 用户接受、忽略、修改建议的反馈记录。

### Phase 6: DomainAgent 升级

目标：各领域 agent 不再只有固定 prompt，而是先选择合适 skill，再生成建议。

交付物：

- finance agent skill 化。
- health agent skill 化。
- movement agent skill 化。
- family agent skill 化。
- skill 选择理由可展示。

### Phase 7: WeeklyReviewAgent

目标：每周主动复盘并生成提醒草案。

交付物：

- 每周数据聚合。
- 各领域复盘。
- ButlerAgent 跨领域冲突处理。
- 周报输出。
- 提醒草案。
- 用户确认后写入 LifeOps。

## 推荐顺序

优先级如下：

```text
1. 开源 lifestyle-skills 仓库
2. schema + registry
3. 12 个 MVP lenses
4. life-butler first-use eval cases
5. lens eval cases + scoreboard
6. local eval prompt builder / runner
7. LifeOps SkillRegistry / SkillRouter
8. DomainAgent lens 化
9. 有足够历史数据后再做每周复盘
```

这个顺序能保证先证明 skill 本身有价值，再把它放进 LifeOps，而不是让 LifeOps 承担所有早期试错成本。

## 参考来源

- Anthropic Skills: https://github.com/anthropics/skills
- SkillsGate: https://skillsgate.ai/
- SkillsMD: https://skillsmd.dev/
- NVIDIA Skills: https://github.com/NVIDIA/skills
- SkillsBench: https://arxiv.org/abs/2602.12670
- OpenSkillEval: https://arxiv.org/abs/2605.23657
- Under the Hood of SKILL.md: https://arxiv.org/abs/2605.11418
