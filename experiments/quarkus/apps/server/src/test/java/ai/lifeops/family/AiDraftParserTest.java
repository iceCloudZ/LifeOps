package ai.lifeops.family;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class AiDraftParserTest {

    @Test
    void parsesValidDraftJson() {
        AiDraftParser parser = new AiDraftParser();

        AiDraftParseResult result = parser.parse("""
                {
                  "drafts": [
                    {
                      "draft_type": "task",
                      "title": "准备彩笔",
                      "description": "周五孩子需要带彩笔",
                      "confidence": 0.91
                    }
                  ]
                }
                """);

        assertTrue(result.ok());
        assertEquals(1, result.drafts().size());
        assertEquals("task", result.drafts().getFirst().draftType());
        assertEquals("准备彩笔", result.drafts().getFirst().title());
        assertEquals(0.91, result.drafts().getFirst().confidence());
    }

    @Test
    void rejectsNonJsonModelOutput() {
        AiDraftParser parser = new AiDraftParser();

        AiDraftParseResult result = parser.parse("我已经帮你整理好了：记得周五带彩笔。");

        assertTrue(result.failed());
        assertEquals("invalid_json", result.failureReason());
    }
}
