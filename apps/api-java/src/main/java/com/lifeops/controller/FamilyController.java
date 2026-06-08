package com.lifeops.controller;

import com.lifeops.entity.FamilyRecord;
import com.lifeops.entity.FamilyStatus;
import com.lifeops.service.FamilyService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/family")
@RequiredArgsConstructor
public class FamilyController {
    private final FamilyService familyService;

    @GetMapping("/status")
    public ResponseEntity<FamilyStatus> getStatus() {
        FamilyStatus status = familyService.getStatus();
        if (status == null) {
            status = new FamilyStatus();
            status.setId("family");
        }
        return ResponseEntity.ok(status);
    }

    @PutMapping("/status")
    public ResponseEntity<FamilyStatus> updateStatus(@RequestBody Map<String, String> body) {
        return ResponseEntity.ok(familyService.updateStatus(body.get("summary")));
    }

    @GetMapping("/records")
    public List<FamilyRecord> listRecords(
        @RequestParam(required = false) String member_id,
        @RequestParam(required = false) String type,
        @RequestParam(required = false) String status) {
        return familyService.listRecords(member_id, type, status);
    }

    @PostMapping("/records")
    public ResponseEntity<FamilyRecord> createRecord(@RequestBody Map<String, Object> body) {
        FamilyRecord record = new FamilyRecord();
        record.setMemberId((String) body.get("member_id"));
        record.setType((String) body.get("type"));
        record.setTitle((String) body.get("title"));
        record.setStatus((String) body.get("status"));
        record.setLocation((String) body.get("location"));
        record.setParticipants((String) body.get("participants"));
        record.setScheduledDate((String) body.get("scheduled_date"));
        record.setNote((String) body.get("note"));
        return ResponseEntity.status(HttpStatus.CREATED).body(familyService.createRecord(record));
    }

    @DeleteMapping("/records/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteRecord(@PathVariable String id) {
        familyService.deleteRecord(id);
    }
}
