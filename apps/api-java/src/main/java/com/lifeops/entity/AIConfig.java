package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
@TableName("ai_config")
public class AIConfig {
    @TableId
    private Integer id;
    private String endpoint;
    private String model;
    @JsonProperty("api_key")
    private String apiKey;
    @JsonProperty("max_tokens")
    private Integer maxTokens;
}
