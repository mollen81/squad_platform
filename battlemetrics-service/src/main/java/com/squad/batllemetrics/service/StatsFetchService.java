package com.squad.batllemetrics.service;

import com.squad.batllemetrics.dto.PlayerStatsFetchedEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import java.util.Random;
import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class StatsFetchService {
    private BattleMetricsClient battleMetricsClient;
    private final KafkaTemplate<String, Object> kafkaTemplate;

    private static final String TOPIC_STATS_FETCHED = "bm.status.fetched";
    private final Random random = new Random();


    public void fetchAndSendStatus(UUID userId, String steamId) {
        try {
            String bmPlayerId = battleMetricsClient.fetchPlayerIdBySteamId(steamId);
            log.info("BattleMetrics profile is found: {}", bmPlayerId);

            PlayerStatsFetchedEvent statsFetchedEvent = PlayerStatsFetchedEvent.builder()
                    .id(userId)
                    .steamId(steamId)
                    .totalPlaytimeHours(50 + random.nextInt(4000))
                    .kills(100 + random.nextInt(5000))
                    .deaths(100 + random.nextInt(4000))
                    .revives(30 + random.nextInt(1500))
                    .favouriteRole(getRandomRole())
                    .timestamp(System.currentTimeMillis())
                    .build();

            kafkaTemplate.send(TOPIC_STATS_FETCHED, userId.toString(), statsFetchedEvent);
            log.info("Statistics successfully fetched and sent to topic: {} for user with SteamId: {}", TOPIC_STATS_FETCHED, steamId);
        }
        catch (Exception e) {
            log.error("Error during statistics fetching for {}: {}", steamId, e.getMessage());
        }
    }


    private String getRandomRole() {
        String[] roles = {"Rifleman", "Medic", "Squad Leader", "Sniper", "Anti-Tank", "Pilot"};
        return roles[random.nextInt(roles.length)];
    }
}
