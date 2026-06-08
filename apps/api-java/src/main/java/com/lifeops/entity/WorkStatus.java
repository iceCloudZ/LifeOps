package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("work_statuses")
public class WorkStatus {
    @TableId
    private String id;
    private String memberId;
    private String summary;
    private String updatedAt;
}
