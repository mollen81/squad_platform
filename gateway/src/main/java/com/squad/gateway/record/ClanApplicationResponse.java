package com.squad.gateway.record;

public record ClanApplicationResponse(
    String id,
    String clanId,
    String applicantSteamId,
    String status
) {}