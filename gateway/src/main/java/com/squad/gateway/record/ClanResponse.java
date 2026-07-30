package com.squad.gateway.record;

public record ClanResponse(
        String id,
        String name,
        String tag,
        String leaderSteamId,
        int totalElo,
        String status,
        int membersCount
) {}
