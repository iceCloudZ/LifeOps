package com.lifeops.controller;

import com.lifeops.entity.MovementRecord;
import com.lifeops.service.MovementService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/movement")
@RequiredArgsConstructor
public class MovementController {
    private final MovementService movementService;

    @GetMapping("/records")
    public List<MovementRecord> listRecords(@RequestParam(required = false) String member_id) {
        return movementService.listRecords(member_id);
    }

    @PostMapping("/records")
    public ResponseEntity<MovementRecord> createRecord(@RequestBody Map<String, Object> body) {
        MovementRecord record = new MovementRecord();
        record.setMemberId((String) body.get("member_id"));
        record.setMetric((String) body.get("metric"));
        record.setValue((String) body.get("value"));
        record.setUnit((String) body.get("unit"));
        record.setNote((String) body.get("note"));
        record.setRecordDate((String) body.get("record_date"));
        return ResponseEntity.status(HttpStatus.CREATED).body(movementService.createRecord(record));
    }

    @DeleteMapping("/records/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteRecord(@PathVariable String id) {
        movementService.deleteRecord(id);
    }
}
