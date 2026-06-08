package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("conversations")
public class Conversation {
    @TableId
    private String id;
    private String title;
    private String createdAt;
    private String updatedAt;
}
