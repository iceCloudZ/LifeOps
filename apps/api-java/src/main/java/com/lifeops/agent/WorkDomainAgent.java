package com.lifeops.agent;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.KnowledgeNote;
import com.lifeops.entity.WorkProfile;
import com.lifeops.entity.WorkRecord;
import com.lifeops.entity.WorkStatus;
import com.lifeops.mapper.KnowledgeNoteMapper;
import com.lifeops.mapper.WorkProfileMapper;
import com.lifeops.mapper.WorkRecordMapper;
import com.lifeops.mapper.WorkStatusMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
@RequiredArgsConstructor
public class WorkDomainAgent implements DomainAgent {

    private final WorkProfileMapper profileMapper;
    private final WorkStatusMapper statusMapper;
    private final WorkRecordMapper recordMapper;
    private final KnowledgeNoteMapper noteMapper;

    @Override
    public String domain() { return "work"; }

    @Override
    public String systemPrompt() {
        return "你是职业规划助手。基于提供的工作数据，分析项目进度、重要节点、工作压力。用中文回答，关注即将到期的deadline，给出时间管理建议。";
    }

    @Override
    public String retrieveContext(String query) {
        List<String> parts = new ArrayList<>();

        List<WorkProfile> profiles = profileMapper.selectList(null);
        for (WorkProfile p : profiles) {
            List<String> pParts = new ArrayList<>();
            pParts.add("[" + p.getMemberId() + "]");
            if (p.getEmploymentStatus() != null && !p.getEmploymentStatus().isEmpty()) pParts.add(p.getEmploymentStatus());
            if (p.getCompany() != null && !p.getCompany().isEmpty()) pParts.add(p.getCompany());
            if (p.getPosition() != null && !p.getPosition().isEmpty()) pParts.add(p.getPosition());
            parts.add(String.join(" ", pParts));
        }

        List<WorkStatus> statuses = statusMapper.selectList(null);
        for (WorkStatus s : statuses) {
            parts.add("[" + s.getMemberId() + "] 状态概要: " + s.getSummary());
        }

        List<WorkRecord> records = recordMapper.selectList(
            new LambdaQueryWrapper<WorkRecord>().orderByDesc(WorkRecord::getCreatedAt).last("LIMIT 50")
        );
        for (WorkRecord r : records) {
            String line = "[" + r.getMemberId() + "] " + r.getType() + " " + r.getTitle() + ": " + r.getProject()
                + " (优先级:" + r.getPriority() + ", 状态:" + r.getStatus() + ")";
            if (r.getDueDate() != null && !r.getDueDate().isEmpty()) line += " 截止:" + r.getDueDate();
            parts.add(line);
        }

        addNotes(parts, "work");
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
