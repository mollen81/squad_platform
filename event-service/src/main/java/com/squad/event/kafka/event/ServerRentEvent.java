package com.squad.event.kafka.event;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

public record ServerRentEvent(
        @JsonProperty("eventId") UUID eventId,
        @JsonProperty("time_start") Instant timeStart,
        @JsonProperty("server_config") ServerConfig serverConfig,
        @JsonProperty("admins") List<AdminEntry> admins,
        @JsonProperty("whitelist_steam_ids") List<String> whitelistSteamIds
        ) {
        public record ServerConfig (
                @JsonProperty("server_name") String serverName,
                @JsonProperty("password") String password,
                @JsonProperty("layer_names") List<String> layerNames,
                @JsonProperty("max_players") Integer maxPlayers
        ) {}
        public record AdminEntry(
                @JsonProperty("steam_id") String steamId,
                @JsonProperty("role") String role
        ) {}
}