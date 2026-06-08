package com.lifeops.controller;

import com.lifeops.entity.FinanceAccount;
import com.lifeops.entity.FinanceRecord;
import com.lifeops.service.FinanceService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/finance")
@RequiredArgsConstructor
public class FinanceController {
    private final FinanceService financeService;

    // Accounts
    @GetMapping("/accounts")
    public List<FinanceAccount> listAccounts() {
        return financeService.listAccounts();
    }

    @PostMapping("/accounts")
    public ResponseEntity<FinanceAccount> createAccount(@RequestBody Map<String, Object> body) {
        FinanceAccount account = new FinanceAccount();
        account.setMemberId((String) body.get("member_id"));
        account.setName((String) body.get("name"));
        account.setType((String) body.get("type"));
        account.setBalance(body.get("balance") != null ? ((Number) body.get("balance")).doubleValue() : 0.0);
        account.setNote((String) body.get("note"));
        return ResponseEntity.status(HttpStatus.CREATED).body(financeService.createAccount(account));
    }

    @PutMapping("/accounts/{id}")
    public ResponseEntity<FinanceAccount> updateAccount(@PathVariable String id, @RequestBody Map<String, Object> body) {
        FinanceAccount account = new FinanceAccount();
        account.setMemberId((String) body.get("member_id"));
        account.setName((String) body.get("name"));
        account.setType((String) body.get("type"));
        account.setNote((String) body.get("note"));
        return ResponseEntity.ok(financeService.updateAccount(id, account));
    }

    @DeleteMapping("/accounts/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteAccount(@PathVariable String id) {
        financeService.deleteAccount(id);
    }

    // Records
    @GetMapping("/records")
    public List<FinanceRecord> listRecords(
        @RequestParam(required = false) String member_id,
        @RequestParam(required = false) String type,
        @RequestParam(required = false) String category,
        @RequestParam(required = false) String from_date,
        @RequestParam(required = false) String to_date) {
        return financeService.listRecords(member_id, type, category, from_date, to_date);
    }

    @PostMapping("/records")
    public ResponseEntity<FinanceRecord> createRecord(@RequestBody Map<String, Object> body) {
        FinanceRecord record = new FinanceRecord();
        record.setMemberId((String) body.get("member_id"));
        record.setType((String) body.get("type"));
        record.setAmount(body.get("amount") != null ? ((Number) body.get("amount")).doubleValue() : 0.0);
        record.setCurrency((String) body.get("currency"));
        record.setCategory((String) body.get("category"));
        record.setNote((String) body.get("note"));
        record.setRecordDate((String) body.get("record_date"));
        return ResponseEntity.status(HttpStatus.CREATED).body(financeService.createRecord(record));
    }

    @DeleteMapping("/records/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteRecord(@PathVariable String id) {
        financeService.deleteRecord(id);
    }
}
