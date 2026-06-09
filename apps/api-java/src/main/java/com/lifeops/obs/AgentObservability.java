package com.lifeops.obs;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Instant;
import java.util.UUID;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.function.Consumer;

/**
 * Lightweight agent trace/span factory with per-conversation context.
 * Reports completed traces to a pluggable consumer (default: SLF4J log).
 */
public class AgentObservability {

    private static final Logger log = LoggerFactory.getLogger(AgentObservability.class);

    private static final ConcurrentLinkedQueue<Trace> recentTraces = new ConcurrentLinkedQueue<>();
    private static final int MAX_RECENT = 500;

    private static volatile Consumer<Trace> reporter = trace -> {
        String summary = String.format(
            "trace=%s conv=%s spans=%d prompt=%d completion=%d total=%d latency=%dms status=%s",
            trace.traceId(), trace.conversationId(), trace.spanCount(),
            trace.aggregateTokens().promptTokens(),
            trace.aggregateTokens().completionTokens(),
            trace.aggregateTokens().totalTokens(),
            trace.durationMs(), trace.status()
        );
        log.info("AgentTrace: {}", summary);
    };

    public static void setReporter(Consumer<Trace> reporter) {
        AgentObservability.reporter = reporter;
    }

    // -- factory methods --

    public static Trace startTrace(String conversationId) {
        return new Trace(UUID.randomUUID().toString().replace("-", "").substring(0, 12), conversationId);
    }

    public static Span startSpan(String traceId, String type, String name) {
        return startSpan(traceId, null, type, name);
    }

    public static Span startSpan(String traceId, String parentSpanId, String type, String name) {
        return new Span(
            UUID.randomUUID().toString().replace("-", "").substring(0, 8),
            parentSpanId, traceId, type, name,
            Instant.now(), null, null, "running", null
        );
    }

    public static Span endSpan(Span span, TokenUsage tokens) {
        return new Span(span.spanId(), span.parentSpanId(), span.traceId(),
            span.type(), span.name(), span.startTime(), Instant.now(),
            tokens, "ok", null);
    }

    public static Span endSpanWithError(Span span, String error) {
        return new Span(span.spanId(), span.parentSpanId(), span.traceId(),
            span.type(), span.name(), span.startTime(), Instant.now(),
            null, "error", error);
    }

    public static void finishTrace(Trace trace) {
        trace.complete();
        recentTraces.add(trace);
        while (recentTraces.size() > MAX_RECENT) recentTraces.poll();
        Consumer<Trace> r = reporter;
        if (r != null) {
            try { r.accept(trace); } catch (Exception ignored) {}
        }
    }

    public static Trace[] recentTraces() {
        return recentTraces.toArray(new Trace[0]);
    }
}
