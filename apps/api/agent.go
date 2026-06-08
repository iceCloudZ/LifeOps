package main

import (
	"fmt"
	"strings"
	"sync"
)

type AgentContext struct {
	Status        string
	RecentRecords []string
	Notes         []string
}

type LensInfo struct {
	ID     string
	Name   string
	Reason string
}

type AgentResponse struct {
	Answer string
	Lens   *LensInfo
}

type Agent interface {
	Name() string
	Domain() string
	SystemPrompt() string
	RetrieveContext(query string, memberID string) (*AgentContext, error)
}

// ButlerAgent

type ButlerAgent struct {
	store  *Store
	llm    *LLMClient
	agents map[string]Agent
	router *SkillRouter
}

func NewButlerAgent(store *Store, llm *LLMClient, router *SkillRouter) *ButlerAgent {
	b := &ButlerAgent{
		store:  store,
		llm:    llm,
		agents: make(map[string]Agent),
		router: router,
	}
	b.agents["finance"] = &FinanceAgent{store: store, llm: llm, router: router}
	b.agents["health"] = &HealthAgent{store: store, llm: llm, router: router}
	b.agents["work"] = &WorkAgent{store: store, llm: llm}
	b.agents["family"] = &FamilyAgent{store: store, llm: llm, router: router}
	b.agents["movement"] = &MovementAgent{store: store, llm: llm, router: router}
	return b
}

func (b *ButlerAgent) Name() string   { return "butler" }
func (b *ButlerAgent) Domain() string  { return "all" }
func (b *ButlerAgent) SystemPrompt() string {
	return "你是LifeOps家庭管家。你的职责是：1.分析用户问题涉及哪些领域（finance/health/work/family/movement）2.如果问题只涉及一个领域，直接给出简洁专业的回答 3.如果涉及多个领域，综合分析后给出整体回答。回答要求：温暖简洁的中文，基于提供的数据回答，不要编造不存在的数据，如果数据不足以回答请明确说明。输出格式：纯文本，不要用markdown。"
}

func (b *ButlerAgent) RetrieveContext(query string, memberID string) (*AgentContext, error) {
	return &AgentContext{}, nil
}

func (b *ButlerAgent) AnalyzeAndRoute(query string) ([]string, error) {
	routingPrompt := "分析以下问题涉及哪些领域，只返回领域名称用逗号分隔（finance,health,work,family,movement），不要返回其他内容。\n问题：" + query

	resp, err := b.llm.Chat([]ChatMessage{
		{Role: "user", Content: routingPrompt},
	})
	if err != nil {
		return nil, fmt.Errorf("routing failed: %w", err)
	}

	resp = strings.TrimSpace(resp)
	if resp == "" {
		return allDomains(), nil
	}

	var matched []string
	for _, part := range strings.Split(resp, ",") {
		domain := strings.TrimSpace(part)
		if _, ok := b.agents[domain]; ok {
			matched = append(matched, domain)
		}
	}

	if len(matched) == 0 {
		return allDomains(), nil
	}
	return matched, nil
}

func allDomains() []string {
	return []string{"finance", "health", "work", "family", "movement"}
}

func (b *ButlerAgent) Answer(query string) (*AgentResponse, error) {
	domains, err := b.AnalyzeAndRoute(query)
	if err != nil {
		return nil, err
	}

	if len(domains) == 1 {
		agent := b.agents[domains[0]]
		return b.answerWithAgent(agent, query, domains[0])
	}

	type agentResult struct {
		domain  string
		 resp   *AgentResponse
		err    error
	}

	var wg sync.WaitGroup
	results := make([]agentResult, len(domains))

	for i, domain := range domains {
		wg.Add(1)
		go func(idx int, d string) {
			defer wg.Done()
			agent := b.agents[d]
			resp, err := b.answerWithAgent(agent, query, d)
			results[idx] = agentResult{domain: d, resp: resp, err: err}
		}(i, domain)
	}
	wg.Wait()

	var parts []string
	var primaryLens *LensInfo
	for _, r := range results {
		if r.err != nil {
			parts = append(parts, fmt.Sprintf("[%s] 获取数据失败: %v", r.domain, r.err))
		} else {
			parts = append(parts, fmt.Sprintf("[%s] %s", r.domain, r.resp.Answer))
			if primaryLens == nil && r.resp.Lens != nil {
				primaryLens = r.resp.Lens
			}
		}
	}

	synthesisPrompt := strings.Join(parts, "\n\n")
	messages := []ChatMessage{
		{Role: "system", Content: b.SystemPrompt()},
		{Role: "user", Content: fmt.Sprintf("以下是对用户问题各个领域的分析结果：\n\n%s\n\n用户原问题：%s\n\n请综合以上分析，给出一个整体回答。", synthesisPrompt, query)},
	}

	finalAnswer, err := b.llm.Chat(messages)
	if err != nil {
		return nil, fmt.Errorf("synthesis failed: %w", err)
	}
	return &AgentResponse{Answer: finalAnswer, Lens: primaryLens}, nil
}

func (b *ButlerAgent) answerWithAgent(agent Agent, query string, domain string) (*AgentResponse, error) {
	ctx, err := agent.RetrieveContext(query, "")
	if err != nil {
		return nil, fmt.Errorf("retrieve context for %s: %w", agent.Name(), err)
	}

	systemPrompt := agent.SystemPrompt()
	var lens *LensInfo

	if b.router != nil {
		result, _ := b.router.Route(query, domain)
		if result != nil {
			lensPrompt := b.router.BuildSystemPrompt(domain, result)
			if lensPrompt != "" {
				systemPrompt = lensPrompt
			}
			lensName := result.PrimaryLensID
			if l := b.router.registry.GetLens(result.PrimaryLensID); l != nil {
				lensName = l.Name
			}
			lens = &LensInfo{
				ID:     result.PrimaryLensID,
				Name:   lensName,
				Reason: result.Reason,
			}
		}
	}

	var contextParts []string
	if ctx.Status != "" {
		contextParts = append(contextParts, "当前状态："+ctx.Status)
	}
	if len(ctx.RecentRecords) > 0 {
		contextParts = append(contextParts, "近期记录：\n"+strings.Join(ctx.RecentRecords, "\n"))
	}
	if len(ctx.Notes) > 0 {
		contextParts = append(contextParts, "相关知识：\n"+strings.Join(ctx.Notes, "\n"))
	}

	contextText := strings.Join(contextParts, "\n\n")
	if contextText == "" {
		contextText = "暂无相关数据。"
	}

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("参考数据：\n%s\n\n用户问题：%s", contextText, query)},
	}

	answer, err := b.llm.Chat(messages)
	if err != nil {
		return nil, err
	}
	return &AgentResponse{Answer: answer, Lens: lens}, nil
}

// FinanceAgent

type FinanceAgent struct {
	store  *Store
	llm    *LLMClient
	router *SkillRouter
}

func (a *FinanceAgent) Name() string  { return "finance" }
func (a *FinanceAgent) Domain() string { return "finance" }
func (a *FinanceAgent) SystemPrompt() string {
	if a.router != nil {
		return a.router.fallbackPrompt("finance")
	}
	return "你是家庭财务顾问。基于提供的家庭财务数据，分析收支情况、资产负债、给出理财建议。要求：用中文回答，数字要准确，指出重要趋势。"
}

func (a *FinanceAgent) RetrieveContext(query string, memberID string) (*AgentContext, error) {
	ctx := &AgentContext{}

	accounts, err := a.store.ListFinanceAccounts()
	if err != nil {
		return nil, err
	}
	var statusParts []string
	var totalAssets, totalLiabilities float64
	for _, acc := range accounts {
		if acc.Balance >= 0 {
			totalAssets += acc.Balance
		} else {
			totalLiabilities += acc.Balance
		}
		statusParts = append(statusParts, fmt.Sprintf("%s: %.0f元", acc.Name, acc.Balance))
	}
	ctx.Status = fmt.Sprintf("总资产: %.0f, 总负债: %.0f, 净资产: %.0f\n账户明细: %s",
		totalAssets, totalLiabilities, totalAssets+totalLiabilities,
		strings.Join(statusParts, "; "))

	records, err := a.store.ListFinanceRecords("", "", "", "", "")
	if err != nil {
		return nil, err
	}
	for i, r := range records {
		if i >= 50 {
			break
		}
		ctx.RecentRecords = append(ctx.RecentRecords,
			fmt.Sprintf("[%s] %s %s %.0f元 (%s) %s", r.RecordDate, r.Type, r.Category, r.Amount, r.Currency, r.Note))
	}

	notes, err := a.store.ListKnowledgeNotes("finance", "")
	if err != nil {
		return nil, err
	}
	for _, n := range notes {
		ctx.Notes = append(ctx.Notes, fmt.Sprintf("%s: %s", n.Title, n.Content))
	}

	return ctx, nil
}

// HealthAgent

type HealthAgent struct {
	store  *Store
	llm    *LLMClient
	router *SkillRouter
}

func (a *HealthAgent) Name() string  { return "health" }
func (a *HealthAgent) Domain() string { return "health" }
func (a *HealthAgent) SystemPrompt() string {
	if a.router != nil {
		return a.router.fallbackPrompt("health")
	}
	return "你是家庭健康助手。基于提供的家庭健康数据，分析健康状况、用药情况、运动习惯。要求：用中文回答，关注异常指标，给出温和的健康建议。不要给出医疗诊断。"
}

func (a *HealthAgent) RetrieveContext(query string, memberID string) (*AgentContext, error) {
	ctx := &AgentContext{}

	profiles, err := a.store.ListHealthProfiles()
	if err != nil {
		return nil, err
	}
	var statusParts []string
	for _, p := range profiles {
		statusParts = append(statusParts, fmt.Sprintf("[%s] %s", p.MemberID, p.Summary))
	}
	ctx.Status = strings.Join(statusParts, "\n")

	records, err := a.store.ListHealthRecords("", "")
	if err != nil {
		return nil, err
	}
	for i, r := range records {
		if i >= 50 {
			break
		}
		parts := fmt.Sprintf("[%s] %s", r.RecordDate, r.MemberID)
		if r.Metric != nil && r.Value != nil {
			parts += fmt.Sprintf(" %s=%s", *r.Metric, *r.Value)
			if r.Unit != nil {
				parts += *r.Unit
			}
		}
		if r.Note != "" {
			parts += " " + r.Note
		}
		ctx.RecentRecords = append(ctx.RecentRecords, parts)
	}

	notes, err := a.store.ListKnowledgeNotes("health", "")
	if err != nil {
		return nil, err
	}
	for _, n := range notes {
		ctx.Notes = append(ctx.Notes, fmt.Sprintf("[%s] %s: %s", derefStr(n.MemberID), n.Title, n.Content))
	}

	return ctx, nil
}

// WorkAgent

type WorkAgent struct {
	store *Store
	llm   *LLMClient
}

func (a *WorkAgent) Name() string  { return "work" }
func (a *WorkAgent) Domain() string { return "work" }
func (a *WorkAgent) SystemPrompt() string {
	return "你是职业规划助手。基于提供的工作数据，分析项目进度、重要节点、工作压力。要求：用中文回答，关注即将到期的deadline，给出时间管理建议。"
}

func (a *WorkAgent) RetrieveContext(query string, memberID string) (*AgentContext, error) {
	ctx := &AgentContext{}

	profiles, err := a.store.ListWorkProfiles()
	if err != nil {
		return nil, err
	}
	var statusParts []string
	for _, p := range profiles {
		parts := []string{fmt.Sprintf("[%s]", p.MemberID)}
		if p.EmploymentStatus != "" {
			parts = append(parts, p.EmploymentStatus)
		}
		if p.Company != "" {
			parts = append(parts, p.Company)
		}
		if p.Position != "" {
			parts = append(parts, p.Position)
		}
		if p.Industry != "" {
			parts = append(parts, p.Industry)
		}
		if p.WorkLocation != "" {
			parts = append(parts, p.WorkLocation)
		}
		if p.IncomeRange != "" {
			parts = append(parts, "收入:"+p.IncomeRange)
		}
		if p.WorkSchedule != "" {
			parts = append(parts, p.WorkSchedule)
		}
		if p.CommuteMinutes > 0 {
			parts = append(parts, fmt.Sprintf("通勤%d分钟", p.CommuteMinutes))
		}
		if p.StartedAt != "" {
			parts = append(parts, "入职:"+p.StartedAt)
		}
		statusParts = append(statusParts, strings.Join(parts, " "))
	}

	statuses, err := a.store.ListWorkStatuses()
	if err != nil {
		return nil, err
	}
	for _, s := range statuses {
		statusParts = append(statusParts, fmt.Sprintf("[%s] 状态概要: %s", s.MemberID, s.Summary))
	}

	ctx.Status = strings.Join(statusParts, "\n")

	records, err := a.store.ListWorkRecords("", "")
	if err != nil {
		return nil, err
	}
	for i, r := range records {
		if i >= 50 {
			break
		}
		line := fmt.Sprintf("[%s] %s %s: %s (优先级:%s, 状态:%s)", r.MemberID, r.Type, r.Title, r.Project, r.Priority, r.Status)
		if r.DueDate != nil {
			line += " 截止:" + *r.DueDate
		}
		ctx.RecentRecords = append(ctx.RecentRecords, line)
	}

	notes, err := a.store.ListKnowledgeNotes("work", "")
	if err != nil {
		return nil, err
	}
	for _, n := range notes {
		ctx.Notes = append(ctx.Notes, fmt.Sprintf("[%s] %s: %s", derefStr(n.MemberID), n.Title, n.Content))
	}

	return ctx, nil
}

// FamilyAgent

type FamilyAgent struct {
	store  *Store
	llm    *LLMClient
	router *SkillRouter
}

func (a *FamilyAgent) Name() string  { return "family" }
func (a *FamilyAgent) Domain() string { return "family" }
func (a *FamilyAgent) SystemPrompt() string {
	if a.router != nil {
		return a.router.fallbackPrompt("family")
	}
	return "你是家庭事务管家。基于提供的家庭事务数据，分析日程安排、家务分工、育儿安排。要求：用中文回答，关注待办事项，提醒重要日程。"
}

func (a *FamilyAgent) RetrieveContext(query string, memberID string) (*AgentContext, error) {
	ctx := &AgentContext{}

	status, err := a.store.GetFamilyStatus()
	if err != nil {
		return nil, err
	}
	if status != nil {
		ctx.Status = status.Summary
	}

	records, err := a.store.ListFamilyRecords("", "", "")
	if err != nil {
		return nil, err
	}
	for i, r := range records {
		if i >= 50 {
			break
		}
		line := fmt.Sprintf("[%s] %s: %s (状态:%s)", derefStr(r.MemberID), r.Type, r.Title, r.Status)
		if r.ScheduledDate != nil {
			line += " 日期:" + *r.ScheduledDate
		}
		ctx.RecentRecords = append(ctx.RecentRecords, line)
	}

	notes, err := a.store.ListKnowledgeNotes("family", "")
	if err != nil {
		return nil, err
	}
	for _, n := range notes {
		ctx.Notes = append(ctx.Notes, fmt.Sprintf("[%s] %s: %s", derefStr(n.MemberID), n.Title, n.Content))
	}

	return ctx, nil
}

// MovementAgent

type MovementAgent struct {
	store  *Store
	llm    *LLMClient
	router *SkillRouter
}

func (a *MovementAgent) Name() string  { return "movement" }
func (a *MovementAgent) Domain() string { return "movement" }
func (a *MovementAgent) SystemPrompt() string {
	if a.router != nil {
		return a.router.fallbackPrompt("movement")
	}
	return "你是运动指导助手。基于提供的运动数据，分析运动习惯、体能变化。要求：用中文回答，给出保守渐进的运动建议。关注安全和持续性。"
}

func (a *MovementAgent) RetrieveContext(query string, memberID string) (*AgentContext, error) {
	ctx := &AgentContext{}

	records, err := a.store.ListMovementRecords("")
	if err != nil {
		return nil, err
	}
	if len(records) > 0 {
		var parts []string
		for i, r := range records {
			if i >= 20 {
				break
			}
			line := fmt.Sprintf("[%s] %s", r.RecordDate, r.MemberID)
			if r.Metric != nil && r.Value != nil {
				line += fmt.Sprintf(" %s=%s", *r.Metric, *r.Value)
			}
			if r.Note != "" {
				line += " " + r.Note
			}
			parts = append(parts, line)
		}
		ctx.RecentRecords = parts
		ctx.Status = fmt.Sprintf("运动记录: %d 条", len(records))
	}

	notes, err := a.store.ListKnowledgeNotes("movement", "")
	if err != nil {
		return nil, err
	}
	for _, n := range notes {
		ctx.Notes = append(ctx.Notes, fmt.Sprintf("[%s] %s: %s", derefStr(n.MemberID), n.Title, n.Content))
	}

	return ctx, nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
