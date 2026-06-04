package ai.lifeops.family;

import io.quarkus.runtime.annotations.RegisterForReflection;

@RegisterForReflection
public record WebhookInboxRequest(
        String source,
        String sender,
        String content
) {
}
