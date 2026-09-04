package com.squad.gateway.controller;

import com.squad.gateway.service.StatsGrpcClientService;
import com.squad.stats.grpc.GetPlayerStatsResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;

import java.util.Map;

@Controller
@RequiredArgsConstructor
public class StatsController {
    private final StatsGrpcClientService statsGrpcClientService;

    @GetMapping("/{userId}")
    public ResponseEntity<?> getPlayerStats(@PathVariable String userId) {
        try {
            GetPlayerStatsResponse grpcResponse = statsGrpcClientService.getPlayerStats(userId);

            StatsResponse response = new StatsResponse(
                    grpcResponse.getEloRating(),
                    grpcResponse.getKills(),
                    grpcResponse.getDeaths(),
                    grpcResponse.getRevives(),
                    grpcResponse.getFavouriteRole(),
                    grpcResponse.getTotalPlaytimeHours()
            );
            return ResponseEntity.ok(response);
        }
        catch (Exception e) {
            return ResponseEntity.status(404).body(Map.of("error", "Stats is not found: " + e.getMessage()));
        }
    }


    public record StatsResponse(
            int eloRating,
            int kills,
            int deaths,
            int revives,
            String favouriteRole,
            int totalPlaytimeHours
    ) {}
}
