package com.lifeops.controller;

import com.lifeops.service.trace.ChatTraceService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/admin/traces")
@RequiredArgsConstructor
public class ChatTraceController {

    private final ChatTraceService traceService;

    @GetMapping("/sessions")
    public Map<String, Object> listSessions(@RequestParam(defaultValue = "1") int page,
                                            @RequestParam(defaultValue = "20") int size) {
        return traceService.listSessions(page, size);
    }

    @GetMapping("/sessions/{conversationId}")
    public Object listTraces(@PathVariable String conversationId) {
        return traceService.listTraces(conversationId);
    }

    @GetMapping("/{traceId}")
    public Map<String, Object> traceDetail(@PathVariable Long traceId) {
        return traceService.getTraceDetail(traceId);
    }
}
