package com.squad.payment.model.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.time.Instant;
import java.util.List;

public record ServerRentEvent(
        @JsonProperty("type") String type,
        @JsonProperty("eventId") String event_id,
        @JsonProperty("time_start") Instant timeStart,
        @JsonProperty("server_config") ServerConfig serverConfig,
        @JsonProperty("admins") List<AdminEntry> admins,
        @JsonProperty("whitelist_steam_ids") List<String> whitelistSteamIds

        ) {
        public record ServerConfig(
                @JsonProperty("server_name") String serverName,
                @JsonProperty("password") String password,
                @JsonProperty("map") String map,
                @JsonProperty("max_players") Integer maxPlayers
        ) {}

        public record AdminEntry(
                @JsonProperty("steam_id") String steamId,
                @JsonProperty("role") String role
        ) {}
}
