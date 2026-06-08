package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.agent.ButlerAgent;
import com.lifeops.entity.ChatMessage;
import com.lifeops.entity.Conversation;
import com.lifeops.mapper.ChatMessageMapper;
import com.lifeops.mapper.ConversationMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class ChatService {
    private final ConversationMapper conversationMapper;
    private final ChatMessageMapper messageMapper;
    private final ButlerAgent butlerAgent;

    public List<Conversation> listConversations() {
        return conversationMapper.selectList(
            new LambdaQueryWrapper<Conversation>().orderByDesc(Conversation::getUpdatedAt)
        );
    }

    public Conversation createConversation(String title) {
        String now = OffsetDateTime.now().toString();
        Conversation conv = new Conversation();
        conv.setId(UUID.randomUUID().toString().replace("-", ""));
        conv.setTitle(title != null ? title.trim() : "");
        conv.setCreatedAt(now);
        conv.setUpdatedAt(now);
        conversationMapper.insert(conv);
        return conv;
    }

    public Conversation getConversation(String id) {
        return conversationMapper.selectById(id);
    }

    public void deleteConversation(String id) {
        conversationMapper.deleteById(id);
    }

    public List<ChatMessage> listMessages(String conversationId) {
        return messageMapper.selectList(
            new LambdaQueryWrapper<ChatMessage>()
                .eq(ChatMessage::getConversationId, conversationId)
                .orderByAsc(ChatMessage::getCreatedAt)
        );
    }

    public ChatMessage sendMessage(String conversationId, String content) {
        // Save user message
        ChatMessage userMsg = new ChatMessage();
        userMsg.setId(UUID.randomUUID().toString().replace("-", ""));
        userMsg.setConversationId(conversationId);
        userMsg.setRole("user");
        userMsg.setContent(content);
        userMsg.setTokensUsed(0);
        userMsg.setCreatedAt(OffsetDateTime.now().toString());
        messageMapper.insert(userMsg);

        // Get AI response
        String answer;
        int tokensUsed = 0;

        try {
            var response = butlerAgent.answer(content, conversationId);
            answer = response.getAnswer();
            tokensUsed = response.getTotalTokens();
        } catch (Exception e) {
            answer = "AI回答失败: " + e.getMessage();
        }

        // Save assistant message
        ChatMessage assistantMsg = new ChatMessage();
        assistantMsg.setId(UUID.randomUUID().toString().replace("-", ""));
        assistantMsg.setConversationId(conversationId);
        assistantMsg.setRole("assistant");
        assistantMsg.setContent(answer);
        assistantMsg.setTokensUsed(tokensUsed);
        assistantMsg.setCreatedAt(OffsetDateTime.now().toString());
        messageMapper.insert(assistantMsg);

        // Update conversation timestamp
        Conversation conv = conversationMapper.selectById(conversationId);
        if (conv != null) {
            conv.setUpdatedAt(OffsetDateTime.now().toString());
            conversationMapper.updateById(conv);
        }

        return assistantMsg;
    }
}
