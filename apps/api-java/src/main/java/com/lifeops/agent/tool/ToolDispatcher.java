package com.lifeops.agent.tool;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import dev.langchain4j.agent.tool.ToolSpecification;
import dev.langchain4j.agent.tool.ToolSpecifications;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.lang.reflect.Method;
import java.lang.reflect.Parameter;
import java.util.*;
import java.util.stream.Stream;

@Slf4j
@Component
public class ToolDispatcher {

    private final List<ToolSpecification> toolSpecs;
    private final Map<String, Method> methodMap;
    private final Map<String, Object> toolInstances;
    private final ObjectMapper objectMapper;

    public ToolDispatcher(LifeOpsReadTools readTools, LifeOpsWriteTools writeTools, ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
        this.toolInstances = new HashMap<>();
        this.methodMap = new LinkedHashMap<>();

        List<Object> allTools = List.of(readTools, writeTools);
        for (Object tool : allTools) {
            for (Method method : tool.getClass().getDeclaredMethods()) {
                if (method.isAnnotationPresent(dev.langchain4j.agent.tool.Tool.class)) {
                    methodMap.put(method.getName(), method);
                    toolInstances.put(method.getName(), tool);
                }
            }
        }

        this.toolSpecs = Stream.concat(
            ToolSpecifications.toolSpecificationsFrom(readTools).stream(),
            ToolSpecifications.toolSpecificationsFrom(writeTools).stream()
        ).toList();

        log.info("Registered {} tools: {}", toolSpecs.size(),
            toolSpecs.stream().map(ToolSpecification::name).toList());
    }

    public List<ToolSpecification> getToolSpecifications() {
        return toolSpecs;
    }

    public String execute(String toolName, String argumentsJson) {
        Method method = methodMap.get(toolName);
        Object instance = toolInstances.get(toolName);
        if (method == null || instance == null) {
            return "{\"error\":\"Unknown tool: " + toolName + "\"}";
        }
        try {
            JsonNode args = objectMapper.readTree(argumentsJson);
            Parameter[] params = method.getParameters();
            Object[] callArgs = new Object[params.length];
            for (int i = 0; i < params.length; i++) {
                String paramName = params[i].getName();
                Class<?> paramType = params[i].getType();
                JsonNode argNode = args.get(paramName);
                if (argNode == null || argNode.isNull()) {
                    callArgs[i] = getDefaultValue(paramType);
                } else if (paramType == String.class) {
                    callArgs[i] = argNode.asText();
                } else if (paramType == double.class || paramType == Double.class) {
                    callArgs[i] = argNode.asDouble();
                } else if (paramType == int.class || paramType == Integer.class) {
                    callArgs[i] = argNode.asInt();
                } else if (paramType == boolean.class || paramType == Boolean.class) {
                    callArgs[i] = argNode.asBoolean();
                } else {
                    callArgs[i] = objectMapper.treeToValue(argNode, paramType);
                }
            }
            Object result = method.invoke(instance, callArgs);
            return result != null ? result.toString() : "null";
        } catch (Exception e) {
            log.error("Tool execution failed: {} with args {}", toolName, argumentsJson, e);
            return "{\"error\":\"" + e.getMessage() + "\"}";
        }
    }

    private Object getDefaultValue(Class<?> type) {
        if (type == String.class) return "";
        if (type == double.class || type == Double.class) return 0.0;
        if (type == int.class || type == Integer.class) return 0;
        if (type == boolean.class || type == Boolean.class) return false;
        return null;
    }
}
