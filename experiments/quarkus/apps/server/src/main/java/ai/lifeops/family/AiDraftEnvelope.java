package ai.lifeops.family;

import io.quarkus.runtime.annotations.RegisterForReflection;

import java.util.List;

@RegisterForReflection
public record AiDraftEnvelope(List<AiDraft> drafts) {
}
