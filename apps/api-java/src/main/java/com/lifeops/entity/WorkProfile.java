package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("work_profiles")
public class WorkProfile {
    @TableId
    private String id;
    private String memberId;
    private String employmentStatus;
    private String company;
    private String position;
    private String industry;
    private String workLocation;
    private String incomeRange;
    private String workSchedule;
    private Integer commuteMinutes;
    private String startedAt;
    private String note;
    private String updatedAt;
}
