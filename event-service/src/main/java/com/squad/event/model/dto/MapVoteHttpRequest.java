package com.squad.event.model.dto;

import java.util.UUID;

public record MapVoteHttpRequest(
        UUID eventId,
        UUID matchId,
        UUID userId,
        String layer
) {}