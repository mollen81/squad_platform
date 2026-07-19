package com.squad.clan.service;

import com.squad.clan.dto.ClanRequests;
import com.squad.clan.entity.Clan;
import com.squad.clan.enums.ClanStatus;
import com.squad.clan.repository.ClanApplicationRepository;
import com.squad.clan.repository.ClanMemberRepository;
import com.squad.clan.repository.ClanRepository;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

@Service
@Slf4j
@RequiredArgsConstructor
public class ClanManagementService {
    private final ClanRepository clanRepository;
    private final ClanMemberRepository clanMemberRepository;
    private final ClanApplicationRepository clanApplicationRepository;

    @Transactional
    public Clan createClan(ClanRequests.CreateClanDto dto) {
        if(clanMemberRepository.findByUserId(dto.getLeaderId()).isPresent()) {
            throw new IllegalStateException("You are already a member of a clan. Leave it first.");
        }

        if(clanRepository.existsByName(dto.getName())) {
            throw new IllegalArgumentException("A clan with that name already exists.");
        }

        Clan clan = Clan.builder()
                .name(dto.getName())
                .tag(dto.getTag())
                .description(dto.getDescription())
                .requirements(dto.getRequirements())
                .avatar_url(dto.getAvatarUrl())
                .clanStatus(ClanStatus.UNVERIFIED)
                .totalElo(0)
                .build();

        clan = clanRepository.save(clan);


    }
}
