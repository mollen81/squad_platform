package com.squad.payment.model.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public record VdsPurchasedEvent (
        @JsonProperty("type") String type, // vps.purchased
        @JsonProperty("event_id") String eventId,
        @JsonProperty("server_id") String serverIp,
        @JsonProperty("server_password") String serverPassword,
        @JsonProperty("server_config") ServerRentEvent.ServerConfig serverConfig,
        @JsonProperty("admins") List<ServerRentEvent.AdminEntry> admins,
        @JsonProperty("whitelist_steam_ids") List<String> whiteListSteamIds
) {}