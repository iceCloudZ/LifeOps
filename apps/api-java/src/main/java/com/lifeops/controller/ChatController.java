package com.lifeops.controller;

import com.lifeops.agent.RoutingResult;
import com.lifeops.entity.ChatMessage;
import com.lifeops.entity.Conversation;
import com.lifeops.service.ChatService;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/chat/conversations")
@RequiredArgsConstructor
public class ChatController {
    private final ChatService chatService;

    @GetMapping
    public List<Conversation> listConversations() {
        return chatService.listConversations();
    }

    @PostMapping
    public ResponseEntity<Conversation> createConversation(@RequestBody Map<String, String> body) {
        Conversation conv = chatService.createConversation(body.get("title"));
        return ResponseEntity.status(HttpStatus.CREATED).body(conv);
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteConversation(@PathVariable String id) {
        chatService.deleteConversation(id);
    }

    @GetMapping("/{id}/messages")
    public List<ChatMessage> listMessages(@PathVariable String id) {
        return chatService.listMessages(id);
    }

    @PostMapping("/{id}/messages")
    public ResponseEntity<ChatMessage> sendMessage(@PathVariable String id, @RequestBody Map<String, String> body) {
        String content = body.get("content");
        if (content == null || content.trim().isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        Conversation conv = chatService.getConversation(id);
        if (conv == null) {
            return ResponseEntity.notFound().build();
        }
        ChatMessage msg = chatService.sendMessage(id, content.trim());
        return ResponseEntity.status(HttpStatus.CREATED).body(msg);
    }

    // Merged route + execute: single SSE stream
    @PostMapping(value = "/{id}/chat", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public SseEmitter chat(@PathVariable String id, @RequestBody Map<String, String> body,
                            HttpServletResponse response) {
        response.setHeader("X-Accel-Buffering", "no");
        response.setHeader("Cache-Control", "no-cache");
        String content = body.get("content");
        String currentMemberId = body.get("currentMemberId");
        if (content == null || content.trim().isEmpty()) return new SseEmitter(0L);
        return chatService.chatStreaming(id, content.trim(), currentMemberId);
    }

    // Backward compat: direct execute with pre-determined routing
    @PostMapping(value = "/{id}/execute", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public SseEmitter execute(@PathVariable String id, @RequestBody Map<String, Object> body,
                               HttpServletResponse response) {
        response.setHeader("X-Accel-Buffering", "no");
        response.setHeader("Cache-Control", "no-cache");
        String content = (String) body.get("content");
        String currentMemberId = (String) body.get("currentMemberId");
        String lens = (String) body.get("lens");
        @SuppressWarnings("unchecked")
        List<String> domain = (List<String>) body.get("domain");
        RoutingResult routing = new RoutingResult(domain != null ? domain : List.of("family"), lens, "", false);
        return chatService.executeStreaming(id, content, routing, currentMemberId);
    }

    // Confirm write operation
    @PostMapping("/{id}/confirm/{messageId}")
    public ResponseEntity<Map<String, Boolean>> confirm(@PathVariable String id,
                                                          @PathVariable String messageId,
                                                          @RequestBody Map<String, Object> body) {
        boolean success = chatService.confirmWrite(body);
        return ResponseEntity.ok(Map.of("success", success));
    }
}
