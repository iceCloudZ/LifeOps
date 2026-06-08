package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type CreateConversationReq struct {
	Title string `json:"title"`
}

type SendMessageReq struct {
	Content string `json:"content"`
}

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		conversations, err := s.store.ListConversations()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to list conversations")
			return
		}
		if conversations == nil {
			conversations = []Conversation{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conversations)
	case http.MethodPost:
		var req CreateConversationReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		conversation := &Conversation{
			ID:    newID(),
			Title: strings.TrimSpace(req.Title),
		}
		if err := s.store.CreateConversation(conversation); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to create conversation")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(conversation)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleConversationRoutes(w http.ResponseWriter, r *http.Request) {
	sub := strings.TrimPrefix(r.URL.Path, "/api/chat/conversations/")
	if sub == "" {
		s.handleConversations(w, r)
		return
	}

	if strings.HasSuffix(sub, "/messages") {
		id := strings.TrimSuffix(sub, "/messages")
		switch r.Method {
		case http.MethodGet:
			messages, err := s.store.ListMessages(id)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to list messages")
				return
			}
			if messages == nil {
				messages = []Message{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(messages)
		case http.MethodPost:
			s.sendMessage(w, r, id)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// Single conversation: DELETE
	if r.Method == http.MethodDelete {
		if err := s.store.DeleteConversation(sub); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to delete conversation")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *Server) sendMessage(w http.ResponseWriter, r *http.Request, conversationID string) {
	var req SendMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "content is required")
		return
	}

	conversation, err := s.store.GetConversation(conversationID)
	if err != nil || conversation == nil {
		writeJSONError(w, http.StatusNotFound, "not_found", "conversation not found")
		return
	}

	userMsg := &Message{
		ID:             newID(),
		ConversationID: conversationID,
		Role:           "user",
		Content:        content,
	}
	if err := s.store.CreateMessage(userMsg); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to save message")
		return
	}

	var answer string
	var lensID, lensName, lensReason string
	var tokensUsed int
	if s.butler != nil {
		resp, err := s.butler.Answer(content, conversationID)
		if err != nil {
			answer = "AI回答失败: " + err.Error()
		} else {
			answer = resp.Answer
			if resp.Lens != nil {
				lensID = resp.Lens.ID
				lensName = resp.Lens.Name
				lensReason = resp.Lens.Reason
			}
			if resp.Usage != nil {
				tokensUsed = resp.Usage.TotalTokens
			}
		}
	} else {
		answer = "AI管家未配置。请在设置页面配置AI服务。"
	}

	assistantMsg := &Message{
		ID:             newID(),
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        answer,
		LensID:         lensID,
		LensName:       lensName,
		LensReason:     lensReason,
		TokensUsed:     tokensUsed,
	}
	if err := s.store.CreateMessage(assistantMsg); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to save response")
		return
	}

	// Update conversation timestamp
	s.store.UpdateConversation(conversationID, conversation.Title)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assistantMsg)
}

// handleQuickEntryRoute routes /api/quick-entry
func (s *Server) handleQuickEntryRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Input string `json:"input"`
		Mode  string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
		return
	}
	if strings.TrimSpace(req.Input) == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "input is required")
		return
	}

	// Use the butler agent to parse natural language input into structured data
	prompt := `你是一个家庭信息整理助手。请将用户输入的自然语言拆分为多个领域的结构化记录。

领域包括：finance（财务）、health（健康）、work（工作）、family（家庭事务）。

请严格按以下JSON格式返回，不要返回其他内容：
{"entries":[{"domain":"finance|health|work|family","title":"标题","content":"详细描述","data":{}}]}

用户输入：` + req.Input

	aiConfig, _ := s.store.GetAIConfig()
	if aiConfig == nil || aiConfig.APIKey == "" {
		writeJSONError(w, http.StatusInternalServerError, "ai_error", "AI not configured")
		return
	}

	llm := NewLLMClient(LLMConfig{
		Endpoint:  aiConfig.Endpoint,
		APIKey:    aiConfig.APIKey,
		Model:     aiConfig.Model,
		MaxTokens: aiConfig.MaxTokens,
	})

	resp, err := llm.Chat([]ChatMessage{
		{Role: "system", Content: "你是家庭信息整理助手，只返回JSON，不要解释。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "ai_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resp.Content))
}

// handleAIConfigRoute routes /api/config/ai
func (s *Server) handleAIConfigRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config, err := s.store.GetAIConfig()
		if err != nil || config == nil {
			config = &AIConfig{
				Endpoint:  "https://api.openai.com/v1",
				Model:     "gpt-4o-mini",
				MaxTokens: 2048,
			}
		}
		masked := *config
		if len(masked.APIKey) > 4 {
			masked.APIKey = "****" + masked.APIKey[len(masked.APIKey)-4:]
		} else if masked.APIKey != "" {
			masked.APIKey = "****"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(masked)
	case http.MethodPut:
		var req AIConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		// Preserve existing API key if masked value sent
		if strings.HasPrefix(req.APIKey, "****") {
			existing, _ := s.store.GetAIConfig()
			if existing != nil {
				req.APIKey = existing.APIKey
			}
		}
		if err := s.store.UpdateAIConfig(&req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to update config")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleNotesRoute routes /api/notes and /api/notes/{id}
func (s *Server) handleNotesRoute(w http.ResponseWriter, r *http.Request) {
	sub := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if sub == "" || r.URL.Path == "/api/notes" {
		switch r.Method {
		case http.MethodGet:
			domain := r.URL.Query().Get("domain")
			memberID := r.URL.Query().Get("member_id")
			notes, err := s.store.ListKnowledgeNotes(domain, memberID)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to list notes")
				return
			}
			if notes == nil {
				notes = []KnowledgeNote{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(notes)
		case http.MethodPost:
			var req struct {
				Domain   *string `json:"domain"`
				MemberID *string `json:"member_id"`
				Title    string  `json:"title"`
				Content  string  `json:"content"`
				Tags     string  `json:"tags"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
				return
			}
			if strings.TrimSpace(req.Title) == "" {
				writeJSONError(w, http.StatusBadRequest, "invalid_request", "title is required")
				return
			}
			note := &KnowledgeNote{
				ID:       newID(),
				Domain:   req.Domain,
				MemberID: req.MemberID,
				Title:    strings.TrimSpace(req.Title),
				Content:  req.Content,
				Tags:     req.Tags,
			}
			if err := s.store.CreateKnowledgeNote(note); err != nil {
				writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to create note")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(note)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// Single note: PUT, DELETE
	switch r.Method {
	case http.MethodPut:
		var req struct {
			Domain   *string `json:"domain"`
			MemberID *string `json:"member_id"`
			Title    string  `json:"title"`
			Content  string  `json:"content"`
			Tags     string  `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if err := s.store.UpdateKnowledgeNote(sub, req.Domain, req.MemberID, strings.TrimSpace(req.Title), req.Content, req.Tags); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to update note")
			return
		}
		result, _ := s.store.GetKnowledgeNote(sub)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	case http.MethodDelete:
		if err := s.store.DeleteKnowledgeNote(sub); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "failed to delete note")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleDomainRoutes routes all domain CRUD endpoints
func (s *Server) handleDomainRoutes(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC().Format(time.RFC3339)
	_ = now

	path := r.URL.Path

	switch {
	// Finance
	case strings.HasPrefix(path, "/api/finance/accounts/"):
		id := strings.TrimPrefix(path, "/api/finance/accounts/")
		s.handleFinanceAccountByID(w, r, id)
	case path == "/api/finance/accounts":
		s.handleFinanceAccounts(w, r)
	case strings.HasPrefix(path, "/api/finance/records/"):
		id := strings.TrimPrefix(path, "/api/finance/records/")
		s.handleFinanceRecordByID(w, r, id)
	case path == "/api/finance/records":
		s.handleFinanceRecords(w, r)

	// Health
	case strings.HasPrefix(path, "/api/health/profiles/"):
		memberID := strings.TrimPrefix(path, "/api/health/profiles/")
		s.handleHealthProfileByMember(w, r, memberID)
	case path == "/api/health/profiles":
		s.handleHealthProfiles(w, r)
	case strings.HasPrefix(path, "/api/health/records/"):
		id := strings.TrimPrefix(path, "/api/health/records/")
		s.handleHealthRecordByID(w, r, id)
	case path == "/api/health/records":
		s.handleHealthRecords(w, r)

	// Work
	case strings.HasPrefix(path, "/api/work/status/"):
		memberID := strings.TrimPrefix(path, "/api/work/status/")
		s.handleWorkStatusByMember(w, r, memberID)
	case path == "/api/work/status":
		s.handleWorkStatuses(w, r)
	case strings.HasPrefix(path, "/api/work/profiles/"):
		memberID := strings.TrimPrefix(path, "/api/work/profiles/")
		s.handleWorkProfileByMember(w, r, memberID)
	case path == "/api/work/profiles":
		s.handleWorkProfiles(w, r)
	case strings.HasPrefix(path, "/api/work/records/"):
		id := strings.TrimPrefix(path, "/api/work/records/")
		s.handleWorkRecordByID(w, r, id)
	case path == "/api/work/records":
		s.handleWorkRecords(w, r)

	// Family
	case path == "/api/family/status":
		s.handleFamilyStatusRoute(w, r)
	case strings.HasPrefix(path, "/api/family/records/"):
		id := strings.TrimPrefix(path, "/api/family/records/")
		s.handleFamilyRecordByID(w, r, id)
	case path == "/api/family/records":
		s.handleFamilyRecords(w, r)

	// Movement (uses dedicated movement_records table)
	case strings.HasPrefix(path, "/api/movement/records/"):
		id := strings.TrimPrefix(path, "/api/movement/records/")
		s.handleMovementRecordByID(w, r, id)
	case path == "/api/movement/records":
		s.handleMovementRecords(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ---- Finance ----

func (s *Server) handleFinanceAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		accounts, _ := s.store.ListFinanceAccounts()
		if accounts == nil {
			accounts = []FinanceAccount{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accounts)
	case http.MethodPost:
		var req struct {
			MemberID *string `json:"member_id"`
			Name     string  `json:"name"`
			Type     string  `json:"type"`
			Balance  float64 `json:"balance"`
			Note     string  `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		acc := &FinanceAccount{ID: newID(), MemberID: req.MemberID, Name: req.Name, Type: req.Type, Balance: req.Balance, Note: req.Note}
		if err := s.store.CreateFinanceAccount(acc); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(acc)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFinanceAccountByID(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodPut:
		var req struct {
			MemberID *string `json:"member_id"`
			Name     string  `json:"name"`
			Type     string  `json:"type"`
			Note     string  `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		s.store.UpdateFinanceAccount(id, req.MemberID, req.Name, req.Type, req.Note)
		acc, _ := s.store.GetFinanceAccount(id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(acc)
	case http.MethodDelete:
		s.store.DeleteFinanceAccount(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFinanceRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		memberID := r.URL.Query().Get("member_id")
		recordType := r.URL.Query().Get("type")
		category := r.URL.Query().Get("category")
		fromDate := r.URL.Query().Get("from_date")
		toDate := r.URL.Query().Get("to_date")
		records, _ := s.store.ListFinanceRecords(memberID, recordType, category, fromDate, toDate)
		if records == nil {
			records = []FinanceRecord{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	case http.MethodPost:
		var req struct {
			MemberID   *string `json:"member_id"`
			Type       string  `json:"type"`
			Amount     float64 `json:"amount"`
			Currency   string  `json:"currency"`
			Category   string  `json:"category"`
			Note       string  `json:"note"`
			RecordDate string  `json:"record_date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		rec := &FinanceRecord{
			ID: newID(), MemberID: req.MemberID, Type: req.Type, Amount: req.Amount,
			Currency: req.Currency, Category: req.Category, Note: req.Note, RecordDate: req.RecordDate,
		}
		if rec.Currency == "" {
			rec.Currency = "CNY"
		}
		if err := s.store.CreateFinanceRecord(rec); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rec)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFinanceRecordByID(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodDelete:
		s.store.DeleteFinanceRecord(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ---- Health ----

func (s *Server) handleHealthProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	profiles, _ := s.store.ListHealthProfiles()
	if profiles == nil {
		profiles = []HealthProfile{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

func (s *Server) handleHealthProfileByMember(w http.ResponseWriter, r *http.Request, memberID string) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
		return
	}
	if err := s.store.UpdateHealthProfile(memberID, req.Summary); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	profile, _ := s.store.GetHealthProfile(memberID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (s *Server) handleHealthRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		memberID := r.URL.Query().Get("member_id")
		recordType := r.URL.Query().Get("type")
		records, _ := s.store.ListHealthRecords(memberID, recordType)
		if records == nil {
			records = []HealthRecord{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	case http.MethodPost:
		var req struct {
			MemberID   string  `json:"member_id"`
			Type       string  `json:"type"`
			Metric     *string `json:"metric"`
			Value      *string `json:"value"`
			Unit       *string `json:"unit"`
			Note       string  `json:"note"`
			RecordDate string  `json:"record_date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		rec := &HealthRecord{
			ID: newID(), MemberID: req.MemberID, Type: req.Type,
			Metric: req.Metric, Value: req.Value, Unit: req.Unit,
			Note: req.Note, RecordDate: req.RecordDate,
		}
		if err := s.store.CreateHealthRecord(rec); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rec)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleHealthRecordByID(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodDelete:
		s.store.DeleteHealthRecord(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ---- Work ----

func (s *Server) handleWorkStatuses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	statuses, _ := s.store.ListWorkStatuses()
	if statuses == nil {
		statuses = []WorkStatus{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

func (s *Server) handleWorkStatusByMember(w http.ResponseWriter, r *http.Request, memberID string) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
		return
	}
	if err := s.store.UpdateWorkStatus(memberID, req.Summary); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	status, _ := s.store.GetWorkStatus(memberID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleWorkProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	profiles, _ := s.store.ListWorkProfiles()
	if profiles == nil {
		profiles = []WorkProfile{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

func (s *Server) handleWorkProfileByMember(w http.ResponseWriter, r *http.Request, memberID string) {
	switch r.Method {
	case http.MethodGet:
		profile, _ := s.store.GetWorkProfile(memberID)
		if profile == nil {
			profile = &WorkProfile{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	case http.MethodPut:
		var req WorkProfile
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if err := s.store.UpdateWorkProfile(memberID, &req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		profile, _ := s.store.GetWorkProfile(memberID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	case http.MethodDelete:
		if err := s.store.DeleteWorkProfile(memberID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleWorkRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		memberID := r.URL.Query().Get("member_id")
		status := r.URL.Query().Get("status")
		records, _ := s.store.ListWorkRecords(memberID, status)
		if records == nil {
			records = []WorkRecord{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	case http.MethodPost:
		var req struct {
			MemberID string  `json:"member_id"`
			Type     string  `json:"type"`
			Title    string  `json:"title"`
			Status   string  `json:"status"`
			Priority string  `json:"priority"`
			Project  string  `json:"project"`
			DueDate  *string `json:"due_date"`
			Note     string  `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		rec := &WorkRecord{
			ID: newID(), MemberID: req.MemberID, Type: req.Type, Title: req.Title,
			Status: req.Status, Priority: req.Priority, Project: req.Project,
			DueDate: req.DueDate, Note: req.Note,
		}
		if rec.Status == "" {
			rec.Status = "active"
		}
		if rec.Priority == "" {
			rec.Priority = "medium"
		}
		if err := s.store.CreateWorkRecord(rec); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rec)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleWorkRecordByID(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodDelete:
		s.store.DeleteWorkRecord(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ---- Family ----

func (s *Server) handleFamilyStatusRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status, _ := s.store.GetFamilyStatus()
		if status == nil {
			status = &FamilyStatus{ID: "family"}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	case http.MethodPut:
		var req struct {
			Summary string `json:"summary"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		if err := s.store.UpdateFamilyStatus(req.Summary); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		status, _ := s.store.GetFamilyStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFamilyRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		memberID := r.URL.Query().Get("member_id")
		recordType := r.URL.Query().Get("type")
		status := r.URL.Query().Get("status")
		records, _ := s.store.ListFamilyRecords(memberID, recordType, status)
		if records == nil {
			records = []FamilyRecord{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	case http.MethodPost:
		var req struct {
			MemberID      *string `json:"member_id"`
			Type          string  `json:"type"`
			Title         string  `json:"title"`
			Status        string  `json:"status"`
			Location      string  `json:"location"`
			Participants  string  `json:"participants"`
			ScheduledDate *string `json:"scheduled_date"`
			Note          string  `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		rec := &FamilyRecord{
			ID: newID(), MemberID: req.MemberID, Type: req.Type, Title: req.Title,
			Status: req.Status, Location: req.Location, Participants: req.Participants,
			ScheduledDate: req.ScheduledDate, Note: req.Note,
		}
		if rec.Status == "" {
			rec.Status = "pending"
		}
		if rec.Participants == "" {
			rec.Participants = "[]"
		}
		if err := s.store.CreateFamilyRecord(rec); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rec)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFamilyRecordByID(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodDelete:
		s.store.DeleteFamilyRecord(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ---- Movement ----

func (s *Server) handleMovementRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		memberID := r.URL.Query().Get("member_id")
		records, _ := s.store.ListMovementRecords(memberID)
		if records == nil {
			records = []MovementRecord{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	case http.MethodPost:
		var req struct {
			MemberID   string  `json:"member_id"`
			Metric     *string `json:"metric"`
			Value      *string `json:"value"`
			Unit       *string `json:"unit"`
			Note       string  `json:"note"`
			RecordDate string  `json:"record_date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid json")
			return
		}
		rec := &MovementRecord{
			ID: newID(), MemberID: req.MemberID,
			Metric: req.Metric, Value: req.Value, Unit: req.Unit,
			Note: req.Note, RecordDate: req.RecordDate,
		}
		if err := s.store.CreateMovementRecord(rec); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rec)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleMovementRecordByID(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodDelete:
		s.store.DeleteMovementRecord(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
