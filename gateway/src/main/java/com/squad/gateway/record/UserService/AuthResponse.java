package com.squad.gateway.record.UserService;

public record AuthResponse(String userId, String steamId, Boolean isNewUser, String token) {}
