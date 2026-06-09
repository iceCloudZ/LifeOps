package com.lifeops.agent;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.HealthProfile;
import com.lifeops.entity.HealthRecord;
import com.lifeops.entity.KnowledgeNote;
import com.lifeops.mapper.HealthProfileMapper;
import com.lifeops.mapper.HealthRecordMapper;
import com.lifeops.mapper.KnowledgeNoteMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
@RequiredArgsConstructor
public class HealthDomainAgent implements DomainAgent {

    private final HealthProfileMapper profileMapper;
    private final HealthRecordMapper recordMapper;
    private final KnowledgeNoteMapper noteMapper;

    @Override
    public String domain() { return "health"; }

    @Override
    public String systemPrompt() {
        return "你是家庭健康助手。基于提供的家庭健康数据，分析健康状况、用药情况、运动习惯。用中文回答，关注异常指标，给出温和的健康建议。不要给出医疗诊断。";
    }

    @Override
    public String retrieveContext(String query) {
        List<String> parts = new ArrayList<>();

        List<HealthProfile> profiles = profileMapper.selectList(null);
        for (HealthProfile p : profiles) {
            parts.add("[" + p.getMemberId() + "] " + p.getSummary());
        }

        List<HealthRecord> records = recordMapper.selectList(
            new LambdaQueryWrapper<HealthRecord>().orderByDesc(HealthRecord::getRecordDate).last("LIMIT 50")
        );
        for (HealthRecord r : records) {
            StringBuilder sb = new StringBuilder("[" + r.getRecordDate() + "] " + r.getMemberId());
            if (r.getMetric() != null && r.getValue() != null) {
                sb.append(" ").append(r.getMetric()).append("=").append(r.getValue());
                if (r.getUnit() != null) sb.append(r.getUnit());
            }
            if (r.getNote() != null && !r.getNote().isEmpty()) sb.append(" ").append(r.getNote());
            parts.add(sb.toString());
        }

        addNotes(parts, "health");
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
