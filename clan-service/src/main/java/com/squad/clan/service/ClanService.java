package com.squad.clan.service;

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
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@Slf4j
@RequiredArgsConstructor
public class ClanService {
    private final ClanRepository clanRepository;
    private final ClanMemberRepository clanMemberRepository;
    private final ClanApplicationRepository clanApplicationRepository;

    @Transactional
    public Clan createClan(ClanRequests.CreateClanDto dto, int leaderElo) {
        if(clanMemberRepository.existsByUserId(dto.getLeaderId())) {
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
                .totalElo(leaderElo)
                .build();
        clan = clanRepository.save(clan);

        ClanMember leader = ClanMember.builder()
                .userId(dto.getLeaderId())
                .clan(clan)
                .role(ClanRole.LEADER)
                .build();
        leader = clanMemberRepository.save(leader);

        clan.getMembers().add(leader);

        log.info("New clan is created: [{}] {}. Basic ELO: {}. Leader: {}",
                clan.getTag(), clan.getName(), leaderElo, leader.getUserId());
        return clan;
    }


    //TODO Facade pattern for userElo fetching from gRPC stats-service
    @Transactional
    public ClanApplication applyToClan(ClanRequests.ApplyToClanDto dto, int userElo) {
        Clan clan = clanRepository.findById(dto.getClanId())
                .orElseThrow(() -> new IllegalArgumentException("Clan not found"));

        if(clanMemberRepository.existsByUserId(dto.getUserId())) {
            String clanName = clanRepository.getReferenceById(dto.getClanId()).getName();
            throw new IllegalStateException("User is already participates the clan: " +
                    clanName + ". He needs to leave current clan.");
        }

        if(clan.getMinElo() > userElo) {
            throw new IllegalStateException("User's ELO is lower, than minimal required ELO for this clan");
        }

        if(clanApplicationRepository.existsByUserIdAndClanIdAndStatus(
                        dto.getUserId(), dto.getClanId(), ApplicationStatus.PENDING)) {
            throw new IllegalStateException("User's application is already being reviewed by the clan");
        }

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

        application = clanApplicationRepository.save(application);

        log.info("User {} is sent application to clan {}",
                application.getUserId(), application.getClan().getName());
        return application;
    }


    @Transactional(readOnly = true)
    public Clan getClanWithMembers(ClanRequests.GetClanWithAllMembersDto dto) {
        Clan clan = clanRepository.findByIdWithAllMembers(dto.getId())
                .orElseThrow(() -> new IllegalArgumentException("Clan is not found"));

        log.info("Clan: {} is successfully fetched", clan.getName());
        return clan;
    }


    @Transactional
    public ClanMember processAcceptance(ClanRequests.AcceptApplicationDto dto, int applicantElo) {
        ClanApplication clanApplication = clanApplicationRepository.findById(dto.applicationId())
                .orElseThrow(() -> new IllegalArgumentException("Application not found"));

        if(clanApplication.getStatus() != ApplicationStatus.PENDING) {
            throw new IllegalStateException("Application is already processed");
        }

        Clan clan = clanRepository.findById(clanApplication.getClan().getId())
                .orElseThrow(() -> new IllegalArgumentException("Clan not found"));

        ClanMember acceptor = clanMemberRepository.findByUserId(dto.accepterId())
                .orElseThrow(() -> new IllegalArgumentException("Acceptor not found"));

        if(!acceptor.getClan().getId().equals(clan.getId()) ||
                (acceptor.getRole().equals(ClanRole.LEADER) || (acceptor.getRole().equals(ClanRole.MODERATOR)))) {
            throw new IllegalStateException(
                    "Acceptor don't have correct rights: he isn't participates traget clan or do not LEADER or MODERATOR of target clan"
            );
        }

        if(applicantElo < clan.getMinElo()) {
            throw new IllegalStateException("User's ELO is lower than minimal clan ELO");
        }

        if(clanApplicationRepository.existsById(dto.applicationId())) {
            clanApplication.setStatus(ApplicationStatus.REJECTED);
            clanApplicationRepository.save(clanApplication);
            throw new IllegalStateException("User is already sent application to another clan");
        }

        clanApplication.setStatus(ApplicationStatus.APPROVED);
        clanApplicationRepository.save(clanApplication);

        List<ClanApplication> otherPendingApplications = clanApplicationRepository.findByUserIdAndStatus(clanApplication.getUserId(), ApplicationStatus.PENDING);
        for(ClanApplication application : otherPendingApplications) {
            application.setStatus(ApplicationStatus.REJECTED);
        }
        clanApplicationRepository.saveAll(otherPendingApplications);

        ClanMember newMember = ClanMember.builder()
                .userId(clanApplication.getUserId())
                .clan(clan)
                .role(ClanRole.MEMBER)
                .build();
        newMember = clanMemberRepository.save(newMember);

        clan.setTotalElo(clan.getTotalElo() + applicantElo);
        clan.getMembers().add(newMember);

        log.info("Application: {} accepted. User: {} joined clan: {}. New total ELO: {}",
                clanApplication.getId(), newMember.getUserId(), clan.getName(), clan.getTotalElo());
        return newMember;
    }
}
