package com.squad.gateway.record.ClanService;

public record ClanApplicationResponse(
    String id,
    String clanId,
    String applicantSteamId,
    String status
) {}