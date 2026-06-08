package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RoutingResult struct {
	PrimaryLensID    string
	SupportingLensID string
	Reason           string
}

type SkillRouter struct {
	registry *SkillRegistry
	llm      *LLMClient
}

func NewSkillRouter(registry *SkillRegistry, llm *LLMClient) *SkillRouter {
	return &SkillRouter{registry: registry, llm: llm}
}

func (sr *SkillRouter) Route(query string, domain string) (*RoutingResult, error) {
	if !sr.registry.Available() || sr.llm == nil {
		return nil, nil
	}

	lenses := sr.registry.ListLenses(domain)
	if len(lenses) == 0 {
		return nil, nil
	}
	if len(lenses) == 1 {
		return &RoutingResult{PrimaryLensID: lenses[0].ID, Reason: "only lens in domain"}, nil
	}

	index := sr.registry.LensIndexText(domain)

	prompt := fmt.Sprintf(`Based on the user's question, select the most suitable lens from the %s domain.

Available lenses:
%s

User question: %s

Respond with JSON only, no explanation:
{"lens_id": "the-best-matching-lens-id", "reason": "one sentence why"}`, domain, index, query)

	chatResult, err := sr.llm.Chat([]ChatMessage{
		{Role: "system", Content: "You select the best lifestyle lens for the user's situation. Return JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("skill routing: %w", err)
	}

	resp := strings.TrimSpace(chatResult.Content)
	resp = strings.TrimPrefix(resp, "```json")
	resp = strings.TrimPrefix(resp, "```")
	resp = strings.TrimSuffix(resp, "```")
	resp = strings.TrimSpace(resp)

	var result struct {
		LensID string `json:"lens_id"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return nil, fmt.Errorf("parse routing response: %w", err)
	}

	if sr.registry.GetLens(result.LensID) == nil {
		return nil, nil
	}

	return &RoutingResult{
		PrimaryLensID: result.LensID,
		Reason:        result.Reason,
	}, nil
}

func (sr *SkillRouter) BuildSystemPrompt(domain string, result *RoutingResult) string {
	if result == nil || result.PrimaryLensID == "" {
		return sr.fallbackPrompt(domain)
	}

	content := sr.registry.GetLensContent(result.PrimaryLensID)
	if content == "" {
		return sr.fallbackPrompt(domain)
	}

	var supportSection string
	if result.SupportingLensID != "" {
		supportLens := sr.registry.GetLens(result.SupportingLensID)
		supportContent := sr.registry.GetLensContent(result.SupportingLensID)
		if supportLens != nil && supportContent != "" {
			supportSection = fmt.Sprintf(`

## Supporting Lens: %s

Apply safety boundaries and output format constraints from this lens where relevant.

%s`, supportLens.Name, firstNLines(supportContent, 20))
		}
	}

	return fmt.Sprintf(`You are a household %s advisor using the following thinking framework.

---

%s
%s

---

Answer in warm, concise Chinese. Follow the lens's Reasoning Flow. If data is insufficient, say so clearly. Do not make up data.`, domain, content, supportSection)
}

func (sr *SkillRouter) fallbackPrompt(domain string) string {
	prompts := map[string]string{
		"finance":  "你是家庭财务顾问。基于提供的家庭财务数据，分析收支情况、资产负债、给出理财建议。要求：用中文回答，数字要准确，指出重要趋势。",
		"health":   "你是家庭健康助手。基于提供的家庭健康数据，分析健康状况、用药情况、运动习惯。要求：用中文回答，关注异常指标，给出温和的健康建议。不要给出医疗诊断。",
		"movement": "你是运动指导助手。基于提供的运动数据，分析运动习惯、体能变化。要求：用中文回答，给出保守渐进的运动建议。关注安全和持续性。",
		"family":   "你是家庭事务管家。基于提供的家庭事务数据，分析日程安排、家务分工、育儿安排。要求：用中文回答，关注待办事项，提醒重要日程。",
	}
	if p, ok := prompts[domain]; ok {
		return p
	}
	return "你是家庭管家助手。用中文回答用户的问题。"
}

func firstNLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n") + "\n..."
}
