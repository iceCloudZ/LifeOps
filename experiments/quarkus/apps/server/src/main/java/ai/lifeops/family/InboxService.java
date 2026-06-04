package ai.lifeops.family;

import jakarta.enterprise.context.ApplicationScoped;

import java.time.Clock;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@ApplicationScoped
public class InboxService {

    private final Clock clock = Clock.systemUTC();
    private final List<InboxItem> items = new ArrayList<>();

    public synchronized InboxItem create(WebhookInboxRequest request) {
        InboxItem item = new InboxItem(
                UUID.randomUUID(),
                request.source(),
                request.sender(),
                request.content(),
                "new",
                clock.instant()
        );
        items.add(item);
        return item;
    }
}
