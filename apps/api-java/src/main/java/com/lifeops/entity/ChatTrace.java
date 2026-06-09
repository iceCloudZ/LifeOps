package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("t_chat_trace")
public class ChatTrace {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String conversationId;
    private Integer traceNo;
    private String inputMessage;
    private String outputMessage;
    private String model;
    private String lensId;
    private Integer totalPromptTokens;
    private Integer totalCompletionTokens;
    private Integer totalTokens;
    private Integer totalCachedTokens;
    private BigDecimal costYuan;
    private Integer totalLatencyMs;
    private Integer llmCallCount;
    private Integer toolCallCount;
    private String status;
    private String errorMessage;
    private String metadata;

    @TableField(fill = FieldFill.INSERT)
    private LocalDateTime createdAt;
}
