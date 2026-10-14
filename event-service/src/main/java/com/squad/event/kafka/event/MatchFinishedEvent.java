package com.squad.event.kafka.event;

import java.util.UUID;

public record MatchFinishedEvent(
        UUID eventId,
        int matchSequenceNumber,
        UUID winnerSideId
) {}