package com.lifeops.agent;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.mapper.ChatMessageMapper;
import dev.langchain4j.data.message.AiMessage;
import dev.langchain4j.data.message.SystemMessage;
import dev.langchain4j.data.message.UserMessage;
import dev.langchain4j.store.memory.chat.ChatMemoryStore;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.Objects;
import java.util.UUID;

@Slf4j
@RequiredArgsConstructor
public class SqliteChatMemoryStore implements ChatMemoryStore {

    private final ChatMessageMapper messageMapper;

    @Override
    public List<dev.langchain4j.data.message.ChatMessage> getMessages(Object memoryId) {
        String conversationId = memoryId.toString();
        List<com.lifeops.entity.ChatMessage> dbMessages = messageMapper.selectList(
            new LambdaQueryWrapper<com.lifeops.entity.ChatMessage>()
                .eq(com.lifeops.entity.ChatMessage::getConversationId, conversationId)
                .orderByAsc(com.lifeops.entity.ChatMessage::getCreatedAt)
        );

        return dbMessages.stream()
            .map(msg -> {
                if ("user".equals(msg.getRole())) {
                    return (dev.langchain4j.data.message.ChatMessage) UserMessage.from(msg.getContent());
                } else if ("assistant".equals(msg.getRole())) {
                    return AiMessage.from(msg.getContent());
                } else if ("system".equals(msg.getRole())) {
                    return SystemMessage.from(msg.getContent());
                }
                return null;
            })
            .filter(Objects::nonNull)
            .toList();
    }

    @Override
    public void updateMessages(Object memoryId, List<dev.langchain4j.data.message.ChatMessage> messages) {
        // LangChain4j calls this after each AI round trip with the full message window.
        // We don't delete-and-reinsert because that would lose metadata (tokensUsed, lensName, etc.)
        // already saved by ChatService. Instead, we do nothing — message persistence is handled
        // by ChatService's onComplete callback which preserves all entity fields.
    }

    @Override
    public void deleteMessages(Object memoryId) {
        messageMapper.delete(
            new LambdaQueryWrapper<com.lifeops.entity.ChatMessage>()
                .eq(com.lifeops.entity.ChatMessage::getConversationId, memoryId.toString())
        );
    }
}
