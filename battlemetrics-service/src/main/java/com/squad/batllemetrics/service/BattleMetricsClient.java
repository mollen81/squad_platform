package com.squad.batllemetrics.service;

import io.github.resilience4j.ratelimiter.annotation.RateLimiter;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;


@Service
@Slf4j
public class BattleMetricsClient {
    private final RestClient restClient;

    public BattleMetricsClient(@Value("${battlemetrics.api.base-url}") String baseUrl) {
        this.restClient = RestClient.builder()
                .baseUrl(baseUrl)
                .build();
    }

    @RateLimiter(name = "battlemetricsAPI", fallbackMethod = "fetchPlayerStatsFallback")
    public String fetchPlayerIdBySteamId(String steamId) {
        log.info("Request sending to BattleMetrics API for search the player with SteamId: {}", steamId);

        try {
            Thread.sleep(1500);
        }
        catch(InterruptedException e) {
            Thread.currentThread().interrupt();
        }

        return "bm_id_" + steamId.substring(steamId.length() - 5);
    }

    public String fetchPlayerStatsFallback(String steamId, Throwable t) {
        log.warn("Rate Limiter hit! Exceeded the limits of requests to BattleMetrics for: {}", steamId);
        throw new RuntimeException("BattleMetrics API rate Limit exceeded", t);
    }
}
