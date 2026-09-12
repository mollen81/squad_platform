package com.squad.stats.service;

import com.squad.stats.grpc.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
@Slf4j
@RequiredArgsConstructor
public class EventGrpcClient {
    @GrpcClient("event-service")
    EventServiceGrpc.EventServiceBlockingStub eventServiceBlockingStub;

    public List<TeamMember> getAllPlayersStats(String eventId) {
        List<Team> teams = getTeamsByEventId(eventId);

        List<TeamMember> stats = new ArrayList<>();

        for(Team team : teams) {
            List<TeamMember> teamStats = getTeamsStats(team.getTeamId());
            stats.addAll(teamStats);
        }

        return stats;
    }

    public List<Team> getTeamsByEventId(String eventId) {
        try {
            GetTeamsByEventIDRequest request = GetTeamsByEventIDRequest.newBuilder()
                    .setEventId(eventId)
                    .build();

            GetTeamsByEventIDResponse response = eventServiceBlockingStub.getTeamsByEventID(request);
            return response.getTeamsList();
        }
        catch (RuntimeException e) {
            log.error("Error while GetTeamsByEventID remote gRPC call for event {}: {}", eventId, e.getMessage(), e);
            throw new RuntimeException("Can't get event teams", e);
        }
    }

    private List<TeamMember> getTeamsStats(String team_id) {
        GetTeamStatsRequest request = GetTeamStatsRequest.newBuilder()
                .setTeamId(team_id)
                .build();

        GetTeamStatsResponse response = eventServiceBlockingStub.getTeamStats(request);

        return response.getStatsList();
    }
}
