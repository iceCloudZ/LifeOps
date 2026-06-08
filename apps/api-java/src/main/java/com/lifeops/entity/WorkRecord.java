package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("work_records")
public class WorkRecord {
    @TableId
    private String id;
    private String memberId;
    private String type;
    private String title;
    private String status;
    private String priority;
    private String project;
    private String dueDate;
    private String note;
    private String createdAt;
    private String updatedAt;
}
