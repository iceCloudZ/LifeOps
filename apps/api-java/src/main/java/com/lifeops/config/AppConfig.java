package com.lifeops.config;

import dev.langchain4j.model.chat.ChatModel;
import dev.langchain4j.model.chat.StreamingChatModel;
import dev.langchain4j.model.openai.OpenAiChatModel;
import dev.langchain4j.model.openai.OpenAiStreamingChatModel;
import com.lifeops.mapper.AIConfigMapper;
import com.lifeops.entity.AIConfig;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.time.Duration;

@Configuration
public class AppConfig {

    @Value("${lifeops.llm.base-url:}")
    private String baseUrl;

    @Value("${lifeops.llm.api-key:}")
    private String apiKey;

    @Value("${lifeops.llm.model:deepseek-chat}")
    private String model;

    @Bean
    @ConditionalOnMissingBean(ChatModel.class)
    public ChatModel chatModel(AIConfigMapper aiConfigMapper) {
        String endpoint = baseUrl;
        String key = apiKey;
        String m = model;

        try {
            AIConfig dbConfig = aiConfigMapper.selectOne(
                new LambdaQueryWrapper<AIConfig>().eq(AIConfig::getId, 1)
            );
            if (dbConfig != null) {
                if (endpoint == null || endpoint.isEmpty()) endpoint = dbConfig.getEndpoint();
                if (key == null || key.isEmpty()) key = dbConfig.getApiKey();
                if (m.equals("deepseek-chat") && dbConfig.getModel() != null) m = dbConfig.getModel();
            }
        } catch (Exception ignored) {}

        return OpenAiChatModel.builder()
                .baseUrl(endpoint)
                .apiKey(key)
                .modelName(m)
                .timeout(Duration.ofSeconds(30))
                .maxRetries(3)
                .build();
    }

    @Bean
    @ConditionalOnMissingBean(StreamingChatModel.class)
    public StreamingChatModel streamingChatModel(AIConfigMapper aiConfigMapper) {
        String endpoint = baseUrl;
        String key = apiKey;
        String m = model;

        try {
            AIConfig dbConfig = aiConfigMapper.selectOne(
                new LambdaQueryWrapper<AIConfig>().eq(AIConfig::getId, 1)
            );
            if (dbConfig != null) {
                if (endpoint == null || endpoint.isEmpty()) endpoint = dbConfig.getEndpoint();
                if (key == null || key.isEmpty()) key = dbConfig.getApiKey();
                if (m.equals("deepseek-chat") && dbConfig.getModel() != null) m = dbConfig.getModel();
            }
        } catch (Exception ignored) {}

        return OpenAiStreamingChatModel.builder()
                .baseUrl(endpoint)
                .apiKey(key)
                .modelName(m)
                .timeout(Duration.ofSeconds(60))
                .build();
    }
}
