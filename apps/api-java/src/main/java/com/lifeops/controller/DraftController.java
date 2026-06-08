package com.lifeops.controller;

import com.lifeops.entity.Draft;
import com.lifeops.service.DraftService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/drafts")
@RequiredArgsConstructor
public class DraftController {
    private final DraftService draftService;

    @GetMapping
    public List<Draft> list(@RequestParam(required = false, defaultValue = "pending") String status) {
        return draftService.listDrafts(status);
    }

    @GetMapping("/{id}")
    public ResponseEntity<Draft> get(@PathVariable String id) {
        Draft draft = draftService.getDraft(id);
        if (draft == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(draft);
    }
}
