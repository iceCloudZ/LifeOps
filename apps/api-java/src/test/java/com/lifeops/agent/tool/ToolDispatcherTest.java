package com.lifeops.agent.tool;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.lifeops.entity.FamilyMember;
import com.lifeops.service.*;
import dev.langchain4j.agent.tool.ToolSpecification;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.stream.Collectors;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ToolDispatcherTest {

    @Mock MemberService memberService;
    @Mock FinanceService financeService;
    @Mock HealthService healthService;
    @Mock MovementService movementService;
    @Mock WorkService workService;
    @Mock FamilyService familyService;
    @Mock NoteService noteService;

    private ObjectMapper objectMapper;
    private ToolDispatcher toolDispatcher;

    @BeforeEach
    void setUp() {
        objectMapper = new ObjectMapper();
        LifeOpsReadTools readTools = new LifeOpsReadTools(
            memberService, financeService, healthService,
            movementService, workService, familyService, noteService);
        LifeOpsWriteTools writeTools = new LifeOpsWriteTools(objectMapper);
        toolDispatcher = new ToolDispatcher(readTools, writeTools, objectMapper);
    }

    @Test
    void getToolSpecifications_returnsNonEmptyList() {
        List<ToolSpecification> specs = toolDispatcher.getToolSpecifications();

        assertFalse(specs.isEmpty(), "Tool specifications should not be empty");
    }

    @Test
    void execute_unknownTool_returnsError() {
        String result = toolDispatcher.execute("nonExistentTool", "{}");

        assertTrue(result.contains("error"), "Result should contain 'error'");
        assertTrue(result.contains("Unknown tool"), "Result should contain 'Unknown tool'");
        assertTrue(result.contains("nonExistentTool"), "Result should mention the tool name");
    }

    @Test
    void execute_listMembers_returnsMemberList() {
        FamilyMember m1 = new FamilyMember();
        m1.setId("abc123");
        m1.setName("Alice");
        m1.setRole("parent");

        FamilyMember m2 = new FamilyMember();
        m2.setId("def456");
        m2.setName("Bob");
        m2.setRole("child");

        when(memberService.listMembers()).thenReturn(List.of(m1, m2));

        String result = toolDispatcher.execute("listMembers", "{}");

        assertTrue(result.contains("Alice"), "Result should contain Alice");
        assertTrue(result.contains("Bob"), "Result should contain Bob");
        assertTrue(result.contains("parent"), "Result should contain parent role");
        assertTrue(result.contains("child"), "Result should contain child role");
    }

    @Test
    void execute_createFinanceRecord_returnsPendingConfirmation() throws Exception {
        String argsJson = """
            {"memberId":"abc123","amount":100.0,"type":"expense","category":"food","note":"lunch"}
            """;

        String result = toolDispatcher.execute("createFinanceRecord", argsJson);

        JsonNode json = objectMapper.readTree(result);
        assertEquals("pending_confirmation", json.get("status").asText());
        assertEquals("createFinanceRecord", json.get("action").asText());
        assertTrue(json.has("summary"), "Result should have 'summary' field");
        assertTrue(json.has("data"), "Result should have 'data' field");
        assertEquals("abc123", json.get("data").get("memberId").asText());
    }

    @Test
    void toolSpecs_containExpectedTools() {
        List<ToolSpecification> specs = toolDispatcher.getToolSpecifications();
        List<String> toolNames = specs.stream()
            .map(ToolSpecification::name)
            .collect(Collectors.toList());

        assertTrue(toolNames.contains("listMembers"),
            "Tools should contain listMembers, got: " + toolNames);
        assertTrue(toolNames.contains("listFinanceSummary"),
            "Tools should contain listFinanceSummary, got: " + toolNames);
        assertTrue(toolNames.contains("createFinanceRecord"),
            "Tools should contain createFinanceRecord, got: " + toolNames);
        assertTrue(toolNames.contains("webSearch"),
            "Tools should contain webSearch, got: " + toolNames);
    }
}
