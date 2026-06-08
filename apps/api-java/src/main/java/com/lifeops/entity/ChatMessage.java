package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("messages")
public class ChatMessage {
    @TableId
    private String id;
    private String conversationId;
    private String role;
    private String content;
    private String agentUsed;
    private Integer tokensUsed;
    private String lensId;
    private String lensName;
    private String lensReason;
    private String createdAt;
}
