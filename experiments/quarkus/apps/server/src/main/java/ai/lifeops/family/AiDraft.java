package ai.lifeops.family;

import com.fasterxml.jackson.annotation.JsonProperty;
import io.quarkus.runtime.annotations.RegisterForReflection;

@RegisterForReflection
public record AiDraft(
        @JsonProperty("draft_type") String draftType,
        String title,
        String description,
        double confidence
) {
}
