package com.lifeops.agent;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.lifeops.agent.tool.ToolDispatcher;
import dev.langchain4j.agent.tool.ToolExecutionRequest;
import dev.langchain4j.data.message.AiMessage;
import dev.langchain4j.data.message.ChatMessage;
import dev.langchain4j.data.message.SystemMessage;
import dev.langchain4j.data.message.ToolExecutionResultMessage;
import dev.langchain4j.data.message.UserMessage;
import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.chat.StreamingChatModel;
import dev.langchain4j.model.chat.request.ChatRequest;
import dev.langchain4j.agent.tool.ToolSpecification;
import dev.langchain4j.model.chat.response.ChatResponse;
import dev.langchain4j.model.chat.response.CompleteToolCall;
import dev.langchain4j.model.chat.response.StreamingChatResponseHandler;
import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CountDownLatch;

@Slf4j
@Component
@RequiredArgsConstructor
public class ButlerAgent {

    private final ChatModel chatModel;
    private final StreamingChatModel streamingChatModel;
    private final List<DomainAgent> domainAgents;
    private final LensRegistry lensRegistry;
    private final ToolDispatcher toolDispatcher;
    private final ObjectMapper objectMapper;

    @PostConstruct
    void init() {
        for (DomainAgent agent : domainAgents) {
            lensRegistry.registerDomainPrompt(agent.domain(), agent.systemPrompt());
        }
        log.info("ButlerAgent initialized with {} domain agents", domainAgents.size());
    }

    /**
     * Backward-compatible synchronous answer method used by ChatService.
     * Will be replaced with the two-phase route+execute pattern in a follow-up task.
     */
    public AgentResponse answer(String query, String conversationId) {
        String sessionId = conversationId != null ? conversationId : java.util.UUID.randomUUID().toString();

        long start = System.currentTimeMillis();
        ChatResponse response = chatModel.chat(
            SystemMessage.from("You are Butler, a helpful personal life management assistant."),
            UserMessage.from(query)
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

        return new AgentResponse(content, promptTokens, completionTokens, totalTokens, latencyMs);
    }

    /**
     * Phase 1: Route the user query to the appropriate domains and lens.
     */
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
            // Strip markdown code fences if present
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

    /**
     * Phase 2: Execute with streaming and tool support.
     */
    public void executeWithTools(SseEmitter emitter, String query, RoutingResult routing,
                                  String conversationId, String currentMemberId) {
        // Build system message from lens content + domain prompts + base instructions
        StringBuilder systemBuilder = new StringBuilder();
        systemBuilder.append("You are Butler, a personal life management AI assistant.\n\n");

        // Add lens content if specified
        if (routing.lens() != null) {
            String lensContent = lensRegistry.getLensContent(routing.lens());
            if (lensContent != null) {
                systemBuilder.append("## Active Lens\n\n").append(lensContent).append("\n\n");
            }
        }

        // Add domain prompts
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

        List<ChatMessage> messages = new ArrayList<>();
        messages.add(SystemMessage.from(systemBuilder.toString()));
        messages.add(UserMessage.from(query));

        List<ToolSpecification> toolSpecs = toolDispatcher.getToolSpecifications();
        streamWithToolLoop(emitter, messages, toolSpecs);
    }

    /**
     * Recursive streaming loop: stream tokens, handle tool calls, recurse.
     */
    private void streamWithToolLoop(SseEmitter emitter, List<ChatMessage> messages,
                                     List<ToolSpecification> toolSpecs) {
        ChatRequest chatRequest = ChatRequest.builder()
                .messages(messages)
                .toolSpecifications(toolSpecs)
                .build();

        StringBuilder tokenBuffer = new StringBuilder();
        List<CompleteToolCall> toolCalls = new ArrayList<>();
        CountDownLatch latch = new CountDownLatch(1);

        streamingChatModel.chat(chatRequest, new StreamingChatResponseHandler() {
            @Override
            public void onPartialResponse(String partialResponse) {
                tokenBuffer.append(partialResponse);
                sendSse(emitter, "token", partialResponse);
            }

            @Override
            public void onCompleteToolCall(CompleteToolCall completeToolCall) {
                toolCalls.add(completeToolCall);
                try {
                    String args = completeToolCall.toolExecutionRequest().arguments();
                    String name = completeToolCall.toolExecutionRequest().name();
                    sendSse(emitter, "tool_start", objectMapper.writeValueAsString(
                        java.util.Map.of("tool", name, "arguments", args != null ? args : "{}")));
                } catch (JsonProcessingException e) {
                    sendSse(emitter, "tool_start", "{\"tool\":\"" +
                        completeToolCall.toolExecutionRequest().name() + "\"}");
                }
            }

            @Override
            public void onCompleteResponse(ChatResponse chatResponse) {
                try {
                    if (!toolCalls.isEmpty()) {
                        // Build AiMessage with text and tool execution requests
                        String text = tokenBuffer.length() > 0 ? tokenBuffer.toString() : null;
                        List<ToolExecutionRequest> requests = toolCalls.stream()
                                .map(CompleteToolCall::toolExecutionRequest)
                                .toList();
                        AiMessage aiMessage = AiMessage.builder()
                                .text(text)
                                .toolExecutionRequests(requests)
                                .build();
                        messages.add(aiMessage);

                        // Execute each tool
                        for (ToolExecutionRequest request : requests) {
                            String toolName = request.name();
                            String arguments = request.arguments();
                            sendSse(emitter, "confirm", "Executing: " + toolName);

                            String result = toolDispatcher.execute(toolName, arguments);
                            ToolExecutionResultMessage resultMsg = ToolExecutionResultMessage.from(request, result);
                            messages.add(resultMsg);

                            sendSse(emitter, "tool_result", result);
                        }

                        // Recurse for next turn
                        streamWithToolLoop(emitter, messages, toolSpecs);
                    } else {
                        // No tool calls — we're done
                        sendSse(emitter, "done", tokenBuffer.toString());
                        emitter.complete();
                    }
                } catch (Exception e) {
                    log.error("Error in onCompleteResponse", e);
                    try {
                        emitter.completeWithError(e);
                    } catch (Exception ignored) {}
                } finally {
                    latch.countDown();
                }
            }

            @Override
            public void onError(Throwable error) {
                log.error("Streaming error", error);
                try {
                    sendSse(emitter, "error", error.getMessage());
                    emitter.completeWithError(error);
                } catch (Exception ignored) {}
                latch.countDown();
            }
        });

        // Block until the streaming response completes (handles async tool loop)
        try {
            latch.await();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            log.warn("Streaming interrupted");
        }
    }

    /**
     * Helper to send SSE events with error handling.
     */
    private void sendSse(SseEmitter emitter, String eventName, String data) {
        try {
            emitter.send(SseEmitter.event().name(eventName).data(data));
        } catch (Exception e) {
            log.warn("Failed to send SSE event '{}': {}", eventName, e.getMessage());
        }
    }
}
