package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("llm_usage")
public class LlmUsage {
    @TableId
    private String id;
    private String conversationId;
    private String model;
    private String lensId;
    private String domain;
    private Integer promptTokens;
    private Integer completionTokens;
    private Integer totalTokens;
    private Integer costCents;
    private Integer latencyMs;
    private String createdAt;
}
