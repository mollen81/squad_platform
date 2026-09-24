package com.squad.deployment.dto;

import java.util.List;

public record VpsPurchasedEvent (
    String eventId,
    String ip,
    String rootPassword,
    ServerConfig serverConfig,
    List<Admin> admins,
    List<String> whitelistSteamIds
)
{
    public record ServerConfig(String serverName, String password, String map) {}
    public record Admin(String steamId, String role) {}
}
