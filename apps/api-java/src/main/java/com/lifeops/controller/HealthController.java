package com.lifeops.controller;

import com.lifeops.entity.HealthProfile;
import com.lifeops.entity.HealthRecord;
import com.lifeops.service.HealthService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/health")
@RequiredArgsConstructor
public class HealthController {
    private final HealthService healthService;

    // Profiles
    @GetMapping("/profiles")
    public List<HealthProfile> listProfiles() {
        return healthService.listProfiles();
    }

    @PutMapping("/profiles/{memberId}")
    public ResponseEntity<HealthProfile> updateProfile(@PathVariable String memberId, @RequestBody Map<String, String> body) {
        HealthProfile profile = healthService.updateProfile(memberId, body.get("summary"));
        return ResponseEntity.ok(profile);
    }

    // Records
    @GetMapping("/records")
    public List<HealthRecord> listRecords(
        @RequestParam(required = false) String member_id,
        @RequestParam(required = false) String type) {
        return healthService.listRecords(member_id, type);
    }

    @PostMapping("/records")
    public ResponseEntity<HealthRecord> createRecord(@RequestBody Map<String, Object> body) {
        HealthRecord record = new HealthRecord();
        record.setMemberId((String) body.get("member_id"));
        record.setType((String) body.get("type"));
        record.setMetric((String) body.get("metric"));
        record.setValue((String) body.get("value"));
        record.setUnit((String) body.get("unit"));
        record.setNote((String) body.get("note"));
        record.setRecordDate((String) body.get("record_date"));
        return ResponseEntity.status(HttpStatus.CREATED).body(healthService.createRecord(record));
    }

    @DeleteMapping("/records/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteRecord(@PathVariable String id) {
        healthService.deleteRecord(id);
    }
}
