package com.squad.stats.service;

import com.squad.stats.dto.PlayerStatsFetchedEvent;
import com.squad.stats.dto.UserStats;
import com.squad.stats.repository.UserStatsRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.Optional;

@Service
@Slf4j
@RequiredArgsConstructor
public class EloCalculationService {
    private final UserStatsRepository userStatsRepository;

    public void processAndSaveStats(PlayerStatsFetchedEvent event) {
        log.info("Starting the ELO calculation for the user {}", event.getSteamId());

        Optional<UserStats> existingStatsOpt = userStatsRepository.findBySteamId(event.getSteamId());

        UserStats stats;
        int newElo = 0;

        if(existingStatsOpt.isEmpty()) {
            log.info("New player. Launching the primary ELO calibration algorithm.");
            newElo = calculateInitialElo(event);

            stats = UserStats.builder()
                    .userId(event.getId())
                    .steamId(event.getSteamId())
                    .build();

        }
        else {
            // after match elo update
            stats = existingStatsOpt.get()
        }

        // data update
        stats.setEloRating(newElo);
        stats.setTotalPlaytimeHours(event.getTotalPlaytimeHours());
        stats.setKills(event.getKills());
        stats.setDeaths(event.getKills());
        stats.setRevives(event.getRevives());
        stats.setFavoriteRole(event.getFavouriteRole());
        stats.setLastUpdatedAt(LocalDateTime.now());

        userStatsRepository.save(stats);
        log.info("The calculation is completed! SteamID: {}. The final ELO: {}", stats.getSteamId(), stats.getEloRating());
    }

    private int calculateInitialElo(PlayerStatsFetchedEvent event) {
        int baseElo = (int) (1000.0 + event.getTotalPlaytimeHours() * 0.1);

        double kdRatio = event.getDeaths() > 0
                ? (double) event.getKills() / event.getDeaths()
                : event.getKills();

        int kdDiff = event.getKills() - event.getDeaths();

        String role = event.getFavouriteRole() != null
                ? event.getFavouriteRole()
                : "Rifleman";

        return switch (role) {

            case "Rifleman", "Ambusher", "Raider", "Automatic Rifleman", "Machine Gunner" ->
                (int) (baseElo + kdDiff * 0.13 + 0.2 * event.getRevives());

            case "Medic" ->
                    (int) (baseElo + kdDiff * 0.035 + 0.65 * event.getRevives());

            case "Sniper", "Marksman" ->
                   (int) (kdRatio / 2 * (baseElo + kdDiff * 0.8));

            case "Grenadier", "Scout" ->
                (int) (kdRatio * (baseElo + kdDiff * 0.2) + 0.1 * event.getRevives());

            case "Squad Leader" ->
                    (int) (baseElo + event.getTotalPlaytimeHours() * 0.05 + kdDiff * 0.1);

            case "Lead Crewman", "Crewman" ->
                    (int) (baseElo + kdDiff * 0.4);

            case "Lead Pilot" ->
                    (int) (baseElo + event.getTotalPlaytimeHours() * 0.075);

            case "Light Anti-Tank" ->
                    (int) (baseElo + kdDiff * 0.1 + event.getDestroyedVehicles() * 0.55);

            case "Heavy Anti-Tank" ->
                    (int) (baseElo + kdDiff * 0.08 + event.getDestroyedVehicles() * 0.9);

            case "Combat Engineer", "Sapper", "Saboteur" ->
                    (int) (baseElo + kdDiff * 0.065 + event.getDestroyedVehicles() * 0.65);

            case "Infiltrator" ->
                    (int) (baseElo + kdDiff * 0.11 + event.getDestroyedVehicles() * 0.25);

            default -> (int) (baseElo + kdDiff * 0.1 + 0.1 * event.getRevives());
        };
    }

    // after mvp
    private int calculateMatchElo(PlayerStatsFetchedEvent event) {
        int eloDelta = 0;

        return eloDelta;
    }
}
