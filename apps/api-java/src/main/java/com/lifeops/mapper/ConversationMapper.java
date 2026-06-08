package com.lifeops.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.lifeops.entity.Conversation;
import org.apache.ibatis.annotations.Mapper;

@Mapper
public interface ConversationMapper extends BaseMapper<Conversation> {
}
