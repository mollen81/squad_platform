package com.squad.gateway.model.dto.request;

import jakarta.validation.constraints.*;

import java.time.Instant;
import java.util.UUID;

public record CreateEventHttpRequest(
        @NotBlank(message = "Event name is required")
        String name,

        UUID creatorUserId,

        @NotNull(message = "Second leader ID is required")
        UUID secondLeaderId,

        @Min(value = 1, message = "Target game count must be at least 1")
        @Max(value = 5, message = "Target game count mustn't be more than 5")
        int targetGameCount,

        @NotNull(message = "Start time is required")
        @Future(message = "Start time must be in the future")
        Instant timeStart,

        @NotBlank(message = "Game mode is required")
        String gameMode, // for example: "RAAS", "AAS", "INVASION"

        @NotBlank(message = "Creator side name is required")
        String creatorSideName,

        @NotBlank(message = "Enemy side name is required")
        String enemySideName
) {}