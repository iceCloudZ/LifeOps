package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;

@Data
@TableName("t_chat_span")
public class ChatSpan {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long traceId;
    private Integer spanNo;
    private String spanType;
    private String spanName;
    private Long parentSpanId;
    private String inputData;
    private String outputData;
    private Integer promptTokens;
    private Integer completionTokens;
    private Integer cachedTokens;
    private Integer latencyMs;
    private String status;
    private String metadata;

    @TableField(fill = FieldFill.INSERT)
    private java.time.LocalDateTime createdAt;
}
