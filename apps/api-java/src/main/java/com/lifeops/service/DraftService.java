package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.Draft;
import com.lifeops.mapper.DraftMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class DraftService {
    private final DraftMapper draftMapper;

    public List<Draft> listDrafts(String status) {
        LambdaQueryWrapper<Draft> wrapper = new LambdaQueryWrapper<>();
        if (status != null && !status.isEmpty()) {
            wrapper.eq(Draft::getStatus, status);
        }
        wrapper.orderByDesc(Draft::getCreatedAt);
        return draftMapper.selectList(wrapper);
    }

    public Draft getDraft(String id) {
        return draftMapper.selectById(id);
    }
}
