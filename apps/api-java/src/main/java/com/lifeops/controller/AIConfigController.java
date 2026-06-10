package com.lifeops.controller;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.AIConfig;
import com.lifeops.mapper.AIConfigMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/config/ai")
@RequiredArgsConstructor
public class AIConfigController {
    private final AIConfigMapper aiConfigMapper;

    @GetMapping
    public ResponseEntity<AIConfig> getConfig() {
        AIConfig config = aiConfigMapper.selectOne(
            new LambdaQueryWrapper<AIConfig>().eq(AIConfig::getId, 1)
        );
        if (config == null) {
            config = new AIConfig();
            config.setEndpoint("https://api.openai.com/v1");
            config.setModel("gpt-4o-mini");
            config.setMaxTokens(2048);
        }
        // Mask the API key
        if (config.getApiKey() != null && config.getApiKey().length() > 4) {
            config.setApiKey("****" + config.getApiKey().substring(config.getApiKey().length() - 4));
        } else if (config.getApiKey() != null && !config.getApiKey().isEmpty()) {
            config.setApiKey("****");
        }
        return ResponseEntity.ok(config);
    }

    @PutMapping
    public ResponseEntity<Map<String, String>> updateConfig(@RequestBody AIConfig config) {
        if (config.getApiKey() == null || config.getApiKey().isEmpty() || config.getApiKey().startsWith("****")) {
            AIConfig existing = aiConfigMapper.selectOne(
                new LambdaQueryWrapper<AIConfig>().eq(AIConfig::getId, 1)
            );
            if (existing != null) {
                config.setApiKey(existing.getApiKey());
            }
        }
        config.setId(1);
        aiConfigMapper.insertOrUpdate(config);
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
