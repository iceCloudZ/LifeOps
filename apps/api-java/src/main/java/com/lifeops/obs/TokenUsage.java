package com.lifeops.obs;

public record TokenUsage(
    int promptTokens,
    int completionTokens,
    int totalTokens
) {
    public static TokenUsage empty() {
        return new TokenUsage(0, 0, 0);
    }

    public static TokenUsage of(int prompt, int completion) {
        return new TokenUsage(prompt, completion, prompt + completion);
    }
}
