package com.lifeops.controller;

import com.lifeops.entity.KnowledgeNote;
import com.lifeops.service.NoteService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/notes")
@RequiredArgsConstructor
public class NoteController {
    private final NoteService noteService;

    @GetMapping
    public List<KnowledgeNote> list(
        @RequestParam(required = false) String domain,
        @RequestParam(required = false) String member_id) {
        return noteService.listNotes(domain, member_id);
    }

    @PostMapping
    public ResponseEntity<KnowledgeNote> create(@RequestBody Map<String, Object> body) {
        String title = (String) body.get("title");
        if (title == null || title.trim().isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        KnowledgeNote note = new KnowledgeNote();
        note.setDomain((String) body.get("domain"));
        note.setMemberId((String) body.get("member_id"));
        note.setTitle(title.trim());
        note.setContent((String) body.get("content"));
        note.setTags((String) body.get("tags"));
        return ResponseEntity.status(HttpStatus.CREATED).body(noteService.createNote(note));
    }

    @PutMapping("/{id}")
    public ResponseEntity<KnowledgeNote> update(@PathVariable String id, @RequestBody Map<String, Object> body) {
        KnowledgeNote note = new KnowledgeNote();
        note.setDomain((String) body.get("domain"));
        note.setMemberId((String) body.get("member_id"));
        note.setTitle(body.get("title") != null ? ((String) body.get("title")).trim() : "");
        note.setContent((String) body.get("content"));
        note.setTags((String) body.get("tags"));
        return ResponseEntity.ok(noteService.updateNote(id, note));
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void delete(@PathVariable String id) {
        noteService.deleteNote(id);
    }
}
