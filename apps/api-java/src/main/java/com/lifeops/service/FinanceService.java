package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.FinanceAccount;
import com.lifeops.entity.FinanceRecord;
import com.lifeops.mapper.FinanceAccountMapper;
import com.lifeops.mapper.FinanceRecordMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class FinanceService {
    private final FinanceAccountMapper accountMapper;
    private final FinanceRecordMapper recordMapper;

    // Account operations
    public List<FinanceAccount> listAccounts() {
        return accountMapper.selectList(
            new LambdaQueryWrapper<FinanceAccount>().orderByDesc(FinanceAccount::getCreatedAt)
        );
    }

    public FinanceAccount getAccount(String id) {
        return accountMapper.selectById(id);
    }

    public FinanceAccount createAccount(FinanceAccount account) {
        String now = OffsetDateTime.now().toString();
        account.setId(UUID.randomUUID().toString().replace("-", ""));
        account.setCreatedAt(now);
        account.setUpdatedAt(now);
        accountMapper.insert(account);
        return account;
    }

    public FinanceAccount updateAccount(String id, FinanceAccount account) {
        account.setId(id);
        account.setUpdatedAt(OffsetDateTime.now().toString());
        accountMapper.updateById(account);
        return accountMapper.selectById(id);
    }

    public void deleteAccount(String id) {
        accountMapper.deleteById(id);
    }

    // Record operations
    public List<FinanceRecord> listRecords(String memberId, String type, String category, String fromDate, String toDate) {
        LambdaQueryWrapper<FinanceRecord> wrapper = new LambdaQueryWrapper<>();
        if (memberId != null && !memberId.isEmpty()) {
            wrapper.eq(FinanceRecord::getMemberId, memberId);
        }
        if (type != null && !type.isEmpty()) {
            wrapper.eq(FinanceRecord::getType, type);
        }
        if (category != null && !category.isEmpty()) {
            wrapper.eq(FinanceRecord::getCategory, category);
        }
        if (fromDate != null && !fromDate.isEmpty()) {
            wrapper.ge(FinanceRecord::getRecordDate, fromDate);
        }
        if (toDate != null && !toDate.isEmpty()) {
            wrapper.le(FinanceRecord::getRecordDate, toDate);
        }
        wrapper.orderByDesc(FinanceRecord::getRecordDate);
        return recordMapper.selectList(wrapper);
    }

    public FinanceRecord getRecord(String id) {
        return recordMapper.selectById(id);
    }

    public FinanceRecord createRecord(FinanceRecord record) {
        String now = OffsetDateTime.now().toString();
        record.setId(UUID.randomUUID().toString().replace("-", ""));
        record.setCreatedAt(now);
        record.setUpdatedAt(now);
        if (record.getCurrency() == null || record.getCurrency().isEmpty()) {
            record.setCurrency("CNY");
        }
        recordMapper.insert(record);
        return record;
    }

    public void deleteRecord(String id) {
        recordMapper.deleteById(id);
    }
}
