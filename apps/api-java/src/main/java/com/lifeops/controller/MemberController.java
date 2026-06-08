package com.lifeops.controller;

import com.lifeops.entity.FamilyMember;
import com.lifeops.service.MemberService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/members")
@RequiredArgsConstructor
public class MemberController {
    private final MemberService memberService;

    @GetMapping
    public List<FamilyMember> list() {
        return memberService.listMembers();
    }

    @PostMapping
    public ResponseEntity<FamilyMember> create(@RequestBody Map<String, Object> body) {
        String name = (String) body.get("name");
        if (name == null || name.trim().isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        FamilyMember member = new FamilyMember();
        member.setName(name.trim());
        member.setRole(body.get("role") != null ? ((String) body.get("role")).trim() : "");
        member.setBirthDate((String) body.get("birth_date"));
        return ResponseEntity.status(HttpStatus.CREATED).body(memberService.createMember(member));
    }

    @PutMapping("/{id}")
    public ResponseEntity<FamilyMember> update(@PathVariable String id, @RequestBody Map<String, Object> body) {
        String name = (String) body.get("name");
        if (name == null || name.trim().isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        FamilyMember member = new FamilyMember();
        member.setName(name.trim());
        member.setRole(body.get("role") != null ? ((String) body.get("role")).trim() : "");
        member.setBirthDate((String) body.get("birth_date"));
        FamilyMember updated = memberService.updateMember(id, member);
        return ResponseEntity.ok(updated);
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void delete(@PathVariable String id) {
        memberService.deleteMember(id);
    }
}
