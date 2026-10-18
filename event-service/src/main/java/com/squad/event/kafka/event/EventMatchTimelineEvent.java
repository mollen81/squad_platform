package com.squad.event.kafka.event;

import com.squad.event.model.enums.TimelineEventType;

import java.util.Map;
import java.util.UUID;

public record EventMatchTimelineEvent(
        UUID eventId,
        int matchSequenceNumber,
        TimelineEventType timelineEventType,
        int gameTimeSeconds,
        Map<String, Object> payload
) {}