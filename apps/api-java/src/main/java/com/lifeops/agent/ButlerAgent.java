package com.lifeops.agent;

import com.lifeops.entity.LlmUsage;
import com.lifeops.mapper.LlmUsageMapper;
import dev.langchain4j.data.message.SystemMessage;
import dev.langchain4j.data.message.UserMessage;
import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.chat.response.ChatResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Slf4j
@Component
@RequiredArgsConstructor
public class ButlerAgent {

    private final ChatModel chatModel;
    private final List<DomainAgent> domainAgents;
    private final LlmUsageMapper llmUsageMapper;

    @Value("${lifeops.agent.system-prompt:}")
    private String systemPrompt;

    public AgentResponse answer(String query, String conversationId) {
        String sessionId = conversationId != null ? conversationId : UUID.randomUUID().toString();

        String context = domainAgents.stream()
                .map(agent -> "## " + agent.domain() + "\n" + agent.retrieveContext(query))
                .collect(Collectors.joining("\n\n"));

        String fullPrompt = "## 家庭数据\n\n" + context + "\n\n## 用户问题\n" + query;

        long start = System.currentTimeMillis();
        ChatResponse response = chatModel.chat(
                SystemMessage.from(systemPrompt),
                UserMessage.from(fullPrompt)
        );
        long latencyMs = System.currentTimeMillis() - start;

        String content = response.aiMessage().text();
        int promptTokens = 0;
        int completionTokens = 0;
        int totalTokens = 0;
        if (response.tokenUsage() != null) {
            promptTokens = response.tokenUsage().inputTokenCount();
            completionTokens = response.tokenUsage().outputTokenCount();
            totalTokens = promptTokens + completionTokens;
        }

        recordUsage(sessionId, promptTokens, completionTokens, totalTokens, latencyMs);

        return new AgentResponse(content, promptTokens, completionTokens, totalTokens, latencyMs);
    }

    private void recordUsage(String conversationId, int promptTokens, int completionTokens,
                             int totalTokens, long latencyMs) {
        try {
            LlmUsage usage = new LlmUsage();
            usage.setId(UUID.randomUUID().toString().replace("-", ""));
            usage.setConversationId(conversationId);
            usage.setModel("deepseek-chat");
            usage.setDomain("butler");
            usage.setPromptTokens(promptTokens);
            usage.setCompletionTokens(completionTokens);
            usage.setTotalTokens(totalTokens);
            usage.setCostCents(0);
            usage.setLatencyMs((int) latencyMs);
            usage.setCreatedAt(OffsetDateTime.now().toString());
            llmUsageMapper.insert(usage);
        } catch (Exception e) {
            log.warn("Failed to record LLM usage", e);
        }
    }
}
