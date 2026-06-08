package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("family_records")
public class FamilyRecord {
    @TableId
    private String id;
    private String memberId;
    private String type;
    private String title;
    private String status;
    private String location;
    private String participants;
    private String scheduledDate;
    private String note;
    private String createdAt;
    private String updatedAt;
}
