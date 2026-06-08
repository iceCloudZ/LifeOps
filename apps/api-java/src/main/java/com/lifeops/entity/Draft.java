package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("drafts")
public class Draft {
    @TableId
    private String id;
    private String inboxItemId;
    private String draftType;
    private String title;
    private String description;
    private Double confidence;
    private String dueAt;
    private String assigneeHint;
    private String quantity;
    private String topic;
    private String status;
    private String entityId;
    private String createdAt;
    private String updatedAt;
}
