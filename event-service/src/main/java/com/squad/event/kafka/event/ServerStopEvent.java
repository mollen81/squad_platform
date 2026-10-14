package com.squad.event.kafka.event;

import java.util.UUID;

public record ServerStopEvent(
        String topic,
        UUID eventId
) {}