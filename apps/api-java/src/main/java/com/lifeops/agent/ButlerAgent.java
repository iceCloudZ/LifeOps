package com.lifeops.agent;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.lifeops.agent.tool.LifeOpsReadTools;
import com.lifeops.agent.tool.LifeOpsWriteTools;
import com.lifeops.agent.tool.ToolDispatcher;
import com.lifeops.entity.ChatTrace;
import com.lifeops.service.trace.ChatTraceService;
import dev.langchain4j.data.message.SystemMessage;
import dev.langchain4j.data.message.UserMessage;
import dev.langchain4j.memory.ChatMemory;
import dev.langchain4j.memory.chat.MessageWindowChatMemory;
import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.chat.StreamingChatModel;
import dev.langchain4j.model.chat.response.ChatResponse;
import dev.langchain4j.service.AiServices;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.math.BigDecimal;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Consumer;

@Slf4j
@Component
@RequiredArgsConstructor
public class ButlerAgent {

    private final ChatModel chatModel;
    private final StreamingChatModel streamingChatModel;
    private final List<DomainAgent> domainAgents;
    private final LensRegistry lensRegistry;
    private final ToolDispatcher toolDispatcher;
    private final LifeOpsReadTools readTools;
    private final LifeOpsWriteTools writeTools;
    private final ObjectMapper objectMapper;
    private final ChatTraceService traceService;
    private final SqliteChatMemoryStore chatMemoryStore;

    @PostConstruct
    void init() {
        for (DomainAgent agent : domainAgents) {
            lensRegistry.registerDomainPrompt(agent.domain(), agent.systemPrompt());
        }
        log.info("ButlerAgent initialized with {} domain agents", domainAgents.size());
    }

    @PreDestroy
    void shutdown() {
        traceService.markRunningTracesAsInterrupted();
    }

    public AgentResponse answer(String query, String conversationId) {
        long start = System.currentTimeMillis();
        ChatResponse response = chatModel.chat(
            SystemMessage.from("You are Butler, a helpful personal life management assistant."),
            UserMessage.from(query)
        );
        long latencyMs = System.currentTimeMillis() - start;

        String content = response.aiMessage().text();
        int promptTokens = 0, completionTokens = 0, totalTokens = 0;
        if (response.tokenUsage() != null) {
            promptTokens = response.tokenUsage().inputTokenCount();
            completionTokens = response.tokenUsage().outputTokenCount();
            totalTokens = promptTokens + completionTokens;
        }

        return new AgentResponse(content, promptTokens, completionTokens, totalTokens, latencyMs);
    }

    public RoutingResult route(String query, String currentMemberId) {
        String lensIndex = lensRegistry.getLensIndexText();

        String routingPrompt = """
            You are a routing assistant for a personal life management AI called Butler.
            Given the user query, determine which domains and lens are most relevant.

            %s

            Respond with ONLY a JSON object (no markdown fences, no explanation):
            {"domain":["<domain>"],"lens":"<lens-id-or-null>","reason":"<brief reason>","needsWebSearch":false}

            If the query does not match any specific domain, use domain=["general"] and lens=null.
            Set needsWebSearch to true only if the query clearly asks about current/recent information
            that would not be available from the user's personal data.
            """.formatted(lensIndex);

        try {
            ChatResponse response = chatModel.chat(
                SystemMessage.from(routingPrompt),
                UserMessage.from(query)
            );
            String raw = response.aiMessage().text();
            if (raw != null) {
                raw = raw.trim();
                if (raw.startsWith("```")) {
                    int end = raw.indexOf('\n');
                    if (end >= 0) raw = raw.substring(end + 1);
                    if (raw.endsWith("```")) raw = raw.substring(0, raw.length() - 3);
                    raw = raw.trim();
                }
            }
            return objectMapper.readValue(raw, RoutingResult.class);
        } catch (Exception e) {
            log.warn("Routing failed, using default route: {}", e.getMessage());
            return new RoutingResult(List.of("general"), null, "fallback: routing parse error", false);
        }
    }

    public void executeWithTools(SseEmitter emitter, String query, RoutingResult routing,
                                  String conversationId, String currentMemberId,
                                  Consumer<String> onComplete) {
        Thread.startVirtualThread(() -> {
            try {
                doExecuteWithTools(emitter, query, routing, conversationId, currentMemberId, onComplete);
            } catch (Exception e) {
                log.error("Execute with tools failed", e);
                sendSse(emitter, "error", e.getMessage());
                emitter.completeWithError(e);
            }
        });
    }

    private void doExecuteWithTools(SseEmitter emitter, String query, RoutingResult routing,
                                     String conversationId, String currentMemberId,
                                     Consumer<String> onComplete) {
        long startMs = System.currentTimeMillis();

        ChatTrace trace = traceService.createTrace(conversationId, query, "deepseek-v4-pro",
            routing.lens());
        Long traceId = trace != null ? trace.getId() : null;
        AtomicInteger spanNo = new AtomicInteger(0);

        if (traceId != null) {
            traceService.addSpan(traceId, spanNo.incrementAndGet(), "routing", "route:" + routing.domain(),
                null, query, routing.toString(), 0, 0, 0, 0, "ok", null);
        }

        String systemPrompt = buildSystemPrompt(routing);
        AtomicInteger totalPrompt = new AtomicInteger(0);
        AtomicInteger totalCompletion = new AtomicInteger(0);
        AtomicInteger llmCount = new AtomicInteger(0);
        AtomicInteger toolCount = new AtomicInteger(0);
        StringBuilder responseBuffer = new StringBuilder();

        ChatMemory chatMemory = MessageWindowChatMemory.builder()
            .id(conversationId)
            .maxMessages(200)
            .chatMemoryStore(chatMemoryStore)
            .build();
        chatMemory.add(SystemMessage.from(systemPrompt));

        ButlerAssistant assistant = AiServices.builder(ButlerAssistant.class)
            .streamingChatModel(streamingChatModel)
            .chatMemory(chatMemory)
            .tools(readTools, writeTools)
            .build();

        assistant.chat(query)
            .onPartialResponse(token -> {
                responseBuffer.append(token);
                sendSse(emitter, "token", token);
            })
            .beforeToolExecution(before -> {
                String toolName = before.request().name();
                String args = before.request().arguments();
                try {
                    sendSse(emitter, "tool_start", objectMapper.writeValueAsString(
                        java.util.Map.of("tool", toolName, "arguments", args != null ? args : "{}")));
                } catch (JsonProcessingException e) {
                    sendSse(emitter, "tool_start", "{\"tool\":\"" + toolName + "\"}");
                }
            })
            .onToolExecuted(execution -> {
                toolCount.incrementAndGet();
                String toolName = execution.request().name();
                String result = execution.result();
                sendSse(emitter, "tool_result", result);

                if (traceId != null) {
                    traceService.addSpan(traceId, spanNo.incrementAndGet(),
                        "tool_execution", toolName, null,
                        truncate(execution.request().arguments(), 500),
                        truncate(result, 2000),
                        0, 0, 0, 0, "ok", null);
                }
            })
            .onIntermediateResponse(response -> {
                llmCount.incrementAndGet();
                int promptTokens = response.tokenUsage() != null ? response.tokenUsage().inputTokenCount() : 0;
                int completionTokens = response.tokenUsage() != null ? response.tokenUsage().outputTokenCount() : 0;
                totalPrompt.addAndGet(promptTokens);
                totalCompletion.addAndGet(completionTokens);

                if (traceId != null) {
                    traceService.addSpan(traceId, spanNo.incrementAndGet(),
                        "llm_call", "execute", null, null,
                        responseBuffer.length() > 0 ? responseBuffer.toString() : null,
                        promptTokens, completionTokens, 0, 0, "ok", null);
                }
            })
            .onCompleteResponse(response -> {
                int latencyMs = (int) (System.currentTimeMillis() - startMs);

                llmCount.incrementAndGet();
                int promptTokens = response.tokenUsage() != null ? response.tokenUsage().inputTokenCount() : 0;
                int completionTokens = response.tokenUsage() != null ? response.tokenUsage().outputTokenCount() : 0;
                totalPrompt.addAndGet(promptTokens);
                totalCompletion.addAndGet(completionTokens);

                if (traceId != null) {
                    traceService.addSpan(traceId, spanNo.incrementAndGet(),
                        "llm_call", "execute", null, null,
                        responseBuffer.length() > 0 ? responseBuffer.toString() : null,
                        promptTokens, completionTokens, 0, 0, "ok", null);

                    traceService.completeTrace(traceId, null,
                        totalPrompt.get(), totalCompletion.get(),
                        totalPrompt.get() + totalCompletion.get(), 0,
                        BigDecimal.ZERO, latencyMs,
                        llmCount.get(), toolCount.get());
                }

                String finalText = responseBuffer.toString();
                sendSse(emitter, "done", finalText);
                if (onComplete != null) onComplete.accept(finalText);
                emitter.complete();
            })
            .onError(error -> {
                log.error("Streaming error", error);
                if (traceId != null) {
                    traceService.addSpan(traceId, spanNo.incrementAndGet(), "llm_call", "execute",
                        null, null, error.getMessage(), 0, 0, 0, 0, "error", null);
                    traceService.failTrace(traceId, error.getMessage());
                }
                sendSse(emitter, "error", error.getMessage());
                emitter.completeWithError(error);
            })
            .start();
    }

    private String buildSystemPrompt(RoutingResult routing) {
        StringBuilder systemBuilder = new StringBuilder();
        systemBuilder.append("You are Butler, a personal life management AI assistant.\n\n");

        if (routing.lens() != null) {
            String lensContent = lensRegistry.getLensContent(routing.lens());
            if (lensContent != null) {
                systemBuilder.append("## Active Lens\n\n").append(lensContent).append("\n\n");
            }
        }

        for (String domain : routing.domain()) {
            String prompt = lensRegistry.getDomainPrompt(domain);
            if (prompt != null) {
                systemBuilder.append("## Domain: ").append(domain).append("\n")
                        .append(prompt).append("\n\n");
            }
        }

        systemBuilder.append("""
            ## Instructions
            - Help the user with their personal life management tasks.
            - Use the available tools when you need to read or modify data.
            - Always confirm before making destructive changes.
            - Respond concisely and helpfully.
            """);

        return systemBuilder.toString();
    }

    private void sendSse(SseEmitter emitter, String eventName, String data) {
        try {
            emitter.send(SseEmitter.event().name(eventName).data(data));
        } catch (Exception e) {
            log.warn("Failed to send SSE event '{}': {}", eventName, e.getMessage());
        }
    }

    private static String truncate(String s, int maxLen) {
        if (s == null) return null;
        return s.length() <= maxLen ? s : s.substring(0, maxLen) + "...";
    }
}
