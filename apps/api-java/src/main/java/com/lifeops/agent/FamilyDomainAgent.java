package com.lifeops.agent;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.FamilyRecord;
import com.lifeops.entity.FamilyStatus;
import com.lifeops.entity.KnowledgeNote;
import com.lifeops.mapper.FamilyRecordMapper;
import com.lifeops.mapper.FamilyStatusMapper;
import com.lifeops.mapper.KnowledgeNoteMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
@RequiredArgsConstructor
public class FamilyDomainAgent implements io.github.icecloudz.c2fe4j.agent.DomainAgent {

    private final FamilyStatusMapper statusMapper;
    private final FamilyRecordMapper recordMapper;
    private final KnowledgeNoteMapper noteMapper;

    @Override
    public String domain() { return "family"; }

    @Override
    public String systemPrompt() {
        return "你是家庭事务管家。基于提供的家庭事务数据，分析日程安排、家务分工、育儿安排。用中文回答，关注待办事项，提醒重要日程。";
    }

    @Override
    public String retrieveContext(String query) {
        List<String> parts = new ArrayList<>();

        FamilyStatus status = statusMapper.selectOne(new LambdaQueryWrapper<FamilyStatus>().last("LIMIT 1"));
        if (status != null) {
            parts.add(status.getSummary());
        }

        List<FamilyRecord> records = recordMapper.selectList(
            new LambdaQueryWrapper<FamilyRecord>().orderByDesc(FamilyRecord::getCreatedAt).last("LIMIT 50")
        );
        for (FamilyRecord r : records) {
            String line = "[" + (r.getMemberId() != null ? r.getMemberId() : "") + "] " + r.getType() + ": " + r.getTitle() + " (状态:" + r.getStatus() + ")";
            if (r.getScheduledDate() != null && !r.getScheduledDate().isEmpty()) line += " 日期:" + r.getScheduledDate();
            parts.add(line);
        }

        addNotes(parts, "family");
        return parts.isEmpty() ? "暂无相关数据。" : String.join("\n", parts);
    }

    private void addNotes(List<String> parts, String domain) {
        List<KnowledgeNote> notes = noteMapper.selectList(
            new LambdaQueryWrapper<KnowledgeNote>().eq(KnowledgeNote::getDomain, domain)
        );
        for (KnowledgeNote n : notes) {
            parts.add((n.getMemberId() != null ? "[" + n.getMemberId() + "] " : "") + n.getTitle() + ": " + n.getContent());
        }
    }
}
