package com.lifeops.service.trace;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.lifeops.entity.ChatSpan;
import com.lifeops.entity.ChatTrace;
import com.lifeops.mapper.ChatSpanMapper;
import com.lifeops.mapper.ChatTraceMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

@Slf4j
@Service
@RequiredArgsConstructor
public class ChatTraceService {

    private static final int MAX_DATA_LENGTH = 8000;

    private final ChatTraceMapper traceMapper;
    private final ChatSpanMapper spanMapper;
    private final ObjectMapper objectMapper;

    public ChatTrace createTrace(String conversationId, String inputMessage, String model, String lensId) {
        try {
            ChatTrace trace = new ChatTrace();
            trace.setConversationId(conversationId);
            trace.setTraceNo(traceMapper.nextTraceNo(conversationId));
            trace.setInputMessage(inputMessage);
            trace.setModel(model);
            trace.setLensId(lensId);
            trace.setStatus("running");
            trace.setTotalPromptTokens(0);
            trace.setTotalCompletionTokens(0);
            trace.setTotalTokens(0);
            trace.setTotalCachedTokens(0);
            trace.setCostYuan(BigDecimal.ZERO);
            trace.setTotalLatencyMs(0);
            trace.setLlmCallCount(0);
            trace.setToolCallCount(0);
            traceMapper.insert(trace);
            return trace;
        } catch (Exception e) {
            log.error("Failed to create trace for conversation {}", conversationId, e);
            return null;
        }
    }

    public Long addSpan(Long traceId, int spanNo, String spanType, String spanName,
                        Long parentSpanId, String inputData, String outputData,
                        int promptTokens, int completionTokens, int cachedTokens,
                        int latencyMs, String status, String metadata) {
        try {
            ChatSpan span = new ChatSpan();
            span.setTraceId(traceId);
            span.setSpanNo(spanNo);
            span.setSpanType(spanType);
            span.setSpanName(spanName);
            span.setParentSpanId(parentSpanId);
            span.setInputData(truncate(inputData));
            span.setOutputData(truncate(outputData));
            span.setPromptTokens(promptTokens);
            span.setCompletionTokens(completionTokens);
            span.setCachedTokens(cachedTokens);
            span.setLatencyMs(latencyMs);
            span.setStatus(status);
            span.setMetadata(metadata);
            spanMapper.insert(span);
            return span.getId();
        } catch (Exception e) {
            log.error("Failed to add span for trace {}", traceId, e);
            return null;
        }
    }

    public void completeTrace(Long traceId, String outputMessage,
                              int totalPromptTokens, int totalCompletionTokens,
                              int totalTokens, int totalCachedTokens,
                              BigDecimal costYuan, int totalLatencyMs,
                              int llmCallCount, int toolCallCount) {
        try {
            ChatTrace trace = traceMapper.selectById(traceId);
            if (trace == null) return;
            trace.setOutputMessage(truncate(outputMessage));
            trace.setTotalPromptTokens(totalPromptTokens);
            trace.setTotalCompletionTokens(totalCompletionTokens);
            trace.setTotalTokens(totalTokens);
            trace.setTotalCachedTokens(totalCachedTokens);
            trace.setCostYuan(costYuan);
            trace.setTotalLatencyMs(totalLatencyMs);
            trace.setLlmCallCount(llmCallCount);
            trace.setToolCallCount(toolCallCount);
            trace.setStatus("ok");
            traceMapper.updateById(trace);
        } catch (Exception e) {
            log.error("Failed to complete trace {}", traceId, e);
        }
    }

    public void failTrace(Long traceId, String errorMessage) {
        try {
            ChatTrace trace = traceMapper.selectById(traceId);
            if (trace == null) return;
            trace.setStatus("error");
            trace.setErrorMessage(truncate(errorMessage));
            traceMapper.updateById(trace);
        } catch (Exception e) {
            log.error("Failed to fail trace {}", traceId, e);
        }
    }

    public Map<String, Object> listSessions(int page, int size) {
        try {
            int offset = (page - 1) * size;

            List<Map<String, Object>> sessions = traceMapper.selectMaps(
                    new QueryWrapper<ChatTrace>()
                            .select("conversation_id", "MAX(created_at) as last_activity")
                            .groupBy("conversation_id")
                            .orderByDesc("last_activity")
                            .last("LIMIT " + size + " OFFSET " + offset));

            List<Map<String, Object>> countResult = traceMapper.selectMaps(
                    new QueryWrapper<ChatTrace>()
                            .select("COUNT(DISTINCT conversation_id) as cnt"));
            long total = countResult.isEmpty() ? 0 : ((Number) countResult.get(0).get("cnt")).longValue();

            for (Map<String, Object> session : sessions) {
                String convId = (String) session.get("conversation_id");
                enrichSessionStats(session, convId);
            }

            Map<String, Object> result = new LinkedHashMap<>();
            result.put("records", sessions);
            result.put("total", total);
            return result;
        } catch (Exception e) {
            log.error("Failed to list sessions", e);
            return Map.of("records", List.of(), "total", 0);
        }
    }

    private void enrichSessionStats(Map<String, Object> session, String conversationId) {
        List<ChatTrace> traces = traceMapper.selectList(
                new LambdaQueryWrapper<ChatTrace>()
                        .eq(ChatTrace::getConversationId, conversationId)
                        .orderByAsc(ChatTrace::getTraceNo));

        int totalTokens = 0;
        BigDecimal totalCost = BigDecimal.ZERO;

        for (ChatTrace t : traces) {
            totalTokens += t.getTotalTokens() != null ? t.getTotalTokens() : 0;
            if (t.getCostYuan() != null) totalCost = totalCost.add(t.getCostYuan());
        }

        session.put("traceCount", traces.size());
        session.put("totalTokens", totalTokens);
        session.put("totalCostYuan", totalCost);
    }

    public List<ChatTrace> listTraces(String conversationId) {
        return traceMapper.selectList(
                new LambdaQueryWrapper<ChatTrace>()
                        .eq(ChatTrace::getConversationId, conversationId)
                        .orderByAsc(ChatTrace::getTraceNo));
    }

    public Map<String, Object> getTraceDetail(Long traceId) {
        ChatTrace trace = traceMapper.selectById(traceId);
        List<ChatSpan> spans = spanMapper.selectList(
                new LambdaQueryWrapper<ChatSpan>()
                        .eq(ChatSpan::getTraceId, traceId)
                        .orderByAsc(ChatSpan::getSpanNo));

        Map<String, Object> result = new LinkedHashMap<>();
        result.put("trace", trace);
        result.put("spans", spans);
        return result;
    }

    public void markRunningTracesAsInterrupted() {
        try {
            List<ChatTrace> running = traceMapper.selectList(
                    new LambdaQueryWrapper<ChatTrace>()
                            .eq(ChatTrace::getStatus, "running"));
            for (ChatTrace trace : running) {
                trace.setStatus("error");
                trace.setErrorMessage("Server shutdown interrupted");
                traceMapper.updateById(trace);
            }
            if (!running.isEmpty()) {
                log.warn("Marked {} running traces as error due to shutdown", running.size());
            }
        } catch (Exception e) {
            log.error("Failed to mark running traces as interrupted", e);
        }
    }

    private String truncate(String data) {
        if (data == null) return null;
        if (data.length() <= MAX_DATA_LENGTH) return data;
        return data.substring(0, MAX_DATA_LENGTH) + "...[truncated]";
    }
}
