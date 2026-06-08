package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.LlmUsage;
import com.lifeops.mapper.LlmUsageMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.*;

@Service
@RequiredArgsConstructor
public class LlmAdminService {
    private final LlmUsageMapper llmUsageMapper;
    private final JdbcTemplate jdbcTemplate;

    public Map<String, Object> getDashboard() {
        String today = OffsetDateTime.now().toString().substring(0, 10);
        String weekStart = OffsetDateTime.now().minusDays(OffsetDateTime.now().getDayOfWeek().getValue() - 1).toString().substring(0, 10);
        String monthStart = OffsetDateTime.now().toString().substring(0, 8) + "01";

        Map<String, Object> stats = new LinkedHashMap<>();
        for (var period : new String[][]{{"today", today}, {"week", weekStart}, {"month", monthStart}, {"total", "2000-01-01"}}) {
            Map<String, Object> periodStats = new LinkedHashMap<>();
            List<Map<String, Object>> rows = jdbcTemplate.queryForList(
                "SELECT COUNT(*) as count, COALESCE(SUM(total_tokens),0) as total_tokens, COALESCE(SUM(cost_cents),0) as cost_cents FROM llm_usage WHERE created_at >= ?",
                period[1]
            );
            if (!rows.isEmpty()) {
                Map<String, Object> row = rows.get(0);
                periodStats.put("calls", row.get("count"));
                periodStats.put("total_tokens", row.get("total_tokens"));
                periodStats.put("cost_cents", row.get("cost_cents"));
            }
            stats.put(period[0], periodStats);
        }
        return stats;
    }

    public List<Map<String, Object>> getUsage(String groupBy, String startDate, String endDate) {
        String selectCols;
        String groupExpr;
        switch (groupBy != null ? groupBy : "day") {
            case "session":
                selectCols = "conversation_id, COUNT(*) as calls, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(cost_cents) as cost_cents";
                groupExpr = "conversation_id";
                break;
            case "lens":
                selectCols = "lens_id, COUNT(*) as calls, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(cost_cents) as cost_cents";
                groupExpr = "lens_id";
                break;
            default:
                selectCols = "DATE(created_at) as period, COUNT(*) as calls, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(cost_cents) as cost_cents";
                groupExpr = "DATE(created_at)";
                break;
        }

        StringBuilder query = new StringBuilder("SELECT " + selectCols + " FROM llm_usage");
        List<Object> args = new ArrayList<>();
        List<String> conditions = new ArrayList<>();
        if (startDate != null && !startDate.isEmpty()) {
            conditions.add("created_at >= ?");
            args.add(startDate);
        }
        if (endDate != null && !endDate.isEmpty()) {
            conditions.add("created_at <= ?");
            args.add(endDate + "T23:59:59");
        }
        if (!conditions.isEmpty()) {
            query.append(" WHERE ").append(String.join(" AND ", conditions));
        }
        query.append(" GROUP BY ").append(groupExpr).append(" ORDER BY 1");

        return jdbcTemplate.queryForList(query.toString(), args.toArray());
    }

    public List<Map<String, Object>> getCostTrend(int days) {
        String since = OffsetDateTime.now().minusDays(days).toString().substring(0, 10);
        return jdbcTemplate.queryForList(
            "SELECT DATE(created_at) as date, COUNT(*) as calls, SUM(total_tokens) as total_tokens, SUM(cost_cents) as cost_cents FROM llm_usage WHERE created_at >= ? GROUP BY DATE(created_at) ORDER BY date",
            since
        );
    }

    public List<Map<String, Object>> getTopSessions(int limit) {
        if (limit <= 0) limit = 10;
        return jdbcTemplate.queryForList(
            "SELECT conversation_id, COUNT(*) as calls, SUM(total_tokens) as total_tokens, SUM(cost_cents) as cost_cents, MIN(created_at) as first_at, MAX(created_at) as last_at FROM llm_usage GROUP BY conversation_id ORDER BY cost_cents DESC LIMIT ?",
            limit
        );
    }
}
