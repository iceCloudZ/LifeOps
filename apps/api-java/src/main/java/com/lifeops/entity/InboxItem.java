package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("inbox_items")
public class InboxItem {
    @TableId
    private String id;
    private String source;
    private String sender;
    private String content;
    private String status;
    private String createdAt;
    private String updatedAt;
}
