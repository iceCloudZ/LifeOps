package com.lifeops.agent.tool;

import com.lifeops.entity.*;
import com.lifeops.service.*;
import dev.langchain4j.agent.tool.P;
import dev.langchain4j.agent.tool.Tool;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

@Component
@RequiredArgsConstructor
public class LifeOpsReadTools {

    private final MemberService memberService;
    private final FinanceService financeService;
    private final HealthService healthService;
    private final MovementService movementService;
    private final WorkService workService;
    private final FamilyService familyService;
    private final NoteService noteService;

    @Tool("查询家庭成员列表，返回所有成员的姓名、角色、ID")
    public String listMembers() {
        List<FamilyMember> members = memberService.listMembers();
        if (members.isEmpty()) return "暂无家庭成员数据";
        return members.stream()
            .map(m -> m.getId() + " " + m.getName() + " (" + m.getRole() + ")")
            .collect(Collectors.joining("\n"));
    }

    @Tool("查询财务概览：总资产、账户明细、近期收支记录。memberId为空时查全家")
    public String listFinanceSummary(
            @P("成员ID，空字符串表示全家") String memberId) {
        StringBuilder sb = new StringBuilder();

        List<FinanceAccount> accounts = financeService.listAccounts();
        double totalAssets = 0, totalLiabilities = 0;
        for (FinanceAccount acc : accounts) {
            if (acc.getBalance() >= 0) totalAssets += acc.getBalance();
            else totalLiabilities += acc.getBalance();
        }
        sb.append("总资产: ").append(String.format("%.0f", totalAssets))
          .append(", 总负债: ").append(String.format("%.0f", totalLiabilities))
          .append(", 净资产: ").append(String.format("%.0f", totalAssets + totalLiabilities)).append("\n");

        List<FinanceRecord> records = financeService.listRecords(
            nullOrEmpty(memberId), null, null, null, null);
        sb.append("近期记录 ").append(records.size()).append(" 条:\n");
        records.stream().limit(20).forEach(r ->
            sb.append("[").append(r.getRecordDate()).append("] ")
              .append(r.getType()).append(" ").append(r.getCategory()).append(" ")
              .append(String.format("%.0f", r.getAmount())).append("元 ")
              .append(r.getNote() != null ? r.getNote() : "").append("\n")
        );

        return sb.toString();
    }

    @Tool("按条件查询财务记录。所有参数均可选")
    public String queryFinanceRecords(
            @P("成员ID") String memberId,
            @P("类型：income/expense") String type,
            @P("分类") String category,
            @P("起始日期 YYYY-MM-DD") String fromDate,
            @P("结束日期 YYYY-MM-DD") String toDate) {
        List<FinanceRecord> records = financeService.listRecords(
            nullOrEmpty(memberId), nullOrEmpty(type), nullOrEmpty(category),
            nullOrEmpty(fromDate), nullOrEmpty(toDate));
        if (records.isEmpty()) return "未找到匹配记录";
        return records.stream().limit(30).map(r ->
            "[" + r.getRecordDate() + "] " + r.getType() + " " + r.getCategory() + " "
            + String.format("%.0f", r.getAmount()) + "元 " + (r.getNote() != null ? r.getNote() : "")
        ).collect(Collectors.joining("\n"));
    }

    @Tool("查询健康概览：成员健康档案和近期健康记录。memberId为空时查全部")
    public String listHealthProfiles(@P("成员ID，空字符串表示全部") String memberId) {
        StringBuilder sb = new StringBuilder();
        List<HealthProfile> profiles = healthService.listProfiles();
        if (!profiles.isEmpty()) {
            profiles.stream()
                .filter(p -> memberId == null || memberId.isEmpty() || memberId.equals(p.getMemberId()))
                .forEach(p -> sb.append("[").append(p.getMemberId()).append("] ").append(p.getSummary()).append("\n"));
        }
        List<HealthRecord> records = healthService.listRecords(
            nullOrEmpty(memberId), null);
        if (!records.isEmpty()) {
            sb.append("近期记录:\n");
            records.stream().limit(20).forEach(r ->
                sb.append("[").append(r.getRecordDate()).append("] ").append(r.getMemberId())
                  .append(" ").append(r.getMetric()).append("=").append(r.getValue())
                  .append(r.getUnit() != null ? r.getUnit() : "")
                  .append(r.getNote() != null ? " " + r.getNote() : "").append("\n")
            );
        }
        return sb.isEmpty() ? "暂无健康数据" : sb.toString();
    }

    @Tool("查询运动记录")
    public String listMovementRecords(@P("成员ID") String memberId) {
        List<MovementRecord> records = movementService.listRecords(nullOrEmpty(memberId));
        if (records.isEmpty()) return "暂无运动记录";
        return records.stream().limit(20).map(r ->
            "[" + r.getRecordDate() + "] " + r.getMemberId()
            + " " + r.getMetric() + "=" + r.getValue()
            + (r.getNote() != null ? " " + r.getNote() : "")
        ).collect(Collectors.joining("\n"));
    }

    @Tool("查询工作状态、档案和项目记录。memberId为空时查全部")
    public String listWorkStatus(@P("成员ID，空字符串表示全部") String memberId) {
        StringBuilder sb = new StringBuilder();
        List<WorkStatus> statuses = workService.listStatuses();
        statuses.stream()
            .filter(s -> memberId == null || memberId.isEmpty() || memberId.equals(s.getMemberId()))
            .forEach(s -> sb.append("[").append(s.getMemberId()).append("] ").append(s.getSummary()).append("\n"));
        List<WorkRecord> records = workService.listRecords(nullOrEmpty(memberId), null);
        if (!records.isEmpty()) {
            sb.append("工作记录:\n");
            records.stream().limit(20).forEach(r ->
                sb.append("[").append(r.getMemberId()).append("] ").append(r.getType())
                  .append(" ").append(r.getTitle()).append(" (").append(r.getStatus()).append(")\n")
            );
        }
        return sb.isEmpty() ? "暂无工作数据" : sb.toString();
    }

    @Tool("查询家庭事务和待办。memberId为空时查全部")
    public String listFamilyRecords(
            @P("成员ID，空字符串表示全部") String memberId,
            @P("类型筛选") String type,
            @P("状态筛选") String status) {
        StringBuilder sb = new StringBuilder();
        FamilyStatus famStatus = familyService.getStatus();
        if (famStatus != null) sb.append(famStatus.getSummary()).append("\n");
        List<FamilyRecord> records = familyService.listRecords(
            nullOrEmpty(memberId), nullOrEmpty(type), nullOrEmpty(status));
        if (!records.isEmpty()) {
            records.stream().limit(20).forEach(r ->
                sb.append("[").append(r.getMemberId() != null ? r.getMemberId() : "").append("] ")
                  .append(r.getType()).append(": ").append(r.getTitle())
                  .append(" (").append(r.getStatus()).append(")")
                  .append(r.getScheduledDate() != null ? " 日期:" + r.getScheduledDate() : "")
                  .append("\n")
            );
        }
        return sb.isEmpty() ? "暂无家庭事务数据" : sb.toString();
    }

    @Tool("查询知识笔记。domain和memberId均可选")
    public String queryNotes(
            @P("领域：finance/health/movement/work/family") String domain,
            @P("成员ID") String memberId) {
        List<KnowledgeNote> notes = noteService.listNotes(nullOrEmpty(domain), nullOrEmpty(memberId));
        if (notes.isEmpty()) return "暂无笔记";
        return notes.stream().map(n ->
            (n.getMemberId() != null ? "[" + n.getMemberId() + "] " : "")
            + n.getTitle() + ": " + n.getContent()
        ).collect(Collectors.joining("\n"));
    }

    @Tool("搜索互联网获取最新信息，用于回答需要外部知识的问题")
    public String webSearch(@P("搜索关键词") String query) {
        // TODO: integrate with DuckDuckGo/Tavily/SearchService
        return "网络搜索暂未集成。关键词: " + query;
    }

    private String nullOrEmpty(String s) {
        return (s == null || s.isEmpty()) ? null : s;
    }
}
