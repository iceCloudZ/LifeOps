package com.lifeops.agent;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.FinanceAccount;
import com.lifeops.entity.FinanceRecord;
import com.lifeops.entity.KnowledgeNote;
import com.lifeops.mapper.FinanceAccountMapper;
import com.lifeops.mapper.FinanceRecordMapper;
import com.lifeops.mapper.KnowledgeNoteMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
@RequiredArgsConstructor
public class FinanceDomainAgent implements DomainAgent {

    private final FinanceAccountMapper accountMapper;
    private final FinanceRecordMapper recordMapper;
    private final KnowledgeNoteMapper noteMapper;

    @Override
    public String domain() { return "finance"; }

    @Override
    public String systemPrompt() {
        return "你是家庭财务顾问。基于提供的家庭财务数据，分析收支情况、资产负债、给出理财建议。用中文回答，数字要准确，指出重要趋势。";
    }

    @Override
    public String retrieveContext(String query) {
        List<String> parts = new ArrayList<>();

        List<FinanceAccount> accounts = accountMapper.selectList(null);
        double totalAssets = 0, totalLiabilities = 0;
        List<String> accParts = new ArrayList<>();
        for (FinanceAccount acc : accounts) {
            if (acc.getBalance() >= 0) totalAssets += acc.getBalance();
            else totalLiabilities += acc.getBalance();
            accParts.add(acc.getName() + ": " + String.format("%.0f", acc.getBalance()) + "元");
        }
        parts.add("总资产: " + String.format("%.0f", totalAssets) + ", 总负债: " + String.format("%.0f", totalLiabilities)
            + ", 净资产: " + String.format("%.0f", totalAssets + totalLiabilities));
        parts.add("账户明细: " + String.join("; ", accParts));

        List<FinanceRecord> records = recordMapper.selectList(
            new LambdaQueryWrapper<FinanceRecord>().orderByDesc(FinanceRecord::getRecordDate).last("LIMIT 50")
        );
        for (FinanceRecord r : records) {
            parts.add("[" + r.getRecordDate() + "] " + r.getType() + " " + r.getCategory() + " "
                + String.format("%.0f", r.getAmount()) + "元 (" + r.getCurrency() + ") " + r.getNote());
        }

        addNotes(parts, "finance");
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
