package com.squad.stats.kafka;

import com.squad.stats.client.SteamApiClient;
import com.squad.stats.dto.UserRegisteredEvent;
import com.squad.stats.dto.UserStats;
import com.squad.stats.repository.UserStatsRepository;
import com.squad.stats.service.EloCalculationService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.time.LocalDateTime;

@Service
@Slf4j
@RequiredArgsConstructor
public class UserRegistrationStatsListener {

    private final SteamApiClient steamApiClient;
    private final UserStatsRepository statsRepository;
    private final EloCalculationService eloCalculationService;

    @Transactional
    @KafkaListener(topics = "user.registered", groupId = "stats-service-group")
    public void onUserRegistered(UserRegisteredEvent event) {
        if(statsRepository.existsById(event.getUserId())) {
            log.info("Stats for user {} is already exists. Skipping.", event.getUserId());
            return;
        }

        int hours = steamApiClient.fetchSquadPlaytimeHours(event.getSteamId());
        int startingElo = eloCalculationService.calculateInitialElo(hours);

        UserStats userStats = UserStats.builder()
                .userId(event.getUserId())
                .steamId(event.getSteamId())
                .totalPlaytimeHours(hours)
                .kills(0)
                .deaths(0)
                .destroyedVehicles(0)
                .revives(0)
                .eloRating(startingElo)
                .favoriteRole("Rifleman")
                .matchesPlayed(0)
                .lastUpdatedAt(LocalDateTime.now())
                .build();
    }
}
