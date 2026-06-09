package com.lifeops.agent.tool;

import com.fasterxml.jackson.databind.ObjectMapper;
import dev.langchain4j.agent.tool.P;
import dev.langchain4j.agent.tool.Tool;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.LinkedHashMap;
import java.util.Map;

@Component
@RequiredArgsConstructor
public class LifeOpsWriteTools {

    private final ObjectMapper objectMapper;

    @Tool("记录一笔财务收支。返回待确认的操作描述")
    public String createFinanceRecord(
            @P("成员ID") String memberId,
            @P("金额") double amount,
            @P("类型：income或expense") String type,
            @P("分类，如：食品、交通、工资") String category,
            @P("备注") String note) {
        return pending("createFinanceRecord",
            "给 " + memberId + " 记账：" + category + " " + String.format("%.0f", amount) + "元",
            Map.of("memberId", memberId, "amount", amount, "type", type,
                   "category", category, "note", note != null ? note : ""));
    }

    @Tool("记录一条健康数据。返回待确认的操作描述")
    public String createHealthRecord(
            @P("成员ID") String memberId,
            @P("指标名，如：血压、体重、心率") String metric,
            @P("值") String value,
            @P("单位，如：mmHg、kg、bpm") String unit,
            @P("备注") String note) {
        return pending("createHealthRecord",
            "记录 " + memberId + " 的" + metric + "：" + value + (unit != null ? unit : ""),
            Map.of("memberId", memberId, "metric", metric, "value", value,
                   "unit", unit != null ? unit : "", "note", note != null ? note : ""));
    }

    @Tool("记录一条运动数据。返回待确认的操作描述")
    public String createMovementRecord(
            @P("成员ID") String memberId,
            @P("运动项目，如：跑步、游泳") String metric,
            @P("值，如：30表示30分钟") String value,
            @P("备注") String note) {
        return pending("createMovementRecord",
            "记录 " + memberId + " 的运动：" + metric + " " + value,
            Map.of("memberId", memberId, "metric", metric, "value", value,
                   "note", note != null ? note : ""));
    }

    @Tool("添加一条工作任务。返回待确认的操作描述")
    public String createWorkRecord(
            @P("成员ID") String memberId,
            @P("任务标题") String title,
            @P("项目名") String project,
            @P("优先级：high/medium/low") String priority) {
        return pending("createWorkRecord",
            "给 " + memberId + " 添加工作任务：" + title,
            Map.of("memberId", memberId, "title", title,
                   "project", project != null ? project : "",
                   "priority", priority != null ? priority : "medium"));
    }

    @Tool("添加一条家庭事务。返回待确认的操作描述")
    public String createFamilyRecord(
            @P("成员ID") String memberId,
            @P("类型：todo/event/note") String type,
            @P("标题") String title,
            @P("计划日期 YYYY-MM-DD") String scheduledDate) {
        return pending("createFamilyRecord",
            "添加" + type + "：" + title,
            Map.of("memberId", memberId, "type", type, "title", title,
                   "scheduledDate", scheduledDate != null ? scheduledDate : ""));
    }

    @Tool("保存一条知识笔记。返回待确认的操作描述")
    public String createNote(
            @P("领域：finance/health/movement/work/family") String domain,
            @P("标题") String title,
            @P("内容") String content,
            @P("成员ID") String memberId) {
        return pending("createNote",
            "保存笔记：" + title,
            Map.of("domain", domain, "title", title, "content", content,
                   "memberId", memberId != null ? memberId : ""));
    }

    private String pending(String action, String summary, Map<String, Object> data) {
        try {
            Map<String, Object> result = new LinkedHashMap<>();
            result.put("status", "pending_confirmation");
            result.put("action", action);
            result.put("summary", summary);
            result.put("data", data);
            return objectMapper.writeValueAsString(result);
        } catch (Exception e) {
            return "{\"status\":\"error\",\"message\":\"" + e.getMessage() + "\"}";
        }
    }
}
