package com.squad.gateway.service;

import com.squad.stats.grpc.GetPlayerStatsByInternalIdRequest;
import com.squad.stats.grpc.GetPlayerStatsResponse;
import com.squad.stats.grpc.StatsServiceGrpc;
import lombok.RequiredArgsConstructor;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;


@Service
@RequiredArgsConstructor
public class StatsGrpcClientService {

    @GrpcClient("stats-service")
    private StatsServiceGrpc.StatsServiceBlockingStub statsServiceBlockingStub;

    public GetPlayerStatsResponse getPlayerStats(String userId) {
        try {
            GetPlayerStatsByInternalIdRequest request = GetPlayerStatsByInternalIdRequest.newBuilder()
                    .setUserId(userId)
                    .build();
            return statsServiceBlockingStub.getPlayerStats(request);
        } catch (Exception e) {
            throw new RuntimeException("Couldn't get statistics: " + e.getMessage(), e);
        }
    }
}
