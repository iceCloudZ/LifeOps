package com.lifeops.entity;

import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("family_members")
public class FamilyMember {
    @TableId
    private String id;
    private String name;
    private String role;
    private String birthDate;
    private String createdAt;
    private String updatedAt;
}
