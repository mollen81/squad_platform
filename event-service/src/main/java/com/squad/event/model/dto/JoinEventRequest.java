package com.squad.event.model.dto;

import java.util.UUID;

public record JoinEventRequest (
        UUID userId,
        UUID eventId,
        UUID sideId,
        UUID clanId
) {}