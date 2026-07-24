package com.squad.clan.service;

import com.squad.clan.client.StatsGrpcClient;
import com.squad.clan.dto.ClanRequests;
import com.squad.clan.entity.Clan;
import com.squad.clan.entity.ClanApplication;
import com.squad.clan.entity.ClanMember;
import com.squad.clan.enums.ApplicationStatus;
import com.squad.clan.enums.ClanRole;
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
    private final StatsGrpcClient statsGrpcClient;

    @Transactional
    public Clan createClan(ClanRequests.CreateClanDto dto) {
        if(clanMemberRepository.findByUserId(dto.getLeaderId()).isPresent()) {
            throw new IllegalStateException("You are already a member of a clan. Leave it first.");
        }

        if(clanRepository.existsByName(dto.getName())) {
            throw new IllegalArgumentException("A clan with that name already exists.");
        }

        int leaderElo = statsGrpcClient.getPlayerElo(dto.getLeaderId());

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

        ClanMember leader = ClanMember.builder()
                .id(dto.getLeaderId())
                .clan(clan)
                .role(ClanRole.LEADER)
                .build();

        clanMemberRepository.save(leader);
        log.info("New clan is created: [{}] {}. Basic ELO: {}. Leader: {}",
                clan.getTag(), clan.getName(), leaderElo, leader.getUserId());
        return clan;
    }


    public ClanApplication applyToClan(ClanRequests.ApplyToClanDto dto) {
        if(!clanRepository.existsById(dto.getClanId())) {
            throw new IllegalArgumentException("Clan isn't exists!");
        }

        if(clanMemberRepository.existsById(dto.getUserId())) {
            String clanName = clanRepository.getReferenceById(dto.getClanId()).getName();
            throw new IllegalStateException("You're already participates the clan: " +
                    clanName + ".You need to leave current clan.");
        }

        if(clanRepository.findById(dto.getClanId()).get().getMinElo() > statsGrpcClient.getPlayerElo(dto.getUserId())) {
            throw new IllegalStateException("Your ELO is lower, than minimal required ELO for this clan");
        }

        if(clanApplicationRepository.existsById(dto.getId()) &&
                clanApplicationRepository.existsByUserIdAndClanIDAndStatus(
                        dto.getUserId(), dto.getClanId(), ApplicationStatus.PENDING)) {
            throw new IllegalStateException("Your application is already being reviewed by the clan");
        }

        Clan clan = clanRepository.findById(
                dto.getClanId()).orElseThrow(
                        () -> new IllegalStateException("Clan is not found"));

        if(!clan.getIsRecruiting()) {
            throw new IllegalArgumentException("Clan isn't recruiting now");
        }

        ClanApplication application = ClanApplication.builder()
                .clan(clan)
                .userId(dto.getUserId())
                .socialLink(dto.getSocialLink())
                .experienceText(dto.getExperienceText())
                .status(ApplicationStatus.PENDING)
                .build();

        log.info("Player {} is sent application to clan {}",
                application.getUserId(), application.getClan().getName());
        return clanApplicationRepository.save(application);
    }
}
