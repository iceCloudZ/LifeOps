package com.lifeops.agent;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.io.Resource;
import org.springframework.core.io.support.PathMatchingResourcePatternResolver;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Spring component that loads the skills registry and lens content at startup.
 * <p>
 * Builds a compact text index of all lenses for Phase 1 routing prompts,
 * and provides cached access to individual lens .md files.
 */
@Component
public class LensRegistry {

    private static final Logger log = LoggerFactory.getLogger(LensRegistry.class);

    private final ObjectMapper objectMapper;
    private final Map<String, String> lensContentCache = new ConcurrentHashMap<>();
    private final Map<String, String> domainPrompts = new ConcurrentHashMap<>();

    /** skill entries keyed by lens id */
    private final Map<String, SkillEntry> skills = new LinkedHashMap<>();

    private String lensIndexText;

    public LensRegistry(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    @PostConstruct
    void init() throws IOException {
        loadRegistry();
        buildIndexText();
        log.info("LensRegistry initialized: {} lenses loaded, index ~{} chars",
                skills.size(), lensIndexText.length());
    }

    // ── public API ──────────────────────────────────────────────────────

    /** Compact text index of all lenses, suitable for Phase 1 routing prompts. */
    public String getLensIndexText() {
        return lensIndexText;
    }

    /** Returns the full .md content for a lens, cached after first read. */
    public String getLensContent(String lensId) {
        return lensContentCache.computeIfAbsent(lensId, id -> {
            SkillEntry entry = skills.get(id);
            if (entry == null) return null;
            try {
                return readClasspathResource(entry.path);
            } catch (IOException e) {
                log.warn("Failed to load lens content for {}: {}", id, e.getMessage());
                return null;
            }
        });
    }

    /** Stores a domain system prompt (e.g. from DomainAgent implementations). */
    public void registerDomainPrompt(String domain, String systemPrompt) {
        domainPrompts.put(domain, systemPrompt);
    }

    /** Retrieves a previously registered domain system prompt. */
    public String getDomainPrompt(String domain) {
        return domainPrompts.get(domain);
    }

    // ── internals ───────────────────────────────────────────────────────

    private void loadRegistry() throws IOException {
        String json = readClasspathResource("skills/registry.json");
        JsonNode root = objectMapper.readTree(json);
        JsonNode skillsArray = root.get("skills");
        if (skillsArray == null) return;

        for (JsonNode skill : skillsArray) {
            String id = skill.path("id").asText();
            String domain = skill.path("domain").asText();
            String path = skill.path("path").asText();
            String description = skill.path("description").asText("");
            List<String> bestFor = toStringList(skill.path("best_for"));
            List<String> avoidIf = toStringList(skill.path("avoid_if"));
            String riskLevel = skill.path("risk_level").asText("");

            skills.put(id, new SkillEntry(id, domain, path, description, bestFor, avoidIf, riskLevel));
        }
    }

    private void buildIndexText() {
        StringBuilder sb = new StringBuilder();
        sb.append("Available lenses by domain:\n\n");

        Map<String, List<SkillEntry>> byDomain = new LinkedHashMap<>();
        for (SkillEntry e : skills.values()) {
            // Skip the entry skill (life-butler) — only index lenses
            if ("entry".equals(getSkillType(e))) continue;
            byDomain.computeIfAbsent(e.domain, k -> new ArrayList<>()).add(e);
        }

        for (Map.Entry<String, List<SkillEntry>> domainEntry : byDomain.entrySet()) {
            sb.append("[").append(domainEntry.getKey()).append("]\n");
            for (SkillEntry lens : domainEntry.getValue()) {
                sb.append("  ").append(lens.id);
                if (!lens.bestFor.isEmpty()) {
                    sb.append(" — best for: ").append(String.join(", ", lens.bestFor));
                }
                if (!lens.avoidIf.isEmpty()) {
                    sb.append("; avoid if: ").append(String.join(", ", lens.avoidIf));
                }
                sb.append("\n");
            }
        }

        lensIndexText = sb.toString();
    }

    private String getSkillType(SkillEntry e) {
        // The path convention distinguishes lenses from entry skills
        // lenses use "lenses/" prefix in path
        return e.path.startsWith("lenses/") ? "lens" : "entry";
    }

    private String readClasspathResource(String location) throws IOException {
        var resolver = new PathMatchingResourcePatternResolver();
        Resource resource = resolver.getResource("classpath:" + location);
        try (InputStream is = resource.getInputStream()) {
            return new String(is.readAllBytes(), StandardCharsets.UTF_8);
        }
    }

    private List<String> toStringList(JsonNode node) {
        List<String> result = new ArrayList<>();
        if (node.isArray()) {
            for (JsonNode item : node) {
                result.add(item.asText());
            }
        }
        return result;
    }

    // ── internal data ───────────────────────────────────────────────────

    private record SkillEntry(
            String id,
            String domain,
            String path,
            String description,
            List<String> bestFor,
            List<String> avoidIf,
            String riskLevel
    ) {}
}
