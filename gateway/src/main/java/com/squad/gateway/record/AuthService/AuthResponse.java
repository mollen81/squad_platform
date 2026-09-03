package com.squad.gateway.record.AuthService;

public record AuthResponse(String userId, String steamId, String token, boolean isNewUser) {}