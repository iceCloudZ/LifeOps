package com.lifeops.controller;

import dev.langchain4j.data.message.SystemMessage;
import dev.langchain4j.data.message.UserMessage;
import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.chat.response.ChatResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/quick-entry")
@RequiredArgsConstructor
public class QuickEntryController {
    private final ChatModel chatModel;

    @PostMapping
    public ResponseEntity<?> quickEntry(@RequestBody QuickEntryRequest body) {
        String input = body.input();
        if (input == null || input.trim().isEmpty()) {
            return ResponseEntity.badRequest().body(java.util.Map.of("error", "invalid_request", "message", "input is required"));
        }

        String prompt = "你是一个家庭信息整理助手。请将用户输入的自然语言拆分为多个领域的结构化记录。\n\n"
            + "领域包括：finance（财务）、health（健康）、work（工作）、family（家庭事务）。\n\n"
            + "请严格按以下JSON格式返回，不要返回其他内容：\n"
            + "{\"entries\":[{\"domain\":\"finance|health|work|family\",\"title\":\"标题\",\"content\":\"详细描述\",\"data\":{}}]}\n\n"
            + "用户输入：" + input;

        try {
            ChatResponse result = chatModel.chat(
                SystemMessage.from("你是家庭信息整理助手，只返回JSON，不要解释。"),
                UserMessage.from(prompt)
            );
            return ResponseEntity.ok().header("Content-Type", "application/json").body(result.aiMessage().text());
        } catch (Exception e) {
            return ResponseEntity.internalServerError().body(java.util.Map.of("error", "ai_error", "message", e.getMessage()));
        }
    }

    record QuickEntryRequest(String input) {}
}
