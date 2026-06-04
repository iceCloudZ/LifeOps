package ai.lifeops.family;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.util.List;

public class AiDraftParser {

    private final ObjectMapper objectMapper = new ObjectMapper();

    public AiDraftParseResult parse(String modelOutput) {
        try {
            AiDraftEnvelope envelope = objectMapper.readValue(modelOutput, AiDraftEnvelope.class);
            if (envelope.drafts() == null || envelope.drafts().isEmpty()) {
                return AiDraftParseResult.failed("missing_drafts");
            }
            if (envelope.drafts().stream().anyMatch(this::invalidDraft)) {
                return AiDraftParseResult.failed("invalid_draft");
            }
            return AiDraftParseResult.ok(List.copyOf(envelope.drafts()));
        } catch (JsonProcessingException e) {
            return AiDraftParseResult.failed("invalid_json");
        }
    }

    private boolean invalidDraft(AiDraft draft) {
        return isBlank(draft.draftType()) || isBlank(draft.title());
    }

    private boolean isBlank(String value) {
        return value == null || value.isBlank();
    }
}
