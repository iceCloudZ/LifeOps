package com.lifeops.obs;

import java.time.Instant;

public record Span(
    String spanId,
    String parentSpanId,
    String traceId,
    String type,
    String name,
    Instant startTime,
    Instant endTime,
    TokenUsage tokenUsage,
    String status,
    String error
) {
    public long durationMs() {
        if (startTime == null || endTime == null) return 0;
        return endTime.toEpochMilli() - startTime.toEpochMilli();
    }
}
