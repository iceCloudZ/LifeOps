package com.lifeops.controller;

import com.lifeops.entity.InboxItem;
import com.lifeops.service.InboxService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/inbox")
@RequiredArgsConstructor
public class InboxController {
    private final InboxService inboxService;

    @PostMapping("/webhook")
    public ResponseEntity<InboxItem> webhookInbox(@RequestBody Map<String, String> body) {
        String source = body.get("source");
        String content = body.get("content");
        if (source == null || source.trim().isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        if (content == null || content.trim().isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        InboxItem item = new InboxItem();
        item.setSource(source);
        item.setSender(body.getOrDefault("sender", ""));
        item.setContent(content);
        return ResponseEntity.status(HttpStatus.CREATED).body(inboxService.createInboxItem(item));
    }
}
