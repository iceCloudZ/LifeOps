package com.lifeops.controller;

import com.lifeops.entity.WorkProfile;
import com.lifeops.entity.WorkRecord;
import com.lifeops.entity.WorkStatus;
import com.lifeops.service.WorkService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/work")
@RequiredArgsConstructor
public class WorkController {
    private final WorkService workService;

    // Statuses
    @GetMapping("/status")
    public List<WorkStatus> listStatuses() {
        return workService.listStatuses();
    }

    @PutMapping("/status/{memberId}")
    public ResponseEntity<WorkStatus> updateStatus(@PathVariable String memberId, @RequestBody Map<String, String> body) {
        return ResponseEntity.ok(workService.updateStatus(memberId, body.get("summary")));
    }

    // Profiles
    @GetMapping("/profiles")
    public List<WorkProfile> listProfiles() {
        return workService.listProfiles();
    }

    @GetMapping("/profiles/{memberId}")
    public ResponseEntity<WorkProfile> getProfile(@PathVariable String memberId) {
        WorkProfile profile = workService.getProfile(memberId);
        return ResponseEntity.ok(profile != null ? profile : new WorkProfile());
    }

    @PutMapping("/profiles/{memberId}")
    public ResponseEntity<WorkProfile> updateProfile(@PathVariable String memberId, @RequestBody WorkProfile profile) {
        return ResponseEntity.ok(workService.updateProfile(memberId, profile));
    }

    @DeleteMapping("/profiles/{memberId}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteProfile(@PathVariable String memberId) {
        workService.deleteProfile(memberId);
    }

    // Records
    @GetMapping("/records")
    public List<WorkRecord> listRecords(
        @RequestParam(required = false) String member_id,
        @RequestParam(required = false) String status) {
        return workService.listRecords(member_id, status);
    }

    @PostMapping("/records")
    public ResponseEntity<WorkRecord> createRecord(@RequestBody Map<String, Object> body) {
        WorkRecord record = new WorkRecord();
        record.setMemberId((String) body.get("member_id"));
        record.setType((String) body.get("type"));
        record.setTitle((String) body.get("title"));
        record.setStatus((String) body.get("status"));
        record.setPriority((String) body.get("priority"));
        record.setProject((String) body.get("project"));
        record.setDueDate((String) body.get("due_date"));
        record.setNote((String) body.get("note"));
        return ResponseEntity.status(HttpStatus.CREATED).body(workService.createRecord(record));
    }

    @DeleteMapping("/records/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteRecord(@PathVariable String id) {
        workService.deleteRecord(id);
    }
}
