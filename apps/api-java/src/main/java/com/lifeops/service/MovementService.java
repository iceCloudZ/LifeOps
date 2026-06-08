package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.MovementRecord;
import com.lifeops.mapper.MovementRecordMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class MovementService {
    private final MovementRecordMapper recordMapper;

    public List<MovementRecord> listRecords(String memberId) {
        LambdaQueryWrapper<MovementRecord> wrapper = new LambdaQueryWrapper<>();
        if (memberId != null && !memberId.isEmpty()) {
            wrapper.eq(MovementRecord::getMemberId, memberId);
        }
        wrapper.orderByDesc(MovementRecord::getRecordDate);
        return recordMapper.selectList(wrapper);
    }

    public MovementRecord getRecord(String id) {
        return recordMapper.selectById(id);
    }

    public MovementRecord createRecord(MovementRecord record) {
        String now = OffsetDateTime.now().toString();
        record.setId(UUID.randomUUID().toString().replace("-", ""));
        record.setCreatedAt(now);
        record.setUpdatedAt(now);
        recordMapper.insert(record);
        return record;
    }

    public void deleteRecord(String id) {
        recordMapper.deleteById(id);
    }
}
