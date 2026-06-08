package com.lifeops.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.lifeops.entity.FamilyMember;
import com.lifeops.mapper.FamilyMemberMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class MemberService {
    private final FamilyMemberMapper memberMapper;

    public List<FamilyMember> listMembers() {
        return memberMapper.selectList(
            new LambdaQueryWrapper<FamilyMember>().orderByDesc(FamilyMember::getCreatedAt)
        );
    }

    public FamilyMember getMember(String id) {
        return memberMapper.selectById(id);
    }

    public FamilyMember createMember(FamilyMember member) {
        String now = OffsetDateTime.now().toString();
        member.setId(UUID.randomUUID().toString().replace("-", ""));
        member.setCreatedAt(now);
        member.setUpdatedAt(now);
        memberMapper.insert(member);
        return member;
    }

    public FamilyMember updateMember(String id, FamilyMember member) {
        member.setId(id);
        member.setUpdatedAt(OffsetDateTime.now().toString());
        memberMapper.updateById(member);
        return memberMapper.selectById(id);
    }

    public void deleteMember(String id) {
        memberMapper.deleteById(id);
    }
}
