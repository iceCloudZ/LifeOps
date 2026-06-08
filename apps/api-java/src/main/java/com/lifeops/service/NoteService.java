package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.KnowledgeNote;
import com.lifeops.mapper.KnowledgeNoteMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class NoteService {
    private final KnowledgeNoteMapper noteMapper;

    public List<KnowledgeNote> listNotes(String domain, String memberId) {
        LambdaQueryWrapper<KnowledgeNote> wrapper = new LambdaQueryWrapper<>();
        if (domain != null && !domain.isEmpty()) {
            wrapper.eq(KnowledgeNote::getDomain, domain);
        }
        if (memberId != null && !memberId.isEmpty()) {
            wrapper.eq(KnowledgeNote::getMemberId, memberId);
        }
        wrapper.orderByDesc(KnowledgeNote::getCreatedAt);
        return noteMapper.selectList(wrapper);
    }

    public KnowledgeNote getNote(String id) {
        return noteMapper.selectById(id);
    }

    public KnowledgeNote createNote(KnowledgeNote note) {
        String now = OffsetDateTime.now().toString();
        note.setId(UUID.randomUUID().toString().replace("-", ""));
        note.setCreatedAt(now);
        note.setUpdatedAt(now);
        noteMapper.insert(note);
        return note;
    }

    public KnowledgeNote updateNote(String id, KnowledgeNote note) {
        note.setId(id);
        note.setUpdatedAt(OffsetDateTime.now().toString());
        noteMapper.updateById(note);
        return noteMapper.selectById(id);
    }

    public void deleteNote(String id) {
        noteMapper.deleteById(id);
    }
}
