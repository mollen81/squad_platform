package com.squad.clan.dto;

import lombok.Data;

import java.util.UUID;

public class ClanRequests {

    @Data
    public static class CreateClanDto {
        private UUID leaderId; // В реальном проекте берется из JWT токена
        private String name;
        private String tag;
        private String description;
        private String requirements;
        private String avatarUrl;
    }

    @Data
    public static class ApplyToClanDto {
        private UUID userId; // В реальном проекте берется из JWT токена
        private UUID clanId;
        private String socialLink;
        private String experienceText;
    }
}
