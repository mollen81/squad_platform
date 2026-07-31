package com.squad.gateway.record.ClanService;

public record ProcessAcceptanceResponse(
    ClanMember newMember,
    String message
) {}