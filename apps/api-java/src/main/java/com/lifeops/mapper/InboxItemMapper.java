package com.lifeops.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.lifeops.entity.InboxItem;
import org.apache.ibatis.annotations.Mapper;

@Mapper
public interface InboxItemMapper extends BaseMapper<InboxItem> {
}
