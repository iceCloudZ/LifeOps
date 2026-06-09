package com.lifeops.agent;

import java.util.List;

/**
 * Result from Phase 1 routing: determines which domains and lens to use
 * for Phase 2 execution.
 *
 * @param domain        one or more domain identifiers (e.g. "finance", "health")
 * @param lens          the selected lens id (e.g. "conscious-spending"), or null
 * @param reason        short explanation of why this route was chosen
 * @param needsWebSearch whether Phase 2 should include a web search tool call
 */
public record RoutingResult(
    List<String> domain,
    String lens,
    String reason,
    boolean needsWebSearch
) {}
