package com.lifeops.agent;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.KnowledgeNote;
import com.lifeops.entity.MovementRecord;
import com.lifeops.mapper.KnowledgeNoteMapper;
import com.lifeops.mapper.MovementRecordMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
@RequiredArgsConstructor
public class MovementDomainAgent implements io.github.icecloudz.c2fe4j.agent.DomainAgent {

    private final MovementRecordMapper recordMapper;
    private final KnowledgeNoteMapper noteMapper;

    @Override
    public String domain() { return "movement"; }

    @Override
    public String systemPrompt() {
        return "你是运动指导助手。基于提供的运动数据，分析运动习惯、体能变化。用中文回答，给出保守渐进的运动建议。关注安全和持续性。";
    }

    @Override
    public String retrieveContext(String query) {
        List<String> parts = new ArrayList<>();

        List<MovementRecord> records = recordMapper.selectList(
            new LambdaQueryWrapper<MovementRecord>().orderByDesc(MovementRecord::getRecordDate).last("LIMIT 20")
        );
        parts.add("运动记录: " + records.size() + " 条");
        for (MovementRecord r : records) {
            StringBuilder sb = new StringBuilder("[" + r.getRecordDate() + "] " + r.getMemberId());
            if (r.getMetric() != null && r.getValue() != null) {
                sb.append(" ").append(r.getMetric()).append("=").append(r.getValue());
            }
            if (r.getNote() != null && !r.getNote().isEmpty()) sb.append(" ").append(r.getNote());
            parts.add(sb.toString());
        }

        addNotes(parts, "movement");
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
