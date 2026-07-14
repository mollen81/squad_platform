package com.squad.gateway.record;

public record AuthResponse(String userId, String steamId, Boolean isNewUser, String token) {}
