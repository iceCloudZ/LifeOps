package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("family_statuses")
public class FamilyStatus {
    @TableId
    private String id;
    private String summary;
    private String updatedAt;
}
