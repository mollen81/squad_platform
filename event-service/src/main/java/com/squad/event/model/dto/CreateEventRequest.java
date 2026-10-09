package com.squad.event.model.dto;

import com.squad.event.model.enums.EventGameMode;

import java.time.Instant;
import java.util.UUID;

public record CreateEventRequest(
        String name,
        UUID creatorUserId,
        UUID secondLeaderId,
        int targetGameCount,
        Instant timeStart,
        EventGameMode gameMode,
        String creatorSideName,
        String enemySideName
) {}