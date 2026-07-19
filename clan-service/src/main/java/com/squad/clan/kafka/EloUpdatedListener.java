package com.squad.clan.kafka;

import com.squad.clan.dto.UserEloUpdatedEvent;
import com.squad.clan.entity.Clan;
import com.squad.clan.entity.ClanMember;
import com.squad.clan.enums.ClanStatus;
import com.squad.clan.repository.ClanMemberRepository;
import com.squad.clan.repository.ClanRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
@Slf4j
@RequiredArgsConstructor
public class EloUpdatedListener {

    private ClanRepository clanRepository;
    private ClanMemberRepository clanMemberRepository;

    @KafkaListener(topics = "user.elo.updated", groupId = "clan-workers")
    public void onUserEloUpdated(UserEloUpdatedEvent event) {
        log.info("[ClanService] Elo update received for player: {}. Old ELO: {}, new ELO: {}",
                event.getUserId(), event.getOldElo(), event.getNewElo());

        Optional<ClanMember> memberOpt = clanMemberRepository.findByUserId(event.getUserId());

        if(memberOpt.isEmpty()) {
            log.info("Player: {} is not clan member. Ignored.", event.getUserId());
            return;
        }

        Clan clan = memberOpt.get().getClan();
        int delta = event.getNewElo() - event.getOldElo();

        clan.setTotalElo(clan.getTotalElo() + delta);
        log.info("Clan [{}] {}: new total_elo: {}", clan.getTag(), clan.getName(), clan.getTotalElo());

        if(clan.getClanStatus() == ClanStatus.UNVERIFIED) {
            if(clan.getTotalElo() >= 11000) {
                int memberCount = clanMemberRepository.countByClanId(clan.getId());
                if(memberCount >= 7) {
                    clan.setClanStatus(ClanStatus.OFFICIAL);
                    // Kafka notification to notification-service about new clan status to members
                    log.info("Clan [{}] {}: received OFFICIAL status!!!", clan.getClanStatus(), clan.getName());
                }
            }
        }

        clanRepository.save(clan);
    }
}
