package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("ai_config")
public class AIConfig {
    @TableId
    private Integer id;
    private String endpoint;
    private String model;
    private String apiKey;
    private Integer maxTokens;
}
