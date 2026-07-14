package com.squad.gateway.controller;

import com.squad.gateway.record.PlayerStats;
import com.squad.gateway.service.StatsGrpcClientService;
import com.squad.stats.grpc.GetPlayerStatsResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.graphql.data.method.annotation.Argument;
import org.springframework.stereotype.Controller;

@Controller
@RequiredArgsConstructor
public class StatsController {
    private final StatsGrpcClientService statsGrpcClientService;

    public PlayerStats getPlayerStats(@Argument String userId) {
        GetPlayerStatsResponse grpcResponse = statsGrpcClientService.getPlayerStats(userId);

        return new PlayerStats(
                grpcResponse.getEloRating(),
                grpcResponse.getTotalPlaytimeHours(),
                grpcResponse.getKills(),
                grpcResponse.getVehiclesDestroyed(),
                grpcResponse.getDeaths(),
                grpcResponse.getRevives(),
                grpcResponse.getFavoriteRole()
        );
    }
}
