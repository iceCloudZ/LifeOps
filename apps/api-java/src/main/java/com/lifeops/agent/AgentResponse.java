package com.lifeops.agent;

import lombok.Data;

@Data
public class AgentResponse {
    private String answer;
    private int promptTokens;
    private int completionTokens;
    private int totalTokens;
    private long latencyMs;

    public AgentResponse(String answer, int promptTokens, int completionTokens, int totalTokens, long latencyMs) {
        this.answer = answer;
        this.promptTokens = promptTokens;
        this.completionTokens = completionTokens;
        this.totalTokens = totalTokens;
        this.latencyMs = latencyMs;
    }
}
