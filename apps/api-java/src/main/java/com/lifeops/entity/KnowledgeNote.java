package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("knowledge_notes")
public class KnowledgeNote {
    @TableId
    private String id;
    private String domain;
    private String memberId;
    private String title;
    private String content;
    private String tags;
    private String createdAt;
    private String updatedAt;
}
