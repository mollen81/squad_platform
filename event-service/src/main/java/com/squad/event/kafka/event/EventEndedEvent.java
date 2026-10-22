package com.squad.event.kafka.event;

import java.time.Instant;
import java.util.UUID;

public record EventEndedEvent (
        UUID eventId,
        UUID winnerSide,
        Instant timeFinish
) {}