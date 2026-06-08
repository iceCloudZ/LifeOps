package com.lifeops.agent;

import com.lifeops.entity.LlmUsage;
import com.lifeops.mapper.LlmUsageMapper;
import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.chat.response.ChatResponse;
import dev.langchain4j.model.output.TokenUsage;
import io.github.icecloudz.c2fe4j.agent.RouterAgent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.time.OffsetDateTime;
import java.util.UUID;

@Slf4j
@Component
@RequiredArgsConstructor
public class ButlerAgent {

    private final RouterAgent routerAgent;
    private final ChatModel chatModel;
    private final LlmUsageMapper llmUsageMapper;

    @Value("${c2fe4j.agent.system-prompt:}")
    private String systemPrompt;

    public AgentResponse answer(String query, String conversationId) {
        String sessionId = conversationId != null ? conversationId : UUID.randomUUID().toString();

        String content = routerAgent.answer(query, sessionId, systemPrompt);

        // Get token usage from a separate call isn't ideal — RouterAgent handles it internally
        // For now, record minimal usage info
        recordUsage(sessionId, content);

        return new AgentResponse(content, 0, 0, 0, 0);
    }

    private void recordUsage(String conversationId, String content) {
        try {
            LlmUsage usage = new LlmUsage();
            usage.setId(UUID.randomUUID().toString().replace("-", ""));
            usage.setConversationId(conversationId);
            usage.setModel("deepseek-chat");
            usage.setDomain("butler");
            usage.setPromptTokens(0);
            usage.setCompletionTokens(0);
            usage.setTotalTokens(0);
            usage.setCostCents(0);
            usage.setLatencyMs(0);
            usage.setCreatedAt(OffsetDateTime.now().toString());
            llmUsageMapper.insert(usage);
        } catch (Exception e) {
            log.warn("Failed to record LLM usage", e);
        }
    }
}
