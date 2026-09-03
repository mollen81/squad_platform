package com.squad.user.client;

import com.fasterxml.jackson.databind.JsonNode;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

@Component
@Slf4j
public class SteamWebApiClient {
    private final RestClient restClient;
    private final String apiKey;

    public SteamWebApiClient(@Value("${steam.api-key}") String apiKey) {
        this.apiKey = apiKey;
        this.restClient = RestClient.builder()
                .baseUrl("https://api.steampowered.com")
                .build();
    }

    public SteamProfile fetchUserProfile(String steamId) {
        try {
            JsonNode response = restClient.get()
                        .uri(uriBuilder -> uriBuilder
                        .path("/ISteamUser/GetPlayerSummaries/v2/")
                        .queryParam("key", apiKey)
                        .queryParam("steamids", steamId)
                        .build())
                    .retrieve()
                    .body(JsonNode.class);

            JsonNode players = response.path("response").path("players");
            if(players.isArray() && players.size() > 0) {
                JsonNode player = players.get(0);
                log.info("Profile fetched successfully for player {}", player.path("personaname").asText());

                return new SteamProfile(
                        player.path("personaname").asText(),
                        player.path("avatarUrl").asText()
                );
            }
        }
        catch (Exception e) {
            log.error("Error while fetching Steam profile for steamId {}: {}", steamId, e.getMessage());
        }

        // fallback if Steam is not accessible
        return new SteamProfile("Player_" + steamId, "");
    }



    public record SteamProfile(String displayName, String avatarUrl) {}
}
