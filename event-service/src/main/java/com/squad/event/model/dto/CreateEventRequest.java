package com.squad.event.model.dto;

import java.time.Instant;
import java.util.UUID;

public record CreateEventRequest(
        String name,
        UUID creatorUserId,
        UUID secondLeaderId,
        int targetGameCount,
        Instant timeStart,
        String creatorSideName,
        String enemySideName
) {}