package com.squad.clan.dto;

import com.squad.clan.entity.ClanMember;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.Builder;
import lombok.Data;

import java.util.List;
import java.util.UUID;

public class ClanRequests {

    @Data
    @Builder
    public static class CreateClanDto {
        private UUID leaderId; // В реальном проекте берется из JWT токена
        @Size(min = 2, max = 40)
        private String name;
        @Size(min = 3, max = 4)
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
    public static class GetClanDto {
        private UUID clanId;
    }

    @Builder
    public record AcceptApplicationDto(
        UUID applicationId,
        UUID accepterId // LEADER or MODERATOR
    ) {}
}
