package com.squad.clan.client;

import com.squad.grpc.stats.GetPlayerStatsByInternalIdRequest;
import com.squad.grpc.stats.GetPlayerStatsResponse;
import com.squad.grpc.stats.StatsServiceGrpc;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class StatsGrpcClient {

    @GrpcClient("stats-service")
    private StatsServiceGrpc.StatsServiceBlockingStub statsServiceStub;

    public int getPlayerElo(UUID userId) {
        try {
            GetPlayerStatsByInternalIdRequest request = GetPlayerStatsByInternalIdRequest.newBuilder()
                    .setUserId(userId.toString())
                    .build();

            GetPlayerStatsResponse response = statsServiceStub.getPlayerStats(request);
            return response.getEloRating();
        }
        catch (Exception e) {
            log.warn("Couldn't get ELO for player {}. We use the basic 1000 ELO. Reason: {}", userId, e.getMessage());
            return 1000;
        }
    }

}
