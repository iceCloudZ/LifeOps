package ai.lifeops.family;

import io.quarkus.runtime.annotations.RegisterForReflection;

import java.time.Instant;
import java.util.UUID;

@RegisterForReflection
public record InboxItem(
        UUID id,
        String source,
        String sender,
        String content,
        String status,
        Instant createdAt
) {
}
