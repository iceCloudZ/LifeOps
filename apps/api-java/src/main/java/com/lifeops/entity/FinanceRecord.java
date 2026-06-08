package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("finance_records")
public class FinanceRecord {
    @TableId
    private String id;
    private String memberId;
    private String type;
    private Double amount;
    private String currency;
    private String category;
    private String note;
    private String recordDate;
    private String createdAt;
    private String updatedAt;
}
