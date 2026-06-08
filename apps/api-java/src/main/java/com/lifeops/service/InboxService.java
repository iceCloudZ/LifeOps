package com.lifeops.service;

import com.lifeops.entity.InboxItem;
import com.lifeops.mapper.InboxItemMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class InboxService {
    private final InboxItemMapper inboxItemMapper;

    public InboxItem createInboxItem(InboxItem item) {
        String now = OffsetDateTime.now().toString();
        item.setId(UUID.randomUUID().toString().replace("-", ""));
        item.setCreatedAt(now);
        item.setUpdatedAt(now);
        item.setStatus("new");
        inboxItemMapper.insert(item);
        return item;
    }
}
