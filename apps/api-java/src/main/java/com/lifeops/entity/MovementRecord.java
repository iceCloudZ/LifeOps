package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("movement_records")
public class MovementRecord {
    @TableId
    private String id;
    private String memberId;
    private String metric;
    private String value;
    private String unit;
    private String note;
    private String recordDate;
    private String createdAt;
    private String updatedAt;
}
