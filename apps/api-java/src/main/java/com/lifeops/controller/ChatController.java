package com.lifeops.controller;

import com.lifeops.entity.ChatMessage;
import com.lifeops.entity.Conversation;
import com.lifeops.service.ChatService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

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
}
