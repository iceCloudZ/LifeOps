package com.lifeops.controller;

import com.lifeops.service.LlmAdminService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/admin/llm")
@RequiredArgsConstructor
public class LlmAdminController {
    private final LlmAdminService llmAdminService;

    @GetMapping("/dashboard")
    public ResponseEntity<Map<String, Object>> dashboard() {
        return ResponseEntity.ok(llmAdminService.getDashboard());
    }

    @GetMapping("/usage")
    public ResponseEntity<List<Map<String, Object>>> usage(
        @RequestParam(required = false, defaultValue = "day") String groupBy,
        @RequestParam(required = false) String startDate,
        @RequestParam(required = false) String endDate) {
        return ResponseEntity.ok(llmAdminService.getUsage(groupBy, startDate, endDate));
    }

    @GetMapping("/cost-trend")
    public ResponseEntity<List<Map<String, Object>>> costTrend(
        @RequestParam(required = false, defaultValue = "7") int days) {
        return ResponseEntity.ok(llmAdminService.getCostTrend(days));
    }

    @GetMapping("/top-sessions")
    public ResponseEntity<List<Map<String, Object>>> topSessions(
        @RequestParam(required = false, defaultValue = "10") int limit) {
        return ResponseEntity.ok(llmAdminService.getTopSessions(limit));
    }
}
