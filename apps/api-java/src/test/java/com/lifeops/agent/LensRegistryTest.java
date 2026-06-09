package com.lifeops.agent;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class LensRegistryTest {

    private LensRegistry registry;

    @BeforeEach
    void setUp() throws Exception {
        registry = new LensRegistry(new ObjectMapper());
        registry.init();
    }

    @Test
    void getLensIndexText_containsAllDomains() {
        String index = registry.getLensIndexText();

        assertNotNull(index);
        assertTrue(index.contains("finance"), "index should contain finance domain");
        assertTrue(index.contains("health"), "index should contain health domain");
        assertTrue(index.contains("movement"), "index should contain movement domain");
        assertTrue(index.contains("family"), "index should contain family domain");
    }

    @Test
    void getLensIndexText_containsLensNames() {
        String index = registry.getLensIndexText();

        assertNotNull(index);
        assertTrue(index.contains("bogleheads-style"), "index should contain bogleheads-style lens");
        assertTrue(index.contains("zone2-longevity"), "index should contain zone2-longevity lens");
    }

    @Test
    void getLensContent_returnsMarkdownContent() {
        String content = registry.getLensContent("bogleheads-style");

        assertNotNull(content, "lens content should not be null");
        assertTrue(content.contains("Reasoning Flow"),
                "lens content should contain 'Reasoning Flow' section");
    }

    @Test
    void getLensContent_cachesResult() {
        String first = registry.getLensContent("bogleheads-style");
        String second = registry.getLensContent("bogleheads-style");

        assertNotNull(first);
        assertSame(first, second, "second call should return the same cached instance");
    }

    @Test
    void registerAndGetDomainPrompt() {
        String domain = "finance";
        String prompt = "You are a finance domain assistant.";

        registry.registerDomainPrompt(domain, prompt);

        assertEquals(prompt, registry.getDomainPrompt(domain),
                "getDomainPrompt should return the registered prompt");
    }

    @Test
    void getLensContent_unknownLens_returnsNull() {
        String content = registry.getLensContent("nonexistent-lens");

        assertNull(content, "content for unknown lens should be null");
    }
}
