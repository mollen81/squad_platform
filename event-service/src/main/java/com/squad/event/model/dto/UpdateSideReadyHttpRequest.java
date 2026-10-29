package com.squad.event.model.dto;

import java.util.UUID;

public record UpdateSideReadyHttpRequest(
        UUID eventId,
        UUID sideId,
        UUID userId,
        Boolean isReady
) {}