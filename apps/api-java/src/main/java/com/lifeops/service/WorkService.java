package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.WorkProfile;
import com.lifeops.entity.WorkRecord;
import com.lifeops.entity.WorkStatus;
import com.lifeops.mapper.WorkProfileMapper;
import com.lifeops.mapper.WorkRecordMapper;
import com.lifeops.mapper.WorkStatusMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class WorkService {
    private final WorkStatusMapper statusMapper;
    private final WorkProfileMapper profileMapper;
    private final WorkRecordMapper recordMapper;

    // Status operations
    public List<WorkStatus> listStatuses() {
        return statusMapper.selectList(
            new LambdaQueryWrapper<WorkStatus>().orderByDesc(WorkStatus::getUpdatedAt)
        );
    }

    public WorkStatus getStatus(String memberId) {
        return statusMapper.selectOne(
            new LambdaQueryWrapper<WorkStatus>().eq(WorkStatus::getMemberId, memberId)
        );
    }

    public WorkStatus updateStatus(String memberId, String summary) {
        WorkStatus existing = getStatus(memberId);
        String now = OffsetDateTime.now().toString();
        if (existing == null) {
            existing = new WorkStatus();
            existing.setId(UUID.randomUUID().toString().replace("-", ""));
            existing.setMemberId(memberId);
        }
        existing.setSummary(summary);
        existing.setUpdatedAt(now);
        statusMapper.insertOrUpdate(existing);
        return getStatus(memberId);
    }

    // Profile operations
    public List<WorkProfile> listProfiles() {
        return profileMapper.selectList(
            new LambdaQueryWrapper<WorkProfile>().orderByDesc(WorkProfile::getUpdatedAt)
        );
    }

    public WorkProfile getProfile(String memberId) {
        return profileMapper.selectOne(
            new LambdaQueryWrapper<WorkProfile>().eq(WorkProfile::getMemberId, memberId)
        );
    }

    public WorkProfile updateProfile(String memberId, WorkProfile profile) {
        WorkProfile existing = getProfile(memberId);
        String now = OffsetDateTime.now().toString();
        if (existing == null) {
            profile.setId(UUID.randomUUID().toString().replace("-", ""));
            profile.setMemberId(memberId);
        } else {
            profile.setId(existing.getId());
            profile.setMemberId(memberId);
        }
        profile.setUpdatedAt(now);
        profileMapper.insertOrUpdate(profile);
        return getProfile(memberId);
    }

    public void deleteProfile(String memberId) {
        profileMapper.delete(
            new LambdaQueryWrapper<WorkProfile>().eq(WorkProfile::getMemberId, memberId)
        );
    }

    // Record operations
    public List<WorkRecord> listRecords(String memberId, String status) {
        LambdaQueryWrapper<WorkRecord> wrapper = new LambdaQueryWrapper<>();
        if (memberId != null && !memberId.isEmpty()) {
            wrapper.eq(WorkRecord::getMemberId, memberId);
        }
        if (status != null && !status.isEmpty()) {
            wrapper.eq(WorkRecord::getStatus, status);
        }
        wrapper.orderByDesc(WorkRecord::getCreatedAt);
        return recordMapper.selectList(wrapper);
    }

    public WorkRecord getRecord(String id) {
        return recordMapper.selectById(id);
    }

    public WorkRecord createRecord(WorkRecord record) {
        String now = OffsetDateTime.now().toString();
        record.setId(UUID.randomUUID().toString().replace("-", ""));
        record.setCreatedAt(now);
        record.setUpdatedAt(now);
        if (record.getStatus() == null || record.getStatus().isEmpty()) {
            record.setStatus("active");
        }
        if (record.getPriority() == null || record.getPriority().isEmpty()) {
            record.setPriority("medium");
        }
        recordMapper.insert(record);
        return record;
    }

    public void deleteRecord(String id) {
        recordMapper.deleteById(id);
    }
}
