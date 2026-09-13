package com.squad.stats.kafka;

import com.squad.stats.dto.PlayerStatsFetchedEvent;
import com.squad.stats.service.EloCalculationService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Service;

@Service
@Slf4j
@RequiredArgsConstructor
public class StatsFetchedEvent {

    private final EloCalculationService eloCalculationService;

    private static final String TOPIC_STATS_FETCHED = "bm.stats.fetched";

    @KafkaListener(topics = "bm.stats.fetched", groupId = "stats-workers")
    public void onStatsFetched(PlayerStatsFetchedEvent event) {
        log.info("[StatsService] fetched rough data from BattleMetrics for SteamId: {}",
                event.getSteamId());
        eloCalculationService.processAndSaveStats(event);
    }
}
