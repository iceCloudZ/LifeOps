package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("health_profiles")
public class HealthProfile {
    @TableId
    private String id;
    private String memberId;
    private String summary;
    private String updatedAt;
}
