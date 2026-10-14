package com.squad.event.kafka.event;

import java.util.UUID;

public record ServerReadyEvent(
        UUID eventId,
        String ipAddress,
        int queryPort,
        String password
) {}