package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("finance_accounts")
public class FinanceAccount {
    @TableId
    private String id;
    private String memberId;
    private String name;
    private String type;
    private Double balance;
    private String note;
    private String createdAt;
    private String updatedAt;
}
