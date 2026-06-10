package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.lifeops.agent.ButlerAgent;
import com.lifeops.agent.RoutingResult;
import com.lifeops.entity.*;
import com.lifeops.mapper.ChatMessageMapper;
import com.lifeops.mapper.ConversationMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;
import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class ChatService {
    private final ConversationMapper conversationMapper;
    private final ChatMessageMapper messageMapper;
    private final ButlerAgent butlerAgent;
    private final ObjectMapper objectMapper;
    private final FinanceService financeService;
    private final HealthService healthService;
    private final MovementService movementService;
    private final WorkService workService;
    private final FamilyService familyService;
    private final NoteService noteService;

    public List<Conversation> listConversations() {
        return conversationMapper.selectList(
            new LambdaQueryWrapper<Conversation>().orderByDesc(Conversation::getUpdatedAt)
        );
    }

    public Conversation createConversation(String title) {
        String now = OffsetDateTime.now().toString();
        Conversation conv = new Conversation();
        conv.setId(UUID.randomUUID().toString().replace("-", ""));
        conv.setTitle(title != null ? title.trim() : "");
        conv.setCreatedAt(now);
        conv.setUpdatedAt(now);
        conversationMapper.insert(conv);
        return conv;
    }

    public Conversation getConversation(String id) {
        return conversationMapper.selectById(id);
    }

    public void deleteConversation(String id) {
        conversationMapper.deleteById(id);
    }

    public List<ChatMessage> listMessages(String conversationId) {
        return messageMapper.selectList(
            new LambdaQueryWrapper<ChatMessage>()
                .eq(ChatMessage::getConversationId, conversationId)
                .orderByAsc(ChatMessage::getCreatedAt)
        );
    }

    public ChatMessage sendMessage(String conversationId, String content) {
        // Save user message
        ChatMessage userMsg = new ChatMessage();
        userMsg.setId(UUID.randomUUID().toString().replace("-", ""));
        userMsg.setConversationId(conversationId);
        userMsg.setRole("user");
        userMsg.setContent(content);
        userMsg.setTokensUsed(0);
        userMsg.setCreatedAt(OffsetDateTime.now().toString());
        messageMapper.insert(userMsg);

        // Get AI response
        String answer;
        int tokensUsed = 0;

        try {
            var response = butlerAgent.answer(content, conversationId);
            answer = response.getAnswer();
            tokensUsed = response.getTotalTokens();
        } catch (Exception e) {
            answer = "AI回答失败: " + e.getMessage();
        }

        // Save assistant message
        ChatMessage assistantMsg = new ChatMessage();
        assistantMsg.setId(UUID.randomUUID().toString().replace("-", ""));
        assistantMsg.setConversationId(conversationId);
        assistantMsg.setRole("assistant");
        assistantMsg.setContent(answer);
        assistantMsg.setTokensUsed(tokensUsed);
        assistantMsg.setCreatedAt(OffsetDateTime.now().toString());
        messageMapper.insert(assistantMsg);

        // Update conversation timestamp
        Conversation conv = conversationMapper.selectById(conversationId);
        if (conv != null) {
            conv.setUpdatedAt(OffsetDateTime.now().toString());
            conversationMapper.updateById(conv);
        }

        return assistantMsg;
    }

    public SseEmitter chatStreaming(String conversationId, String content, String currentMemberId) {
        SseEmitter emitter = new SseEmitter(300_000L);
        emitter.onCompletion(() -> log.debug("SSE completed for {}", conversationId));
        emitter.onTimeout(() -> { log.warn("SSE timeout for {}", conversationId); emitter.complete(); });

        // Save user message
        ChatMessage userMsg = new ChatMessage();
        userMsg.setId(UUID.randomUUID().toString().replace("-", ""));
        userMsg.setConversationId(conversationId);
        userMsg.setRole("user");
        userMsg.setContent(content);
        userMsg.setTokensUsed(0);
        userMsg.setCreatedAt(OffsetDateTime.now().toString());
        messageMapper.insert(userMsg);

        Thread.startVirtualThread(() -> {
            try {
                // Phase 1: Route
                RoutingResult routing = butlerAgent.route(content, currentMemberId);
                // Send routing result as SSE event for frontend
                emitter.send(SseEmitter.event().name("route")
                    .data(objectMapper.writeValueAsString(routing)));

                // Phase 2: Execute with streaming
                butlerAgent.executeWithTools(emitter, content, routing, conversationId, currentMemberId,
                    finalText -> {
                        String now = OffsetDateTime.now().toString();
                        ChatMessage assistantMsg = new ChatMessage();
                        assistantMsg.setId(UUID.randomUUID().toString().replace("-", ""));
                        assistantMsg.setConversationId(conversationId);
                        assistantMsg.setRole("assistant");
                        assistantMsg.setContent(finalText);
                        assistantMsg.setLensName(routing.lens());
                        assistantMsg.setLensReason("");
                        assistantMsg.setTokensUsed(0);
                        assistantMsg.setCreatedAt(now);
                        messageMapper.insert(assistantMsg);
                        Conversation conv = conversationMapper.selectById(conversationId);
                        if (conv != null) {
                            conv.setUpdatedAt(now);
                            conversationMapper.updateById(conv);
                        }
                    });
            } catch (Exception e) {
                log.error("Chat streaming failed", e);
                try { emitter.completeWithError(e); } catch (Exception ignored) {}
            }
        });

        // Update conversation timestamp
        Conversation conv = conversationMapper.selectById(conversationId);
        if (conv != null) {
            conv.setUpdatedAt(OffsetDateTime.now().toString());
            conversationMapper.updateById(conv);
        }

        return emitter;
    }

    public SseEmitter executeStreaming(String conversationId, String content,
                                        RoutingResult routing, String currentMemberId) {
        SseEmitter emitter = new SseEmitter(300_000L);
        emitter.onCompletion(() -> log.debug("SSE completed for {}", conversationId));
        emitter.onTimeout(() -> { log.warn("SSE timeout for {}", conversationId); emitter.complete(); });
        butlerAgent.executeWithTools(emitter, content, routing, conversationId, currentMemberId,
            finalText -> {
                String now = OffsetDateTime.now().toString();
                ChatMessage assistantMsg = new ChatMessage();
                assistantMsg.setId(UUID.randomUUID().toString().replace("-", ""));
                assistantMsg.setConversationId(conversationId);
                assistantMsg.setRole("assistant");
                assistantMsg.setContent(finalText);
                assistantMsg.setLensName(routing.lens());
                assistantMsg.setLensReason("");
                assistantMsg.setTokensUsed(0);
                assistantMsg.setCreatedAt(now);
                messageMapper.insert(assistantMsg);
                Conversation conv = conversationMapper.selectById(conversationId);
                if (conv != null) {
                    conv.setUpdatedAt(now);
                    conversationMapper.updateById(conv);
                }
            });
        return emitter;
    }

    public boolean confirmWrite(Map<String, Object> body) {
        try {
            @SuppressWarnings("unchecked")
            Map<String, Object> data = (Map<String, Object>) body.get("data");
            String action = (String) body.get("action");
            if (action == null || data == null) return false;
            switch (action) {
                case "createFinanceRecord" -> {
                    FinanceRecord r = new FinanceRecord();
                    r.setMemberId((String) data.get("memberId"));
                    r.setAmount(((Number) data.get("amount")).doubleValue());
                    r.setType((String) data.get("type"));
                    r.setCategory((String) data.get("category"));
                    r.setNote((String) data.get("note"));
                    r.setRecordDate(java.time.LocalDate.now().toString());
                    financeService.createRecord(r);
                    return true;
                }
                case "createHealthRecord" -> {
                    HealthRecord r = new HealthRecord();
                    r.setMemberId((String) data.get("memberId"));
                    r.setMetric((String) data.get("metric"));
                    r.setValue((String) data.get("value"));
                    r.setUnit((String) data.get("unit"));
                    r.setNote((String) data.get("note"));
                    r.setRecordDate(java.time.LocalDate.now().toString());
                    healthService.createRecord(r);
                    return true;
                }
                case "createMovementRecord" -> {
                    MovementRecord r = new MovementRecord();
                    r.setMemberId((String) data.get("memberId"));
                    r.setMetric((String) data.get("metric"));
                    r.setValue((String) data.get("value"));
                    r.setNote((String) data.get("note"));
                    r.setRecordDate(java.time.LocalDate.now().toString());
                    movementService.createRecord(r);
                    return true;
                }
                case "createWorkRecord" -> {
                    WorkRecord r = new WorkRecord();
                    r.setMemberId((String) data.get("memberId"));
                    r.setTitle((String) data.get("title"));
                    r.setProject((String) data.get("project"));
                    r.setPriority((String) data.get("priority"));
                    r.setStatus("active");
                    workService.createRecord(r);
                    return true;
                }
                case "createFamilyRecord" -> {
                    FamilyRecord r = new FamilyRecord();
                    r.setMemberId((String) data.get("memberId"));
                    r.setType((String) data.get("type"));
                    r.setTitle((String) data.get("title"));
                    r.setScheduledDate((String) data.get("scheduledDate"));
                    r.setStatus("pending");
                    familyService.createRecord(r);
                    return true;
                }
                case "createNote" -> {
                    KnowledgeNote n = new KnowledgeNote();
                    n.setDomain((String) data.get("domain"));
                    n.setTitle((String) data.get("title"));
                    n.setContent((String) data.get("content"));
                    n.setMemberId((String) data.get("memberId"));
                    noteService.createNote(n);
                    return true;
                }
                default -> { return false; }
            }
        } catch (Exception e) {
            log.error("Confirm write failed for action: {}", body.get("action"), e);
            return false;
        }
    }

}
