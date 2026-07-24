package com.squad.clan.dto;

import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
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
        @Id
        @GeneratedValue(strategy = GenerationType.UUID)
        private UUID id; // application id
        private UUID userId; // В реальном проекте берется из JWT токена
        private UUID clanId;
        private String socialLink;
        private String experienceText;
    }
}
