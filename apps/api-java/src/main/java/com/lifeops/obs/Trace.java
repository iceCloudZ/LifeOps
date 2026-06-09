package com.lifeops.obs;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

public class Trace {
    private final String traceId;
    private final String conversationId;
    private final Instant startTime;
    private final List<Span> spans = new ArrayList<>();
    private Instant endTime;
    private String status = "ok";

    public Trace(String traceId, String conversationId) {
        this.traceId = traceId;
        this.conversationId = conversationId;
        this.startTime = Instant.now();
    }

    public void addSpan(Span span) { spans.add(span); }

    public void complete() {
        this.endTime = Instant.now();
    }

    public void fail(String error) {
        this.status = "error: " + error;
        this.endTime = Instant.now();
    }

    public TokenUsage aggregateTokens() {
        int prompt = 0, completion = 0;
        for (Span s : spans) {
            if (s.tokenUsage() != null) {
                prompt += s.tokenUsage().promptTokens();
                completion += s.tokenUsage().completionTokens();
            }
        }
        return new TokenUsage(prompt, completion, prompt + completion);
    }

    public long durationMs() {
        if (endTime == null) return System.currentTimeMillis() - startTime.toEpochMilli();
        return endTime.toEpochMilli() - startTime.toEpochMilli();
    }

    // -- accessors --
    public String traceId() { return traceId; }
    public String conversationId() { return conversationId; }
    public Instant startTime() { return startTime; }
    public Instant endTime() { return endTime; }
    public List<Span> spans() { return List.copyOf(spans); }
    public String status() { return status; }
    public int spanCount() { return spans.size(); }
}
