package com.squad.event.kafka.event;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.UUID;

public record ServerStopEvent(
        @JsonProperty("event_id") UUID eventId
) {}