package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.HealthProfile;
import com.lifeops.entity.HealthRecord;
import com.lifeops.mapper.HealthProfileMapper;
import com.lifeops.mapper.HealthRecordMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class HealthService {
    private final HealthProfileMapper profileMapper;
    private final HealthRecordMapper recordMapper;

    // Profile operations
    public List<HealthProfile> listProfiles() {
        return profileMapper.selectList(
            new LambdaQueryWrapper<HealthProfile>().orderByDesc(HealthProfile::getUpdatedAt)
        );
    }

    public HealthProfile getProfile(String memberId) {
        return profileMapper.selectOne(
            new LambdaQueryWrapper<HealthProfile>().eq(HealthProfile::getMemberId, memberId)
        );
    }

    public HealthProfile updateProfile(String memberId, String summary) {
        HealthProfile existing = getProfile(memberId);
        String now = OffsetDateTime.now().toString();
        if (existing == null) {
            existing = new HealthProfile();
            existing.setId(UUID.randomUUID().toString().replace("-", ""));
            existing.setMemberId(memberId);
        }
        existing.setSummary(summary);
        existing.setUpdatedAt(now);
        profileMapper.insertOrUpdate(existing);
        return getProfile(memberId);
    }

    public void deleteProfile(String memberId) {
        profileMapper.delete(
            new LambdaQueryWrapper<HealthProfile>().eq(HealthProfile::getMemberId, memberId)
        );
    }

    // Record operations
    public List<HealthRecord> listRecords(String memberId, String type) {
        LambdaQueryWrapper<HealthRecord> wrapper = new LambdaQueryWrapper<>();
        if (memberId != null && !memberId.isEmpty()) {
            wrapper.eq(HealthRecord::getMemberId, memberId);
        }
        if (type != null && !type.isEmpty()) {
            wrapper.eq(HealthRecord::getType, type);
        }
        wrapper.orderByDesc(HealthRecord::getRecordDate);
        return recordMapper.selectList(wrapper);
    }

    public HealthRecord getRecord(String id) {
        return recordMapper.selectById(id);
    }

    public HealthRecord createRecord(HealthRecord record) {
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
