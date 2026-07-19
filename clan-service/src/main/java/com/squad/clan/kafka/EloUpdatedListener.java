package com.squad.clan.kafka;

import com.squad.clan.dto.UserEloUpdatedEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Service;

@Service
@Slf4j
@RequiredArgsConstructor
public class EloUpdatedListener {

    @KafkaListener(topics = "user.elo.updated", groupId = "clan-workers")
    public void onUserEloUpdated(UserEloUpdatedEvent event) {
        log.info("[ClanService] Elo update received for player: {}. Old ELO: {}, new ELO: {}",
                event.getUserId(), event.getOldElo(), event.getNewElo());

    }
}
