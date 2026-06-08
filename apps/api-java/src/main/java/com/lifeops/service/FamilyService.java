package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.FamilyRecord;
import com.lifeops.entity.FamilyStatus;
import com.lifeops.mapper.FamilyRecordMapper;
import com.lifeops.mapper.FamilyStatusMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class FamilyService {
    private final FamilyStatusMapper statusMapper;
    private final FamilyRecordMapper recordMapper;

    public FamilyStatus getStatus() {
        return statusMapper.selectOne(
            new LambdaQueryWrapper<FamilyStatus>().last("LIMIT 1")
        );
    }

    public FamilyStatus updateStatus(String summary) {
        FamilyStatus existing = getStatus();
        String now = OffsetDateTime.now().toString();
        if (existing == null) {
            existing = new FamilyStatus();
            existing.setId("family");
            existing.setUpdatedAt(now);
        }
        existing.setSummary(summary);
        existing.setUpdatedAt(now);
        statusMapper.insertOrUpdate(existing);
        return getStatus();
    }

    public List<FamilyRecord> listRecords(String memberId, String type, String status) {
        LambdaQueryWrapper<FamilyRecord> wrapper = new LambdaQueryWrapper<>();
        if (memberId != null && !memberId.isEmpty()) {
            wrapper.eq(FamilyRecord::getMemberId, memberId);
        }
        if (type != null && !type.isEmpty()) {
            wrapper.eq(FamilyRecord::getType, type);
        }
        if (status != null && !status.isEmpty()) {
            wrapper.eq(FamilyRecord::getStatus, status);
        }
        wrapper.orderByDesc(FamilyRecord::getCreatedAt);
        return recordMapper.selectList(wrapper);
    }

    public FamilyRecord createRecord(FamilyRecord record) {
        String now = OffsetDateTime.now().toString();
        record.setId(UUID.randomUUID().toString().replace("-", ""));
        record.setCreatedAt(now);
        record.setUpdatedAt(now);
        if (record.getStatus() == null || record.getStatus().isEmpty()) {
            record.setStatus("pending");
        }
        if (record.getParticipants() == null || record.getParticipants().isEmpty()) {
            record.setParticipants("[]");
        }
        recordMapper.insert(record);
        return record;
    }

    public void deleteRecord(String id) {
        recordMapper.deleteById(id);
    }
}
