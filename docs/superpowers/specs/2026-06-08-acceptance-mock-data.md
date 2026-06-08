# 验收 Mock 数据

## 家庭成员

| id | name | role | birth_date |
|----|------|------|------------|
| dad | 赵志豪 | dad | 1990-03-15 |
| mom | 林小婉 | mom | 1992-07-22 |
| son | 赵一鸣 | son | 2018-09-01 |
| daughter | 赵一朵 | daughter | 2021-04-18 |
| grandpa | 赵德明 | grandpa | 1965-11-08 |
| grandma | 王秀兰 | grandma | 1968-02-14 |

## 财务数据

### 账户 (finance_accounts)

| id | member_id | name | type | balance |
|----|-----------|------|------|---------|
| acc1 | dad | 工商银行工资卡 | bank | 45000 |
| acc2 | dad | 招商银行储蓄 | bank | 120000 |
| acc3 | mom | 支付宝 | ewallet | 8500 |
| acc4 | mom | 微信钱包 | ewallet | 3200 |
| acc5 | — | 家庭基金-稳健 | investment | 350000 |
| acc6 | — | 家庭基金-指数 | investment | 180000 |
| acc7 | dad | 房贷 | loan | -1680000 |
| acc8 | dad | 车贷 | loan | -85000 |

### 流水 (finance_records)

近3个月数据：

| member_id | type | amount | category | note | record_date |
|-----------|------|--------|----------|------|-------------|
| dad | income | 28000 | salary | 4月工资 | 2026-04-01 |
| dad | income | 28000 | salary | 5月工资 | 2026-05-01 |
| dad | income | 28000 | salary | 6月工资 | 2026-06-01 |
| mom | income | 16000 | salary | 4月工资 | 2026-04-05 |
| mom | income | 16000 | salary | 5月工资 | 2026-05-05 |
| mom | income | 16000 | salary | 6月工资 | 2026-06-05 |
| dad | income | 5000 | other | 4月项目奖金 | 2026-04-15 |
| dad | income | 8000 | other | 5月年终奖补发 | 2026-05-20 |
| dad | expense | 6500 | housing | 4月房贷月供 | 2026-04-05 |
| dad | expense | 6500 | housing | 5月房贷月供 | 2026-05-05 |
| dad | expense | 6500 | housing | 6月房贷月供 | 2026-06-05 |
| dad | expense | 2800 | transport | 4月车贷月供 | 2026-04-10 |
| dad | expense | 2800 | transport | 5月车贷月供 | 2026-05-10 |
| dad | expense | 2800 | transport | 6月车贷月供 | 2026-06-10 |
| mom | expense | 1500 | food | 4月超市采购 | 2026-04-08 |
| mom | expense | 1800 | food | 5月超市采购 | 2026-05-06 |
| mom | expense | 1650 | food | 6月超市采购 | 2026-06-03 |
| mom | expense | 600 | food | 6月水果店 | 2026-06-06 |
| dad | expense | 3500 | education | 一鸣春季兴趣班 | 2026-04-01 |
| dad | expense | 4200 | education | 一鸣暑期游泳班 | 2026-06-01 |
| mom | expense | 800 | medical | 一朵体检费 | 2026-05-12 |
| dad | expense | 2200 | entertainment | 全家五一出游 | 2026-05-01 |
| mom | expense | 600 | entertainment | 一鸣生日蛋糕+礼物 | 2026-06-01 |
| mom | expense | 300 | entertainment | 电影票 | 2026-06-07 |
| dad | expense | 450 | transport | 6月加油 | 2026-06-04 |
| mom | expense | 1500 | medical | 爸爸体检费 | 2026-06-02 |
| grandpa | expense | 320 | medical | 慢性病药费 | 2026-06-05 |
| dad | expense | 8000 | other | 家庭基金定投 | 2026-06-01 |

## 健康数据

### 健康状态 (health_profiles)

| member_id | summary |
|-----------|---------|
| dad | 整体健康，血压略高（135/88），每日服降压药缬沙坦1粒，血糖正常5.1，体重78kg偏重，久坐需多运动 |
| mom | 近期血压偏高（142/92），正在控制饮食减盐，无慢性病，体重55kg正常，每周瑜伽2次 |
| son | 健康活泼，身高118cm，体重23kg，视力略下降需关注，花生过敏 |
| daughter | 健康正常，身高95cm，体重14kg，发育正常 |
| grandpa | 高血压（服药中），血糖偏高6.8，轻度脂肪肝，体重72kg，每日散步30分钟 |
| grandma | 轻度骨质疏松，膝关节偶尔疼痛，体重60kg，血压正常 |

### 健康记录 (health_records)

| member_id | type | metric | value | unit | note | record_date |
|-----------|------|--------|-------|------|------|-------------|
| dad | vitals | blood_pressure | 135/88 | mmHg | 略高 | 2026-06-07 |
| dad | vitals | blood_pressure | 138/90 | mmHg | 比上次高 | 2026-06-01 |
| dad | vitals | blood_pressure | 132/85 | mmHg | 还行 | 2026-05-25 |
| dad | vitals | weight | 78 | kg | | 2026-06-07 |
| dad | vitals | weight | 78.5 | kg | | 2026-05-25 |
| dad | checkup | blood_sugar | 5.1 | mmol/L | 正常 | 2026-05-20 |
| dad | medication | — | 缬沙坦 | 80mg | 降压药每天1粒 | 2026-06-01 |
| mom | vitals | blood_pressure | 142/92 | mmHg | 偏高需关注 | 2026-06-05 |
| mom | vitals | blood_pressure | 138/88 | mmHg | | 2026-05-28 |
| mom | vitals | blood_pressure | 135/86 | mmHg | 还行 | 2026-05-20 |
| mom | vitals | weight | 55 | kg | | 2026-06-05 |
| mom | exercise | yoga | 1 | hour | 瑜伽课 | 2026-06-04 |
| mom | exercise | yoga | 1 | hour | 瑜伽课 | 2026-06-01 |
| son | vitals | height | 118 | cm | | 2026-06-01 |
| son | vitals | weight | 23 | kg | | 2026-06-01 |
| son | checkup | vision | 4.8/4.9 | — | 略下降 | 2026-05-15 |
| daughter | vitals | height | 95 | cm | | 2026-06-01 |
| daughter | vitals | weight | 14 | kg | | 2026-06-01 |
| grandpa | vitals | blood_pressure | 150/95 | mmHg | 控制中 | 2026-06-06 |
| grandpa | vitals | blood_pressure | 145/92 | mmHg | | 2026-06-01 |
| grandpa | checkup | blood_sugar | 6.8 | mmol/L | 偏高 | 2026-05-20 |
| grandpa | exercise | walking | 30 | min | 每日散步 | 2026-06-07 |
| grandma | vitals | blood_pressure | 125/80 | mmHg | 正常 | 2026-06-05 |
| grandma | medication | — | 钙片+维D | — | 骨质疏松 | 2026-06-01 |

## 工作数据

### 工作状态 (work_status)

| member_id | summary |
|-----------|---------|
| dad | Q3系统迁移项目进行中，7/15方案评审是关键节点，每周一上午技术例会，最近在赶方案文档 |
| mom | 上半年收尾阶段，6/30季度汇报截止，常规教学工作，下学期课表待定 |

### 工作记录 (work_records)

| member_id | type | title | status | priority | project | due_date | note |
|-----------|------|-------|--------|----------|---------|----------|------|
| dad | project | Q3系统迁移 | active | high | infra | — | 核心项目，涉及3个系统 |
| dad | deadline | 迁移方案评审 | pending | high | infra | 2026-07-15 | 需提前准备PPT和架构文档 |
| dad | deadline | 数据库迁移脚本 | pending | high | infra | 2026-07-20 | 依赖方案评审结果 |
| dad | meeting | 技术周会 | active | medium | infra | 每周一10:00 | 全组参加 |
| dad | meeting | 迁移方案讨论 | active | high | infra | 2026-06-12 | 与架构师对齐 |
| dad | milestone | 测试环境搭建 | completed | medium | infra | 2026-06-01 | 已完成 |
| dad | deadline | 团队OKR自评 | pending | medium | — | 2026-06-25 | HR催了 |
| dad | project | 个人技术博客 | active | low | — | — | 月更1篇，本月未写 |
| mom | deadline | 季度教学汇报 | pending | high | — | 2026-06-30 | 需要准备课件 |
| mom | meeting | 教研组会议 | active | medium | — | 2026-06-13 | 讨论下学期教学计划 |
| mom | deadline | 期末考试出卷 | pending | high | — | 2026-06-20 | 高二数学卷 |
| mom | project | 在线课程录制 | active | medium | — | — | 暑假前完成前5节 |

## 家庭事务数据

### 家庭状态 (family_status)

| summary |
|---------|
| 下周三家长会待确认谁去，空调该清洗了，暑假看护待安排（7-8月），外婆生日6/20要准备，冰箱里的菜可能不够了 |

### 家庭事务记录 (family_records)

| member_id | type | title | status | scheduled_date | participants | note |
|-----------|------|-------|--------|---------------|-------------|------|
| mom | schedule | 家长会 | pending | 2026-06-11 | ["mom","dad"] | 班主任通知，一鸣班级 |
| mom | shopping | 周末采购 | pending | 2026-06-08 | ["mom"] | 牛奶、鸡蛋、纸巾、水果 |
| dad | chore | 约空调清洗 | pending | 2026-06-10 | ["dad"] | 找张师傅，去年洗过的 |
| mom | activity | 外婆生日 | pending | 2026-06-20 | ["all"] | 准备礼物和蛋糕 |
| dad | childcare | 暑假看护安排 | pending | 2026-07-01 | ["dad","mom"] | 一鸣一朵谁来带，考虑托管班 |
| mom | chore | 深度保洁 | done | 2026-06-01 | ["mom"] | 找了家政 |
| dad | schedule | 车辆保养 | pending | 2026-06-15 | ["dad"] | 该换机油了，里程到了 |
| mom | schedule | 一鸣视力复查 | pending | 2026-06-18 | ["mom","son"] | 儿童医院眼科 |
| dad | shopping | 父亲节礼物 | done | 2026-06-08 | ["mom","son"] | 给爷爷买的血压计 |
| — | activity | 全家出游 | pending | 2026-07-05 | ["all"] | 暑假短途旅行，待定目的地 |

## 知识笔记 (knowledge_notes)

| domain | member_id | title | content | tags |
|--------|-----------|-------|---------|------|
| health | dad | 饮食禁忌 | 不能吃太咸太辣，医生嘱咐低钠饮食，少喝酒 | ["饮食","医嘱"] |
| health | son | 过敏信息 | 花生过敏，注意零食和餐厅菜品，随身带抗过敏药 | ["过敏","重要"] |
| health | grandpa | 用药清单 | 缬沙坦(降压)、二甲双胍(降糖)、阿托伐他汀(降脂) | ["用药","慢性病"] |
| health | grandma | 骨质疏松 | 轻度骨质疏松，每天吃钙片+维D，避免提重物和剧烈运动 | ["骨骼","用药"] |
| finance | — | 理财偏好 | 妈偏好稳健型基金(债基为主)，不碰股票。爸定投指数基金，每月8号扣款 | ["理财","偏好"] |
| finance | — | 保险 | 全家重疾险(每年1.2万)，一鸣学平险，车辆交强+商业险(每年6800) | ["保险","固定支出"] |
| family | — | 家庭习惯 | 周末一般周六去外婆家吃饭，周日上午全家活动下午在家休息 | ["习惯","周末"] |
| family | — | 学校信息 | 一鸣在实验小学二年级(2)班，班主任王老师，一朵在幼儿园小班 | ["学校","孩子"] |
| work | dad | 通勤 | 上班单程40分钟，开车。公司朝九晚六，弹性半小时 | ["通勤","作息"] |
| work | mom | 通勤 | 学校离家近，骑车10分钟，寒暑假不上班 | ["通勤","作息"] |

---

## 验收测试问题

### 财务领域 (10题)

1. "家里现在一共有多少钱？"
2. "这个月花了多少？跟上个月比呢？"
3. "房贷还剩多少？每个月还多少？"
4. "家庭净资产大概是多少？"
5. "这个月最大的开支是什么？"
6. "爷爷最近有什么医疗支出吗？"
7. "我们每月固定支出大概是多少？"
8. "基金投了多少钱？"
9. "今年上半年收入大概多少？"
10. "有什么可以省钱的地方吗？"

### 健康领域 (10题)

1. "妈最近血压怎么样？"
2. "爸在吃什么药？"
3. "一鸣的身高体重是多少？"
4. "爷爷的血糖控制得怎么样？"
5. "家里谁最近在运动？"
6. "一鸣的视力怎么样了？"
7. "奶奶的膝盖好点了吗？"
8. "妈的血压有没有在好转？"
9. "一鸣有什么要注意的健康问题吗？"
10. "全家人的体重都是多少？"

### 工作领域 (10题)

1. "最近有什么重要的deadline？"
2. "志豪的工作项目进展怎么样？"
3. "小婉这个月有什么工作安排？"
4. "最紧急的工作事项是什么？"
5. "志豪每周有什么固定会议？"
6. "小婉的期末出卷准备得怎么样了？"
7. "有什么已经完成的工作里程碑吗？"
8. "志豪的博客更新了吗？"
9. "两人的通勤情况是怎样的？"
10. "下个月工作方面有什么重点？"

### 家庭事务领域 (10题)

1. "这周有什么安排？"
2. "下周三家长会谁去？"
3. "暑假孩子的看护安排好了吗？"
4. "外婆生日什么时候？准备了吗？"
5. "最近有什么家务要做？"
6. "一鸣什么时候去复查视力？"
7. "车子该保养了吗？"
8. "父亲节礼物准备了吗？"
9. "孩子们在哪个学校上学？"
10. "暑假全家出游计划得怎么样了？"

### 混合领域 (5题)

1. "最近家里整体情况怎么样？"
2. "这个月家庭有什么大事需要注意？"
3. "志豪最近忙不忙？工作和健康怎么样？"
4. "一鸣最近各方面情况怎么样？"
5. "帮我看一下下个月需要提前准备什么"
