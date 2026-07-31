package com.squad.gateway.record.StatsService;

public record PlayerStats(
        int eloRating,
        int totalPlaytimeHours,
        int kills,
        int vehiclesDestroyed,
        int deaths,
        int revives,
        String favouriteRole ) {}