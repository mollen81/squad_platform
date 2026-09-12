package com.squad.stats.service;

import com.squad.stats.dto.UserStats;
import com.squad.stats.grpc.Team;
import com.squad.stats.grpc.TeamMember;
import com.squad.stats.repository.UserStatsRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class EventResultsProcessingService {
    private final EventGrpcClient eventGrpcClient;
    private final EloCalculationService eloCalculationService;
    private final UserStatsRepository userStatsRepository;

    @Transactional
    public void processEventResults(String eventId) {
        List<TeamMember> participants = eventGrpcClient.getAllPlayersStats(eventId);
        List<Team> teams = eventGrpcClient.getTeamsByEventId(eventId);

        if(participants.isEmpty()) {
            return;
        }

        for(TeamMember member : participants) {
            UUID userId = UUID.fromString(member.getUserId());
            Optional<UserStats> userStats = userStatsRepository.findById(userId);

            if(userStats.isEmpty()) {
                log.warn("User with userId {} is not found.", userId);
                return;
            }

            boolean isWinner = (member.getTeamId().equals(teams.getFirst().getTeamId())) == teams.getFirst().getWinner();

            int newElo = eloCalculationService.calculateMatchElo(
                    userStats.get().getEloRating(),
                    isWinner,
                    member.getRole(),
                    (int) member.getKills(), // long values in input :(
                    (int) member.getDeaths(),
                    (int) member.getRevives(),
                    (int) member.getDestroyedVehicles()
            );

            userStats.get().setEloRating(newElo);
            userStats.get().setKills((int) (userStats.get().getKills() + member.getKills()));
            userStats.get().setDeaths((int) (userStats.get().getDeaths() + member.getDeaths()));
            userStats.get().setRevives((int) (userStats.get().getRevives() + member.getRevives()));
            userStats.get().setDestroyedVehicles((int) (userStats.get().getDestroyedVehicles() + member.getDestroyedVehicles()));
            userStats.get().setMatchesPlayed(userStats.get().getMatchesPlayed() + 1);
            userStats.get().setLastUpdatedAt(LocalDateTime.now());

            userStatsRepository.save(userStats.get());
            log.info("Event: {}. Stats for user {} is updated.", eventId, userId);
        }
        log.info("Stats for event {} is fetched.", eventId);
    }
}
