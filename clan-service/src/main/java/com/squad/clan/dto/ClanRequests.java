package com.squad.clan.dto;

import com.squad.clan.entity.ClanMember;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import lombok.Builder;
import lombok.Data;

import java.util.List;
import java.util.UUID;

public class ClanRequests {

    @Data
    @Builder
    public static class CreateClanDto {
        private UUID leaderId; // В реальном проекте берется из JWT токена
        private String name;
        private String tag;
        private String description;
        private String requirements;
        private String avatarUrl;
    }

    @Data
    @Builder
    public static class ApplyToClanDto {
        @Id
        @GeneratedValue(strategy = GenerationType.UUID)
        private UUID id; // application id
        private UUID userId; // В реальном проекте берется из JWT токена
        private UUID clanId;
        private String socialLink;
        private String experienceText;
    }

    @Data
    @Builder
    public static class GetClanWithAllMembersDto {
        @Id
        @GeneratedValue(strategy = GenerationType.UUID)
        private UUID id;
        private String name;
        private String tag;
        private String status;
        private int totalElo;
        private List<ClanMember> members;
    }

    public record AcceptApplicationDto(
            UUID applicationId,
            UUID accepterId // LEADER or MODERATOR
    ) {}
}
