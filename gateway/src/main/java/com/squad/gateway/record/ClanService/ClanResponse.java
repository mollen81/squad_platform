package com.squad.gateway.record.ClanService;

public record ClanResponse(
        String id,
        String name,
        String tag,
        String leaderId,
        int totalElo,
        String status,
        int membersCount
) {}
