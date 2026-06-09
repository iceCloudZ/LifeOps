package com.lifeops.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.lifeops.entity.ChatTrace;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface ChatTraceMapper extends BaseMapper<ChatTrace> {

    @Select("SELECT COALESCE(MAX(trace_no), 0) + 1 FROM t_chat_trace WHERE conversation_id = #{conversationId}")
    int nextTraceNo(String conversationId);
}
