package com.squad.gateway.model.dto.response;

import java.util.UUID;

public record TimelineEventHttpResponse(
        UUID id,
        UUID matchId,
        String eventType,
        int gameTimeSeconds,
        String payloadJson,
        long createdAtEpochMilli
) {}