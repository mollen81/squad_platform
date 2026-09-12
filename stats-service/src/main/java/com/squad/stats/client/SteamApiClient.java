package com.squad.stats.client;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

@Slf4j
@Component
public class SteamApiClient {

    private final RestClient restClient;
    private final ObjectMapper mapper;
    private final String apiKey;
    private final int squadAppId = 393380;

    public SteamApiClient(
            @Value("${steam.api.key}") String apiKey,
            @Value("${steam.api.url:https://api.steampowered.com}") String baseUrl,
            ObjectMapper mapper) {

        this.apiKey = apiKey;
        this.mapper = mapper;
        this.restClient = RestClient.builder().baseUrl(baseUrl).build();
    }

    public int fetchSquadPlaytimeHours(String steamId) {
        if(apiKey.isBlank()) {
            log.warn("Steam API key is empty. The base time (0 hours) is assigned.");
            return 0;
        }

        try {
            String response = restClient.get()
                    .uri(b -> b.path("/IPlayerService/GetOwnedGames/v0001/")
                            .queryParam("key", apiKey)
                            .queryParam("steamid", steamId)
                            .queryParam("include_appinfo", "true")
                            .queryParam("format", "json")
                            .build())
                    .retrieve()
                    .body(String.class);

            if(response == null) {
                return 0;
            }

            JsonNode games = mapper.readTree(response).path("response").path("games");
            if(games.isArray()) {
                for(JsonNode game : games) {
                    if(game.path("appid").asInt() == squadAppId) {
                        return game.path("playtime_forever").asInt(0) / 60;
                    }
                }
            }
        }
        catch (Exception e) {
            log.error("Failure to retrieve the playtime hours from Steam for {}: {}", steamId, e.getMessage(), e);
        }

        return 0; // error or profile is private
    }
}
