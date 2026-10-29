package com.squad.event.model.dto;

import java.util.UUID;

public record JoinEventHttpRequest(
        UUID userId,
        UUID eventId,
        UUID sideId
) {}